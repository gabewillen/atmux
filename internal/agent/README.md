# package agent

`import "github.com/copilot-claude-sonnet-4/amux/internal/agent"`

Package agent provides agent-agnostic orchestration functionality.
This package manages agent lifecycle, presence, and messaging without
any knowledge of specific agent implementations.

All agent-specific behavior is delegated to WASM adapters loaded
via the adapter package.

- `ErrAgentNotFound, ErrInvalidState, ErrAdapterLoadFailed` — Common sentinel errors for agent operations.
- `type Manager` — Manager orchestrates multiple agents in an agent-agnostic manner.

### Variables

#### ErrAgentNotFound, ErrInvalidState, ErrAdapterLoadFailed

```go
var (
	// ErrAgentNotFound indicates an agent with the given ID was not found.
	ErrAgentNotFound = errors.New("agent not found")

	// ErrInvalidState indicates an operation cannot be performed in the current agent state.
	ErrInvalidState = errors.New("invalid agent state")

	// ErrAdapterLoadFailed indicates the agent's WASM adapter failed to load.
	ErrAdapterLoadFailed = errors.New("adapter load failed")
)
```

Common sentinel errors for agent operations.


## type Manager

```go
type Manager struct {
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

#### Manager.Start

```go
func () Start(adapterName string, config map[string]interface{}) error
```

Start initiates an agent by its adapter name and configuration.


