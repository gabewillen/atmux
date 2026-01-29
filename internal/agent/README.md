# package agent

`import "github.com/stateforward/amux/internal/agent"`

Package agent implements agent orchestration (lifecycle, presence, messaging)

Package agent implements agent orchestration (lifecycle, presence, messaging)

Package agent implements agent orchestration (lifecycle, presence, messaging)

Package agent implements agent orchestration (lifecycle, presence, messaging)

The agent package provides core functionality for managing agents including:
  - Lifecycle management (Pending → Starting → Running → Terminated/Errored)
  - Presence management (Online ↔ Busy ↔ Offline ↔ Away)
  - Roster maintenance for tracking all agents and their states
  - Inter-agent messaging capabilities

- `BroadcastID` — BroadcastID is a special ID for broadcast to all participants
- `ErrInvalidAgent` — ErrInvalidAgent is returned when an agent is invalid
- `type AgentActor` — AgentActor manages an agent's lifecycle and presence state machines
- `type AgentManager` — AgentManager manages multiple agents and their lifecycles
- `type AgentMessage` — AgentMessage represents a message between agents
- `type AgentProcess` — AgentProcess represents a running agent process
- `type LifecycleState` — LifecycleState represents the agent lifecycle state
- `type MessageRouter` — MessageRouter handles routing of messages between agents
- `type PresenceChangeCallback` — PresenceChangeCallback is a function that gets called when an agent's presence changes
- `type PresenceState` — PresenceState represents the agent presence state
- `type RosterEntry` — RosterEntry represents an entry in the roster
- `type Roster` — Add a presence subscriptions field to the Roster
- `type Subscription` — Subscription represents a subscription to presence changes
- `type presenceSubscriptions` — presenceSubscriptions holds all presence change subscriptions

### Constants

#### BroadcastID

```go
const BroadcastID muid.MUID = 0
```

BroadcastID is a special ID for broadcast to all participants


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


## type AgentMessage

```go
type AgentMessage struct {
	ID        muid.MUID `json:"id"`
	From      muid.MUID `json:"from"`    // Sender runtime ID (set by publishing component)
	To        muid.MUID `json:"to"`      // Recipient runtime ID (set by publishing component, or BroadcastID)
	ToSlug    string    `json:"to_slug"` // Recipient token captured from text (typically agent_slug); case-insensitive
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}
```

AgentMessage represents a message between agents

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


## type MessageRouter

```go
type MessageRouter struct {
	agents map[muid.MUID]*AgentActor
	roster *Roster
	// In a real implementation, this would connect to NATS for distributed messaging
	natsEnabled bool
}
```

MessageRouter handles routing of messages between agents

### Functions returning MessageRouter

#### NewMessageRouter

```go
func NewMessageRouter(roster *Roster) *MessageRouter
```

NewMessageRouter creates a new message router


### Methods

#### MessageRouter.DisableNATS

```go
func () DisableNATS()
```

DisableNATS disables NATS-based messaging (fallback to local routing)

#### MessageRouter.EnableNATS

```go
func () EnableNATS()
```

EnableNATS enables NATS-based messaging

#### MessageRouter.RegisterAgent

```go
func () RegisterAgent(agent *AgentActor)
```

RegisterAgent registers an agent with the message router

#### MessageRouter.SendMessage

```go
func () SendMessage(ctx context.Context, msg *AgentMessage) error
```

SendMessage sends a message to a specific agent or broadcasts to all
Implements the inter-agent messaging routes as specified in the spec

#### MessageRouter.UnregisterAgent

```go
func () UnregisterAgent(agentID muid.MUID)
```

UnregisterAgent removes an agent from the message router

#### MessageRouter.broadcastMessage

```go
func () broadcastMessage(ctx context.Context, msg *AgentMessage) error
```

broadcastMessage sends a message to all registered agents

#### MessageRouter.handleReceivedMessage

```go
func () handleReceivedMessage(ctx context.Context, agent *AgentActor, msg *AgentMessage)
```

handleReceivedMessage simulates handling of a received message by an agent

#### MessageRouter.publishToNATS

```go
func () publishToNATS(ctx context.Context, msg *AgentMessage) error
```

publishToNATS publishes the message to NATS subjects for distributed messaging
This is a placeholder that will be fully implemented in Phase 7

#### MessageRouter.sendToAgent

```go
func () sendToAgent(ctx context.Context, msg *AgentMessage) error
```

sendToAgent sends a message to a specific agent

#### MessageRouter.sendToSlug

```go
func () sendToSlug(ctx context.Context, msg *AgentMessage) error
```

sendToSlug finds an agent by slug and sends the message


## type PresenceChangeCallback

```go
type PresenceChangeCallback func(*RosterEntry)
```

PresenceChangeCallback is a function that gets called when an agent's presence changes

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


## type Roster

```go
type Roster struct {
	agents       map[muid.MUID]*RosterEntry
	mutex        sync.RWMutex
	presenceSubs *presenceSubscriptions
}
```

Add a presence subscriptions field to the Roster

### Functions returning Roster

#### NewRoster

```go
func NewRoster() *Roster
```

NewRoster creates a new roster instance


### Methods

#### Roster.AddAgent

```go
func () AddAgent(agent *api.Agent, presence PresenceState)
```

AddAgent adds an agent to the roster

#### Roster.GetAgent

```go
func () GetAgent(agentID muid.MUID) (*RosterEntry, bool)
```

GetAgent returns the roster entry for an agent

#### Roster.GetAgentsByPresence

```go
func () GetAgentsByPresence(presence PresenceState) []*RosterEntry
```

GetAgentsByPresence returns agents filtered by presence state

#### Roster.GetAllAgents

```go
func () GetAllAgents() []*RosterEntry
```

GetAllAgents returns all agents in the roster

#### Roster.GetAwayAgents

```go
func () GetAwayAgents() []*RosterEntry
```

GetAwayAgents returns all agents that are currently away

#### Roster.GetBusyAgents

```go
func () GetBusyAgents() []*RosterEntry
```

GetBusyAgents returns all agents that are currently busy

#### Roster.GetOfflineAgents

```go
func () GetOfflineAgents() []*RosterEntry
```

GetOfflineAgents returns all agents that are currently offline

#### Roster.GetOnlineAgents

```go
func () GetOnlineAgents() []*RosterEntry
```

GetOnlineAgents returns all agents that are currently online

#### Roster.RemoveAgent

```go
func () RemoveAgent(agentID muid.MUID)
```

RemoveAgent removes an agent from the roster

#### Roster.Size

```go
func () Size() int
```

Size returns the number of agents in the roster

#### Roster.SubscribeToPresenceChanges

```go
func () SubscribeToPresenceChanges(ctx context.Context, handler func(*RosterEntry)) string
```

SubscribeToPresenceChanges allows components to subscribe to roster/presence changes

#### Roster.UnsubscribeFromPresenceChanges

```go
func () UnsubscribeFromPresenceChanges(subID string)
```

UnsubscribeFromPresenceChanges removes a subscription to presence changes

#### Roster.UpdatePresence

```go
func () UpdatePresence(agentID muid.MUID, presence PresenceState)
```

UpdatePresence updates the presence state of an agent in the roster

#### Roster.UpdateTask

```go
func () UpdateTask(agentID muid.MUID, task string)
```

UpdateTask updates the current task of an agent in the roster

#### Roster.notifyPresenceChange

```go
func () notifyPresenceChange(agentID muid.MUID)
```

notifyPresenceChange notifies all subscribers about a presence change for an agent


## type RosterEntry

```go
type RosterEntry struct {
	ID       muid.MUID     `json:"id"`
	Name     string        `json:"name"`
	Adapter  string        `json:"adapter"`
	Presence PresenceState `json:"presence"`
	RepoRoot string        `json:"repo_root"`
	HostID   muid.MUID     `json:"host_id,omitempty"`
	Task     string        `json:"task,omitempty"` // Current task if agent is busy
}
```

RosterEntry represents an entry in the roster

## type Subscription

```go
type Subscription struct {
	id       string
	callback PresenceChangeCallback
}
```

Subscription represents a subscription to presence changes

## type presenceSubscriptions

```go
type presenceSubscriptions struct {
	subs   map[string]PresenceChangeCallback
	nextID int
	mutex  sync.RWMutex
}
```

presenceSubscriptions holds all presence change subscriptions

### Functions returning presenceSubscriptions

#### newPresenceSubscriptions

```go
func newPresenceSubscriptions() *presenceSubscriptions
```

newPresenceSubscriptions creates a new presenceSubscriptions instance


### Methods

#### presenceSubscriptions.NotifyAll

```go
func () NotifyAll(entry *RosterEntry)
```

NotifyAll notifies all subscribers about a presence change

#### presenceSubscriptions.Subscribe

```go
func () Subscribe(callback PresenceChangeCallback) string
```

Subscribe adds a new subscription to presence changes

#### presenceSubscriptions.Unsubscribe

```go
func () Unsubscribe(id string)
```

Unsubscribe removes a subscription


