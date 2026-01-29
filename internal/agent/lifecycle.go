// Package agent implements agent lifecycle and presence state machines
// using hsm-go library. This package contains the core state machine logic
// for managing agent instances and their presence states.
package agent

import (
	"context"
	"log/slog"

	"github.com/stateforward/hsm-go"

	"github.com/agentflare-ai/amux/pkg/api"
)

// AgentLifecycle represents the agent lifecycle state machine.
// The lifecycle follows: Pending → Starting → Running → Terminated/Errored
type AgentLifecycle struct {
	hsm.HSM
	agent *api.Agent
	model *hsm.Model
	ctx   context.Context
}

// AgentLifecycleEvent represents events that can trigger state transitions
// in the agent lifecycle state machine.
var (
	// EventStart is triggered when an agent should be started.
	EventStart = hsm.Event{Name: "start"}

	// EventStarted is triggered when an agent process has started.
	EventStarted = hsm.Event{Name: "started"}

	// EventStop is triggered when an agent should be stopped.
	EventStop = hsm.Event{Name: "stop"}

	// EventStopped is triggered when an agent process has stopped normally.
	EventStopped = hsm.Event{Name: "stopped"}

	// EventError is triggered when an error occurs during agent operation.
	EventError = hsm.Event{Name: "error"}

	// EventRestart is triggered when an agent should be restarted.
	EventRestart = hsm.Event{Name: "restart"}
)

// NewAgentLifecycle creates a new agent lifecycle state machine.
func NewAgentLifecycle(ctx context.Context, agent *api.Agent) *AgentLifecycle {
	lifecycle := &AgentLifecycle{
		agent: agent,
		ctx:   ctx,
	}

	// Build the state machine model
	lifecycle.model = lifecycle.buildStateMachine()

	// Start the state machine
	hsm.Started(ctx, lifecycle, lifecycle.model)

	return lifecycle
}

// buildStateMachine constructs the agent lifecycle state machine model.
func (al *AgentLifecycle) buildStateMachine() *hsm.Model {
	agentID := api.IDToString(al.agent.ID)

	return hsm.Define(
		"agent_lifecycle_"+agentID,

		// Define states
		hsm.State("pending",
			hsm.Transition(
				hsm.On(EventStart),
				hsm.Source("pending"),
				hsm.Target("starting"),
				hsm.Effect(func(ctx context.Context, hsm *AgentLifecycle, event hsm.Event) {
					al.logStateChange("pending", "starting")
				}),
			),
		),

		hsm.State("starting",
			hsm.Transition(
				hsm.On(EventStarted),
				hsm.Source("starting"),
				hsm.Target("running"),
				hsm.Effect(func(ctx context.Context, hsm *AgentLifecycle, event hsm.Event) {
					al.logStateChange("starting", "running")
				}),
			),
			hsm.Transition(
				hsm.On(EventError),
				hsm.Source("starting"),
				hsm.Target("errored"),
				hsm.Effect(func(ctx context.Context, hsm *AgentLifecycle, event hsm.Event) {
					al.logStateChange("starting", "errored")
				}),
			),
		),

		hsm.State("running",
			hsm.Transition(
				hsm.On(EventStop),
				hsm.Source("running"),
				hsm.Target("pending"),
				hsm.Effect(func(ctx context.Context, hsm *AgentLifecycle, event hsm.Event) {
					al.logStateChange("running", "pending")
				}),
			),
			hsm.Transition(
				hsm.On(EventStopped),
				hsm.Source("running"),
				hsm.Target("terminated"),
				hsm.Effect(func(ctx context.Context, hsm *AgentLifecycle, event hsm.Event) {
					al.logStateChange("running", "terminated")
				}),
			),
			hsm.Transition(
				hsm.On(EventError),
				hsm.Source("running"),
				hsm.Target("errored"),
				hsm.Effect(func(ctx context.Context, hsm *AgentLifecycle, event hsm.Event) {
					al.logStateChange("running", "errored")
				}),
			),
			hsm.Transition(
				hsm.On(EventRestart),
				hsm.Source("running"),
				hsm.Target("starting"),
				hsm.Effect(func(ctx context.Context, hsm *AgentLifecycle, event hsm.Event) {
					al.logStateChange("running", "starting")
				}),
			),
		),

		hsm.State("terminated"),
		hsm.State("errored",
			hsm.Transition(
				hsm.On(EventRestart),
				hsm.Source("errored"),
				hsm.Target("starting"),
				hsm.Effect(func(ctx context.Context, hsm *AgentLifecycle, event hsm.Event) {
					al.logStateChange("errored", "starting")
				}),
			),
		),

		// Set the initial state
		hsm.Initial(hsm.Target("pending")),
	)
}

// logStateChange logs state transitions for debugging.
func (al *AgentLifecycle) logStateChange(from, to string) {
	slog.Info("Agent lifecycle state transition",
		"agent_id", al.agent.ID,
		"agent_name", al.agent.Name,
		"from", from,
		"to", to,
	)
}

// CurrentState returns the current state of the agent lifecycle.
func (al *AgentLifecycle) CurrentState() string {
	return hsm.State(al)
}

// IsRunning returns true if the agent is in the running state.
func (al *AgentLifecycle) IsRunning() bool {
	return hsm.State(al) == "running"
}

// IsTerminated returns true if the agent is in a terminal state (terminated or errored).
func (al *AgentLifecycle) IsTerminated() bool {
	state := hsm.State(al)
	return state == "terminated" || state == "errored"
}

// Start initiates the agent startup process.
func (al *AgentLifecycle) Start() error {
	<-hsm.Dispatch(al.ctx, al, EventStart)
	return nil
}

// Started notifies that the agent has started successfully.
func (al *AgentLifecycle) Started() error {
	<-hsm.Dispatch(al.ctx, al, EventStarted)
	return nil
}

// Stop initiates the agent shutdown process.
func (al *AgentLifecycle) Stop() error {
	<-hsm.Dispatch(al.ctx, al, EventStop)
	return nil
}

// Stopped notifies that the agent has stopped.
func (al *AgentLifecycle) Stopped() error {
	<-hsm.Dispatch(al.ctx, al, EventStopped)
	return nil
}

// Error notifies that an error occurred.
func (al *AgentLifecycle) Error() error {
	<-hsm.Dispatch(al.ctx, al, EventError)
	return nil
}

// Restart initiates an agent restart.
func (al *AgentLifecycle) Restart() error {
	<-hsm.Dispatch(al.ctx, al, EventRestart)
	return nil
}

// Agent returns the agent associated with this lifecycle.
func (al *AgentLifecycle) Agent() *api.Agent {
	return al.agent
}
