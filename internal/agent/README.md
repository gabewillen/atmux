# package agent

`import "github.com/agentflare-ai/amux/internal/agent"`

Package agent provides agent orchestration: lifecycle, presence, and messaging.
actor.go composes lifecycle and presence HSMs and wires dispatch.

Package agent provides agent orchestration: lifecycle, presence, and messaging.
lifecycle.go implements the Agent lifecycle HSM per spec §4.2.3, §5.4.

Package agent provides agent orchestration: lifecycle, presence, and messaging.
presence.go implements the Presence HSM per spec §4.2.3, §6.1, §6.5.

- `EventLifecycleStart, EventLifecycleReady, EventLifecycleStop, EventLifecycleError` — Lifecycle event names for dispatch (spec §5.4).
- `EventPresenceTaskAssigned, EventPresenceTaskCompleted, EventPresencePromptDetected, EventPresenceRateLimit, EventPresenceRateCleared, EventPresenceStuckDetected, EventPresenceActivityDetected` — Presence event names for dispatch (spec §6.5).
- `LifecycleModel` — LifecycleModel defines the agent lifecycle HSM (spec §5.4).
- `LifecyclePending, LifecycleStarting, LifecycleRunning, LifecycleTerminated, LifecycleErrored` — Lifecycle state names (spec §5.4); HSM returns qualified names like /agent.lifecycle/pending.
- `PresenceModel` — PresenceModel defines the presence HSM (spec §6.5).
- `PresenceOnline, PresenceBusy, PresenceOffline, PresenceAway` — Presence state names (spec §6.1); HSM returns qualified names like /agent.presence/online.
- `func emitLifecycleChanged(ctx context.Context, d protocol.Dispatcher, agentID api.ID, state string)`
- `func emitPresenceChanged(ctx context.Context, d protocol.Dispatcher, agentID api.ID, state string)`
- `type Actor` — Actor holds an agent's data and its lifecycle and presence state machines.
- `type lifecycleActor` — lifecycleActor holds HSM state and dispatch hook for agent lifecycle.
- `type presenceActor` — presenceActor holds HSM state and dispatch hook for agent presence.

### Constants

#### LifecyclePending, LifecycleStarting, LifecycleRunning, LifecycleTerminated, LifecycleErrored

```go
const (
	LifecyclePending    = "/agent.lifecycle/pending"
	LifecycleStarting   = "/agent.lifecycle/starting"
	LifecycleRunning    = "/agent.lifecycle/running"
	LifecycleTerminated = "/agent.lifecycle/terminated"
	LifecycleErrored    = "/agent.lifecycle/errored"
)
```

Lifecycle state names (spec §5.4); HSM returns qualified names like /agent.lifecycle/pending.

#### EventLifecycleStart, EventLifecycleReady, EventLifecycleStop, EventLifecycleError

```go
const (
	EventLifecycleStart = "start"
	EventLifecycleReady = "ready"
	EventLifecycleStop  = "stop"
	EventLifecycleError = "error"
)
```

Lifecycle event names for dispatch (spec §5.4).

#### PresenceOnline, PresenceBusy, PresenceOffline, PresenceAway

```go
const (
	PresenceOnline  = "/agent.presence/online"
	PresenceBusy    = "/agent.presence/busy"
	PresenceOffline = "/agent.presence/offline"
	PresenceAway    = "/agent.presence/away"
)
```

Presence state names (spec §6.1); HSM returns qualified names like /agent.presence/online.

#### EventPresenceTaskAssigned, EventPresenceTaskCompleted, EventPresencePromptDetected, EventPresenceRateLimit, EventPresenceRateCleared, EventPresenceStuckDetected, EventPresenceActivityDetected

```go
const (
	EventPresenceTaskAssigned     = "task.assigned"
	EventPresenceTaskCompleted    = "task.completed"
	EventPresencePromptDetected   = "prompt.detected"
	EventPresenceRateLimit        = "rate.limit"
	EventPresenceRateCleared      = "rate.cleared"
	EventPresenceStuckDetected    = "stuck.detected"
	EventPresenceActivityDetected = "activity.detected"
)
```

Presence event names for dispatch (spec §6.5).


### Variables

#### LifecycleModel

```go
var LifecycleModel = hsm.Define("agent.lifecycle",
	hsm.State("pending"),
	hsm.State("starting"),
	hsm.State("running"),
	hsm.State("terminated", hsm.Final("terminated")),
	hsm.State("errored", hsm.Final("errored")),

	hsm.Transition(hsm.On(hsm.Event{Name: EventLifecycleStart}), hsm.Source("pending"), hsm.Target("starting"),
		hsm.Effect(func(ctx context.Context, a *lifecycleActor, _ hsm.Event) {
			emitLifecycleChanged(ctx, a.Dispatcher, a.AgentID, LifecycleStarting)
		})),
	hsm.Transition(hsm.On(hsm.Event{Name: EventLifecycleReady}), hsm.Source("starting"), hsm.Target("running"),
		hsm.Effect(func(ctx context.Context, a *lifecycleActor, _ hsm.Event) {
			emitLifecycleChanged(ctx, a.Dispatcher, a.AgentID, LifecycleRunning)
		})),
	hsm.Transition(hsm.On(hsm.Event{Name: EventLifecycleStop}), hsm.Source("running"), hsm.Target("terminated"),
		hsm.Effect(func(ctx context.Context, a *lifecycleActor, _ hsm.Event) {
			emitLifecycleChanged(ctx, a.Dispatcher, a.AgentID, LifecycleTerminated)
		})),

	hsm.Transition(hsm.On(hsm.Event{Name: EventLifecycleError}), hsm.Source("pending"), hsm.Target("errored"),
		hsm.Effect(func(ctx context.Context, a *lifecycleActor, _ hsm.Event) {
			emitLifecycleChanged(ctx, a.Dispatcher, a.AgentID, LifecycleErrored)
		})),
	hsm.Transition(hsm.On(hsm.Event{Name: EventLifecycleError}), hsm.Source("starting"), hsm.Target("errored"),
		hsm.Effect(func(ctx context.Context, a *lifecycleActor, _ hsm.Event) {
			emitLifecycleChanged(ctx, a.Dispatcher, a.AgentID, LifecycleErrored)
		})),
	hsm.Transition(hsm.On(hsm.Event{Name: EventLifecycleError}), hsm.Source("running"), hsm.Target("errored"),
		hsm.Effect(func(ctx context.Context, a *lifecycleActor, _ hsm.Event) {
			emitLifecycleChanged(ctx, a.Dispatcher, a.AgentID, LifecycleErrored)
		})),

	hsm.Initial(hsm.Target("pending")),
)
```

LifecycleModel defines the agent lifecycle HSM (spec §5.4).
Pending → Starting → Running → Terminated/Errored.

#### PresenceModel

```go
var PresenceModel = hsm.Define("agent.presence",
	hsm.State("online"),
	hsm.State("busy"),
	hsm.State("offline"),
	hsm.State("away"),

	hsm.Transition(hsm.On(hsm.Event{Name: EventPresenceTaskAssigned}), hsm.Source("online"), hsm.Target("busy"),
		hsm.Effect(func(ctx context.Context, a *presenceActor, _ hsm.Event) {
			emitPresenceChanged(ctx, a.Dispatcher, a.AgentID, PresenceBusy)
		})),
	hsm.Transition(hsm.On(hsm.Event{Name: EventPresenceTaskCompleted}), hsm.Source("busy"), hsm.Target("online"),
		hsm.Effect(func(ctx context.Context, a *presenceActor, _ hsm.Event) {
			emitPresenceChanged(ctx, a.Dispatcher, a.AgentID, PresenceOnline)
		})),
	hsm.Transition(hsm.On(hsm.Event{Name: EventPresencePromptDetected}), hsm.Source("busy"), hsm.Target("online"),
		hsm.Effect(func(ctx context.Context, a *presenceActor, _ hsm.Event) {
			emitPresenceChanged(ctx, a.Dispatcher, a.AgentID, PresenceOnline)
		})),

	hsm.Transition(hsm.On(hsm.Event{Name: EventPresenceRateLimit}), hsm.Source("busy"), hsm.Target("offline"),
		hsm.Effect(func(ctx context.Context, a *presenceActor, _ hsm.Event) {
			emitPresenceChanged(ctx, a.Dispatcher, a.AgentID, PresenceOffline)
		})),
	hsm.Transition(hsm.On(hsm.Event{Name: EventPresenceRateLimit}), hsm.Source("online"), hsm.Target("offline"),
		hsm.Effect(func(ctx context.Context, a *presenceActor, _ hsm.Event) {
			emitPresenceChanged(ctx, a.Dispatcher, a.AgentID, PresenceOffline)
		})),
	hsm.Transition(hsm.On(hsm.Event{Name: EventPresenceRateCleared}), hsm.Source("offline"), hsm.Target("online"),
		hsm.Effect(func(ctx context.Context, a *presenceActor, _ hsm.Event) {
			emitPresenceChanged(ctx, a.Dispatcher, a.AgentID, PresenceOnline)
		})),

	hsm.Transition(hsm.On(hsm.Event{Name: EventPresenceStuckDetected}), hsm.Source("online"), hsm.Target("away"),
		hsm.Effect(func(ctx context.Context, a *presenceActor, _ hsm.Event) {
			emitPresenceChanged(ctx, a.Dispatcher, a.AgentID, PresenceAway)
		})),
	hsm.Transition(hsm.On(hsm.Event{Name: EventPresenceStuckDetected}), hsm.Source("busy"), hsm.Target("away"),
		hsm.Effect(func(ctx context.Context, a *presenceActor, _ hsm.Event) {
			emitPresenceChanged(ctx, a.Dispatcher, a.AgentID, PresenceAway)
		})),
	hsm.Transition(hsm.On(hsm.Event{Name: EventPresenceStuckDetected}), hsm.Source("offline"), hsm.Target("away"),
		hsm.Effect(func(ctx context.Context, a *presenceActor, _ hsm.Event) {
			emitPresenceChanged(ctx, a.Dispatcher, a.AgentID, PresenceAway)
		})),
	hsm.Transition(hsm.On(hsm.Event{Name: EventPresenceActivityDetected}), hsm.Source("away"), hsm.Target("online"),
		hsm.Effect(func(ctx context.Context, a *presenceActor, _ hsm.Event) {
			emitPresenceChanged(ctx, a.Dispatcher, a.AgentID, PresenceOnline)
		})),

	hsm.Initial(hsm.Target("online")),
)
```

PresenceModel defines the presence HSM (spec §6.5).
Online ↔ Busy ↔ Offline ↔ Away.


### Functions

#### emitLifecycleChanged

```go
func emitLifecycleChanged(ctx context.Context, d protocol.Dispatcher, agentID api.ID, state string)
```

#### emitPresenceChanged

```go
func emitPresenceChanged(ctx context.Context, d protocol.Dispatcher, agentID api.ID, state string)
```


## type Actor

```go
type Actor struct {
	Agent      *api.Agent
	Dispatcher protocol.Dispatcher

	lifecycle hsm.Instance
	presence  hsm.Instance
	mu        sync.RWMutex
}
```

Actor holds an agent's data and its lifecycle and presence state machines.
Lifecycle and presence transitions emit events via the configured Dispatcher.

### Functions returning Actor

#### NewActor

```go
func NewActor(agent *api.Agent, d protocol.Dispatcher) (*Actor, error)
```

NewActor creates an actor for the given agent and dispatcher.
The agent ID MUST be a valid runtime ID (non-zero). Call Start to run the HSMs.


### Methods

#### Actor.DispatchLifecycle

```go
func () DispatchLifecycle(ctx context.Context, eventName string, data interface{})
```

DispatchLifecycle sends an event to the lifecycle HSM (start, ready, stop, error).
It blocks until the transition is processed.

#### Actor.DispatchPresence

```go
func () DispatchPresence(ctx context.Context, eventName string, data interface{})
```

DispatchPresence sends an event to the presence HSM (task.assigned, activity.detected, etc.).
It blocks until the transition is processed.

#### Actor.LifecycleState

```go
func () LifecycleState() string
```

LifecycleState returns the current lifecycle state name (e.g. agent.lifecycle.pending).

#### Actor.PresenceState

```go
func () PresenceState() string
```

PresenceState returns the current presence state name (e.g. agent.presence.online).

#### Actor.Start

```go
func () Start(ctx context.Context)
```

Start starts the lifecycle and presence HSMs. Call once after NewActor.


## type lifecycleActor

```go
type lifecycleActor struct {
	hsm.HSM
	AgentID    api.ID
	Dispatcher protocol.Dispatcher
	mu         sync.Mutex
}
```

lifecycleActor holds HSM state and dispatch hook for agent lifecycle.

## type presenceActor

```go
type presenceActor struct {
	hsm.HSM
	AgentID    api.ID
	Dispatcher protocol.Dispatcher
	mu         sync.Mutex
}
```

presenceActor holds HSM state and dispatch hook for agent presence.

