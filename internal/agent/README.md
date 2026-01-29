# package agent

`import "github.com/agentflare-ai/amux/internal/agent"`

Package agent provides agent orchestration for amux.

This package implements agent lifecycle management, presence tracking,
and messaging. All operations are agent-agnostic; agent-specific behavior
is delegated to adapters.

The Add flow validates that the agent's repository exists, computes a
unique slug, creates an isolated git worktree, and registers the agent
for lifecycle management.

See spec §5 for agent management requirements.

Control provides the control plane that wires lifecycle HSM transitions
to actual session spawn/stop/kill operations.

The SessionSpawner interface breaks the import cycle between agent and
session packages: agent defines the interface, session/adapter.go adapts
*session.Manager to satisfy it.

See spec §5.4 for lifecycle state machine and §5.6 for shutdown behavior.

Lifecycle provides the HSM-based agent lifecycle state machine.

The lifecycle HSM implements the state transitions defined in spec §5.4:

	┌─────────┐    ┌─────────┐    ┌─────────┐    ┌────────────┐
	│ Pending │───▶│ Starting│───▶│ Running │───▶│ Terminated │
	└─────────┘    └─────────┘    └─────────┘    └────────────┘
	                                  │
	                                  ▼
	                             ┌─────────┐
	                             │ Errored │
	                             └─────────┘

Transitions:
  - "start" event: Pending → Starting
  - "ready" event: Starting → Running
  - "stop" event: Running → Terminated
  - "error" event: Any → Errored

Messaging provides inter-agent messaging for amux.

This package implements the messaging system defined in spec §6.4,
including message routing, ToSlug resolution, and message events.

Messages flow through NATS P.comm.* subjects:
  - P.comm.director: director channel
  - P.comm.manager.<host_id>: host manager channel
  - P.comm.agent.<host_id>.<agent_id>: agent channel
  - P.comm.broadcast: broadcast to all participants

See spec §6.4 for the complete messaging specification.

Package agent - persist.go provides disk persistence for agent definitions.

Per spec, agent definitions must survive daemon restarts. This file implements
a persistence layer that saves agents to ~/.amux/agents.json using the
paths.Resolver for directory resolution.

The persistence format is JSON with api.Agent types. On startup, the manager
loads persisted agents. On Add/Remove, the persistence file is updated.

Presence provides the HSM-based agent presence state machine.

The presence HSM implements the state transitions defined in spec §6.1 and §6.5:

	                    ┌──────────────────┐
	                    ▼                  │
	┌────────┐    ┌─────────┐    ┌────────┐
	│ Online │◀──▶│  Busy   │───▶│ Offline│
	└────────┘    └─────────┘    └────────┘
	     ▲              │              │
	     │              ▼              │
	     │         ┌────────┐          │
	     └─────────│  Away  │◀─────────┘
	               └────────┘

Transitions:
  - task.assigned: Online → Busy
  - task.completed: Busy → Online
  - prompt.detected: Busy → Online
  - rate.limit: * → Offline
  - rate.cleared: Offline → Online
  - stuck.detected: * → Away
  - activity.detected: Away → Online

Roster provides the roster management for amux.

The roster contains all participants in a session: agents, host managers,
and the director. It is updated in real-time as presence changes occur
and broadcast via roster.updated events.

See spec §6.2 for roster requirements and §6.3 for presence awareness.

- `LifecycleEventStart, LifecycleEventReady, LifecycleEventStop, LifecycleEventError` — LifecycleEvent names for lifecycle state transitions.
- `LifecycleModel` — LifecycleModel defines the HSM model for agent lifecycle.
- `PresenceEventTaskAssigned, PresenceEventTaskCompleted, PresenceEventPromptDetected, PresenceEventRateLimit, PresenceEventRateCleared, PresenceEventStuckDetected, PresenceEventActivityDetected` — PresenceEvent names for presence state transitions.
- `PresenceModel` — PresenceModel defines the HSM model for agent presence.
- `func DispatchActivityDetected(ctx context.Context, instance hsm.Instance) <-chan struct{}` — DispatchActivityDetected sends an "activity.detected" event to transition from Away to Online.
- `func DispatchError(ctx context.Context, instance hsm.Instance, err error) <-chan struct{}` — DispatchError sends an "error" event to transition to Errored state.
- `func DispatchPromptDetected(ctx context.Context, instance hsm.Instance) <-chan struct{}` — DispatchPromptDetected sends a "prompt.detected" event to transition from Busy to Online.
- `func DispatchRateCleared(ctx context.Context, instance hsm.Instance) <-chan struct{}` — DispatchRateCleared sends a "rate.cleared" event to transition from Offline to Online.
- `func DispatchRateLimit(ctx context.Context, instance hsm.Instance) <-chan struct{}` — DispatchRateLimit sends a "rate.limit" event to transition to Offline state.
- `func DispatchReady(ctx context.Context, instance hsm.Instance) <-chan struct{}` — DispatchReady sends a "ready" event to transition from Starting to Running.
- `func DispatchStart(ctx context.Context, instance hsm.Instance) <-chan struct{}` — DispatchStart sends a "start" event to transition from Pending to Starting.
- `func DispatchStop(ctx context.Context, instance hsm.Instance) <-chan struct{}` — DispatchStop sends a "stop" event to transition from Running to Terminated.
- `func DispatchStuckDetected(ctx context.Context, instance hsm.Instance) <-chan struct{}` — DispatchStuckDetected sends a "stuck.detected" event to transition to Away state.
- `func DispatchTaskAssigned(ctx context.Context, instance hsm.Instance) <-chan struct{}` — DispatchTaskAssigned sends a "task.assigned" event to transition from Online to Busy.
- `func DispatchTaskCompleted(ctx context.Context, instance hsm.Instance) <-chan struct{}` — DispatchTaskCompleted sends a "task.completed" event to transition from Busy to Online.
- `func FromEnvelope(env *MessageEnvelope) (*api.AgentMessage, error)` — FromEnvelope converts a wire-format envelope to an AgentMessage.
- `func IsBroadcast(msg *api.AgentMessage) bool` — IsBroadcast returns true if the message is a broadcast message.
- `func equalFoldASCII(a, b string) bool` — equalFoldASCII compares two strings case-insensitively (ASCII only).
- `persistFilename` — persistFilename is the name of the agents persistence file.
- `type Agent` — Agent represents a managed agent instance.
- `type LifecycleHSM` — LifecycleHSM wraps an agent with HSM-driven lifecycle management.
- `type Manager` — Manager manages agents, including worktree isolation and lifecycle tracking.
- `type MessageEnvelope` — MessageEnvelope wraps an AgentMessage for wire transmission.
- `type MessageRouter` — MessageRouter handles inter-agent message routing per spec §6.4.
- `type PresenceHSM` — PresenceHSM wraps an agent with HSM-driven presence management.
- `type Roster` — Roster manages all participants in a session.
- `type SessionHandle` — SessionHandle represents a running session.
- `type SessionSpawner` — SessionSpawner is the interface that the agent control plane uses to spawn, stop, and kill sessions.
- `type agentHSMs` — agentHSMs holds the per-agent lifecycle and presence HSMs plus control state.
- `type persister` — persister handles reading and writing agent definitions to disk.

### Constants

#### LifecycleEventStart, LifecycleEventReady, LifecycleEventStop, LifecycleEventError

```go
const (
	LifecycleEventStart = "start" // Pending → Starting
	LifecycleEventReady = "ready" // Starting → Running
	LifecycleEventStop  = "stop"  // Running → Terminated
	LifecycleEventError = "error" // Any → Errored
)
```

LifecycleEvent names for lifecycle state transitions.

#### PresenceEventTaskAssigned, PresenceEventTaskCompleted, PresenceEventPromptDetected, PresenceEventRateLimit, PresenceEventRateCleared, PresenceEventStuckDetected, PresenceEventActivityDetected

```go
const (
	PresenceEventTaskAssigned     = "task.assigned"     // Online → Busy
	PresenceEventTaskCompleted    = "task.completed"    // Busy → Online
	PresenceEventPromptDetected   = "prompt.detected"   // Busy → Online
	PresenceEventRateLimit        = "rate.limit"        // * → Offline
	PresenceEventRateCleared      = "rate.cleared"      // Offline → Online
	PresenceEventStuckDetected    = "stuck.detected"    // * → Away
	PresenceEventActivityDetected = "activity.detected" // Away → Online
)
```

PresenceEvent names for presence state transitions.

#### persistFilename

```go
const persistFilename = "agents.json"
```

persistFilename is the name of the agents persistence file.


### Variables

#### LifecycleModel

```go
var LifecycleModel = hsm.Define(
	"agent.lifecycle",

	hsm.State("pending"),
	hsm.State("starting",
		hsm.Entry(func(ctx context.Context, l *LifecycleHSM, e hsm.Event) {
			l.onEnterStarting(ctx)
		}),
	),
	hsm.State("running",
		hsm.Entry(func(ctx context.Context, l *LifecycleHSM, e hsm.Event) {
			l.onEnterRunning(ctx)
		}),
		hsm.Exit(func(ctx context.Context, l *LifecycleHSM, e hsm.Event) {
			l.onExitRunning(ctx)
		}),
	),
	hsm.State("terminated",
		hsm.Entry(func(ctx context.Context, l *LifecycleHSM, e hsm.Event) {
			l.onEnterTerminated(ctx)
		}),
	),
	hsm.State("errored",
		hsm.Entry(func(ctx context.Context, l *LifecycleHSM, e hsm.Event) {
			l.onEnterErrored(ctx, e)
		}),
	),

	hsm.Transition(
		hsm.On(hsm.Event{Name: LifecycleEventStart}),
		hsm.Source("pending"),
		hsm.Target("starting"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: LifecycleEventReady}),
		hsm.Source("starting"),
		hsm.Target("running"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: LifecycleEventStop}),
		hsm.Source("running"),
		hsm.Target("terminated"),
	),

	hsm.Transition(
		hsm.On(hsm.Event{Name: LifecycleEventError}),
		hsm.Source("pending"),
		hsm.Target("errored"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: LifecycleEventError}),
		hsm.Source("starting"),
		hsm.Target("errored"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: LifecycleEventError}),
		hsm.Source("running"),
		hsm.Target("errored"),
	),

	hsm.Initial(
		hsm.Target("pending"),
	),
)
```

LifecycleModel defines the HSM model for agent lifecycle.
See spec §5.4.

#### PresenceModel

```go
var PresenceModel = hsm.Define(
	"agent.presence",

	hsm.State("online",
		hsm.Entry(func(ctx context.Context, p *PresenceHSM, e hsm.Event) {
			p.onEnterOnline(ctx, e)
		}),
	),
	hsm.State("busy",
		hsm.Entry(func(ctx context.Context, p *PresenceHSM, e hsm.Event) {
			p.onEnterBusy(ctx, e)
		}),
	),
	hsm.State("offline",
		hsm.Entry(func(ctx context.Context, p *PresenceHSM, e hsm.Event) {
			p.onEnterOffline(ctx, e)
		}),
	),
	hsm.State("away",
		hsm.Entry(func(ctx context.Context, p *PresenceHSM, e hsm.Event) {
			p.onEnterAway(ctx, e)
		}),
	),

	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventTaskAssigned}),
		hsm.Source("online"),
		hsm.Target("busy"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventTaskCompleted}),
		hsm.Source("busy"),
		hsm.Target("online"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventPromptDetected}),
		hsm.Source("busy"),
		hsm.Target("online"),
	),

	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventRateLimit}),
		hsm.Source("online"),
		hsm.Target("offline"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventRateLimit}),
		hsm.Source("busy"),
		hsm.Target("offline"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventRateCleared}),
		hsm.Source("offline"),
		hsm.Target("online"),
	),

	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventStuckDetected}),
		hsm.Source("online"),
		hsm.Target("away"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventStuckDetected}),
		hsm.Source("busy"),
		hsm.Target("away"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventStuckDetected}),
		hsm.Source("offline"),
		hsm.Target("away"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventActivityDetected}),
		hsm.Source("away"),
		hsm.Target("online"),
	),

	hsm.Initial(
		hsm.Target("online"),
	),
)
```

PresenceModel defines the HSM model for agent presence.
See spec §6.1 and §6.5.


### Functions

#### DispatchActivityDetected

```go
func DispatchActivityDetected(ctx context.Context, instance hsm.Instance) <-chan struct{}
```

DispatchActivityDetected sends an "activity.detected" event to transition from Away to Online.

#### DispatchError

```go
func DispatchError(ctx context.Context, instance hsm.Instance, err error) <-chan struct{}
```

DispatchError sends an "error" event to transition to Errored state.
The error parameter is stored and can be retrieved via LastError().

#### DispatchPromptDetected

```go
func DispatchPromptDetected(ctx context.Context, instance hsm.Instance) <-chan struct{}
```

DispatchPromptDetected sends a "prompt.detected" event to transition from Busy to Online.

#### DispatchRateCleared

```go
func DispatchRateCleared(ctx context.Context, instance hsm.Instance) <-chan struct{}
```

DispatchRateCleared sends a "rate.cleared" event to transition from Offline to Online.

#### DispatchRateLimit

```go
func DispatchRateLimit(ctx context.Context, instance hsm.Instance) <-chan struct{}
```

DispatchRateLimit sends a "rate.limit" event to transition to Offline state.

#### DispatchReady

```go
func DispatchReady(ctx context.Context, instance hsm.Instance) <-chan struct{}
```

DispatchReady sends a "ready" event to transition from Starting to Running.

#### DispatchStart

```go
func DispatchStart(ctx context.Context, instance hsm.Instance) <-chan struct{}
```

DispatchStart sends a "start" event to transition from Pending to Starting.

#### DispatchStop

```go
func DispatchStop(ctx context.Context, instance hsm.Instance) <-chan struct{}
```

DispatchStop sends a "stop" event to transition from Running to Terminated.

#### DispatchStuckDetected

```go
func DispatchStuckDetected(ctx context.Context, instance hsm.Instance) <-chan struct{}
```

DispatchStuckDetected sends a "stuck.detected" event to transition to Away state.

#### DispatchTaskAssigned

```go
func DispatchTaskAssigned(ctx context.Context, instance hsm.Instance) <-chan struct{}
```

DispatchTaskAssigned sends a "task.assigned" event to transition from Online to Busy.

#### DispatchTaskCompleted

```go
func DispatchTaskCompleted(ctx context.Context, instance hsm.Instance) <-chan struct{}
```

DispatchTaskCompleted sends a "task.completed" event to transition from Busy to Online.

#### FromEnvelope

```go
func FromEnvelope(env *MessageEnvelope) (*api.AgentMessage, error)
```

FromEnvelope converts a wire-format envelope to an AgentMessage.

#### IsBroadcast

```go
func IsBroadcast(msg *api.AgentMessage) bool
```

IsBroadcast returns true if the message is a broadcast message.

#### equalFoldASCII

```go
func equalFoldASCII(a, b string) bool
```

equalFoldASCII compares two strings case-insensitively (ASCII only).
This is faster than strings.EqualFold for ASCII-only slugs.


## type Agent

```go
type Agent struct {
	mu sync.RWMutex
	api.Agent

	// Lifecycle state
	lifecycle api.LifecycleState

	// Presence state
	presence api.PresenceState
}
```

Agent represents a managed agent instance.

### Methods

#### Agent.Lifecycle

```go
func () Lifecycle() api.LifecycleState
```

Lifecycle returns the agent's lifecycle state.

#### Agent.Presence

```go
func () Presence() api.PresenceState
```

Presence returns the agent's presence state.

#### Agent.SetLifecycle

```go
func () SetLifecycle(state api.LifecycleState)
```

SetLifecycle sets the agent's lifecycle state.

#### Agent.SetPresence

```go
func () SetPresence(state api.PresenceState)
```

SetPresence sets the agent's presence state.


## type LifecycleHSM

```go
type LifecycleHSM struct {
	hsm.HSM

	mu             sync.RWMutex
	agent          *Agent
	lifecycleState api.LifecycleState
	dispatcher     event.Dispatcher
	lastError      error
}
```

LifecycleHSM wraps an agent with HSM-driven lifecycle management.

### Functions returning LifecycleHSM

#### NewLifecycleHSM

```go
func NewLifecycleHSM(agent *Agent, dispatcher event.Dispatcher) *LifecycleHSM
```

NewLifecycleHSM creates a new lifecycle HSM for an agent.


### Methods

#### LifecycleHSM.Agent

```go
func () Agent() *Agent
```

Agent returns the associated agent.

#### LifecycleHSM.LastError

```go
func () LastError() error
```

LastError returns the last error that caused the errored state.

#### LifecycleHSM.LifecycleState

```go
func () LifecycleState() api.LifecycleState
```

LifecycleState returns the current lifecycle state.

#### LifecycleHSM.Start

```go
func () Start(ctx context.Context) *LifecycleHSM
```

Start initializes and starts the lifecycle HSM.
Returns the started HSM instance.

#### LifecycleHSM.onEnterErrored

```go
func () onEnterErrored(ctx context.Context, e hsm.Event)
```

Entry action for Errored state

#### LifecycleHSM.onEnterRunning

```go
func () onEnterRunning(ctx context.Context)
```

Entry action for Running state

#### LifecycleHSM.onEnterStarting

```go
func () onEnterStarting(ctx context.Context)
```

Entry action for Starting state

#### LifecycleHSM.onEnterTerminated

```go
func () onEnterTerminated(ctx context.Context)
```

Entry action for Terminated state

#### LifecycleHSM.onExitRunning

```go
func () onExitRunning(ctx context.Context)
```

Exit action for Running state

#### LifecycleHSM.setLifecycleState

```go
func () setLifecycleState(state api.LifecycleState)
```

setLifecycleState updates the internal state and synchronizes with the agent.


## type Manager

```go
type Manager struct {
	mu         sync.RWMutex
	agents     map[muid.MUID]*Agent
	slugs      map[string]muid.MUID // slug -> agent ID for collision detection
	hsms       map[muid.MUID]*agentHSMs
	sessions   SessionSpawner
	dispatcher event.Dispatcher
	resolver   *paths.Resolver
	worktrees  *worktree.Manager
	persist    *persister

	// baseBranches tracks the base_branch per repo_root, recorded at the time
	// the first agent for that repository is added (spec §5.7.1).
	baseBranches map[string]string

	// mergeTargetBranch is the configured git.merge.target_branch fallback.
	// Per spec §5.7.1, when git symbolic-ref fails (detached HEAD), base_branch
	// MUST be set to this value. If this is also empty, the add operation MUST fail.
	mergeTargetBranch string

	// monitorUnsub is the unsubscribe function for the monitor event subscription.
	monitorUnsub func()

	// connectionUnsub is the unsubscribe function for the connection event subscription.
	connectionUnsub func()
}
```

Manager manages agents, including worktree isolation and lifecycle tracking.

### Functions returning Manager

#### NewManager

```go
func NewManager(dispatcher event.Dispatcher) *Manager
```

NewManager creates a new agent manager.

#### NewManagerWithResolver

```go
func NewManagerWithResolver(dispatcher event.Dispatcher, resolver *paths.Resolver) *Manager
```

NewManagerWithResolver creates a new agent manager with a specific path resolver.


### Methods

#### Manager.Add

```go
func () Add(ctx context.Context, cfg api.Agent) (*Agent, error)
```

Add adds a new agent. The agent's Slug is computed from Name, a worktree is
created for isolation, and the agent is registered for lifecycle management.

For local agents, if RepoRoot is empty, it is resolved from the Location.RepoPath
or from the current working directory. The repo_root must be a valid git repository.

See spec §5.2 for the agent add flow.

#### Manager.BaseBranch

```go
func () BaseBranch(repoRoot string) (string, bool)
```

BaseBranch returns the recorded base branch for a repository.

#### Manager.Get

```go
func () Get(id muid.MUID) *Agent
```

Get returns an agent by ID.

#### Manager.GetBySlug

```go
func () GetBySlug(slug string) *Agent
```

GetBySlug returns an agent by its slug.

#### Manager.Kill

```go
func () Kill(ctx context.Context, agentID muid.MUID) error
```

Kill forcefully terminates an agent's session.

Like Stop, the stopping flag is set so watchSession transitions to
Terminated rather than Errored.

#### Manager.LifecycleHSMFor

```go
func () LifecycleHSMFor(agentID muid.MUID) *LifecycleHSM
```

LifecycleHSMFor returns the lifecycle HSM for an agent, or nil if not found.

#### Manager.List

```go
func () List() []*Agent
```

List returns all agents.

#### Manager.LoadPersisted

```go
func () LoadPersisted(ctx context.Context) error
```

LoadPersisted loads agent definitions from disk and registers them.
This should be called on daemon startup to restore agents that survived a restart.
Agents are loaded in their persisted state (lifecycle=pending, presence=online).

#### Manager.PresenceHSMFor

```go
func () PresenceHSMFor(agentID muid.MUID) *PresenceHSM
```

PresenceHSMFor returns the presence HSM for an agent, or nil if not found.

#### Manager.Remove

```go
func () Remove(ctx context.Context, id muid.MUID, deleteBranch bool) error
```

Remove removes an agent, cleaning up its worktree if configured.
The deleteBranch parameter controls whether the agent's git branch is deleted.
If the agent has a running session, it is stopped first.

#### Manager.Roster

```go
func () Roster() []api.RosterEntry
```

Roster returns the roster entries for all agents.

#### Manager.SetMergeTargetBranch

```go
func () SetMergeTargetBranch(branch string)
```

SetMergeTargetBranch sets the configured git.merge.target_branch fallback.
Per spec §5.7.1, this value is used as base_branch when the repository is
in detached HEAD state (git symbolic-ref fails).

#### Manager.SetSessionSpawner

```go
func () SetSessionSpawner(s SessionSpawner)
```

SetSessionSpawner sets the session spawner used by control plane methods.

#### Manager.SlugExists

```go
func () SlugExists(slug string) bool
```

SlugExists returns true if an agent with the given slug exists.

#### Manager.Start

```go
func () Start(ctx context.Context, agentID muid.MUID, shell string, args ...string) error
```

Start transitions an agent from Pending to Running by spawning a session.

The lifecycle HSM is driven through: Pending → Starting → Running.
If the spawn fails, the lifecycle transitions to Errored.
A watchSession goroutine is launched to monitor the session.

#### Manager.Stop

```go
func () Stop(ctx context.Context, agentID muid.MUID) error
```

Stop gracefully stops an agent's session and waits for it to exit.

The stopping flag is set so watchSession knows this was intentional
and transitions to Terminated (not Errored).

#### Manager.handleConnectionEvent

```go
func () handleConnectionEvent(ctx context.Context, evt event.Event)
```

handleConnectionEvent maps connection events to presence HSM transitions
for remote agents per spec §5.5.8 and §6.5.

#### Manager.handleMonitorEvent

```go
func () handleMonitorEvent(ctx context.Context, evt event.Event)
```

handleMonitorEvent maps PTY monitor events to presence HSM transitions.
Per spec §7.6:
  - TypePTYActivity  -> task.assigned   (Online -> Busy)
  - TypePTYIdle      -> prompt.detected (Busy -> Online)
  - TypePTYStuck     -> stuck.detected  (* -> Away)

#### Manager.persistAgents

```go
func () persistAgents()
```

persistAgents saves the current agent definitions to disk.
Must be called with m.mu held (at least RLock).

#### Manager.resolveLocalRepoRoot

```go
func () resolveLocalRepoRoot(loc api.Location) (string, error)
```

resolveLocalRepoRoot resolves the repo_root for a local agent.
If location.RepoPath is set, it is validated; otherwise the current
working directory's repo root is used.

#### Manager.setupConnectionSubscription

```go
func () setupConnectionSubscription()
```

setupConnectionSubscription subscribes to connection events and dispatches
the appropriate presence HSM transitions per spec §5.5.8 and §6.5:
  - connection.lost      -> stuck.detected  (* -> Away)
  - connection.recovered -> activity.detected (Away -> Online)

Remote agents transition to Away when hub connection is lost and return
to Online when the connection is recovered and replay is complete.

#### Manager.setupMonitorSubscription

```go
func () setupMonitorSubscription()
```

setupMonitorSubscription subscribes to PTY monitor events and dispatches
the appropriate presence HSM transitions per spec §7.6:
  - pty.activity  -> ActivityDetected -> Busy
  - pty.idle      -> PromptDetected   -> Online
  - pty.stuck     -> StuckDetected    -> Away

#### Manager.watchSession

```go
func () watchSession(agentID muid.MUID, handle SessionHandle)
```

watchSession monitors a session and drives the lifecycle HSM when it exits.

Uses context.Background() because this goroutine outlives the Start call
that launched it. The caller's context may be canceled independently.

If stopping is true (intentional Stop/Kill), lifecycle → Terminated.
If stopping is false (unexpected crash), lifecycle → Errored.


## type MessageEnvelope

```go
type MessageEnvelope struct {
	// ID is the message ID (base-10 string per spec §9.1.3.1).
	ID string `json:"id"`

	// From is the sender runtime ID (base-10 string).
	From string `json:"from"`

	// To is the recipient runtime ID (base-10 string).
	// "0" indicates broadcast.
	To string `json:"to"`

	// ToSlug is the original recipient token from the message.
	ToSlug string `json:"to_slug"`

	// Content is the message body.
	Content string `json:"content"`

	// Timestamp is the message timestamp (RFC 3339 UTC).
	Timestamp string `json:"timestamp"`
}
```

MessageEnvelope wraps an AgentMessage for wire transmission.
This is the JSON format used on NATS P.comm.* subjects.

See spec §5.5.7.1 and §9.1.3.1 for wire format requirements.

### Functions returning MessageEnvelope

#### ToEnvelope

```go
func ToEnvelope(msg *api.AgentMessage) *MessageEnvelope
```

ToEnvelope converts an AgentMessage to a wire-format envelope.


## type MessageRouter

```go
type MessageRouter struct {
	roster     *Roster
	dispatcher event.Dispatcher
	localID    muid.MUID // ID of local participant (manager or director)
	hostID     string    // host_id for this router (empty for director)
}
```

MessageRouter handles inter-agent message routing per spec §6.4.

The router resolves ToSlug to runtime IDs, enriches messages with
sender information, and dispatches to the appropriate channels.

### Functions returning MessageRouter

#### NewMessageRouter

```go
func NewMessageRouter(roster *Roster, dispatcher event.Dispatcher, localID muid.MUID, hostID string) *MessageRouter
```

NewMessageRouter creates a new message router.
localID is the runtime ID of the local participant (manager or director).
hostID is the host identifier (empty string for director).


### Methods

#### MessageRouter.BroadcastMessage

```go
func () BroadcastMessage(ctx context.Context, senderID muid.MUID, content string) (*api.AgentMessage, error)
```

BroadcastMessage broadcasts a message to all participants.
This is typically used by the director.

Dispatches message.broadcast event with the message data.

See spec §6.4.1.

#### MessageRouter.DeliverMessage

```go
func () DeliverMessage(ctx context.Context, msg *api.AgentMessage) error
```

DeliverMessage delivers an inbound message to a recipient.
This is called by the host manager when a message arrives for a local participant.

Dispatches message.inbound event with the message data.

See spec §6.4.1.

#### MessageRouter.ResolveToSlug

```go
func () ResolveToSlug(toSlug string) (muid.MUID, error)
```

ResolveToSlug resolves a ToSlug string to a recipient runtime ID.
Returns (recipient ID, error). BroadcastID (0) is returned for broadcast targets.

Resolution rules per spec §6.4.1.3:
  - "all", "broadcast", "*" -> BroadcastID
  - "director" -> director runtime ID
  - "manager" -> local host manager runtime ID
  - "manager@<host_id>" -> specific host manager runtime ID
  - Otherwise -> agent_slug lookup

#### MessageRouter.RouteMessage

```go
func () RouteMessage(ctx context.Context, senderID muid.MUID, toSlug, content string) (*api.AgentMessage, error)
```

RouteMessage routes an outbound message from an agent.
This is called by the host manager when it detects an outbound message
from an agent's PTY output (via adapter pattern matching).

The router:
 1. Sets From to the sender runtime ID
 2. Generates a unique message ID
 3. Sets Timestamp to current time (UTC)
 4. Resolves ToSlug to a recipient runtime ID
 5. Dispatches message.outbound event

Returns the enriched message or an error if resolution fails.

See spec §6.4.1.


## type PresenceHSM

```go
type PresenceHSM struct {
	hsm.HSM

	mu            sync.RWMutex
	agent         *Agent
	presenceState api.PresenceState
	dispatcher    event.Dispatcher
}
```

PresenceHSM wraps an agent with HSM-driven presence management.

### Functions returning PresenceHSM

#### NewPresenceHSM

```go
func NewPresenceHSM(agent *Agent, dispatcher event.Dispatcher) *PresenceHSM
```

NewPresenceHSM creates a new presence HSM for an agent.


### Methods

#### PresenceHSM.Agent

```go
func () Agent() *Agent
```

Agent returns the associated agent.

#### PresenceHSM.PresenceState

```go
func () PresenceState() api.PresenceState
```

PresenceState returns the current presence state.

#### PresenceHSM.Start

```go
func () Start(ctx context.Context) *PresenceHSM
```

Start initializes and starts the presence HSM.
Returns the started HSM instance.

#### PresenceHSM.onEnterAway

```go
func () onEnterAway(ctx context.Context, e hsm.Event)
```

Entry action for Away state

#### PresenceHSM.onEnterBusy

```go
func () onEnterBusy(ctx context.Context, e hsm.Event)
```

Entry action for Busy state

#### PresenceHSM.onEnterOffline

```go
func () onEnterOffline(ctx context.Context, e hsm.Event)
```

Entry action for Offline state

#### PresenceHSM.onEnterOnline

```go
func () onEnterOnline(ctx context.Context, e hsm.Event)
```

Entry action for Online state

#### PresenceHSM.setPresenceState

```go
func () setPresenceState(state api.PresenceState)
```

setPresenceState updates the internal state and synchronizes with the agent.


## type Roster

```go
type Roster struct {
	mu           sync.RWMutex
	participants map[muid.MUID]*api.Participant
	slugIndex    map[string]muid.MUID // slug -> participant ID for lookups
	dispatcher   event.Dispatcher
	directorID   muid.MUID
}
```

Roster manages all participants in a session.
The roster includes agents, host managers (manager agents), and the director.

See spec §6.2.

### Functions returning Roster

#### NewRoster

```go
func NewRoster(dispatcher event.Dispatcher) *Roster
```

NewRoster creates a new roster with the given event dispatcher.


### Methods

#### Roster.AddAgent

```go
func () AddAgent(agent *api.Agent, lifecycle api.LifecycleState, presence api.PresenceState)
```

AddAgent registers an agent in the roster.

#### Roster.AddManager

```go
func () AddManager(id muid.MUID, name, hostID, about string)
```

AddManager registers a host manager in the roster.
The slug is "manager@<hostID>" for remote managers, or "manager" for local.

#### Roster.DirectorID

```go
func () DirectorID() muid.MUID
```

DirectorID returns the director's runtime ID, or 0 if no director is registered.

#### Roster.Get

```go
func () Get(id muid.MUID) *api.Participant
```

Get returns a participant by ID.

#### Roster.GetBySlug

```go
func () GetBySlug(slug string) *api.Participant
```

GetBySlug returns a participant by slug.
Slug lookup is case-insensitive per spec §6.4.1.3.

#### Roster.List

```go
func () List() []api.Participant
```

List returns all participants in the roster.

#### Roster.ListAgents

```go
func () ListAgents() []api.Participant
```

ListAgents returns all agent participants.

#### Roster.ListManagers

```go
func () ListManagers() []api.Participant
```

ListManagers returns all manager participants.

#### Roster.RemoveParticipant

```go
func () RemoveParticipant(id muid.MUID)
```

RemoveParticipant removes a participant from the roster.

#### Roster.SetDirector

```go
func () SetDirector(id muid.MUID, name, about string)
```

SetDirector registers the director in the roster.
There can only be one director; calling this replaces any existing director.

#### Roster.UpdateCurrentTask

```go
func () UpdateCurrentTask(id muid.MUID, task string)
```

UpdateCurrentTask updates the current task of a busy participant.
Per spec §6.3, this enables other agents to know what busy agents are working on.

#### Roster.UpdateLifecycle

```go
func () UpdateLifecycle(id muid.MUID, lifecycle api.LifecycleState)
```

UpdateLifecycle updates the lifecycle state of an agent participant.

#### Roster.UpdatePresence

```go
func () UpdatePresence(id muid.MUID, presence api.PresenceState)
```

UpdatePresence updates the presence state of a participant.

#### Roster.emitRosterUpdated

```go
func () emitRosterUpdated()
```

emitRosterUpdated dispatches a roster.updated event.
Must be called with r.mu held.


## type SessionHandle

```go
type SessionHandle interface {
	// Done returns a channel that is closed when the session exits.
	Done() <-chan struct{}

	// ExitErr returns the process exit error, or nil if exited cleanly.
	ExitErr() error
}
```

SessionHandle represents a running session. The agent package uses this
to monitor session lifetime without importing the session package.

## type SessionSpawner

```go
type SessionSpawner interface {
	// SpawnAgent creates and starts a new PTY session for an agent.
	SpawnAgent(ctx context.Context, ag *Agent, shell string, args ...string) (SessionHandle, error)

	// StopAgent gracefully stops the session for an agent.
	StopAgent(ctx context.Context, agentID muid.MUID) error

	// KillAgent forcefully terminates the session for an agent.
	KillAgent(ctx context.Context, agentID muid.MUID) error

	// RemoveSession removes a session from the session manager.
	RemoveSession(agentID muid.MUID)
}
```

SessionSpawner is the interface that the agent control plane uses to
spawn, stop, and kill sessions. It is satisfied by session.Adapter.

## type agentHSMs

```go
type agentHSMs struct {
	lifecycle *LifecycleHSM
	presence  *PresenceHSM

	// lifecycleInstance is the started HSM instance for lifecycle.
	lifecycleInstance hsm.Instance

	// presenceInstance is the started HSM instance for presence.
	presenceInstance hsm.Instance

	// stopping is set to true when Stop or Kill is called intentionally.
	// watchSession uses this to distinguish intentional shutdown from crash.
	stopping bool

	// mu protects the stopping flag.
	mu sync.Mutex
}
```

agentHSMs holds the per-agent lifecycle and presence HSMs plus control state.

### Methods

#### agentHSMs.isStopping

```go
func () isStopping() bool
```

isStopping atomically reads the stopping flag.

#### agentHSMs.setStopping

```go
func () setStopping(v bool)
```

setStopping atomically sets the stopping flag.


## type persister

```go
type persister struct {
	mu       sync.Mutex
	resolver *paths.Resolver
}
```

persister handles reading and writing agent definitions to disk.

### Functions returning persister

#### newPersister

```go
func newPersister(resolver *paths.Resolver) *persister
```

newPersister creates a new persister with the given resolver.


### Methods

#### persister.filePath

```go
func () filePath() string
```

filePath returns the full path to the persistence file.

#### persister.load

```go
func () load() ([]api.Agent, error)
```

load reads persisted agent definitions from disk.
Returns an empty slice (not an error) if the file does not exist.

#### persister.save

```go
func () save(agents []api.Agent) error
```

save writes agent definitions to disk atomically.
It creates the data directory if it does not exist.


