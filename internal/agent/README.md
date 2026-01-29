# package agent

`import "github.com/agentflare-ai/amux/internal/agent"`

- `EventSpawn, EventStarted, EventExited, EventError, EventStop, EventConnect, EventDisconnect, EventBusy, EventIdle, EventAway, EventBack` — Events
- `func NewLifecycleHSM(agent *Agent) hsm.Instance` — NewLifecycleHSM creates a new lifecycle HSM for the agent.
- `func NewPresenceHSM(agent *Agent) hsm.Instance` — NewPresenceHSM creates a new presence HSM for the agent.
- `lifecycleModel`
- `presenceModel`
- `type Agent` — Agent represents the runtime state of an agent.
- `type LifecycleHSM` — LifecycleHSM manages the agent lifecycle.
- `type LifecycleState` — LifecycleState represents the lifecycle state of an agent.
- `type PresenceHSM` — PresenceHSM manages the agent presence.
- `type PresenceState` — PresenceState represents the presence state of an agent.
- `type Session` — Session represents a running session of an agent.

### Constants

#### EventSpawn, EventStarted, EventExited, EventError, EventStop, EventConnect, EventDisconnect, EventBusy, EventIdle, EventAway, EventBack

```go
const (
	EventSpawn   = "spawn"
	EventStarted = "started"
	EventExited  = "exited"
	EventError   = "error"
	EventStop    = "stop"

	EventConnect    = "connect"
	EventDisconnect = "disconnect"
	EventBusy       = "busy"
	EventIdle       = "idle"
	EventAway       = "away"
	EventBack       = "back"
)
```

Events


### Variables

#### lifecycleModel

```go
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
```

#### presenceModel

```go
var presenceModel = hsm.Define("presence",
	hsm.State(string(PresenceOffline)),
	hsm.State(string(PresenceOnline)),
	hsm.State(string(PresenceBusy)),
	hsm.State(string(PresenceAway)),

	hsm.Initial(hsm.Target(string(PresenceOffline))),

	hsm.Transition(
		hsm.On(hsm.Event{Name: EventConnect}),
		hsm.Source(string(PresenceOffline)),
		hsm.Target(string(PresenceOnline)),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventDisconnect}),
		hsm.Source(string(PresenceOnline)),
		hsm.Target(string(PresenceOffline)),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventDisconnect}),
		hsm.Source(string(PresenceBusy)),
		hsm.Target(string(PresenceOffline)),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventDisconnect}),
		hsm.Source(string(PresenceAway)),
		hsm.Target(string(PresenceOffline)),
	),
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
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventAway}),
		hsm.Source(string(PresenceOnline)),
		hsm.Target(string(PresenceAway)),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: EventBack}),
		hsm.Source(string(PresenceAway)),
		hsm.Target(string(PresenceOnline)),
	),
)
```


### Functions

#### NewLifecycleHSM

```go
func NewLifecycleHSM(agent *Agent) hsm.Instance
```

NewLifecycleHSM creates a new lifecycle HSM for the agent.

#### NewPresenceHSM

```go
func NewPresenceHSM(agent *Agent) hsm.Instance
```

NewPresenceHSM creates a new presence HSM for the agent.


## type Agent

```go
type Agent struct {
	ID       api.AgentID
	Slug     api.AgentSlug
	Name     string
	RepoRoot api.RepoRoot
	Config   config.AgentConfig

	// State machines
	Lifecycle hsm.Instance
	Presence  hsm.Instance

	// Sessions active for this agent
	Sessions map[api.SessionID]*Session
}
```

Agent represents the runtime state of an agent.

### Functions returning Agent

#### NewAgent

```go
func NewAgent(cfg config.AgentConfig, repoRoot api.RepoRoot) (*Agent, error)
```

NewAgent creates a new Agent instance.


## type LifecycleHSM

```go
type LifecycleHSM struct {
	hsm.HSM
	Agent *Agent
}
```

LifecycleHSM manages the agent lifecycle.

## type LifecycleState

```go
type LifecycleState string
```

LifecycleState represents the lifecycle state of an agent.

### Constants

#### LifecyclePending, LifecycleStarting, LifecycleRunning, LifecycleTerminated, LifecycleErrored

```go
const (
	LifecyclePending    LifecycleState = "Pending"
	LifecycleStarting   LifecycleState = "Starting"
	LifecycleRunning    LifecycleState = "Running"
	LifecycleTerminated LifecycleState = "Terminated"
	LifecycleErrored    LifecycleState = "Errored"
)
```


## type PresenceHSM

```go
type PresenceHSM struct {
	hsm.HSM
	Agent *Agent
}
```

PresenceHSM manages the agent presence.

## type PresenceState

```go
type PresenceState string
```

PresenceState represents the presence state of an agent.

### Constants

#### PresenceOnline, PresenceBusy, PresenceOffline, PresenceAway

```go
const (
	PresenceOnline  PresenceState = "Online"
	PresenceBusy    PresenceState = "Busy"
	PresenceOffline PresenceState = "Offline"
	PresenceAway    PresenceState = "Away"
)
```


## type Session

```go
type Session struct {
	ID        api.SessionID
	AgentID   api.AgentID
	HostID    api.HostID
	StartedAt time.Time
}
```

Session represents a running session of an agent.

