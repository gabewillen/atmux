# package agent

`import "github.com/agentflare-ai/amux/internal/agent"`

Package agent manages agent lifecycle and presence state machines.

- `ErrDispatcherRequired` — ErrDispatcherRequired is returned when a dispatcher is required but missing.
- `LifecycleModel` — LifecycleModel defines the agent lifecycle state machine.
- `LifecyclePending, LifecycleStarting, LifecycleRunning, LifecycleTerminated, LifecycleErrored, EventStart, EventReady, EventStop, EventError, EventShutdownInitiated, EventShutdownForce, PresenceOnline, PresenceBusy, PresenceOffline, PresenceAway, EventTaskAssigned, EventTaskCompleted, EventTaskCancel, EventPromptDetected, EventRateLimit, EventRateCleared, EventStuckDetected, EventActivity, EventAgentStarted, EventAgentStopped, EventPresenceChanged`
- `PresenceModel` — PresenceModel defines the agent presence state machine.
- `type Agent` — Agent represents a runtime agent instance with lifecycle and presence state machines.
- `type EventRouter` — EventRouter routes lifecycle and presence events through NATS subjects.
- `type LifecycleEvent` — LifecycleEvent describes lifecycle state changes.
- `type Lifecycle` — Lifecycle drives the agent lifecycle state machine.
- `type PresenceEvent` — PresenceEvent describes presence state changes.
- `type Presence` — Presence drives the agent presence state machine.

### Constants

#### LifecyclePending, LifecycleStarting, LifecycleRunning, LifecycleTerminated, LifecycleErrored, EventStart, EventReady, EventStop, EventError, EventShutdownInitiated, EventShutdownForce, PresenceOnline, PresenceBusy, PresenceOffline, PresenceAway, EventTaskAssigned, EventTaskCompleted, EventTaskCancel, EventPromptDetected, EventRateLimit, EventRateCleared, EventStuckDetected, EventActivity, EventAgentStarted, EventAgentStopped, EventPresenceChanged

```go
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
```


### Variables

#### ErrDispatcherRequired

```go
var ErrDispatcherRequired = errors.New("dispatcher required")
```

ErrDispatcherRequired is returned when a dispatcher is required but missing.

#### LifecycleModel

```go
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
```

LifecycleModel defines the agent lifecycle state machine.

#### PresenceModel

```go
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

	hsm.Transition(hsm.On(hsm.Event{Name: EventTaskAssigned}), hsm.Source(PresenceOnline), hsm.Target(PresenceBusy)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventTaskCompleted}), hsm.Source(PresenceBusy), hsm.Target(PresenceOnline)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventPromptDetected}), hsm.Source(PresenceBusy), hsm.Target(PresenceOnline)),

	hsm.Transition(hsm.On(hsm.Event{Name: EventRateLimit}), hsm.Source(PresenceBusy), hsm.Target(PresenceOffline)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventRateLimit}), hsm.Source(PresenceOnline), hsm.Target(PresenceOffline)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventRateCleared}), hsm.Source(PresenceOffline), hsm.Target(PresenceOnline)),

	hsm.Transition(hsm.On(hsm.Event{Name: EventStuckDetected}), hsm.Source(PresenceOnline), hsm.Target(PresenceAway)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventStuckDetected}), hsm.Source(PresenceBusy), hsm.Target(PresenceAway)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventStuckDetected}), hsm.Source(PresenceOffline), hsm.Target(PresenceAway)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventStuckDetected}), hsm.Source(PresenceAway), hsm.Target(PresenceAway)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventActivity}), hsm.Source(PresenceAway), hsm.Target(PresenceOnline)),

	hsm.Initial(hsm.Target(PresenceOnline)),
)
```

PresenceModel defines the agent presence state machine.


## type Agent

```go
type Agent struct {
	api.Agent
	Lifecycle  *Lifecycle
	Presence   *Presence
	router     *EventRouter
	dispatcher protocol.Dispatcher
	mu         sync.RWMutex
	lastErr    error
}
```

Agent represents a runtime agent instance with lifecycle and presence state machines.

### Functions returning Agent

#### NewAgent

```go
func NewAgent(meta api.Agent, dispatcher protocol.Dispatcher) (*Agent, error)
```

NewAgent constructs a new agent with lifecycle and presence state machines.


### Methods

#### Agent.EmitLifecycle

```go
func () EmitLifecycle(ctx context.Context, name string, payload any) error
```

EmitLifecycle publishes a lifecycle event through the dispatcher.

#### Agent.EmitPresence

```go
func () EmitPresence(ctx context.Context, name string, payload any) error
```

EmitPresence publishes a presence event through the dispatcher.

#### Agent.LastError

```go
func () LastError() error
```

LastError returns the last error observed by the agent state machines.

#### Agent.Start

```go
func () Start(ctx context.Context)
```

Start starts the lifecycle and presence state machines.

#### Agent.recordError

```go
func () recordError(err error)
```


## type EventRouter

```go
type EventRouter struct {
	agent      *Agent
	dispatcher protocol.Dispatcher
	mu         sync.Mutex
	started    bool
	subs       []protocol.Subscription
}
```

EventRouter routes lifecycle and presence events through NATS subjects.

### Functions returning EventRouter

#### NewEventRouter

```go
func NewEventRouter(agent *Agent, dispatcher protocol.Dispatcher) *EventRouter
```

NewEventRouter constructs a router for an agent.


### Methods

#### EventRouter.EmitLifecycle

```go
func () EmitLifecycle(ctx context.Context, name string, payload any) error
```

EmitLifecycle publishes a lifecycle event.

#### EventRouter.EmitPresence

```go
func () EmitPresence(ctx context.Context, name string, payload any) error
```

EmitPresence publishes a presence event.

#### EventRouter.Start

```go
func () Start(ctx context.Context) error
```

Start subscribes to agent event subjects.

#### EventRouter.emit

```go
func () emit(ctx context.Context, subject string, name string, payload any) error
```


## type Lifecycle

```go
type Lifecycle struct {
	hsm.HSM
	agent      *Agent
	dispatcher protocol.Dispatcher
}
```

Lifecycle drives the agent lifecycle state machine.

### Functions returning Lifecycle

#### NewLifecycle

```go
func NewLifecycle(agent *Agent, dispatcher protocol.Dispatcher) (*Lifecycle, error)
```

NewLifecycle constructs a lifecycle state machine bound to an agent.


### Methods

#### Lifecycle.Start

```go
func () Start(ctx context.Context)
```

Start starts the lifecycle state machine.

#### Lifecycle.emit

```go
func () emit(ctx context.Context, name string, payload any)
```

#### Lifecycle.onErrored

```go
func () onErrored(ctx context.Context, event hsm.Event)
```

#### Lifecycle.onRunning

```go
func () onRunning(ctx context.Context)
```

#### Lifecycle.onStarting

```go
func () onStarting(ctx context.Context)
```

#### Lifecycle.onTerminated

```go
func () onTerminated(ctx context.Context)
```


## type LifecycleEvent

```go
type LifecycleEvent struct {
	AgentID api.AgentID `json:"agent_id"`
	State   string      `json:"state"`
	Error   string      `json:"error,omitempty"`
}
```

LifecycleEvent describes lifecycle state changes.

## type Presence

```go
type Presence struct {
	hsm.HSM
	agent      *Agent
	dispatcher protocol.Dispatcher
}
```

Presence drives the agent presence state machine.

### Functions returning Presence

#### NewPresence

```go
func NewPresence(agent *Agent, dispatcher protocol.Dispatcher) (*Presence, error)
```

NewPresence constructs a presence state machine bound to an agent.


### Methods

#### Presence.Start

```go
func () Start(ctx context.Context)
```

Start starts the presence state machine.

#### Presence.emitChanged

```go
func () emitChanged(ctx context.Context, state string)
```


## type PresenceEvent

```go
type PresenceEvent struct {
	AgentID  api.AgentID `json:"agent_id"`
	Presence string      `json:"presence"`
}
```

PresenceEvent describes presence state changes.

