# package agent

`import "github.com/copilot-claude-sonnet-4/amux/internal/agent"`

Package agent provides agent-agnostic orchestration functionality.
This package manages agent lifecycle, presence, and messaging without
any knowledge of specific agent implementations.

All agent-specific behavior is delegated to WASM adapters loaded
via the adapter package.

- `ErrAgentNotFound, ErrInvalidState, ErrAdapterLoadFailed, ErrInvalidTransition` — Common sentinel errors for agent operations.
- `type AgentActor` — AgentActor wraps an Agent with simple state machines for lifecycle and presence.
- `type LifecycleEvent` — LifecycleEvent represents events that can trigger agent lifecycle transitions.
- `type Manager` — Manager orchestrates multiple agents in an agent-agnostic manner.
- `type PresenceEvent` — PresenceEvent represents events that can trigger presence transitions.

### Variables

#### ErrAgentNotFound, ErrInvalidState, ErrAdapterLoadFailed, ErrInvalidTransition

```go
var (
	// ErrAgentNotFound indicates an agent with the given ID was not found.
	ErrAgentNotFound = errors.New("agent not found")

	// ErrInvalidState indicates an operation cannot be performed in the current agent state.
	ErrInvalidState = errors.New("invalid agent state")

	// ErrAdapterLoadFailed indicates the agent's WASM adapter failed to load.
	ErrAdapterLoadFailed = errors.New("adapter load failed")

	// ErrInvalidTransition indicates an invalid state transition was attempted.
	ErrInvalidTransition = errors.New("invalid state transition")
)
```

Common sentinel errors for agent operations.


## type AgentActor

```go
type AgentActor struct {
	// Agent is the underlying agent data.
	Agent *api.Agent

	// mu protects concurrent access to the actor.
	mu sync.RWMutex

	// eventHandlers contains callbacks for state transitions.
	eventHandlers map[string][]func(*api.Agent, interface{})
}
```

AgentActor wraps an Agent with simple state machines for lifecycle and presence.
This is a simplified implementation until full HSM integration is complete.

### Functions returning AgentActor

#### NewAgentActor

```go
func NewAgentActor(name, adapter, repoRoot string, config map[string]interface{}) (*AgentActor, error)
```

NewAgentActor creates a new agent actor with initialized state machines.


### Methods

#### AgentActor.GetAgent

```go
func () GetAgent() api.Agent
```

GetAgent returns a copy of the agent data safely.

#### AgentActor.GetPresence

```go
func () GetPresence() api.PresenceState
```

GetPresence returns the current agent presence safely.

#### AgentActor.GetState

```go
func () GetState() api.AgentState
```

GetState returns the current agent state safely.

#### AgentActor.OnEvent

```go
func () OnEvent(eventType string, handler func(*api.Agent, interface{}))
```

OnEvent registers an event handler for state changes.

#### AgentActor.SendLifecycleEvent

```go
func () SendLifecycleEvent(event LifecycleEvent) error
```

SendLifecycleEvent dispatches a lifecycle event to trigger state transitions.

#### AgentActor.SendPresenceEvent

```go
func () SendPresenceEvent(event PresenceEvent) error
```

SendPresenceEvent dispatches a presence event to trigger state transitions.

#### AgentActor.triggerEventHandlers

```go
func () triggerEventHandlers(eventType string, data interface{})
```

triggerEventHandlers invokes registered event handlers.


## type LifecycleEvent

```go
type LifecycleEvent string
```

LifecycleEvent represents events that can trigger agent lifecycle transitions.

### Constants

#### EventStart, EventStartupComplete, EventTerminate, EventError, EventRestart

```go
const (
	// EventStart triggers transition from Pending to Starting.
	EventStart LifecycleEvent = "start"

	// EventStartupComplete triggers transition from Starting to Running.
	EventStartupComplete LifecycleEvent = "startup_complete"

	// EventTerminate triggers transition to Terminated.
	EventTerminate LifecycleEvent = "terminate"

	// EventError triggers transition to Errored.
	EventError LifecycleEvent = "error"

	// EventRestart triggers transition from Terminated/Errored back to Pending.
	EventRestart LifecycleEvent = "restart"
)
```


## type Manager

```go
type Manager struct {
	// agents maps agent IDs to their actors.
	agents map[muid.MUID]*AgentActor

	// mu protects concurrent access to the manager.
	mu sync.RWMutex
}
```

Manager orchestrates multiple agents in an agent-agnostic manner.
It treats all agents uniformly through the adapter interface.

### Functions returning Manager

#### NewManager

```go
func NewManager() (*Manager, error)
```

NewManager creates a new agent manager instance.


### Methods

#### Manager.AddAgent

```go
func () AddAgent(name, adapter, repoRoot string, config map[string]interface{}) (*api.Agent, error)
```

AddAgent creates and adds a new agent to the manager.

#### Manager.GetAgent

```go
func () GetAgent(id muid.MUID) (*api.Agent, error)
```

GetAgent returns an agent by ID.

#### Manager.ListAgents

```go
func () ListAgents() []api.Agent
```

ListAgents returns all agents managed by this manager.

#### Manager.StartAgent

```go
func () StartAgent(id muid.MUID) error
```

StartAgent starts an agent by sending a start lifecycle event.

#### Manager.TerminateAgent

```go
func () TerminateAgent(id muid.MUID) error
```

TerminateAgent terminates an agent by sending a terminate lifecycle event.

#### Manager.UpdatePresence

```go
func () UpdatePresence(id muid.MUID, event PresenceEvent) error
```

UpdatePresence updates an agent's presence state.


## type PresenceEvent

```go
type PresenceEvent string
```

PresenceEvent represents events that can trigger presence transitions.

### Constants

#### EventGoOnline, EventGoBusy, EventGoOffline, EventGoAway, EventActivity

```go
const (
	// EventGoOnline triggers transition to Online.
	EventGoOnline PresenceEvent = "go_online"

	// EventGoBusy triggers transition to Busy.
	EventGoBusy PresenceEvent = "go_busy"

	// EventGoOffline triggers transition to Offline.
	EventGoOffline PresenceEvent = "go_offline"

	// EventGoAway triggers transition to Away.
	EventGoAway PresenceEvent = "go_away"

	// EventActivity triggers transition from Away back to Online.
	EventActivity PresenceEvent = "activity"
)
```


