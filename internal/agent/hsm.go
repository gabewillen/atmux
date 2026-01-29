package agent

import (
	"context"

	"github.com/agentflare-ai/amux/internal/pty"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go"
)

// Event constants for agent state machines.
const (
	EventStart            = "agent.lifecycle.start"
	EventStarted          = "agent.lifecycle.started" // Emitted when process actually runs
	EventStop             = "agent.lifecycle.stop"
	EventTerminated       = "agent.lifecycle.terminated"
	EventError            = "agent.lifecycle.error"
	EventActivityDetected = "agent.presence.activity"
	EventInactivity       = "agent.presence.inactivity"
	EventSetPresence      = "agent.presence.set" // Manual override
	EventRemoteDisconnect = "remote.disconnect"  // Remote host disconnected
)

// Guards helpers
func presenceIs(p api.Presence) func(context.Context, *AgentActor, hsm.Event) bool {
	return func(ctx context.Context, a *AgentActor, e hsm.Event) bool {
		if val, ok := e.Data.(api.Presence); ok {
			return val == p
		}
		// Also support string data for flexibility
		if val, ok := e.Data.(string); ok {
			return api.Presence(val) == p
		}
		return false
	}
}

// Helpers for event definition
func on(name string) hsm.RedefinableElement {
	return hsm.On(hsm.Event{Name: name})
}

// StartAction handles the agent start sequence.
func StartAction(ctx context.Context, a *AgentActor, event hsm.Event) {
	if a.worktree == nil {
		// In tests with nil manager, just simulate immediate success
		hsm.Dispatch(ctx, a, hsm.Event{Name: EventStarted})
		return
	}

	// 1. Ensure Worktree
	path, err := a.worktree.Ensure(a.data)
	if err != nil {
		hsm.Dispatch(ctx, a, hsm.Event{Name: EventError, Data: err})
		return
	}
	a.data.Worktree = path

	// 2. Spawn PTY (default to bash for now)
	ptmx, err := pty.Start("bash", nil, path)
	if err != nil {
		hsm.Dispatch(ctx, a, hsm.Event{Name: EventError, Data: err})
		return
	}
	a.ptyFile = ptmx

	// 3. Emit Started
	hsm.Dispatch(ctx, a, hsm.Event{Name: EventStarted})
}

// StopAction handles the agent stop sequence.
func StopAction(ctx context.Context, a *AgentActor, event hsm.Event) {
	if a.monitor != nil {
		a.monitor.Stop()
		a.monitor = nil
	}
	if a.ptyFile != nil {
		pty.Close(a.ptyFile)
		a.ptyFile = nil
	}
	// We don't remove worktree on stop, only on agent remove.
}

// SetPresenceAction updates the agent's presence data.
func SetPresenceAction(p api.Presence) func(context.Context, *AgentActor, hsm.Event) {
	return func(ctx context.Context, a *AgentActor, e hsm.Event) {
		a.data.Presence = p
	}
}

// AgentModel defines the combined agent state machine.
// Pending -> Starting -> Running (containing Presence) -> Terminated / Errored
var AgentModel = hsm.Define("agent",
	hsm.State("pending",
		hsm.Transition(on(EventStart), hsm.Target("/agent/starting")),
	),
	hsm.State("starting",
		hsm.Entry[*AgentActor](StartAction),
		hsm.Transition(on(EventStarted), hsm.Target("/agent/running")),
		hsm.Transition(on(EventError), hsm.Target("/agent/errored")),
		hsm.Transition(on(EventStop), hsm.Target("/agent/terminated")),
	),
	hsm.State("running",
		// Presence Sub-states
		hsm.State("online",
			hsm.Entry(SetPresenceAction(api.PresenceOnline)),
			hsm.Transition(on(EventActivityDetected), hsm.Target("/agent/running/busy")),
			hsm.Transition(on(EventInactivity), hsm.Target("/agent/running/away")),
			hsm.Transition(on(EventRemoteDisconnect), hsm.Target("/agent/running/away")),
		),
		hsm.State("busy",
			hsm.Entry(SetPresenceAction(api.PresenceBusy)),
			hsm.Transition(on(EventInactivity), hsm.Target("/agent/running/online")),
			hsm.Transition(on(EventRemoteDisconnect), hsm.Target("/agent/running/away")),
		),
		hsm.State("away",
			hsm.Entry(SetPresenceAction(api.PresenceAway)),
			hsm.Transition(on(EventActivityDetected), hsm.Target("/agent/running/online")),
			// Remote disconnect while away stays away
		),
		// Manual presence overrides using Guards
		hsm.Transition(
			on(EventSetPresence), hsm.Target("/agent/running/online"),
			hsm.Guard(presenceIs(api.PresenceOnline)),
		),
		hsm.Transition(
			on(EventSetPresence), hsm.Target("/agent/running/busy"),
			hsm.Guard(presenceIs(api.PresenceBusy)),
		),
		hsm.Transition(
			on(EventSetPresence), hsm.Target("/agent/running/away"),
			hsm.Guard(presenceIs(api.PresenceAway)),
		),
		// Note: "offline" is effectively represented by non-running states or handled by parent.
		// If explicit "offline" state is needed within running (e.g. connected but hidden), add it.
		// For now, adhering to Online/Busy/Away as active presence states.

		hsm.Initial(hsm.Target("/agent/running/online")),

		// Parent transitions
		// Use Effect for actions on transition
		hsm.Transition(on(EventStop), hsm.Target("/agent/terminated"), hsm.Effect[*AgentActor](StopAction)),
		hsm.Transition(on(EventError), hsm.Target("/agent/errored"), hsm.Effect[*AgentActor](StopAction)),
		hsm.Transition(on(EventTerminated), hsm.Target("/agent/terminated"), hsm.Effect[*AgentActor](StopAction)),
	),
	hsm.State("terminated",
		hsm.Entry[*AgentActor](StopAction), // Ensure stopped if entering directly
	),
	hsm.State("errored",
		hsm.Entry[*AgentActor](StopAction), // Ensure stopped if entering directly
	),

	hsm.Initial(hsm.Target("pending")),
)
