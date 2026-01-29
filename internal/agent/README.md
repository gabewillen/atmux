# package agent

`import "github.com/stateforward/amux/internal/agent"`

Package agent implements agent orchestration (lifecycle, presence, messaging)

Package agent implements agent orchestration (lifecycle, presence, messaging)

- `ErrInvalidAgent` — ErrInvalidAgent is returned when an agent is invalid
- `type AgentActor` — AgentActor manages an agent's lifecycle and presence state machines
- `type AgentManager` — AgentManager manages multiple agents and their lifecycles
- `type AgentProcess` — AgentProcess represents a running agent process
- `type LifecycleState` — LifecycleState represents the agent lifecycle state
- `type PresenceState` — PresenceState represents the agent presence state

### Variables

#### ErrInvalidAgent

```go
var ErrInvalidAgent = errors.New("invalid agent")
```

ErrInvalidAgent is returned when an agent is invalid


## type AgentActor

```go
type AgentActor struct {
	ID             muid.MUID
	Agent          *api.Agent
	lifecycleState LifecycleState
	presenceState  PresenceState
	stateMutex     sync.RWMutex
	eventHandler   func(event interface{})
}
```

AgentActor manages an agent's lifecycle and presence state machines

### Functions returning AgentActor

#### NewAgentActor

```go
func NewAgentActor(agent *api.Agent, eventHandler func(event interface{})) (*AgentActor, error)
```

NewAgentActor creates a new agent actor with initialized state machines


### Methods

#### AgentActor.Connect

```go
func () Connect(ctx context.Context) error
```

Connect brings the agent online, transitioning from Offline to Online

#### AgentActor.CurrentLifecycleState

```go
func () CurrentLifecycleState() LifecycleState
```

CurrentLifecycleState returns the current lifecycle state

#### AgentActor.CurrentPresenceState

```go
func () CurrentPresenceState() PresenceState
```

CurrentPresenceState returns the current presence state

#### AgentActor.Disconnect

```go
func () Disconnect(ctx context.Context) error
```

Disconnect takes the agent offline, transitioning to Offline state

#### AgentActor.Error

```go
func () Error(ctx context.Context, err error) error
```

Error signals an error condition, transitioning to Errored state

#### AgentActor.FatalError

```go
func () FatalError(ctx context.Context, err error) error
```

FatalError signals a fatal error, transitioning to Errored state

#### AgentActor.HandleEvent

```go
func () HandleEvent(ctx context.Context, eventType string, eventData map[string]interface{}) error
```

HandleEvent processes an event that may trigger state transitions
This satisfies the spec requirement that events from the PTY monitor
can trigger state transitions via hsm.Dispatch() or equivalent

#### AgentActor.Ready

```go
func () Ready(ctx context.Context) error
```

Ready signals that the agent is ready, transitioning from Starting to Running

#### AgentActor.SetAvailable

```go
func () SetAvailable(ctx context.Context) error
```

SetAvailable marks the agent as available, transitioning from Busy to Online

#### AgentActor.SetAway

```go
func () SetAway(ctx context.Context) error
```

SetAway marks the agent as away, transitioning from Online to Away

#### AgentActor.SetBack

```go
func () SetBack(ctx context.Context) error
```

SetBack marks the agent as back, transitioning from Away to Online

#### AgentActor.SetBusy

```go
func () SetBusy(ctx context.Context) error
```

SetBusy marks the agent as busy, transitioning from Online to Busy

#### AgentActor.Start

```go
func () Start(ctx context.Context) error
```

Start initiates the agent lifecycle by transitioning from Pending to Starting

#### AgentActor.SubscribeToEvents

```go
func () SubscribeToEvents() error
```

SubscribeToEvents subscribes this agent to relevant events from the event system

#### AgentActor.Terminate

```go
func () Terminate(ctx context.Context) error
```

Terminate signals graceful termination, transitioning to Terminated state


## type AgentManager

```go
type AgentManager struct {
	agents      map[muid.MUID]*AgentActor
	processes   map[muid.MUID]*AgentProcess
	agentsMutex sync.RWMutex
	config      *config.Config
	resolver    *paths.Resolver
}
```

AgentManager manages multiple agents and their lifecycles

### Functions returning AgentManager

#### NewAgentManager

```go
func NewAgentManager(cfg *config.Config) (*AgentManager, error)
```

NewAgentManager creates a new agent manager


### Methods

#### AgentManager.AddAgent

```go
func () AddAgent(ctx context.Context, name, about, adapter string, location *api.Location) error
```

AddAgent adds a new agent to the manager

#### AgentManager.AttachAgent

```go
func () AttachAgent(ctx context.Context, agentID muid.MUID) error
```

AttachAgent connects to an existing agent process

#### AgentManager.GetAgent

```go
func () GetAgent(agentID muid.MUID) (*AgentActor, bool)
```

GetAgent returns an agent by ID

#### AgentManager.GetBaseBranchForRepo

```go
func () GetBaseBranchForRepo(repoPath, targetBranch string) (string, error)
```

GetBaseBranchForRepo determines the base branch for a repository
Following the spec: run `git symbolic-ref --quiet --short HEAD` and use the output
If that fails, use targetBranch if provided, otherwise return an error

#### AgentManager.GetDefaultMergeStrategy

```go
func () GetDefaultMergeStrategy() git.MergeStrategy
```

GetDefaultMergeStrategy returns the default merge strategy

#### AgentManager.KillAgent

```go
func () KillAgent(ctx context.Context, agentID muid.MUID) error
```

KillAgent forcefully kills an agent process

#### AgentManager.ListAgents

```go
func () ListAgents() []*api.Agent
```

ListAgents returns a list of all agents

#### AgentManager.MergeAgentChanges

```go
func () MergeAgentChanges(ctx context.Context, agentID muid.MUID, targetBranch string, strategy git.MergeStrategy) error
```

MergeAgentChanges performs a git merge of the agent's worktree changes to the target branch

#### AgentManager.RemoveAgent

```go
func () RemoveAgent(ctx context.Context, agentID muid.MUID) error
```

RemoveAgent removes an agent from the manager

#### AgentManager.RestartAgent

```go
func () RestartAgent(ctx context.Context, agentID muid.MUID) error
```

RestartAgent stops and then starts an agent

#### AgentManager.SetDefaultMergeStrategy

```go
func () SetDefaultMergeStrategy(strategy git.MergeStrategy)
```

SetDefaultMergeStrategy sets the default merge strategy for the manager

#### AgentManager.SpawnAgent

```go
func () SpawnAgent(ctx context.Context, agentID muid.MUID) error
```

SpawnAgent starts a new agent process in its worktree

#### AgentManager.StartAgent

```go
func () StartAgent(ctx context.Context, agentID muid.MUID) error
```

StartAgent transitions an agent to the Running state

#### AgentManager.StopAgent

```go
func () StopAgent(ctx context.Context, agentID muid.MUID) error
```

StopAgent gracefully stops an agent process

#### AgentManager.createOrReuseWorktree

```go
func () createOrReuseWorktree(repoRoot, worktreePath, agentSlug string) error
```

createOrReuseWorktree creates or reuses a git worktree for the agent

#### AgentManager.findGitRepo

```go
func () findGitRepo(location *api.Location) (string, error)
```

findGitRepo determines the git repository root based on the location

#### AgentManager.isGitRepo

```go
func () isGitRepo(path string) bool
```

isGitRepo checks if a directory is a git repository


## type AgentProcess

```go
type AgentProcess struct {
	ID      muid.MUID
	Cmd     *exec.Cmd
	PTY     *os.File
	WorkDir string
}
```

AgentProcess represents a running agent process

## type LifecycleState

```go
type LifecycleState string
```

LifecycleState represents the agent lifecycle state

### Constants

#### LifecyclePending, LifecycleStarting, LifecycleRunning, LifecycleTerminated, LifecycleErrored

```go
const (
	LifecyclePending    LifecycleState = "pending"
	LifecycleStarting   LifecycleState = "starting"
	LifecycleRunning    LifecycleState = "running"
	LifecycleTerminated LifecycleState = "terminated"
	LifecycleErrored    LifecycleState = "errored"
)
```


## type PresenceState

```go
type PresenceState string
```

PresenceState represents the agent presence state

### Constants

#### PresenceOnline, PresenceBusy, PresenceOffline, PresenceAway

```go
const (
	PresenceOnline  PresenceState = "online"
	PresenceBusy    PresenceState = "busy"
	PresenceOffline PresenceState = "offline"
	PresenceAway    PresenceState = "away"
)
```


