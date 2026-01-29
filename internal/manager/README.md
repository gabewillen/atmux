# package manager

`import "github.com/agentflare-ai/amux/internal/manager"`

Package manager manages local agents, worktrees, and sessions.

- `ErrAgentNotFound, ErrAgentAmbiguous, ErrAgentInvalid`
- `func encodeAgents(agents []config.AgentConfig) []any`
- `func ensureGitRepo(repoRoot string) error`
- `func extractAgents(raw map[string]any) []config.AgentConfig`
- `func lastStateSegment(state string) string`
- `func sameAgent(a, b config.AgentConfig) bool`
- `func statePresence(runtime *agent.Agent) string`
- `type AddRequest` — AddRequest describes a local agent add request.
- `type AgentRecord` — AgentRecord describes a managed agent.
- `type LocalManager` — LocalManager manages local agents and sessions.
- `type RemoveRequest` — RemoveRequest describes an agent removal request.
- `type agentState`

### Variables

#### ErrAgentNotFound, ErrAgentAmbiguous, ErrAgentInvalid

```go
var (
	// ErrAgentNotFound is returned when an agent cannot be found.
	ErrAgentNotFound = errors.New("agent not found")
	// ErrAgentAmbiguous is returned when a name matches multiple agents.
	ErrAgentAmbiguous = errors.New("agent name is ambiguous")
	// ErrAgentInvalid is returned when an agent request is invalid.
	ErrAgentInvalid = errors.New("agent invalid")
)
```


### Functions

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

#### sameAgent

```go
func sameAgent(a, b config.AgentConfig) bool
```

#### statePresence

```go
func statePresence(runtime *agent.Agent) string
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

AddRequest describes a local agent add request.

## type AgentRecord

```go
type AgentRecord struct {
	ID       api.AgentID
	Name     string
	About    string
	Adapter  string
	RepoRoot string
	Worktree string
	Slug     string
	Presence string
	Location api.Location
}
```

AgentRecord describes a managed agent.

## type LocalManager

```go
type LocalManager struct {
	resolver        *paths.Resolver
	dispatcher      protocol.Dispatcher
	cfg             config.Config
	git             *git.Runner
	mu              sync.Mutex
	agents          map[api.AgentID]*agentState
	nameIndex       map[string][]api.AgentID
	bases           map[string]string
	registries      map[string]adapter.Registry
	registryFactory func(*paths.Resolver) (adapter.Registry, error)
}
```

LocalManager manages local agents and sessions.

### Functions returning LocalManager

#### NewLocalManager

```go
func NewLocalManager(ctx context.Context, resolver *paths.Resolver, cfg config.Config, dispatcher protocol.Dispatcher) (*LocalManager, error)
```

NewLocalManager constructs a LocalManager.


### Methods

#### LocalManager.AddAgent

```go
func () AddAgent(ctx context.Context, req AddRequest) (AgentRecord, error)
```

AddAgent adds and starts a local agent.

#### LocalManager.AttachAgent

```go
func () AttachAgent(id api.AgentID) (*os.File, error)
```

AttachAgent attaches to a running agent PTY.

#### LocalManager.ListAgents

```go
func () ListAgents() ([]AgentRecord, error)
```

ListAgents returns the current roster.

#### LocalManager.MergeAgent

```go
func () MergeAgent(ctx context.Context, id api.AgentID, strategy git.MergeStrategy, targetBranch string) (git.MergeResult, error)
```

MergeAgent integrates an agent branch into a target branch.

#### LocalManager.RemoveAgent

```go
func () RemoveAgent(ctx context.Context, req RemoveRequest) error
```

RemoveAgent removes an agent and its worktree.

#### LocalManager.RestartAgent

```go
func () RestartAgent(ctx context.Context, id api.AgentID) error
```

RestartAgent restarts a running agent session.

#### LocalManager.SetRegistryFactory

```go
func () SetRegistryFactory(factory func(*paths.Resolver) (adapter.Registry, error))
```

SetRegistryFactory overrides the adapter registry factory.

#### LocalManager.StartAgent

```go
func () StartAgent(ctx context.Context, id api.AgentID) error
```

StartAgent starts an existing agent session.

#### LocalManager.StopAgent

```go
func () StopAgent(ctx context.Context, id api.AgentID) error
```

StopAgent stops a running agent session.

#### LocalManager.appendAgentConfig

```go
func () appendAgentConfig(entry config.AgentConfig) error
```

#### LocalManager.baseBranch

```go
func () baseBranch(ctx context.Context, repoRoot string) (string, error)
```

#### LocalManager.emit

```go
func () emit(ctx context.Context, name string, payload any)
```

#### LocalManager.findAgent

```go
func () findAgent(req RemoveRequest) (*agentState, api.AgentID, error)
```

#### LocalManager.loadFromConfig

```go
func () loadFromConfig(ctx context.Context) error
```

#### LocalManager.registry

```go
func () registry(resolver *paths.Resolver) (adapter.Registry, error)
```

#### LocalManager.removeAgentConfig

```go
func () removeAgentConfig(entry config.AgentConfig) error
```

#### LocalManager.removeConfigEntryLocked

```go
func () removeConfigEntryLocked(entry config.AgentConfig)
```

#### LocalManager.removeNameIndexLocked

```go
func () removeNameIndexLocked(name string, id api.AgentID)
```

#### LocalManager.resolveLocation

```go
func () resolveLocation(req AddRequest) (api.Location, string, error)
```

#### LocalManager.startSession

```go
func () startSession(ctx context.Context, id api.AgentID) (*session.LocalSession, error)
```

#### LocalManager.stopSession

```go
func () stopSession(ctx context.Context, id api.AgentID) error
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
	runtime  *agent.Agent
	slug     string
	repoRoot string
	worktree string
	session  *session.LocalSession
	config   config.AgentConfig
}
```

