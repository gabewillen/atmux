# package agent

`import "github.com/copilot-claude-sonnet-4/amux/internal/agent"`

Package agent implements HSM-based agent lifecycle and presence management.
This implementation uses stateforward/hsm-go for proper hierarchical state machines
as required by the amux specification.

Package agent provides agent-agnostic orchestration functionality.
This package manages agent lifecycle, presence, and messaging without
any knowledge of specific agent implementations.

All agent-specific behavior is delegated to WASM adapters loaded
via the adapter package.

This package implements HSM-based state management per spec requirements.

- `ErrAgentNotFound, ErrInvalidState, ErrAdapterLoadFailed, ErrInvalidTransition` — Common sentinel errors for agent operations.
- `EvtGoOnline, EvtGoBusy, EvtGoOffline, EvtGoAway, EvtActivity` — PresenceEvents define the events that can trigger presence transitions.
- `EvtStart, EvtStartupComplete, EvtTerminate, EvtError, EvtRestart` — LifecycleEvents define the events that can trigger agent lifecycle transitions.
- `type AgentActor` — AgentActor wraps an Agent with simple state machines for lifecycle and presence.
- `type AgentHSMActor` — AgentHSMActor wraps an Agent with proper HSM-based state machines for lifecycle and presence.
- `type AgentHSM` — AgentHSM represents an agent with HSM-based state management.
- `type LifecycleEvent` — LifecycleEvent represents events that can trigger agent lifecycle transitions.
- `type Manager` — Manager orchestrates multiple agents in an agent-agnostic manner.
- `type PresenceEvent` — PresenceEvent represents events that can trigger presence transitions.

### Constants

#### EvtStart, EvtStartupComplete, EvtTerminate, EvtError, EvtRestart

```go
const (
	// EvtStart triggers transition from Pending to Starting.
	EvtStart = "start"
	// EvtStartupComplete triggers transition from Starting to Running.
	EvtStartupComplete = "startup_complete"
	// EvtTerminate triggers transition to Terminated.
	EvtTerminate = "terminate"
	// EvtError triggers transition to Errored.
	EvtError = "error"
	// EvtRestart triggers transition from Terminated/Errored back to Pending.
	EvtRestart = "restart"
)
```

LifecycleEvents define the events that can trigger agent lifecycle transitions.

#### EvtGoOnline, EvtGoBusy, EvtGoOffline, EvtGoAway, EvtActivity

```go
const (
	// EvtGoOnline triggers transition to Online.
	EvtGoOnline = "go_online"
	// EvtGoBusy triggers transition to Busy.
	EvtGoBusy = "go_busy"
	// EvtGoOffline triggers transition to Offline.
	EvtGoOffline = "go_offline"
	// EvtGoAway triggers transition to Away.
	EvtGoAway = "go_away"
	// EvtActivity triggers transition from Away back to Online.
	EvtActivity = "activity"
)
```

PresenceEvents define the events that can trigger presence transitions.


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


## type AgentHSM

```go
type AgentHSM struct {
	hsm.HSM

	// Agent is the underlying agent data.
	Agent *api.Agent

	// mu protects concurrent access to the agent.
	mu sync.RWMutex
}
```

AgentHSM represents an agent with HSM-based state management.

## type AgentHSMActor

```go
type AgentHSMActor struct {
	// lifecycleHSM manages agent lifecycle state transitions.
	lifecycleHSM *AgentHSM

	// presenceHSM manages agent presence state transitions.
	presenceHSM *AgentHSM

	// ctx is the context for HSM operations.
	ctx context.Context

	// cancel cancels the HSM context.
	cancel context.CancelFunc
}
```

AgentHSMActor wraps an Agent with proper HSM-based state machines for lifecycle and presence.
This replaces the simple state machine implementation per spec requirements.

### Functions returning AgentHSMActor

#### NewAgentHSMActor

```go
func NewAgentHSMActor(name, adapter, repoRoot string, config map[string]interface{}) (*AgentHSMActor, error)
```

NewAgentHSMActor creates a new HSM-based agent actor with proper state machines.


### Methods

#### AgentHSMActor.Close

```go
func () Close() error
```

Close gracefully shuts down the HSM actor.

#### AgentHSMActor.Dispatch

```go
func () Dispatch(eventName string, data interface{}) error
```

Dispatch sends an event to the appropriate HSM using hsm.Dispatch() as required by spec.

#### AgentHSMActor.GetAgent

```go
func () GetAgent() api.Agent
```

GetAgent returns a copy of the agent data safely.

#### AgentHSMActor.GetPresence

```go
func () GetPresence() api.PresenceState
```

GetPresence returns the current agent presence safely.

#### AgentHSMActor.GetState

```go
func () GetState() api.AgentState
```

GetState returns the current agent state safely.

#### AgentHSMActor.initLifecycleHSM

```go
func () initLifecycleHSM(agent *api.Agent) (*AgentHSM, error)
```

initLifecycleHSM initializes the agent lifecycle state machine.
Per spec: Pending → Starting → Running → Terminated/Errored

#### AgentHSMActor.initPresenceHSM

```go
func () initPresenceHSM(agent *api.Agent) (*AgentHSM, error)
```

initPresenceHSM initializes the agent presence state machine.
Per spec: Online ↔ Busy ↔ Offline ↔ Away


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
	// agents maps agent IDs to their HSM actors.
	agents map[muid.MUID]*AgentHSMActor

	// legacyAgents maps agent IDs to legacy actors (for backward compatibility during transition).
	legacyAgents map[muid.MUID]*AgentActor

	// useHSM determines whether to use HSM actors (true) or legacy actors (false).
	useHSM bool

	// mu protects concurrent access to the manager.
	mu sync.RWMutex
}
```

Manager orchestrates multiple agents in an agent-agnostic manner.
It treats all agents uniformly through the adapter interface.
Updated to use HSM-based actors per spec requirements.

### Functions returning Manager

#### NewLegacyManager

```go
func NewLegacyManager() (*Manager, error)
```

NewLegacyManager creates a manager that uses legacy simple state machines.
This is provided for backward compatibility during transition.

#### NewManager

```go
func NewManager() (*Manager, error)
```

NewManager creates a new agent manager instance.
By default, uses HSM-based actors per spec requirements.


### Methods

#### Manager.AddAgent

```go
func () AddAgent(name, adapter, repoRoot string, config map[string]interface{}) (*api.Agent, error)
```

AddAgent creates and adds a new agent to the manager.

#### Manager.DispatchEvent

```go
func () DispatchEvent(id muid.MUID, eventName string, data interface{}) error
```

DispatchEvent dispatches an event to an agent using hsm.Dispatch() per spec requirements.
This is the primary method for triggering state transitions.

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
Uses hsm.Dispatch() for HSM actors per spec requirements.

#### Manager.TerminateAgent

```go
func () TerminateAgent(id muid.MUID) error
```

TerminateAgent terminates an agent by sending a terminate lifecycle event.
Uses hsm.Dispatch() for HSM actors per spec requirements.

#### Manager.UpdatePresence

```go
func () UpdatePresence(id muid.MUID, eventName string) error
```

UpdatePresence updates an agent's presence state.
Uses hsm.Dispatch() for HSM actors per spec requirements.


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


