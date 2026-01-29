# package agent

`import "github.com/agentflare-ai/amux/internal/agent"`

Package agent provides agent orchestration: lifecycle, presence, and messaging.
actor.go composes lifecycle and presence HSMs and wires dispatch.

Package agent provides agent orchestration: lifecycle, presence, and messaging.
add.go implements agent add validation and config building (spec §5.2, §5.3.1).

Package agent provides agent orchestration: lifecycle, presence, and messaging.
lifecycle.go implements the Agent lifecycle HSM per spec §4.2.3, §5.4.

Package agent provides agent orchestration: lifecycle, presence, and messaging.
local.go implements local agent lifecycle operations: spawn, stop, restart (spec §5.4, §5.6).

Package agent provides agent orchestration: lifecycle, presence, and messaging.
presence.go implements the Presence HSM per spec §4.2.3, §6.1, §6.5.

- `ErrInvalidLocation` — ErrInvalidLocation is returned when location type is invalid.
- `ErrNotInRepo` — ErrNotInRepo is returned when adding an agent outside a git repository.
- `EventLifecycleStart, EventLifecycleReady, EventLifecycleStop, EventLifecycleError` — Lifecycle event names for dispatch (spec §5.4).
- `EventPresenceTaskAssigned, EventPresenceTaskCompleted, EventPresencePromptDetected, EventPresenceRateLimit, EventPresenceRateCleared, EventPresenceStuckDetected, EventPresenceActivityDetected` — Presence event names for dispatch (spec §6.5).
- `LifecycleModel` — LifecycleModel defines the agent lifecycle HSM (spec §5.4).
- `LifecyclePending, LifecycleStarting, LifecycleRunning, LifecycleTerminated, LifecycleErrored` — Lifecycle state names (spec §5.4); HSM returns qualified names like /agent.lifecycle/pending.
- `PresenceModel` — PresenceModel defines the presence HSM (spec §6.5).
- `PresenceOnline, PresenceBusy, PresenceOffline, PresenceAway` — Presence state names (spec §6.1); HSM returns qualified names like /agent.presence/online.
- `func BuildAgentConfig(in *AddInput, agentSlug string) config.AgentConfig` — BuildAgentConfig builds an AgentConfig for persistence from AddInput.
- `func ResolveRepoRoot(homeDir, cwd, repoPath string) (string, error)` — ResolveRepoRoot returns the canonical repo root for a local add.
- `func emitLifecycleChanged(ctx context.Context, d protocol.Dispatcher, agentID api.ID, state string)`
- `func emitPresenceChanged(ctx context.Context, d protocol.Dispatcher, agentID api.ID, state string)`
- `type Actor` — Actor holds an agent's data and its lifecycle and presence state machines.
- `type AddInput` — AddInput holds validated inputs for adding an agent (spec §5.2).
- `type LocalSession` — LocalSession holds a local agent's actor, PTY session, and worktree path (spec §5.4, §7).
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

#### ErrInvalidLocation

```go
var ErrInvalidLocation = errors.New("invalid location type")
```

ErrInvalidLocation is returned when location type is invalid.

#### ErrNotInRepo

```go
var ErrNotInRepo = errors.New("not in a git repository")
```

ErrNotInRepo is returned when adding an agent outside a git repository.

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

#### BuildAgentConfig

```go
func BuildAgentConfig(in *AddInput, agentSlug string) config.AgentConfig
```

BuildAgentConfig builds an AgentConfig for persistence from AddInput.
agentSlug is the uniquified slug assigned to this agent (for worktree path and persistence).

#### ResolveRepoRoot

```go
func ResolveRepoRoot(homeDir, cwd, repoPath string) (string, error)
```

ResolveRepoRoot returns the canonical repo root for a local add.
If repoPath is non-empty, canonicalizes it; otherwise uses git.Root(cwd).

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


## type AddInput

```go
type AddInput struct {
	Name     string
	About    string
	Adapter  string
	RepoRoot string // Canonical repo root; must be a git repo
	Location config.AgentLocationConfig
}
```

AddInput holds validated inputs for adding an agent (spec §5.2).

### Functions returning AddInput

#### ValidateAddInput

```go
func ValidateAddInput(repoRoot, name, about, adapter string, location config.AgentLocationConfig) (*AddInput, error)
```

ValidateAddInput validates inputs for adding an agent.
Adding an agent outside a git repo fails (spec §1.3, §5.2).
For local agents, repoRoot is used; if empty, the caller must resolve from cwd.


## type LocalSession

```go
type LocalSession struct {
	AgentConfig config.AgentConfig
	RepoRoot    string

	actor *Actor
	sess  *pty.Session
	agent *api.Agent
	disp  protocol.Dispatcher
	mu    sync.Mutex
}
```

LocalSession holds a local agent's actor, PTY session, and worktree path (spec §5.4, §7).
Lifecycle HSM transitions align to Spawn/Stop; PTY is started in the agent workdir.

### Functions returning LocalSession

#### NewLocalSession

```go
func NewLocalSession(ac config.AgentConfig, repoRoot string, disp protocol.Dispatcher) *LocalSession
```

NewLocalSession creates a session holder for a local agent. Call Spawn to start.


### Methods

#### LocalSession.Actor

```go
func () Actor() *Actor
```

Actor returns the agent actor, or nil if not spawned.

#### LocalSession.Agent

```go
func () Agent() *api.Agent
```

Agent returns the api.Agent for this session, or nil if not spawned.

#### LocalSession.PTY

```go
func () PTY() *pty.Session
```

PTY returns the PTY session, or nil if not spawned.

#### LocalSession.Restart

```go
func () Restart(ctx context.Context, command []string, env []string) error
```

Restart stops then spawns again with the same command. Caller must pass the same command and env as Spawn.

#### LocalSession.Spawn

```go
func () Spawn(ctx context.Context, command []string, env []string) error
```

Spawn ensures the worktree exists, starts the PTY in the workdir, and runs the lifecycle to Running (spec §5.4, §5.6).
command is the argv to run in the PTY (e.g. []string{"bash"} or adapter CLI). env is the environment; TERM etc. may be set by caller.

#### LocalSession.Stop

```go
func () Stop(ctx context.Context) error
```

Stop drains the lifecycle to Terminated and closes the PTY (spec §5.6).


## type lifecycleActor

```go
type lifecycleActor struct {
	hsm.HSM
	AgentID    api.ID
	Dispatcher protocol.Dispatcher
}
```

lifecycleActor holds HSM state and dispatch hook for agent lifecycle.

## type presenceActor

```go
type presenceActor struct {
	hsm.HSM
	AgentID    api.ID
	Dispatcher protocol.Dispatcher
}
```

presenceActor holds HSM state and dispatch hook for agent presence.

