# package agent

`import "github.com/stateforward/amux/internal/agent"`

Package agent implements agent orchestration (lifecycle, presence, messaging)

- `ErrInvalidAgent` — ErrInvalidAgent is returned when an agent is invalid
- `type AgentActor` — AgentActor manages an agent's lifecycle and presence state machines
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

#### AgentActor.Terminate

```go
func () Terminate(ctx context.Context) error
```

Terminate signals graceful termination, transitioning to Terminated state


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


