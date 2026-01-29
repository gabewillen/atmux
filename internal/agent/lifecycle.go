// Lifecycle provides the HSM-based agent lifecycle state machine.
//
// The lifecycle HSM implements the state transitions defined in spec §5.4:
//
//	┌─────────┐    ┌─────────┐    ┌─────────┐    ┌────────────┐
//	│ Pending │───▶│ Starting│───▶│ Running │───▶│ Terminated │
//	└─────────┘    └─────────┘    └─────────┘    └────────────┘
//	                                  │
//	                                  ▼
//	                             ┌─────────┐
//	                             │ Errored │
//	                             └─────────┘
//
// Transitions:
//   - "start" event: Pending → Starting
//   - "ready" event: Starting → Running
//   - "stop" event: Running → Terminated
//   - "error" event: Any → Errored
package agent

import (
	"context"
	"sync"

	hsm "github.com/stateforward/hsm-go"

	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/pkg/api"
)

// LifecycleEvent names for lifecycle state transitions.
const (
	LifecycleEventStart = "start" // Pending → Starting
	LifecycleEventReady = "ready" // Starting → Running
	LifecycleEventStop  = "stop"  // Running → Terminated
	LifecycleEventError = "error" // Any → Errored
)

// LifecycleHSM wraps an agent with HSM-driven lifecycle management.
type LifecycleHSM struct {
	hsm.HSM

	mu            sync.RWMutex
	agent         *Agent
	lifecycleState api.LifecycleState
	dispatcher    event.Dispatcher
	lastError     error
}

// LifecycleModel defines the HSM model for agent lifecycle.
// See spec §5.4.
var LifecycleModel = hsm.Define(
	"agent.lifecycle",

	// States
	hsm.State("pending"),
	hsm.State("starting",
		hsm.Entry(func(ctx context.Context, l *LifecycleHSM, e hsm.Event) {
			l.onEnterStarting(ctx)
		}),
	),
	hsm.State("running",
		hsm.Entry(func(ctx context.Context, l *LifecycleHSM, e hsm.Event) {
			l.onEnterRunning(ctx)
		}),
		hsm.Exit(func(ctx context.Context, l *LifecycleHSM, e hsm.Event) {
			l.onExitRunning(ctx)
		}),
	),
	hsm.State("terminated",
		hsm.Entry(func(ctx context.Context, l *LifecycleHSM, e hsm.Event) {
			l.onEnterTerminated(ctx)
		}),
	),
	hsm.State("errored",
		hsm.Entry(func(ctx context.Context, l *LifecycleHSM, e hsm.Event) {
			l.onEnterErrored(ctx, e)
		}),
	),

	// Transitions
	hsm.Transition(
		hsm.On(hsm.Event{Name: LifecycleEventStart}),
		hsm.Source("pending"),
		hsm.Target("starting"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: LifecycleEventReady}),
		hsm.Source("starting"),
		hsm.Target("running"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: LifecycleEventStop}),
		hsm.Source("running"),
		hsm.Target("terminated"),
	),
	// Error transitions from any non-final state
	hsm.Transition(
		hsm.On(hsm.Event{Name: LifecycleEventError}),
		hsm.Source("pending"),
		hsm.Target("errored"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: LifecycleEventError}),
		hsm.Source("starting"),
		hsm.Target("errored"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: LifecycleEventError}),
		hsm.Source("running"),
		hsm.Target("errored"),
	),

	// Initial state
	hsm.Initial(
		hsm.Target("pending"),
	),
)

// NewLifecycleHSM creates a new lifecycle HSM for an agent.
func NewLifecycleHSM(agent *Agent, dispatcher event.Dispatcher) *LifecycleHSM {
	if dispatcher == nil {
		dispatcher = event.NewNoopDispatcher()
	}

	return &LifecycleHSM{
		agent:          agent,
		lifecycleState: api.LifecyclePending,
		dispatcher:     dispatcher,
	}
}

// Start initializes and starts the lifecycle HSM.
// Returns the started HSM instance.
func (l *LifecycleHSM) Start(ctx context.Context) *LifecycleHSM {
	return hsm.Started(ctx, l, &LifecycleModel)
}

// LifecycleState returns the current lifecycle state.
func (l *LifecycleHSM) LifecycleState() api.LifecycleState {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.lifecycleState
}

// LastError returns the last error that caused the errored state.
func (l *LifecycleHSM) LastError() error {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.lastError
}

// Agent returns the associated agent.
func (l *LifecycleHSM) Agent() *Agent {
	return l.agent
}

// setLifecycleState updates the internal state and synchronizes with the agent.
func (l *LifecycleHSM) setLifecycleState(state api.LifecycleState) {
	l.mu.Lock()
	l.lifecycleState = state
	l.mu.Unlock()

	if l.agent != nil {
		l.agent.SetLifecycle(state)
	}
}

// Entry action for Starting state
func (l *LifecycleHSM) onEnterStarting(ctx context.Context) {
	l.setLifecycleState(api.LifecycleStarting)

	if l.dispatcher != nil && l.agent != nil {
		_ = l.dispatcher.Dispatch(ctx, event.NewEvent(
			event.TypeAgentStarting,
			l.agent.ID,
			map[string]any{"state": string(api.LifecycleStarting)},
		))
	}
}

// Entry action for Running state
func (l *LifecycleHSM) onEnterRunning(ctx context.Context) {
	l.setLifecycleState(api.LifecycleRunning)

	if l.dispatcher != nil && l.agent != nil {
		_ = l.dispatcher.Dispatch(ctx, event.NewEvent(
			event.TypeAgentStarted,
			l.agent.ID,
			map[string]any{"state": string(api.LifecycleRunning)},
		))
	}
}

// Exit action for Running state
func (l *LifecycleHSM) onExitRunning(ctx context.Context) {
	if l.dispatcher != nil && l.agent != nil {
		_ = l.dispatcher.Dispatch(ctx, event.NewEvent(
			event.TypeAgentStopping,
			l.agent.ID,
			map[string]any{"state": string(l.lifecycleState)},
		))
	}
}

// Entry action for Terminated state
func (l *LifecycleHSM) onEnterTerminated(ctx context.Context) {
	l.setLifecycleState(api.LifecycleTerminated)

	if l.dispatcher != nil && l.agent != nil {
		_ = l.dispatcher.Dispatch(ctx, event.NewEvent(
			event.TypeAgentTerminated,
			l.agent.ID,
			map[string]any{"state": string(api.LifecycleTerminated)},
		))
	}
}

// Entry action for Errored state
func (l *LifecycleHSM) onEnterErrored(ctx context.Context, e hsm.Event) {
	l.mu.Lock()
	l.lifecycleState = api.LifecycleErrored
	// Extract error from event data if present
	if err, ok := e.Data.(error); ok {
		l.lastError = err
	}
	l.mu.Unlock()

	if l.agent != nil {
		l.agent.SetLifecycle(api.LifecycleErrored)
	}

	if l.dispatcher != nil && l.agent != nil {
		data := map[string]any{"state": string(api.LifecycleErrored)}
		if l.lastError != nil {
			data["error"] = l.lastError.Error()
		}
		_ = l.dispatcher.Dispatch(ctx, event.NewEvent(
			event.TypeAgentErrored,
			l.agent.ID,
			data,
		))
	}
}

// DispatchStart sends a "start" event to transition from Pending to Starting.
func DispatchStart(ctx context.Context, instance hsm.Instance) <-chan struct{} {
	return hsm.Dispatch(ctx, instance, hsm.Event{Name: LifecycleEventStart})
}

// DispatchReady sends a "ready" event to transition from Starting to Running.
func DispatchReady(ctx context.Context, instance hsm.Instance) <-chan struct{} {
	return hsm.Dispatch(ctx, instance, hsm.Event{Name: LifecycleEventReady})
}

// DispatchStop sends a "stop" event to transition from Running to Terminated.
func DispatchStop(ctx context.Context, instance hsm.Instance) <-chan struct{} {
	return hsm.Dispatch(ctx, instance, hsm.Event{Name: LifecycleEventStop})
}

// DispatchError sends an "error" event to transition to Errored state.
// The error parameter is stored and can be retrieved via LastError().
func DispatchError(ctx context.Context, instance hsm.Instance, err error) <-chan struct{} {
	return hsm.Dispatch(ctx, instance, hsm.Event{Name: LifecycleEventError}.WithData(err))
}
