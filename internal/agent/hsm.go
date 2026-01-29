package agent

import (
	"context"

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

// AgentModel defines the combined agent state machine.
// Pending -> Starting -> Running (containing Presence) -> Terminated / Errored
var AgentModel = hsm.Define("agent",
	hsm.State("pending",
		hsm.Transition(on(EventStart), hsm.Target("/agent/starting")),
	),
	hsm.State("starting",
		hsm.Transition(on(EventStarted), hsm.Target("/agent/running")),
		hsm.Transition(on(EventError), hsm.Target("/agent/errored")),
		hsm.Transition(on(EventStop), hsm.Target("/agent/terminated")),
	),
	hsm.State("running",
		// Presence Sub-states
		hsm.State("online",
			hsm.Transition(on(EventActivityDetected), hsm.Target("/agent/running/busy")),
			hsm.Transition(on(EventInactivity), hsm.Target("/agent/running/away")),
		),
		hsm.State("busy",
			hsm.Transition(on(EventInactivity), hsm.Target("/agent/running/online")),
		),
		hsm.State("away",
			hsm.Transition(on(EventActivityDetected), hsm.Target("/agent/running/online")),
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
		hsm.Transition(on(EventStop), hsm.Target("/agent/terminated")),
		hsm.Transition(on(EventError), hsm.Target("/agent/errored")),
		hsm.Transition(on(EventTerminated), hsm.Target("/agent/terminated")),
	),
	hsm.State("terminated"),
	hsm.State("errored"),

	hsm.Initial(hsm.Target("pending")),
)
