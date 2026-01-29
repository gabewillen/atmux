// Package agent provides the agent actor model and HSM-driven lifecycle and presence state machines.
package agent

import (
	"context"
	"strings"

	"github.com/stateforward/hsm-go"

	"github.com/stateforward/amux/pkg/api"
)

// Lifecycle state constants matching spec §5.4.
const (
	StatePending    = "pending"
	StateStarting   = "starting"
	StateRunning    = "running"
	StateTerminated = "terminated"
	StateErrored    = "errored"
)

// Lifecycle event constants.
const (
	EventStart = "start"
	EventReady = "ready"
	EventStop  = "stop"
	EventError = "error"
)

// AgentActor wraps an Agent with HSM-driven lifecycle and presence state machines.
// Per spec §5.4, the lifecycle is managed as an HSM.
type AgentActor struct {
	hsm.HSM
	*api.Agent

	// Presence will be added in next task
}

// NewAgentActor creates a new AgentActor with initialized HSMs.
// Per spec §5.4, the lifecycle starts in the "pending" state.
func NewAgentActor(ctx context.Context, agent *api.Agent) *AgentActor {
	actor := &AgentActor{
		Agent: agent,
	}

	// Initialize and start lifecycle HSM
	actor = hsm.Started(ctx, actor, &LifecycleModel)

	return actor
}

// LifecycleModel defines the agent lifecycle state machine per spec §5.4.
//
// State diagram:
//
//	┌─────────┐    ┌─────────┐    ┌─────────┐    ┌────────────┐
//	│ Pending │───▶│ Starting│───▶│ Running │───▶│ Terminated │
//	└─────────┘    └─────────┘    └─────────┘    └────────────┘
//	                                   │
//	                                   ▼
//	                              ┌─────────┐
//	                              │ Errored │
//	                              └─────────┘
var LifecycleModel = hsm.Define("agent.lifecycle",
	hsm.State(StatePending),

	hsm.State(StateStarting,
		hsm.Entry(func(ctx context.Context, a *AgentActor, e hsm.Event) {
			// Per spec: a.bootstrap() would be called here
			// This will be implemented in Phase 2 when we add agent spawn/attach
		}),
	),

	hsm.State(StateRunning,
		hsm.Entry(func(ctx context.Context, a *AgentActor, e hsm.Event) {
			// Per spec: a.startMonitoring() would be called here
			// This will be implemented in Phase 5 when we add PTY monitoring
		}),
		hsm.Exit(func(ctx context.Context, a *AgentActor, e hsm.Event) {
			// Per spec: a.stopMonitoring() would be called here
			// This will be implemented in Phase 5 when we add PTY monitoring
		}),
	),

	hsm.State(StateTerminated, hsm.Final(StateTerminated)),
	hsm.State(StateErrored, hsm.Final(StateErrored)),

	// Transitions per spec §5.4
	hsm.Transition(hsm.On(hsm.Event{Name: EventStart}), hsm.Source(StatePending), hsm.Target(StateStarting)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventReady}), hsm.Source(StateStarting), hsm.Target(StateRunning)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventStop}), hsm.Source(StateRunning), hsm.Target(StateTerminated)),
	// Error can be triggered from any state
	hsm.Transition(hsm.On(hsm.Event{Name: EventError}), hsm.Source(StatePending), hsm.Target(StateErrored)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventError}), hsm.Source(StateStarting), hsm.Target(StateErrored)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventError}), hsm.Source(StateRunning), hsm.Target(StateErrored)),

	hsm.Initial(hsm.Target(StatePending)),
)

// GetState returns the current lifecycle state.
func (a *AgentActor) GetState() string {
	return a.State()
}

// GetSimpleState returns the current lifecycle state without the qualified name prefix.
// For example, returns "pending" instead of "/agent.lifecycle/pending".
func (a *AgentActor) GetSimpleState() string {
	state := a.State()
	// Extract the last component after the final slash
	if idx := strings.LastIndex(state, "/"); idx >= 0 {
		return state[idx+1:]
	}
	return state
}

// StartAgent transitions the agent from Pending to Starting.
// Per spec §5.4, this is triggered by the "start" event.
func (a *AgentActor) StartAgent(ctx context.Context) {
	done := hsm.Dispatch(ctx, a, hsm.Event{
		Name: EventStart,
	})
	<-done // Wait for transition to complete
}

// Ready transitions the agent from Starting to Running.
// Per spec §5.4, this is triggered by the "ready" event after bootstrap completes.
func (a *AgentActor) Ready(ctx context.Context) {
	done := hsm.Dispatch(ctx, a, hsm.Event{
		Name: EventReady,
	})
	<-done // Wait for transition to complete
}

// StopAgent transitions the agent from Running to Terminated.
// Per spec §5.4, this is triggered by the "stop" event.
func (a *AgentActor) StopAgent(ctx context.Context) {
	done := hsm.Dispatch(ctx, a, hsm.Event{
		Name: EventStop,
	})
	<-done // Wait for transition to complete
}

// ErrorAgent transitions the agent to the Errored state from any state.
// Per spec §5.4, this can be triggered from any state.
func (a *AgentActor) ErrorAgent(ctx context.Context, err error) {
	done := hsm.Dispatch(ctx, a, hsm.Event{
		Name: EventError,
		Data: err,
	})
	<-done // Wait for transition to complete
}
