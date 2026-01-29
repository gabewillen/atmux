# package agent

`import "github.com/stateforward/amux/internal/agent"`

Package agent provides the agent actor model and HSM-driven lifecycle and presence state machines.

Package agent provides the agent actor model and local management helpers.

This file implements Phase 2 local agent management for:
- Agent add flow (validation, repo required, config persistence) per spec §5.2
- Worktree isolation and slug-based path layout per spec §5.3.1
- Git merge strategy selection scaffolding per spec §5.7

Package agent provides the agent actor model, presence state machines, and
roster management helpers.

- `EventStart, EventReady, EventStop, EventError` — Lifecycle event constants.
- `EventTaskAssigned, EventTaskCompleted, EventPromptDetected, EventRateLimit, EventRateCleared, EventStuckDetected, EventActivityDetected` — Presence event constants per spec §6.5.
- `EventTypeMessageOutbound, EventTypeMessageInbound, EventTypeMessageBroadcast` — Message event type constants per spec §6.4 and §9.3.
- `EventTypePresenceChanged, EventTypeRosterUpdated` — Event type constants for presence and roster updates.
- `LifecycleModel` — LifecycleModel defines the agent lifecycle state machine per spec §5.4.
- `PresenceModel` — PresenceModel defines the agent presence state machine per spec §6.5.
- `StateOnline, StateBusy, StateOffline, StateAway` — Presence state constants matching spec §6.1.
- `StatePending, StateStarting, StateRunning, StateTerminated, StateErrored` — Lifecycle state constants matching spec §5.4.
- `func AddLocalAgent(ctx context.Context, cfg *config.Config, opts AddLocalAgentOptions) (*api.Agent, string, error)` — AddLocalAgent adds a new local agent for the given repository root.
- `func EmitBroadcastMessage(ctx context.Context, dispatcher event.Dispatcher, msg api.AgentMessage) error` — EmitBroadcastMessage emits a message.broadcast event carrying the given message.
- `func EmitInboundMessage(ctx context.Context, dispatcher event.Dispatcher, msg api.AgentMessage) error` — EmitInboundMessage emits a message.inbound event carrying the given message.
- `func EmitOutboundMessage(ctx context.Context, dispatcher event.Dispatcher, msg api.AgentMessage) error` — EmitOutboundMessage emits a message.outbound event carrying the given message.
- `func deriveUniqueAgentSlug(cfg *config.Config, name string) string` — deriveUniqueAgentSlug derives a unique agent_slug given the existing config and desired agent name.
- `func ensureLocalWorktree(ctx context.Context, repoRoot, slug, worktreePath string) error` — ensureLocalWorktree ensures a git worktree exists for the given agent slug.
- `func flattenEnv(m map[string]string) []string` — flattenEnv flattens a map[string]string into KEY=VALUE strings.
- `func slugExists(cfg *config.Config, slug string) bool` — slugExists checks whether a slug is already in use by any configured agent.
- `func verifyGitRepository(ctx context.Context, repoRoot string) error` — verifyGitRepository ensures the given path is a git repository.
- `func worktreeListed(output, worktreePath string) bool` — worktreeListed checks whether a worktreePath appears in the `git worktree list --porcelain` output.
- `type AddLocalAgentOptions` — AddLocalAgentOptions holds input parameters for adding a local agent.
- `type AgentActor` — AgentActor wraps an Agent with HSM-driven lifecycle and presence state machines.
- `type LocalSession` — LocalSession represents a local PTY-backed agent session.
- `type MergeStrategy` — AddLocalAgentOptions holds input parameters for adding a local agent.
- `type MessagePayload` — MessagePayload wraps an AgentMessage for event dispatch.
- `type PresenceActor` — PresenceActor wraps an Agent with a presence state machine.
- `type PresenceChangedPayload` — PresenceChangedPayload is the payload for presence.changed events.
- `type RosterStore` — RosterStore maintains the current roster of agents and their presence.

### Constants

#### StatePending, StateStarting, StateRunning, StateTerminated, StateErrored

```go
const (
	StatePending    = "pending"
	StateStarting   = "starting"
	StateRunning    = "running"
	StateTerminated = "terminated"
	StateErrored    = "errored"
)
```

Lifecycle state constants matching spec §5.4.

#### EventStart, EventReady, EventStop, EventError

```go
const (
	EventStart = "start"
	EventReady = "ready"
	EventStop  = "stop"
	EventError = "error"
)
```

Lifecycle event constants.

#### EventTypeMessageOutbound, EventTypeMessageInbound, EventTypeMessageBroadcast

```go
const (
	EventTypeMessageOutbound  = "message.outbound"
	EventTypeMessageInbound   = "message.inbound"
	EventTypeMessageBroadcast = "message.broadcast"
)
```

Message event type constants per spec §6.4 and §9.3.

#### StateOnline, StateBusy, StateOffline, StateAway

```go
const (
	StateOnline  = "online"
	StateBusy    = "busy"
	StateOffline = "offline"
	StateAway    = "away"
)
```

Presence state constants matching spec §6.1.

#### EventTaskAssigned, EventTaskCompleted, EventPromptDetected, EventRateLimit, EventRateCleared, EventStuckDetected, EventActivityDetected

```go
const (
	EventTaskAssigned     = "task.assigned"
	EventTaskCompleted    = "task.completed"
	EventPromptDetected   = "prompt.detected"
	EventRateLimit        = "rate.limit"
	EventRateCleared      = "rate.cleared"
	EventStuckDetected    = "stuck.detected"
	EventActivityDetected = "activity.detected"
)
```

Presence event constants per spec §6.5.

#### EventTypePresenceChanged, EventTypeRosterUpdated

```go
const (
	EventTypePresenceChanged = "presence.changed"
	EventTypeRosterUpdated   = "roster.updated"
)
```

Event type constants for presence and roster updates.


### Variables

#### LifecycleModel

```go
var LifecycleModel = hsm.Define("agent.lifecycle",
	hsm.State(StatePending),

	hsm.State(StateStarting,
		hsm.Entry(func(ctx context.Context, a *AgentActor, e hsm.Event) {

		}),
	),

	hsm.State(StateRunning,
		hsm.Entry(func(ctx context.Context, a *AgentActor, e hsm.Event) {

		}),
		hsm.Exit(func(ctx context.Context, a *AgentActor, e hsm.Event) {

		}),
	),

	hsm.State(StateTerminated, hsm.Final(StateTerminated)),
	hsm.State(StateErrored, hsm.Final(StateErrored)),

	hsm.Transition(hsm.On(hsm.Event{Name: EventStart}), hsm.Source(StatePending), hsm.Target(StateStarting)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventReady}), hsm.Source(StateStarting), hsm.Target(StateRunning)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventStop}), hsm.Source(StateRunning), hsm.Target(StateTerminated)),

	hsm.Transition(hsm.On(hsm.Event{Name: EventError}), hsm.Source(StatePending), hsm.Target(StateErrored)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventError}), hsm.Source(StateStarting), hsm.Target(StateErrored)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventError}), hsm.Source(StateRunning), hsm.Target(StateErrored)),

	hsm.Initial(hsm.Target(StatePending)),
)
```

LifecycleModel defines the agent lifecycle state machine per spec §5.4.

State diagram:

	┌─────────┐    ┌─────────┐    ┌─────────┐    ┌────────────┐
	│ Pending │───▶│ Starting│───▶│ Running │───▶│ Terminated │
	└─────────┘    └─────────┘    └─────────┘    └────────────┘
	                                   │
	                                   ▼
	                              ┌─────────┐
	                              │ Errored │
	                              └─────────┘

#### PresenceModel

```go
var PresenceModel = hsm.Define("agent.presence",
	hsm.State(StateOnline),
	hsm.State(StateBusy),
	hsm.State(StateOffline),
	hsm.State(StateAway),

	hsm.Transition(hsm.On(hsm.Event{Name: EventTaskAssigned}), hsm.Source(StateOnline), hsm.Target(StateBusy)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventTaskCompleted}), hsm.Source(StateBusy), hsm.Target(StateOnline)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventPromptDetected}), hsm.Source(StateBusy), hsm.Target(StateOnline)),

	hsm.Transition(hsm.On(hsm.Event{Name: EventRateLimit}), hsm.Source(StateBusy), hsm.Target(StateOffline)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventRateLimit}), hsm.Source(StateOnline), hsm.Target(StateOffline)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventRateCleared}), hsm.Source(StateOffline), hsm.Target(StateOnline)),

	hsm.Transition(hsm.On(hsm.Event{Name: EventStuckDetected}), hsm.Source(StateOnline), hsm.Target(StateAway)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventStuckDetected}), hsm.Source(StateBusy), hsm.Target(StateAway)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventStuckDetected}), hsm.Source(StateOffline), hsm.Target(StateAway)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventActivityDetected}), hsm.Source(StateAway), hsm.Target(StateOnline)),

	hsm.Initial(hsm.Target(StateOnline)),
)
```

PresenceModel defines the agent presence state machine per spec §6.5.

State diagram:

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


### Functions

#### AddLocalAgent

```go
func AddLocalAgent(ctx context.Context, cfg *config.Config, opts AddLocalAgentOptions) (*api.Agent, string, error)
```

AddLocalAgent adds a new local agent for the given repository root.

Responsibilities (Phase 2, spec §5.2, §5.3.1, §5.7.1):
- Validate input and ensure the repoRoot is a git repository
- Derive a unique agent_slug from Name using NormalizeAgentSlug
- Ensure a git worktree exists at .amux/worktrees/{agent_slug}/ under repoRoot
- Append the agent to the provided configuration (Agents slice)
- Return the instantiated api.Agent with canonical RepoRoot and Worktree

#### EmitBroadcastMessage

```go
func EmitBroadcastMessage(ctx context.Context, dispatcher event.Dispatcher, msg api.AgentMessage) error
```

EmitBroadcastMessage emits a message.broadcast event carrying the given message.

#### EmitInboundMessage

```go
func EmitInboundMessage(ctx context.Context, dispatcher event.Dispatcher, msg api.AgentMessage) error
```

EmitInboundMessage emits a message.inbound event carrying the given message.

#### EmitOutboundMessage

```go
func EmitOutboundMessage(ctx context.Context, dispatcher event.Dispatcher, msg api.AgentMessage) error
```

EmitOutboundMessage emits a message.outbound event carrying the given message.
Phase 4 uses the local event dispatcher; Phase 7 will route these over hsmnet
and NATS per spec §5.5.7.1 and §6.4.

#### deriveUniqueAgentSlug

```go
func deriveUniqueAgentSlug(cfg *config.Config, name string) string
```

deriveUniqueAgentSlug derives a unique agent_slug given the existing config
and desired agent name. Per spec §5.3.1, collisions are resolved by
appending a numeric suffix -2, -3, ... until unique.

#### ensureLocalWorktree

```go
func ensureLocalWorktree(ctx context.Context, repoRoot, slug, worktreePath string) error
```

ensureLocalWorktree ensures a git worktree exists for the given agent slug.
If the worktree already exists, the function is idempotent and returns nil.

#### flattenEnv

```go
func flattenEnv(m map[string]string) []string
```

flattenEnv flattens a map[string]string into KEY=VALUE strings.

#### slugExists

```go
func slugExists(cfg *config.Config, slug string) bool
```

slugExists checks whether a slug is already in use by any configured agent.

#### verifyGitRepository

```go
func verifyGitRepository(ctx context.Context, repoRoot string) error
```

verifyGitRepository ensures the given path is a git repository.
It runs `git rev-parse --show-toplevel` and verifies success.

#### worktreeListed

```go
func worktreeListed(output, worktreePath string) bool
```

worktreeListed checks whether a worktreePath appears in the
`git worktree list --porcelain` output.


## type AddLocalAgentOptions

```go
type AddLocalAgentOptions struct {
	Name     string
	About    string
	Adapter  string
	RepoRoot string
}
```

AddLocalAgentOptions holds input parameters for adding a local agent.

## type AgentActor

```go
type AgentActor struct {
	hsm.HSM
	*api.Agent
}
```

AgentActor wraps an Agent with HSM-driven lifecycle and presence state machines.
Per spec §5.4, the lifecycle is managed as an HSM.

### Functions returning AgentActor

#### NewAgentActor

```go
func NewAgentActor(ctx context.Context, agent *api.Agent) *AgentActor
```

NewAgentActor creates a new AgentActor with initialized HSMs.
Per spec §5.4, the lifecycle starts in the "pending" state.


### Methods

#### AgentActor.ErrorAgent

```go
func () ErrorAgent(ctx context.Context, err error)
```

ErrorAgent transitions the agent to the Errored state from any state.
Per spec §5.4, this can be triggered from any state.

#### AgentActor.GetSimpleState

```go
func () GetSimpleState() string
```

GetSimpleState returns the current lifecycle state without the qualified name prefix.
For example, returns "pending" instead of "/agent.lifecycle/pending".

#### AgentActor.GetState

```go
func () GetState() string
```

GetState returns the current lifecycle state.

#### AgentActor.Ready

```go
func () Ready(ctx context.Context)
```

Ready transitions the agent from Starting to Running.
Per spec §5.4, this is triggered by the "ready" event after bootstrap completes.

#### AgentActor.StartAgent

```go
func () StartAgent(ctx context.Context)
```

StartAgent transitions the agent from Pending to Starting.
Per spec §5.4, this is triggered by the "start" event.

#### AgentActor.StopAgent

```go
func () StopAgent(ctx context.Context)
```

StopAgent transitions the agent from Running to Terminated.
Per spec §5.4, this is triggered by the "stop" event.


## type LocalSession

```go
type LocalSession struct {
	Cmd *exec.Cmd
	PTY *os.File
}
```

LocalSession represents a local PTY-backed agent session.

Phase 2 provides basic PTY ownership for local agents; later phases add
monitoring and process tracking.

### Functions returning LocalSession

#### RestartLocalSession

```go
func RestartLocalSession(ctx context.Context, ag *api.Agent, prev *LocalSession, command []string, env map[string]string) (*LocalSession, error)
```

RestartLocalSession stops the previous session (if any) and starts a new one.

#### StartLocalSession

```go
func StartLocalSession(ctx context.Context, ag *api.Agent, command []string, env map[string]string) (*LocalSession, error)
```

StartLocalSession starts a new local PTY session for the given agent.
The process working directory is set to agent.Worktree per spec §5.3.1.


### Methods

#### LocalSession.Stop

```go
func () Stop() error
```

Stop terminates the local session and waits for the process to exit.


## type MergeStrategy

```go
type MergeStrategy string
```

AddLocalAgentOptions holds input parameters for adding a local agent.

MergeStrategy represents supported git merge strategies for agent worktrees.

### Constants

#### MergeStrategyMergeCommit, MergeStrategySquash, MergeStrategyRebase, MergeStrategyFFOnly

```go
const (
	MergeStrategyMergeCommit MergeStrategy = "merge-commit"
	MergeStrategySquash      MergeStrategy = "squash"
	MergeStrategyRebase      MergeStrategy = "rebase"
	MergeStrategyFFOnly      MergeStrategy = "ff-only"
)
```


### Functions returning MergeStrategy

#### SelectMergeStrategy

```go
func SelectMergeStrategy(cfg *config.Config) MergeStrategy
```

SelectMergeStrategy returns the effective merge strategy based on config,
defaulting to squash when an unknown or empty value is provided.


## type MessagePayload

```go
type MessagePayload struct {
	Message api.AgentMessage
}
```

MessagePayload wraps an AgentMessage for event dispatch.
Per spec §6.4, agents, host managers, and the director communicate using
AgentMessage payloads routed over the event system and, in later phases,
NATS participant channels.

## type PresenceActor

```go
type PresenceActor struct {
	hsm.HSM
	AgentID string // For logging/debugging
}
```

PresenceActor wraps an Agent with a presence state machine.
Per spec §6.1, presence indicates whether an agent can accept tasks.

### Functions returning PresenceActor

#### NewPresenceActor

```go
func NewPresenceActor(ctx context.Context, agentID string) *PresenceActor
```

NewPresenceActor creates a new PresenceActor with initialized presence HSM.
Per spec §6.1, presence starts in the "online" state.


### Methods

#### PresenceActor.ActivityDetected

```go
func () ActivityDetected(ctx context.Context)
```

ActivityDetected transitions from Away to Online.

#### PresenceActor.GetPresenceState

```go
func () GetPresenceState() string
```

GetPresenceState returns the current presence state.

#### PresenceActor.GetSimplePresenceState

```go
func () GetSimplePresenceState() string
```

GetSimplePresenceState returns the current presence state without the qualified name prefix.
For example, returns "online" instead of "/agent.presence/online".

#### PresenceActor.PromptDetected

```go
func () PromptDetected(ctx context.Context)
```

PromptDetected transitions from Busy to Online.

#### PresenceActor.RateCleared

```go
func () RateCleared(ctx context.Context)
```

RateCleared transitions from Offline to Online.

#### PresenceActor.RateLimit

```go
func () RateLimit(ctx context.Context)
```

RateLimit transitions to Offline.

#### PresenceActor.StuckDetected

```go
func () StuckDetected(ctx context.Context)
```

StuckDetected transitions to Away from any state.

#### PresenceActor.TaskAssigned

```go
func () TaskAssigned(ctx context.Context)
```

TaskAssigned transitions from Online to Busy.

#### PresenceActor.TaskCompleted

```go
func () TaskCompleted(ctx context.Context)
```

TaskCompleted transitions from Busy to Online.


## type PresenceChangedPayload

```go
type PresenceChangedPayload struct {
	AgentID  muid.MUID
	Presence string
}
```

PresenceChangedPayload is the payload for presence.changed events.
Per spec §6.1 and §6.5, presence indicates whether an agent can accept tasks.
IDs are kept as muid.MUID here; JSON encoding is handled at the event system
or control-plane layer per spec §9.1.3.1.

## type RosterStore

```go
type RosterStore struct {
	mu         sync.RWMutex
	entries    map[muid.MUID]api.RosterEntry
	dispatcher event.Dispatcher
}
```

RosterStore maintains the current roster of agents and their presence.
Per spec §6.2 and §6.3, the roster MUST be updated in real time as presence
changes occur and MUST be broadcast via presence.changed events.

### Functions returning RosterStore

#### NewRosterStore

```go
func NewRosterStore(dispatcher event.Dispatcher) *RosterStore
```

NewRosterStore creates a new RosterStore backed by the provided dispatcher.
If dispatcher is nil, a local in-process dispatcher is created.


### Methods

#### RosterStore.Dispatcher

```go
func () Dispatcher() event.Dispatcher
```

Dispatcher returns the underlying event dispatcher used by the store.
This is primarily exposed for tests and integration wiring.

#### RosterStore.List

```go
func () List() []api.RosterEntry
```

List returns a snapshot of the current roster entries, ordered deterministically
by name (then by AgentID) for stable listings.

#### RosterStore.RemoveAgent

```go
func () RemoveAgent(ctx context.Context, id muid.MUID) error
```

RemoveAgent removes an agent from the roster and emits a roster.updated event.

#### RosterStore.UpsertAgent

```go
func () UpsertAgent(ctx context.Context, ag *api.Agent, presence string) error
```

UpsertAgent inserts or updates an agent in the roster and emits presence and
roster events. Presence defaults to StateOnline when empty.

#### RosterStore.snapshotLocked

```go
func () snapshotLocked() []api.RosterEntry
```

snapshotLocked builds a sorted slice of roster entries. Caller must hold
either a read or write lock.


