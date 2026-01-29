# package agent

`import "github.com/agentflare-ai/amux/internal/agent"`

- `EventSpawn, EventStarted, EventExited, EventError, EventStop, EventConnect, EventDisconnect, EventBusy, EventIdle, EventAway, EventBack` — Events
- `func AddAgent(cfg *config.Config, newAgent config.AgentConfig) error` — AddAgent validates and persists a new agent configuration.
- `func EnsureWorktree(repoRoot api.RepoRoot, slug api.AgentSlug, targetBranch string) (string, error)` — EnsureWorktree creates or reuses a worktree for the given agent.
- `func MessageError(format string, a ...any) error` — MessageError creates a formatted error.
- `func NewLifecycleHSM(agent *Agent) hsm.Instance` — NewLifecycleHSM creates a new lifecycle HSM for the agent.
- `func NewPresenceHSM(agent *Agent) hsm.Instance` — NewPresenceHSM creates a new presence HSM for the agent.
- `func RemoveWorktree(repoRoot api.RepoRoot, slug api.AgentSlug) error` — RemoveWorktree removes the worktree for the given agent.
- `func SpawnAgent(ctx context.Context, a *Agent) error` — SpawnAgent starts the agent process in a new PTY session.
- `func StopAgent(ctx context.Context, a *Agent) error` — StopAgent stops the agent process.
- `func ValidateAgentConfig(c config.AgentConfig) error` — ValidateAgentConfig checks required fields.
- `func isGitRepo(path string) bool`
- `func updatePresence(state api.PresenceState) func(context.Context, *PresenceHSM, hsm.Event)`
- `lifecycleModel`
- `presenceModel`
- `type AgentRegistry` — AgentRegistry tracks active agents.
- `type Agent` — Agent represents the runtime state of an agent.
- `type BusEvent` — BusEvent represents an event on the bus.
- `type EventBus` — EventBus manages subscriptions and event distribution.
- `type EventType` — EventType represents the type of event.
- `type LifecycleHSM` — LifecycleHSM manages the agent lifecycle.
- `type LifecycleState` — LifecycleState represents the lifecycle state of an agent.
- `type MergeStrategy` — MergeStrategy represents a git merge strategy.
- `type PresenceHSM` — PresenceHSM manages the agent presence.
- `type PresenceState` — PresenceState represents the presence state of an agent.
- `type RosterEntry` — RosterEntry represents an agent in the roster.
- `type Session` — Session represents a running session of an agent.
- `type Subscription` — Subscription is a channel for receiving events.

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
	hsm.State(string(PresenceOffline), hsm.Entry(updatePresence(api.PresenceOffline))),
	hsm.State(string(PresenceOnline), hsm.Entry(updatePresence(api.PresenceOnline))),
	hsm.State(string(PresenceBusy), hsm.Entry(updatePresence(api.PresenceBusy))),
	hsm.State(string(PresenceAway), hsm.Entry(updatePresence(api.PresenceAway))),

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

#### AddAgent

```go
func AddAgent(cfg *config.Config, newAgent config.AgentConfig) error
```

AddAgent validates and persists a new agent configuration.
It requires the location.repo_path to be a valid git repository.

#### EnsureWorktree

```go
func EnsureWorktree(repoRoot api.RepoRoot, slug api.AgentSlug, targetBranch string) (string, error)
```

EnsureWorktree creates or reuses a worktree for the given agent.
It returns the path to the worktree.

#### MessageError

```go
func MessageError(format string, a ...any) error
```

MessageError creates a formatted error.

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

#### RemoveWorktree

```go
func RemoveWorktree(repoRoot api.RepoRoot, slug api.AgentSlug) error
```

RemoveWorktree removes the worktree for the given agent.

#### SpawnAgent

```go
func SpawnAgent(ctx context.Context, a *Agent) error
```

SpawnAgent starts the agent process in a new PTY session.

#### StopAgent

```go
func StopAgent(ctx context.Context, a *Agent) error
```

StopAgent stops the agent process.

#### ValidateAgentConfig

```go
func ValidateAgentConfig(c config.AgentConfig) error
```

ValidateAgentConfig checks required fields.

#### isGitRepo

```go
func isGitRepo(path string) bool
```

#### updatePresence

```go
func updatePresence(state api.PresenceState) func(context.Context, *PresenceHSM, hsm.Event)
```


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

	// CurrentPresence tracks the current presence state (updated by HSM).
	CurrentPresence api.PresenceState

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


### Methods

#### Agent.GetPresence

```go
func () GetPresence() api.PresenceState
```

GetPresence returns the current presence state of the agent.


## type AgentRegistry

```go
type AgentRegistry struct {
	Agents map[api.AgentID]*Agent
}
```

AgentRegistry tracks active agents.

### Functions returning AgentRegistry

#### NewRegistry

```go
func NewRegistry() *AgentRegistry
```

NewRegistry creates a new registry.


### Methods

#### AgentRegistry.GetRoster

```go
func () GetRoster() []RosterEntry
```

GetRoster returns the list of agents.

#### AgentRegistry.Register

```go
func () Register(a *Agent)
```

Register adds an agent to the registry.


## type BusEvent

```go
type BusEvent struct {
	Type    EventType
	Source  api.AgentID
	Payload interface{}
}
```

BusEvent represents an event on the bus.

## type EventBus

```go
type EventBus struct {
	mu   sync.RWMutex
	subs map[*Subscription]struct{}
}
```

EventBus manages subscriptions and event distribution.

### Functions returning EventBus

#### NewEventBus

```go
func NewEventBus() *EventBus
```

NewEventBus creates a new EventBus.


### Methods

#### EventBus.Publish

```go
func () Publish(event BusEvent)
```

Publish sends an event to all subscribers.

#### EventBus.Subscribe

```go
func () Subscribe() *Subscription
```

Subscribe returns a subscription for all events (for now).
In a real implementation, we'd filter by topic.

#### EventBus.unsubscribe

```go
func () unsubscribe(sub *Subscription)
```


## type EventType

```go
type EventType string
```

EventType represents the type of event.

### Constants

#### EventPresenceUpdate, EventMessage, EventActivity

```go
const (
	EventPresenceUpdate EventType = "presence.update"
	EventMessage        EventType = "message"
	EventActivity       EventType = "activity"
)
```


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


## type MergeStrategy

```go
type MergeStrategy string
```

MergeStrategy represents a git merge strategy.

### Constants

#### MergeSquash, MergeRebase, MergeFFOnly

```go
const (
	MergeSquash MergeStrategy = "squash"
	MergeRebase MergeStrategy = "rebase"
	MergeFFOnly MergeStrategy = "ff-only"
)
```


### Functions returning MergeStrategy

#### SelectMergeStrategy

```go
func SelectMergeStrategy(cfg config.GitConfig, repoRoot api.RepoRoot) (MergeStrategy, string, error)
```

SelectMergeStrategy determines the merge strategy and target branch.
It checks repo_root for current HEAD if target_branch is not configured.


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


## type RosterEntry

```go
type RosterEntry struct {
	AgentID  api.AgentID       `json:"agent_id"`
	Name     string            `json:"name"`
	Adapter  string            `json:"adapter"`
	Presence api.PresenceState `json:"presence"`
	RepoRoot api.RepoRoot      `json:"repo_root"`
}
```

RosterEntry represents an agent in the roster.

## type Session

```go
type Session struct {
	ID        api.SessionID
	AgentID   api.AgentID
	HostID    api.HostID
	StartedAt time.Time

	// Runtime
	Cmd *exec.Cmd
	PTY *os.File
}
```

Session represents a running session of an agent.

## type Subscription

```go
type Subscription struct {
	C      chan BusEvent
	cancel func()
}
```

Subscription is a channel for receiving events.

### Methods

#### Subscription.Close

```go
func () Close()
```

Close unsubscribes.


