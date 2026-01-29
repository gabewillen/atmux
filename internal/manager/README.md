# package manager

`import "github.com/agentflare-ai/amux/internal/manager"`

Package manager manages local agents, worktrees, and sessions.

- `ErrAgentNotFound, ErrAgentAmbiguous, ErrAgentInvalid, ErrRepoPathRequired`
- `func buildAdapterBundle(resolver *paths.Resolver, name string) (remote.AdapterBundle, error)`
- `func decodeEventPayload(payload any, dest any) error`
- `func encodeAgents(agents []config.AgentConfig) []any`
- `func ensureGitRepo(repoRoot string) error`
- `func extractAgents(raw map[string]any) []config.AgentConfig`
- `func lastStateSegment(state string) string`
- `func locationForState(state *agentState) *api.Location`
- `func rosterOrder(kind api.RosterKind) int`
- `func sameAgent(a, b config.AgentConfig) bool`
- `func sortRoster(entries []api.RosterEntry)`
- `func statePresence(state *agentState) string`
- `shutdownModel`
- `shutdownStateRunning, shutdownStateDraining, shutdownStateTerminating, shutdownStateStopped, shutdownEventRequest, shutdownEventForce, shutdownEventDrainComplete, shutdownEventDrainTimeout, shutdownEventTerminateComplete`
- `type AddRequest` — AddRequest describes an agent add request.
- `type Manager` — Manager manages local and remote agents and sessions.
- `type RemoveRequest` — RemoveRequest describes an agent removal request.
- `type agentState`
- `type shutdownController`
- `type shutdownTarget`

### Constants

#### shutdownStateRunning, shutdownStateDraining, shutdownStateTerminating, shutdownStateStopped, shutdownEventRequest, shutdownEventForce, shutdownEventDrainComplete, shutdownEventDrainTimeout, shutdownEventTerminateComplete

```go
const (
	shutdownStateRunning     = "running"
	shutdownStateDraining    = "draining"
	shutdownStateTerminating = "terminating"
	shutdownStateStopped     = "stopped"

	shutdownEventRequest           = "shutdown.request"
	shutdownEventForce             = "shutdown.force"
	shutdownEventDrainComplete     = "drain.complete"
	shutdownEventDrainTimeout      = "drain.timeout"
	shutdownEventTerminateComplete = "terminate.complete"
)
```


### Variables

#### ErrAgentNotFound, ErrAgentAmbiguous, ErrAgentInvalid, ErrRepoPathRequired

```go
var (
	// ErrAgentNotFound is returned when an agent cannot be found.
	ErrAgentNotFound = errors.New("agent not found")
	// ErrAgentAmbiguous is returned when a name matches multiple agents.
	ErrAgentAmbiguous = errors.New("agent name is ambiguous")
	// ErrAgentInvalid is returned when an agent request is invalid.
	ErrAgentInvalid = errors.New("agent invalid")
	// ErrRepoPathRequired is returned when repo_path is required by the spec.
	ErrRepoPathRequired = errors.New("repo path required")
)
```

#### shutdownModel

```go
var shutdownModel = hsm.Define(
	"system.shutdown",
	hsm.State(shutdownStateRunning),
	hsm.State(
		shutdownStateDraining,
		hsm.Entry(func(ctx context.Context, actor *shutdownController, event hsm.Event) {
			actor.onDraining(ctx)
		}),
	),
	hsm.State(
		shutdownStateTerminating,
		hsm.Entry(func(ctx context.Context, actor *shutdownController, event hsm.Event) {
			actor.onTerminating(ctx)
		}),
	),
	hsm.Final(shutdownStateStopped),

	hsm.Transition(hsm.On(hsm.Event{Name: shutdownEventRequest}), hsm.Source(shutdownStateRunning), hsm.Target(shutdownStateDraining)),
	hsm.Transition(hsm.On(hsm.Event{Name: shutdownEventForce}), hsm.Source(shutdownStateRunning), hsm.Target(shutdownStateTerminating)),
	hsm.Transition(hsm.On(hsm.Event{Name: shutdownEventForce}), hsm.Source(shutdownStateDraining), hsm.Target(shutdownStateTerminating)),
	hsm.Transition(
		hsm.On(hsm.Event{Name: shutdownEventDrainComplete}),
		hsm.Source(shutdownStateDraining),
		hsm.Target(shutdownStateStopped),
		hsm.Effect(func(ctx context.Context, actor *shutdownController, event hsm.Event) {
			actor.onStopped()
		}),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: shutdownEventDrainTimeout}),
		hsm.Source(shutdownStateDraining),
		hsm.Target(shutdownStateTerminating),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: shutdownEventTerminateComplete}),
		hsm.Source(shutdownStateTerminating),
		hsm.Target(shutdownStateStopped),
		hsm.Effect(func(ctx context.Context, actor *shutdownController, event hsm.Event) {
			actor.onStopped()
		}),
	),

	hsm.Initial(hsm.Target(shutdownStateRunning)),
)
```


### Functions

#### buildAdapterBundle

```go
func buildAdapterBundle(resolver *paths.Resolver, name string) (remote.AdapterBundle, error)
```

#### decodeEventPayload

```go
func decodeEventPayload(payload any, dest any) error
```

#### encodeAgents

```go
func encodeAgents(agents []config.AgentConfig) []any
```

#### ensureGitRepo

```go
func ensureGitRepo(repoRoot string) error
```

#### extractAgents

```go
func extractAgents(raw map[string]any) []config.AgentConfig
```

#### lastStateSegment

```go
func lastStateSegment(state string) string
```

#### locationForState

```go
func locationForState(state *agentState) *api.Location
```

#### rosterOrder

```go
func rosterOrder(kind api.RosterKind) int
```

#### sameAgent

```go
func sameAgent(a, b config.AgentConfig) bool
```

#### sortRoster

```go
func sortRoster(entries []api.RosterEntry)
```

#### statePresence

```go
func statePresence(state *agentState) string
```


## type AddRequest

```go
type AddRequest struct {
	Name     string
	About    string
	Adapter  string
	Location api.Location
	Cwd      string
}
```

AddRequest describes an agent add request.

## type Manager

```go
type Manager struct {
	resolver        *paths.Resolver
	dispatcher      protocol.Dispatcher
	cfg             config.Config
	git             *git.Runner
	remoteDirector  *remote.Director
	mu              sync.Mutex
	agents          map[api.AgentID]*agentState
	nameIndex       map[string][]api.AgentID
	bases           map[string]string
	registries      map[string]adapter.Registry
	registryFactory func(*paths.Resolver) (adapter.Registry, error)
	subs            []protocol.Subscription
	shutdownMu      sync.Mutex
	shutdown        *shutdownController
}
```

Manager manages local and remote agents and sessions.

### Functions returning Manager

#### NewManager

```go
func NewManager(ctx context.Context, resolver *paths.Resolver, cfg config.Config, dispatcher protocol.Dispatcher, version string) (*Manager, error)
```

NewManager constructs a Manager.


### Methods

#### Manager.AddAgent

```go
func () AddAgent(ctx context.Context, req AddRequest) (api.RosterEntry, error)
```

AddAgent adds and starts a local agent.

#### Manager.AttachAgent

```go
func () AttachAgent(id api.AgentID) (net.Conn, error)
```

AttachAgent attaches to a running agent PTY.

#### Manager.KillAgent

```go
func () KillAgent(ctx context.Context, id api.AgentID) error
```

KillAgent forces a running agent session to stop.

#### Manager.ListAgents

```go
func () ListAgents() ([]api.RosterEntry, error)
```

ListAgents returns the current roster entries.

#### Manager.MergeAgent

```go
func () MergeAgent(ctx context.Context, id api.AgentID, strategy git.MergeStrategy, targetBranch string) (git.MergeResult, error)
```

MergeAgent integrates an agent branch into a target branch.

#### Manager.RemoveAgent

```go
func () RemoveAgent(ctx context.Context, req RemoveRequest) error
```

RemoveAgent removes an agent and its worktree.

#### Manager.RestartAgent

```go
func () RestartAgent(ctx context.Context, id api.AgentID) error
```

RestartAgent restarts a running agent session.

#### Manager.SetRegistryFactory

```go
func () SetRegistryFactory(factory func(*paths.Resolver) (adapter.Registry, error))
```

SetRegistryFactory overrides the adapter registry factory.

#### Manager.Shutdown

```go
func () Shutdown(ctx context.Context, force bool) error
```

Shutdown drains all running sessions and optionally forces termination.

#### Manager.StartAgent

```go
func () StartAgent(ctx context.Context, id api.AgentID) error
```

StartAgent starts an existing agent session.

#### Manager.StopAgent

```go
func () StopAgent(ctx context.Context, id api.AgentID) error
```

StopAgent stops a running agent session.

#### Manager.addRemoteAgent

```go
func () addRemoteAgent(ctx context.Context, req AddRequest, location api.Location, repoRoot string, explicitRepoPath bool) (api.RosterEntry, error)
```

#### Manager.appendAgentConfig

```go
func () appendAgentConfig(entry config.AgentConfig) error
```

#### Manager.baseBranch

```go
func () baseBranch(ctx context.Context, repoRoot string) (string, error)
```

#### Manager.buildAgentMessage

```go
func () buildAgentMessage(sender api.AgentID, payload api.OutboundMessage) (api.AgentMessage, bool)
```

#### Manager.cleanupWorktrees

```go
func () cleanupWorktrees(ctx context.Context, targets []shutdownTarget) error
```

#### Manager.clearSessions

```go
func () clearSessions(targets []shutdownTarget)
```

#### Manager.commSubjectForTarget

```go
func () commSubjectForTarget(target api.TargetID) string
```

#### Manager.deliverBroadcast

```go
func () deliverBroadcast(payload api.AgentMessage)
```

#### Manager.deliverMessage

```go
func () deliverMessage(state *agentState, payload api.AgentMessage)
```

#### Manager.deliverToTarget

```go
func () deliverToTarget(payload api.AgentMessage) bool
```

#### Manager.directorPeerID

```go
func () directorPeerID() api.PeerID
```

#### Manager.dispatchAgentLifecycle

```go
func () dispatchAgentLifecycle(ctx context.Context, targets []shutdownTarget, name string)
```

#### Manager.drainSessions

```go
func () drainSessions(ctx context.Context, targets []shutdownTarget) (bool, error)
```

#### Manager.emitAgentEvent

```go
func () emitAgentEvent(ctx context.Context, name string, payload any)
```

#### Manager.emitEvent

```go
func () emitEvent(ctx context.Context, category string, name string, payload any)
```

#### Manager.emitMessageEvent

```go
func () emitMessageEvent(name string, payload api.AgentMessage)
```

#### Manager.emitPresenceChanged

```go
func () emitPresenceChanged(ctx context.Context, id api.AgentID, presence string)
```

#### Manager.emitRosterUpdated

```go
func () emitRosterUpdated(ctx context.Context)
```

#### Manager.emitSystemEvent

```go
func () emitSystemEvent(ctx context.Context, name string, payload any)
```

#### Manager.ensureShutdownController

```go
func () ensureShutdownController() *shutdownController
```

#### Manager.findAgent

```go
func () findAgent(req RemoveRequest) (*agentState, api.AgentID, error)
```

#### Manager.forceTerminate

```go
func () forceTerminate(ctx context.Context, targets []shutdownTarget) error
```

#### Manager.handleCommMessage

```go
func () handleCommMessage(msg protocol.Message)
```

#### Manager.handleOutboundMessage

```go
func () handleOutboundMessage(event protocol.Event)
```

#### Manager.handlePresenceEvent

```go
func () handlePresenceEvent(event protocol.Event)
```

#### Manager.handleRemoteEvent

```go
func () handleRemoteEvent(msg protocol.Message)
```

#### Manager.loadFromConfig

```go
func () loadFromConfig(ctx context.Context) error
```

#### Manager.localHostID

```go
func () localHostID() api.HostID
```

#### Manager.peerForHost

```go
func () peerForHost(hostID api.HostID) (api.PeerID, bool)
```

#### Manager.publishAgentMessage

```go
func () publishAgentMessage(sender api.AgentID, msg api.AgentMessage)
```

#### Manager.publishComm

```go
func () publishComm(subject string, payload []byte)
```

#### Manager.registry

```go
func () registry(resolver *paths.Resolver) (adapter.Registry, error)
```

#### Manager.releaseShutdownController

```go
func () releaseShutdownController(controller *shutdownController)
```

#### Manager.removeAgentConfig

```go
func () removeAgentConfig(entry config.AgentConfig) error
```

#### Manager.removeConfigEntryLocked

```go
func () removeConfigEntryLocked(entry config.AgentConfig)
```

#### Manager.removeNameIndexLocked

```go
func () removeNameIndexLocked(name string, id api.AgentID)
```

#### Manager.resolveLocation

```go
func () resolveLocation(req AddRequest) (api.Location, string, error)
```

#### Manager.resolveSender

```go
func () resolveSender(payload api.OutboundMessage) (api.AgentID, *agentState, bool)
```

#### Manager.resolveToID

```go
func () resolveToID(slug string) (api.TargetID, bool)
```

#### Manager.rosterEntry

```go
func () rosterEntry(id api.AgentID, state *agentState) api.RosterEntry
```

#### Manager.routeOutboundMessage

```go
func () routeOutboundMessage(ctx context.Context, payload api.OutboundMessage)
```

#### Manager.setPresence

```go
func () setPresence(ctx context.Context, id api.AgentID, presence string, emit bool) bool
```

#### Manager.shutdownTargets

```go
func () shutdownTargets() []shutdownTarget
```

#### Manager.spawnRemote

```go
func () spawnRemote(ctx context.Context, hostID api.HostID, req remote.SpawnRequest) (remote.SpawnResponse, error)
```

#### Manager.startMessageRouting

```go
func () startMessageRouting(ctx context.Context) error
```

#### Manager.startPresenceRouting

```go
func () startPresenceRouting(ctx context.Context) error
```

#### Manager.startRemoteSession

```go
func () startRemoteSession(ctx context.Context, id api.AgentID, state *agentState) error
```

#### Manager.startSession

```go
func () startSession(ctx context.Context, id api.AgentID) (*session.LocalSession, error)
```

#### Manager.stopSession

```go
func () stopSession(ctx context.Context, id api.AgentID) error
```

#### Manager.systemRosterLocked

```go
func () systemRosterLocked() []api.RosterEntry
```

#### Manager.updateRemotePresence

```go
func () updateRemotePresence(ctx context.Context, hostID api.HostID, presence string, onlyIfAway bool)
```

#### Manager.validateMultiRepo

```go
func () validateMultiRepo(repoRoot string, explicitRepoPath bool) error
```


## type RemoveRequest

```go
type RemoveRequest struct {
	AgentID api.AgentID
	Name    string
}
```

RemoveRequest describes an agent removal request.

## type agentState

```go
type agentState struct {
	runtime          *agent.Agent
	slug             string
	repoRoot         string
	worktree         string
	session          *session.LocalSession
	formatter        adapter.ActionFormatter
	remoteHost       api.HostID
	remoteSession    api.SessionID
	remote           bool
	config           config.AgentConfig
	explicitRepoPath bool
	presence         string
	task             string
}
```

## type shutdownController

```go
type shutdownController struct {
	hsm.HSM
	manager *Manager
	done    chan struct{}
	errMu   sync.Mutex
	err     error
	once    sync.Once
}
```

### Functions returning shutdownController

#### newShutdownController

```go
func newShutdownController(m *Manager) *shutdownController
```


### Methods

#### shutdownController.error

```go
func () error() error
```

#### shutdownController.onDraining

```go
func () onDraining(ctx context.Context)
```

#### shutdownController.onStopped

```go
func () onStopped()
```

#### shutdownController.onTerminating

```go
func () onTerminating(ctx context.Context)
```

#### shutdownController.recordError

```go
func () recordError(err error)
```

#### shutdownController.signal

```go
func () signal(ctx context.Context, name string, payload any)
```

#### shutdownController.wait

```go
func () wait(ctx context.Context) error
```


## type shutdownTarget

```go
type shutdownTarget struct {
	id       api.AgentID
	repoRoot string
	slug     string
	session  *session.LocalSession
	runtime  *agent.Agent
}
```

