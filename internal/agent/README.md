# package agent

`import "github.com/agentflare-ai/amux/internal/agent"`

Package agent provides agent orchestration for amux.

This package implements agent lifecycle management, presence tracking,
and messaging. All operations are agent-agnostic; agent-specific behavior
is delegated to adapters.

See spec §5 for agent management requirements.

- `type Agent` — Agent represents a managed agent instance.
- `type Manager` — Manager manages agents.

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


## type Manager

```go
type Manager struct {
	mu         sync.RWMutex
	agents     map[muid.MUID]*Agent
	dispatcher event.Dispatcher
}
```

Manager manages agents.

### Functions returning Manager

#### NewManager

```go
func NewManager(dispatcher event.Dispatcher) *Manager
```

NewManager creates a new agent manager.


### Methods

#### Manager.Add

```go
func () Add(ctx context.Context, cfg api.Agent) (*Agent, error)
```

Add adds an agent.

#### Manager.Get

```go
func () Get(id muid.MUID) *Agent
```

Get returns an agent by ID.

#### Manager.List

```go
func () List() []*Agent
```

List returns all agents.

#### Manager.Remove

```go
func () Remove(ctx context.Context, id muid.MUID) error
```

Remove removes an agent.

#### Manager.Roster

```go
func () Roster() []api.RosterEntry
```

Roster returns the roster entries for all agents.


