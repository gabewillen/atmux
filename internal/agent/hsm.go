package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go"
)

const (
	// LifecyclePending is the initial lifecycle state.
	LifecyclePending = "pending"
	// LifecycleStarting indicates the agent is starting.
	LifecycleStarting = "starting"
	// LifecycleRunning indicates the agent is running.
	LifecycleRunning = "running"
	// LifecycleTerminated indicates the agent terminated.
	LifecycleTerminated = "terminated"
	// LifecycleErrored indicates the agent errored.
	LifecycleErrored = "errored"

	// EventStart triggers the lifecycle start transition.
	EventStart = "start"
	// EventReady triggers the lifecycle ready transition.
	EventReady = "ready"
	// EventStop triggers the lifecycle stop transition.
	EventStop = "stop"
	// EventError triggers the lifecycle error transition.
	EventError = "error"
	// EventShutdownInitiated triggers graceful shutdown.
	EventShutdownInitiated = "shutdown.initiated"
	// EventShutdownForce triggers forced shutdown.
	EventShutdownForce = "shutdown.force"

	// PresenceOnline indicates the agent is available.
	PresenceOnline = "online"
	// PresenceBusy indicates the agent is working.
	PresenceBusy = "busy"
	// PresenceOffline indicates the agent is offline.
	PresenceOffline = "offline"
	// PresenceAway indicates the agent is away or unresponsive.
	PresenceAway = "away"

	// EventTaskAssigned marks task assignment.
	EventTaskAssigned = "task.assigned"
	// EventTaskCompleted marks task completion.
	EventTaskCompleted = "task.completed"
	// EventTaskCancel requests task cancellation.
	EventTaskCancel = "task.cancel"
	// EventPromptDetected indicates a prompt was detected.
	EventPromptDetected = "prompt.detected"
	// EventRateLimit indicates rate limiting.
	EventRateLimit = "rate.limit"
	// EventRateCleared clears a rate limit.
	EventRateCleared = "rate.cleared"
	// EventStuckDetected marks a stuck agent.
	EventStuckDetected = "stuck.detected"
	// EventActivity marks agent activity.
	EventActivity = "activity.detected"

	// EventAgentStarted is emitted when the agent starts running.
	EventAgentStarted = "agent.started"
	// EventAgentStopped is emitted when the agent stops.
	EventAgentStopped = "agent.stopped"
	// EventPresenceChanged is emitted when presence changes.
	EventPresenceChanged = "presence.changed"
)

// LifecycleModel defines the agent lifecycle state machine.
var LifecycleModel = hsm.Define(
	"agent.lifecycle",
	hsm.State(LifecyclePending),
	hsm.State(
		LifecycleStarting,
		hsm.Entry(func(ctx context.Context, actor *Lifecycle, event hsm.Event) {
			actor.onStarting(ctx)
		}),
	),
	hsm.State(
		LifecycleRunning,
		hsm.Entry(func(ctx context.Context, actor *Lifecycle, event hsm.Event) {
			actor.onRunning(ctx)
		}),
	),
	hsm.Final(LifecycleTerminated),
	hsm.Final(LifecycleErrored),

	hsm.Transition(hsm.On(hsm.Event{Name: EventStart}), hsm.Source(LifecyclePending), hsm.Target(LifecycleStarting)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventReady}), hsm.Source(LifecycleStarting), hsm.Target(LifecycleRunning)),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventStop}),
		hsm.Source(LifecycleRunning),
		hsm.Target(LifecycleTerminated),
		hsm.Effect(func(ctx context.Context, actor *Lifecycle, event hsm.Event) {
			actor.onTerminated(ctx)
		}),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventShutdownInitiated}),
		hsm.Source(LifecyclePending),
		hsm.Target(LifecycleTerminated),
		hsm.Effect(func(ctx context.Context, actor *Lifecycle, event hsm.Event) {
			actor.onTerminated(ctx)
		}),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventShutdownInitiated}),
		hsm.Source(LifecycleStarting),
		hsm.Target(LifecycleTerminated),
		hsm.Effect(func(ctx context.Context, actor *Lifecycle, event hsm.Event) {
			actor.onTerminated(ctx)
		}),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventShutdownInitiated}),
		hsm.Source(LifecycleRunning),
		hsm.Target(LifecycleTerminated),
		hsm.Effect(func(ctx context.Context, actor *Lifecycle, event hsm.Event) {
			actor.onTerminated(ctx)
		}),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventError}),
		hsm.Source(LifecyclePending),
		hsm.Target(LifecycleErrored),
		hsm.Effect(func(ctx context.Context, actor *Lifecycle, event hsm.Event) {
			actor.onErrored(ctx, event)
		}),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventShutdownForce}),
		hsm.Source(LifecyclePending),
		hsm.Target(LifecycleTerminated),
		hsm.Effect(func(ctx context.Context, actor *Lifecycle, event hsm.Event) {
			actor.onTerminated(ctx)
		}),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventShutdownForce}),
		hsm.Source(LifecycleStarting),
		hsm.Target(LifecycleTerminated),
		hsm.Effect(func(ctx context.Context, actor *Lifecycle, event hsm.Event) {
			actor.onTerminated(ctx)
		}),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventShutdownForce}),
		hsm.Source(LifecycleRunning),
		hsm.Target(LifecycleTerminated),
		hsm.Effect(func(ctx context.Context, actor *Lifecycle, event hsm.Event) {
			actor.onTerminated(ctx)
		}),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventError}),
		hsm.Source(LifecycleStarting),
		hsm.Target(LifecycleErrored),
		hsm.Effect(func(ctx context.Context, actor *Lifecycle, event hsm.Event) {
			actor.onErrored(ctx, event)
		}),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventError}),
		hsm.Source(LifecycleRunning),
		hsm.Target(LifecycleErrored),
		hsm.Effect(func(ctx context.Context, actor *Lifecycle, event hsm.Event) {
			actor.onErrored(ctx, event)
		}),
	),

	hsm.Initial(hsm.Target(LifecyclePending)),
)

// PresenceModel defines the agent presence state machine.
var PresenceModel = hsm.Define(
	"agent.presence",
	hsm.State(
		PresenceOnline,
		hsm.Entry(func(ctx context.Context, actor *Presence, event hsm.Event) {
			actor.emitChanged(ctx, PresenceOnline)
		}),
	),
	hsm.State(
		PresenceBusy,
		hsm.Entry(func(ctx context.Context, actor *Presence, event hsm.Event) {
			actor.emitChanged(ctx, PresenceBusy)
		}),
	),
	hsm.State(
		PresenceOffline,
		hsm.Entry(func(ctx context.Context, actor *Presence, event hsm.Event) {
			actor.emitChanged(ctx, PresenceOffline)
		}),
	),
	hsm.State(
		PresenceAway,
		hsm.Entry(func(ctx context.Context, actor *Presence, event hsm.Event) {
			actor.emitChanged(ctx, PresenceAway)
		}),
	),

	// Online ↔ Busy
	hsm.Transition(hsm.On(hsm.Event{Name: EventTaskAssigned}), hsm.Source(PresenceOnline), hsm.Target(PresenceBusy)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventTaskCompleted}), hsm.Source(PresenceBusy), hsm.Target(PresenceOnline)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventPromptDetected}), hsm.Source(PresenceBusy), hsm.Target(PresenceOnline)),

	// → Offline
	hsm.Transition(hsm.On(hsm.Event{Name: EventRateLimit}), hsm.Source(PresenceBusy), hsm.Target(PresenceOffline)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventRateLimit}), hsm.Source(PresenceOnline), hsm.Target(PresenceOffline)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventRateCleared}), hsm.Source(PresenceOffline), hsm.Target(PresenceOnline)),

	// → Away
	hsm.Transition(hsm.On(hsm.Event{Name: EventStuckDetected}), hsm.Source(PresenceOnline), hsm.Target(PresenceAway)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventStuckDetected}), hsm.Source(PresenceBusy), hsm.Target(PresenceAway)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventStuckDetected}), hsm.Source(PresenceOffline), hsm.Target(PresenceAway)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventStuckDetected}), hsm.Source(PresenceAway), hsm.Target(PresenceAway)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventActivity}), hsm.Source(PresenceAway), hsm.Target(PresenceOnline)),

	hsm.Initial(hsm.Target(PresenceOnline)),
)

// Lifecycle drives the agent lifecycle state machine.
type Lifecycle struct {
	hsm.HSM
	agent      *Agent
	dispatcher protocol.Dispatcher
}

// Presence drives the agent presence state machine.
type Presence struct {
	hsm.HSM
	agent      *Agent
	dispatcher protocol.Dispatcher
}

// NewLifecycle constructs a lifecycle state machine bound to an agent.
func NewLifecycle(agent *Agent, dispatcher protocol.Dispatcher) (*Lifecycle, error) {
	if dispatcher == nil {
		return nil, fmt.Errorf("new lifecycle: %w", ErrDispatcherRequired)
	}
	return &Lifecycle{agent: agent, dispatcher: dispatcher}, nil
}

// NewPresence constructs a presence state machine bound to an agent.
func NewPresence(agent *Agent, dispatcher protocol.Dispatcher) (*Presence, error) {
	if dispatcher == nil {
		return nil, fmt.Errorf("new presence: %w", ErrDispatcherRequired)
	}
	return &Presence{agent: agent, dispatcher: dispatcher}, nil
}

// Start starts the lifecycle state machine.
func (l *Lifecycle) Start(ctx context.Context) {
	hsm.Started(ctx, l, &LifecycleModel)
}

// Start starts the presence state machine.
func (p *Presence) Start(ctx context.Context) {
	hsm.Started(ctx, p, &PresenceModel)
}

func (l *Lifecycle) onStarting(ctx context.Context) {
	// Placeholder for startup hooks.
	_ = ctx
}

func (l *Lifecycle) onRunning(ctx context.Context) {
	l.emit(ctx, EventAgentStarted, LifecycleEvent{AgentID: l.agent.ID, State: LifecycleRunning})
}

func (l *Lifecycle) onTerminated(ctx context.Context) {
	l.emit(ctx, EventAgentStopped, LifecycleEvent{AgentID: l.agent.ID, State: LifecycleTerminated})
}

func (l *Lifecycle) onErrored(ctx context.Context, event hsm.Event) {
	payload := LifecycleEvent{AgentID: l.agent.ID, State: LifecycleErrored}
	if event.Data != nil {
		payload.Error = fmt.Sprintf("%v", event.Data)
	}
	l.emit(ctx, EventAgentStopped, payload)
}

func (l *Lifecycle) emit(ctx context.Context, name string, payload any) {
	event := protocol.Event{Name: name, Payload: payload, OccurredAt: time.Now().UTC()}
	if err := l.dispatcher.Publish(ctx, protocol.Subject("events", "agent"), event); err != nil {
		l.agent.recordError(fmt.Errorf("emit lifecycle event: %w", err))
	}
}

func (p *Presence) emitChanged(ctx context.Context, state string) {
	payload := PresenceEvent{AgentID: p.agent.ID, Presence: state}
	event := protocol.Event{Name: EventPresenceChanged, Payload: payload, OccurredAt: time.Now().UTC()}
	if err := p.dispatcher.Publish(ctx, protocol.Subject("events", "presence"), event); err != nil {
		p.agent.recordError(fmt.Errorf("emit presence event: %w", err))
	}
}

// LifecycleEvent describes lifecycle state changes.
type LifecycleEvent struct {
	AgentID api.AgentID `json:"agent_id"`
	State   string      `json:"state"`
	Error   string      `json:"error,omitempty"`
}

// PresenceEvent describes presence state changes.
type PresenceEvent struct {
	AgentID  api.AgentID `json:"agent_id"`
	Presence string      `json:"presence"`
}
