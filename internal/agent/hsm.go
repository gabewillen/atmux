package agent

import (
	"context"

	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go"
)

// Events
const (
	EventSpawn      = "spawn"
	EventStarted    = "started"
	EventExited     = "exited"
	EventError      = "error"
	EventStop       = "stop"

	EventConnect          = "connect"
	EventDisconnect       = "disconnect"
	EventBusy             = "busy"
	EventIdle             = "idle"
	EventAway             = "away"
	EventBack             = "back"
	EventRateLimit        = "rate.limit"
	EventRateCleared      = "rate.cleared"
	EventStuck            = "stuck.detected"
	EventActivityDetected = "activity.detected"
)

// LifecycleHSM manages the agent lifecycle.
type LifecycleHSM struct {
	hsm.HSM
	Agent *Agent
}

// PresenceHSM manages the agent presence.
type PresenceHSM struct {
	hsm.HSM
	Agent *Agent
	Bus   *EventBus
}

func updatePresence(state api.PresenceState) func(context.Context, *PresenceHSM, hsm.Event) {
	return func(_ context.Context, sm *PresenceHSM, _ hsm.Event) {
		if sm.Agent != nil {
			sm.Agent.CurrentPresence = state
			if sm.Bus != nil {
				sm.Bus.Publish(BusEvent{
					Type:    EventPresenceUpdate,
					Source:  sm.Agent.ID,
					Payload: state,
				})
			}
		}
	}
}

var lifecycleModel = hsm.Define("lifecycle",
	hsm.State(string(LifecyclePending)),
	hsm.State(string(LifecycleStarting)),
	hsm.State(string(LifecycleRunning)),
	hsm.State(string(LifecycleTerminated)),
	hsm.State(string(LifecycleErrored)),

	hsm.Initial(hsm.Target(string(LifecyclePending))),

	hsm.Transition(
		hsm.On(hsm.Event{Name: EventSpawn}),
		hsm.Source(string(LifecyclePending)),
		hsm.Target(string(LifecycleStarting)),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventStarted}),
		hsm.Source(string(LifecycleStarting)),
		hsm.Target(string(LifecycleRunning)),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventError}),
		hsm.Source(string(LifecycleStarting)),
		hsm.Target(string(LifecycleErrored)),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventExited}),
		hsm.Source(string(LifecycleRunning)),
		hsm.Target(string(LifecycleTerminated)),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventError}),
		hsm.Source(string(LifecycleRunning)),
		hsm.Target(string(LifecycleErrored)),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventStop}),
		hsm.Source(string(LifecycleRunning)),
		hsm.Target(string(LifecycleTerminated)),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventStop}),
		hsm.Source(string(LifecycleStarting)),
		hsm.Target(string(LifecycleTerminated)),
	),
)

var presenceModel = hsm.Define("presence",
	hsm.State(string(PresenceOffline), hsm.Entry(updatePresence(api.PresenceOffline))),
	hsm.State(string(PresenceOnline), hsm.Entry(updatePresence(api.PresenceOnline))),
	hsm.State(string(PresenceBusy), hsm.Entry(updatePresence(api.PresenceBusy))),
	hsm.State(string(PresenceAway), hsm.Entry(updatePresence(api.PresenceAway))),

	hsm.Initial(hsm.Target(string(PresenceOffline))),

	// Connect/Disconnect (Disconnect -> Away per Spec §5.5.8)
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventConnect}),
		hsm.Source(string(PresenceOffline)),
		hsm.Target(string(PresenceOnline)),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventDisconnect}),
		hsm.Source(string(PresenceOnline)),
		hsm.Target(string(PresenceAway)),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventDisconnect}),
		hsm.Source(string(PresenceBusy)),
		hsm.Target(string(PresenceAway)),
	),
	// Explicitly handle Away->Away for idempotency or ignore
	
	// Busy/Idle
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventBusy}),
		hsm.Source(string(PresenceOnline)),
		hsm.Target(string(PresenceBusy)),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventIdle}),
		hsm.Source(string(PresenceBusy)),
		hsm.Target(string(PresenceOnline)),
	),

	// Away (Stuck) / Back (Activity)
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventAway}),
		hsm.Source(string(PresenceOnline)),
		hsm.Target(string(PresenceAway)),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventStuck}),
		hsm.Source(string(PresenceOnline)),
		hsm.Target(string(PresenceAway)),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventStuck}),
		hsm.Source(string(PresenceBusy)),
		hsm.Target(string(PresenceAway)),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventBack}),
		hsm.Source(string(PresenceAway)),
		hsm.Target(string(PresenceOnline)),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventActivityDetected}),
		hsm.Source(string(PresenceAway)),
		hsm.Target(string(PresenceOnline)),
	),

	// Rate Limits (Offline)
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventRateLimit}),
		hsm.Source(string(PresenceOnline)),
		hsm.Target(string(PresenceOffline)),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventRateLimit}),
		hsm.Source(string(PresenceBusy)),
		hsm.Target(string(PresenceOffline)),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventRateCleared}),
		hsm.Source(string(PresenceOffline)),
		hsm.Target(string(PresenceOnline)),
	),
)

// NewLifecycleHSM creates a new lifecycle HSM for the agent.
func NewLifecycleHSM(agent *Agent) hsm.Instance {
	sm := &LifecycleHSM{Agent: agent}
	return hsm.Started(context.Background(), sm, &lifecycleModel)
}

// NewPresenceHSM creates a new presence HSM for the agent.
func NewPresenceHSM(agent *Agent, bus *EventBus) hsm.Instance {
	sm := &PresenceHSM{Agent: agent, Bus: bus}
	return hsm.Started(context.Background(), sm, &presenceModel)
}

