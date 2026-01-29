// Package agent implements HSM-based agent lifecycle and presence management.
// This implementation uses stateforward/hsm-go for proper hierarchical state machines
// as required by the amux specification.
package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/stateforward/hsm-go"

	"github.com/copilot-claude-sonnet-4/amux/pkg/api"
)

// LifecycleEvents define the events that can trigger agent lifecycle transitions.
const (
	// EvtStart triggers transition from Pending to Starting.
	EvtStart = "start"
	// EvtStartupComplete triggers transition from Starting to Running.
	EvtStartupComplete = "startup_complete"
	// EvtTerminate triggers transition to Terminated.
	EvtTerminate = "terminate"
	// EvtError triggers transition to Errored.
	EvtError = "error"
	// EvtRestart triggers transition from Terminated/Errored back to Pending.
	EvtRestart = "restart"
)

// PresenceEvents define the events that can trigger presence transitions.
const (
	// EvtGoOnline triggers transition to Online.
	EvtGoOnline = "go_online"
	// EvtGoBusy triggers transition to Busy.
	EvtGoBusy = "go_busy"
	// EvtGoOffline triggers transition to Offline.
	EvtGoOffline = "go_offline"
	// EvtGoAway triggers transition to Away.
	EvtGoAway = "go_away"
	// EvtActivity triggers transition from Away back to Online.
	EvtActivity = "activity"
)

// AgentHSM represents an agent with HSM-based state management.
type AgentHSM struct {
	hsm.HSM

	// Agent is the underlying agent data.
	Agent *api.Agent

	// mu protects concurrent access to the agent.
	mu sync.RWMutex
}

// AgentHSMActor wraps an Agent with proper HSM-based state machines for lifecycle and presence.
// This replaces the simple state machine implementation per spec requirements.
type AgentHSMActor struct {
	// lifecycleHSM manages agent lifecycle state transitions.
	lifecycleHSM *AgentHSM

	// presenceHSM manages agent presence state transitions.
	presenceHSM *AgentHSM

	// ctx is the context for HSM operations.
	ctx context.Context

	// cancel cancels the HSM context.
	cancel context.CancelFunc
}

// NewAgentHSMActor creates a new HSM-based agent actor with proper state machines.
func NewAgentHSMActor(name, adapter, repoRoot string, config map[string]interface{}) (*AgentHSMActor, error) {
	// Validate inputs using the same logic as the old implementation
	tempAgent, err := NewAgentActor(name, adapter, repoRoot, config)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create HSM context
	ctx, cancel := context.WithCancel(context.Background())

	// Create the HSM actor
	hsmActor := &AgentHSMActor{
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize lifecycle HSM
	lifecycleHSM, err := hsmActor.initLifecycleHSM(tempAgent.Agent)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize lifecycle HSM: %w", err)
	}
	hsmActor.lifecycleHSM = lifecycleHSM

	// Initialize presence HSM
	presenceHSM, err := hsmActor.initPresenceHSM(tempAgent.Agent)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize presence HSM: %w", err)
	}
	hsmActor.presenceHSM = presenceHSM

	return hsmActor, nil
}

// initLifecycleHSM initializes the agent lifecycle state machine.
// Per spec: Pending → Starting → Running → Terminated/Errored
func (a *AgentHSMActor) initLifecycleHSM(agent *api.Agent) (*AgentHSM, error) {
	// Define lifecycle HSM model
	model := hsm.Define("lifecycle",
		hsm.State("pending",
			hsm.Entry[*AgentHSM](func(ctx context.Context, hsm *AgentHSM, event hsm.Event) {
				hsm.mu.Lock()
				defer hsm.mu.Unlock()
				hsm.Agent.State = api.AgentStatePending
				hsm.Agent.UpdatedAt = time.Now().UTC()
			}),
		),
		hsm.State("starting",
			hsm.Entry[*AgentHSM](func(ctx context.Context, hsm *AgentHSM, event hsm.Event) {
				hsm.mu.Lock()
				defer hsm.mu.Unlock()
				hsm.Agent.State = api.AgentStateStarting
				hsm.Agent.UpdatedAt = time.Now().UTC()
			}),
		),
		hsm.State("running",
			hsm.Entry[*AgentHSM](func(ctx context.Context, hsm *AgentHSM, event hsm.Event) {
				hsm.mu.Lock()
				defer hsm.mu.Unlock()
				hsm.Agent.State = api.AgentStateRunning
				hsm.Agent.UpdatedAt = time.Now().UTC()
			}),
		),
		hsm.State("terminated",
			hsm.Entry[*AgentHSM](func(ctx context.Context, hsm *AgentHSM, event hsm.Event) {
				hsm.mu.Lock()
				defer hsm.mu.Unlock()
				hsm.Agent.State = api.AgentStateTerminated
				hsm.Agent.UpdatedAt = time.Now().UTC()
			}),
		),
		hsm.State("errored",
			hsm.Entry[*AgentHSM](func(ctx context.Context, hsm *AgentHSM, event hsm.Event) {
				hsm.mu.Lock()
				defer hsm.mu.Unlock()
				hsm.Agent.State = api.AgentStateErrored
				hsm.Agent.UpdatedAt = time.Now().UTC()
			}),
		),
		hsm.Transition("start_transition",
			hsm.Source("pending"),
			hsm.Target("starting"),
			hsm.On(hsm.Event{Name: EvtStart}),
		),
		hsm.Transition("startup_complete_transition",
			hsm.Source("starting"),
			hsm.Target("running"),
			hsm.On(hsm.Event{Name: EvtStartupComplete}),
		),
		hsm.Transition("startup_error_transition",
			hsm.Source("starting"),
			hsm.Target("errored"),
			hsm.On(hsm.Event{Name: EvtError}),
		),
		hsm.Transition("terminate_transition",
			hsm.Source("running"),
			hsm.Target("terminated"),
			hsm.On(hsm.Event{Name: EvtTerminate}),
		),
		hsm.Transition("runtime_error_transition",
			hsm.Source("running"),
			hsm.Target("errored"),
			hsm.On(hsm.Event{Name: EvtError}),
		),
		hsm.Transition("restart_from_terminated",
			hsm.Source("terminated"),
			hsm.Target("pending"),
			hsm.On(hsm.Event{Name: EvtRestart}),
		),
		hsm.Transition("restart_from_errored",
			hsm.Source("errored"),
			hsm.Target("pending"),
			hsm.On(hsm.Event{Name: EvtRestart}),
		),
		hsm.Initial(hsm.Target("pending")),
	)

	// Create the HSM instance
	hsmInstance := &AgentHSM{Agent: agent}
	hsmInstance = hsm.Started(a.ctx, hsmInstance, &model)
	return hsmInstance, nil
}

// initPresenceHSM initializes the agent presence state machine.
// Per spec: Online ↔ Busy ↔ Offline ↔ Away
func (a *AgentHSMActor) initPresenceHSM(agent *api.Agent) (*AgentHSM, error) {
	// Define presence HSM model
	model := hsm.Define("presence",
		hsm.State("offline",
			hsm.Entry[*AgentHSM](func(ctx context.Context, hsm *AgentHSM, event hsm.Event) {
				hsm.mu.Lock()
				defer hsm.mu.Unlock()
				hsm.Agent.Presence = api.PresenceOffline
				hsm.Agent.UpdatedAt = time.Now().UTC()
			}),
		),
		hsm.State("online",
			hsm.Entry[*AgentHSM](func(ctx context.Context, hsm *AgentHSM, event hsm.Event) {
				hsm.mu.Lock()
				defer hsm.mu.Unlock()
				hsm.Agent.Presence = api.PresenceOnline
				hsm.Agent.UpdatedAt = time.Now().UTC()
			}),
		),
		hsm.State("busy",
			hsm.Entry[*AgentHSM](func(ctx context.Context, hsm *AgentHSM, event hsm.Event) {
				hsm.mu.Lock()
				defer hsm.mu.Unlock()
				hsm.Agent.Presence = api.PresenceBusy
				hsm.Agent.UpdatedAt = time.Now().UTC()
			}),
		),
		hsm.State("away",
			hsm.Entry[*AgentHSM](func(ctx context.Context, hsm *AgentHSM, event hsm.Event) {
				hsm.mu.Lock()
				defer hsm.mu.Unlock()
				hsm.Agent.Presence = api.PresenceAway
				hsm.Agent.UpdatedAt = time.Now().UTC()
			}),
		),
		hsm.Transition("go_online",
			hsm.Source("offline"),
			hsm.Target("online"),
			hsm.On(hsm.Event{Name: EvtGoOnline}),
		),
		hsm.Transition("go_busy",
			hsm.Source("online"),
			hsm.Target("busy"),
			hsm.On(hsm.Event{Name: EvtGoBusy}),
		),
		hsm.Transition("go_away",
			hsm.Source("online"),
			hsm.Target("away"),
			hsm.On(hsm.Event{Name: EvtGoAway}),
		),
		hsm.Transition("go_offline_from_online",
			hsm.Source("online"),
			hsm.Target("offline"),
			hsm.On(hsm.Event{Name: EvtGoOffline}),
		),
		hsm.Transition("busy_to_online",
			hsm.Source("busy"),
			hsm.Target("online"),
			hsm.On(hsm.Event{Name: EvtGoOnline}),
		),
		hsm.Transition("busy_to_offline",
			hsm.Source("busy"),
			hsm.Target("offline"),
			hsm.On(hsm.Event{Name: EvtGoOffline}),
		),
		hsm.Transition("away_activity",
			hsm.Source("away"),
			hsm.Target("online"),
			hsm.On(hsm.Event{Name: EvtActivity}),
		),
		hsm.Transition("away_to_offline",
			hsm.Source("away"),
			hsm.Target("offline"),
			hsm.On(hsm.Event{Name: EvtGoOffline}),
		),
		hsm.Initial(hsm.Target("offline")),
	)

	// Create the HSM instance (use a copy of agent for presence)
	hsmInstance := &AgentHSM{Agent: agent}
	hsmInstance = hsm.Started(a.ctx, hsmInstance, &model)
	return hsmInstance, nil
}

// Dispatch sends an event to the appropriate HSM using hsm.Dispatch() as required by spec.
func (a *AgentHSMActor) Dispatch(eventName string, data interface{}) error {
	event := hsm.Event{Name: eventName, Data: data}

	// Determine which HSM should handle this event
	switch eventName {
	case EvtStart, EvtStartupComplete, EvtTerminate, EvtError, EvtRestart:
		// Use HSM dispatch and wait for completion
		ch := hsm.Dispatch(a.ctx, a.lifecycleHSM, event)
		<-ch // Wait for dispatch to complete
		return nil
	case EvtGoOnline, EvtGoBusy, EvtGoOffline, EvtGoAway, EvtActivity:
		// Use HSM dispatch and wait for completion
		ch := hsm.Dispatch(a.ctx, a.presenceHSM, event)
		<-ch // Wait for dispatch to complete
		return nil
	default:
		return fmt.Errorf("unknown event: %s", eventName)
	}
}

// GetState returns the current agent state safely.
func (a *AgentHSMActor) GetState() api.AgentState {
	a.lifecycleHSM.mu.RLock()
	defer a.lifecycleHSM.mu.RUnlock()
	return a.lifecycleHSM.Agent.State
}

// GetPresence returns the current agent presence safely.
func (a *AgentHSMActor) GetPresence() api.PresenceState {
	a.presenceHSM.mu.RLock()
	defer a.presenceHSM.mu.RUnlock()
	return a.presenceHSM.Agent.Presence
}

// GetAgent returns a copy of the agent data safely.
func (a *AgentHSMActor) GetAgent() api.Agent {
	// Both HSMs share the same agent data, so use lifecycle HSM
	a.lifecycleHSM.mu.RLock()
	defer a.lifecycleHSM.mu.RUnlock()
	return *a.lifecycleHSM.Agent
}

// Close gracefully shuts down the HSM actor.
func (a *AgentHSMActor) Close() error {
	if a.cancel != nil {
		a.cancel()
	}
	return nil
}