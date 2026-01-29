// Presence provides the HSM-based agent presence state machine.
//
// The presence HSM implements the state transitions defined in spec §6.1 and §6.5:
//
//	                    ┌──────────────────┐
//	                    ▼                  │
//	┌────────┐    ┌─────────┐    ┌────────┐
//	│ Online │◀──▶│  Busy   │───▶│ Offline│
//	└────────┘    └─────────┘    └────────┘
//	     ▲              │              │
//	     │              ▼              │
//	     │         ┌────────┐          │
//	     └─────────│  Away  │◀─────────┘
//	               └────────┘
//
// Transitions:
//   - task.assigned: Online → Busy
//   - task.completed: Busy → Online
//   - prompt.detected: Busy → Online
//   - rate.limit: * → Offline
//   - rate.cleared: Offline → Online
//   - stuck.detected: * → Away
//   - activity.detected: Away → Online
package agent

import (
	"context"
	"sync"

	hsm "github.com/stateforward/hsm-go"

	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/pkg/api"
)

// PresenceEvent names for presence state transitions.
const (
	PresenceEventTaskAssigned     = "task.assigned"     // Online → Busy
	PresenceEventTaskCompleted    = "task.completed"    // Busy → Online
	PresenceEventPromptDetected   = "prompt.detected"   // Busy → Online
	PresenceEventRateLimit        = "rate.limit"        // * → Offline
	PresenceEventRateCleared      = "rate.cleared"      // Offline → Online
	PresenceEventStuckDetected    = "stuck.detected"    // * → Away
	PresenceEventActivityDetected = "activity.detected" // Away → Online
)

// PresenceHSM wraps an agent with HSM-driven presence management.
type PresenceHSM struct {
	hsm.HSM

	mu            sync.RWMutex
	agent         *Agent
	presenceState api.PresenceState
	dispatcher    event.Dispatcher
}

// PresenceModel defines the HSM model for agent presence.
// See spec §6.1 and §6.5.
var PresenceModel = hsm.Define(
	"agent.presence",

	// States
	hsm.State("online",
		hsm.Entry(func(ctx context.Context, p *PresenceHSM, e hsm.Event) {
			p.onEnterOnline(ctx, e)
		}),
	),
	hsm.State("busy",
		hsm.Entry(func(ctx context.Context, p *PresenceHSM, e hsm.Event) {
			p.onEnterBusy(ctx, e)
		}),
	),
	hsm.State("offline",
		hsm.Entry(func(ctx context.Context, p *PresenceHSM, e hsm.Event) {
			p.onEnterOffline(ctx, e)
		}),
	),
	hsm.State("away",
		hsm.Entry(func(ctx context.Context, p *PresenceHSM, e hsm.Event) {
			p.onEnterAway(ctx, e)
		}),
	),

	// Transitions: Online ↔ Busy
	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventTaskAssigned}),
		hsm.Source("online"),
		hsm.Target("busy"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventTaskCompleted}),
		hsm.Source("busy"),
		hsm.Target("online"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventPromptDetected}),
		hsm.Source("busy"),
		hsm.Target("online"),
	),

	// Transitions: → Offline (rate limited)
	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventRateLimit}),
		hsm.Source("online"),
		hsm.Target("offline"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventRateLimit}),
		hsm.Source("busy"),
		hsm.Target("offline"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventRateCleared}),
		hsm.Source("offline"),
		hsm.Target("online"),
	),

	// Transitions: → Away (unresponsive/stuck)
	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventStuckDetected}),
		hsm.Source("online"),
		hsm.Target("away"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventStuckDetected}),
		hsm.Source("busy"),
		hsm.Target("away"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventStuckDetected}),
		hsm.Source("offline"),
		hsm.Target("away"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventActivityDetected}),
		hsm.Source("away"),
		hsm.Target("online"),
	),

	// Initial state
	hsm.Initial(
		hsm.Target("online"),
	),
)

// NewPresenceHSM creates a new presence HSM for an agent.
func NewPresenceHSM(agent *Agent, dispatcher event.Dispatcher) *PresenceHSM {
	if dispatcher == nil {
		dispatcher = event.NewNoopDispatcher()
	}

	return &PresenceHSM{
		agent:         agent,
		presenceState: api.PresenceOnline,
		dispatcher:    dispatcher,
	}
}

// Start initializes and starts the presence HSM.
// Returns the started HSM instance.
func (p *PresenceHSM) Start(ctx context.Context) *PresenceHSM {
	return hsm.Started(ctx, p, &PresenceModel)
}

// PresenceState returns the current presence state.
func (p *PresenceHSM) PresenceState() api.PresenceState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.presenceState
}

// Agent returns the associated agent.
func (p *PresenceHSM) Agent() *Agent {
	return p.agent
}

// setPresenceState updates the internal state and synchronizes with the agent.
func (p *PresenceHSM) setPresenceState(state api.PresenceState) {
	p.mu.Lock()
	old := p.presenceState
	p.presenceState = state
	p.mu.Unlock()

	if p.agent != nil {
		p.agent.SetPresence(state)
	}

	// Dispatch presence changed event
	if p.dispatcher != nil && p.agent != nil && old != state {
		_ = p.dispatcher.Dispatch(context.Background(), event.NewEvent(
			event.TypePresenceChanged,
			p.agent.ID,
			map[string]any{
				"old":   string(old),
				"new":   string(state),
				"state": string(state),
			},
		))
	}
}

// Entry action for Online state
func (p *PresenceHSM) onEnterOnline(ctx context.Context, e hsm.Event) {
	p.setPresenceState(api.PresenceOnline)
}

// Entry action for Busy state
func (p *PresenceHSM) onEnterBusy(ctx context.Context, e hsm.Event) {
	p.setPresenceState(api.PresenceBusy)
}

// Entry action for Offline state
func (p *PresenceHSM) onEnterOffline(ctx context.Context, e hsm.Event) {
	p.setPresenceState(api.PresenceOffline)
}

// Entry action for Away state
func (p *PresenceHSM) onEnterAway(ctx context.Context, e hsm.Event) {
	p.setPresenceState(api.PresenceAway)
}

// DispatchTaskAssigned sends a "task.assigned" event to transition from Online to Busy.
func DispatchTaskAssigned(ctx context.Context, instance hsm.Instance) <-chan struct{} {
	return hsm.Dispatch(ctx, instance, hsm.Event{Name: PresenceEventTaskAssigned})
}

// DispatchTaskCompleted sends a "task.completed" event to transition from Busy to Online.
func DispatchTaskCompleted(ctx context.Context, instance hsm.Instance) <-chan struct{} {
	return hsm.Dispatch(ctx, instance, hsm.Event{Name: PresenceEventTaskCompleted})
}

// DispatchPromptDetected sends a "prompt.detected" event to transition from Busy to Online.
func DispatchPromptDetected(ctx context.Context, instance hsm.Instance) <-chan struct{} {
	return hsm.Dispatch(ctx, instance, hsm.Event{Name: PresenceEventPromptDetected})
}

// DispatchRateLimit sends a "rate.limit" event to transition to Offline state.
func DispatchRateLimit(ctx context.Context, instance hsm.Instance) <-chan struct{} {
	return hsm.Dispatch(ctx, instance, hsm.Event{Name: PresenceEventRateLimit})
}

// DispatchRateCleared sends a "rate.cleared" event to transition from Offline to Online.
func DispatchRateCleared(ctx context.Context, instance hsm.Instance) <-chan struct{} {
	return hsm.Dispatch(ctx, instance, hsm.Event{Name: PresenceEventRateCleared})
}

// DispatchStuckDetected sends a "stuck.detected" event to transition to Away state.
func DispatchStuckDetected(ctx context.Context, instance hsm.Instance) <-chan struct{} {
	return hsm.Dispatch(ctx, instance, hsm.Event{Name: PresenceEventStuckDetected})
}

// DispatchActivityDetected sends an "activity.detected" event to transition from Away to Online.
func DispatchActivityDetected(ctx context.Context, instance hsm.Instance) <-chan struct{} {
	return hsm.Dispatch(ctx, instance, hsm.Event{Name: PresenceEventActivityDetected})
}
