# package adapter

`import "github.com/agentflare-ai/amux/internal/adapter"`

Package adapter provides stable interfaces for adapters (to be fully implemented in Phase 8).

- `type Action` — Action represents an action that adapters can perform.
- `type ManagerInfo` — ManagerInfo provides information about the adapter manager.
- `type Manager` — Manager manages adapter lifecycles and interactions.
- `type Match` — Match represents a successful pattern match.
- `type NoopManager` — NoopManager provides a no-op implementation for Phase 0.
- `type PTYSnapshot` — PTYSnapshot represents a snapshot of PTY output.
- `type Pattern` — Pattern represents a pattern that adapters can match against.
- `type ProcessSnapshot` — ProcessSnapshot represents a snapshot of process information.
- `type RuntimeInfo` — RuntimeInfo provides information about the adapter runtime.
- `type Runtime` — Runtime provides adapter pattern matching and action execution.

## type Action

```go
type Action struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"` // "input", "command", "message", "screenshot"
	Parameters map[string]interface{} `json:"parameters"`
	Priority   int                    `json:"priority,omitempty"` // Higher = more important
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}
```

Action represents an action that adapters can perform.

## type Manager

```go
type Manager interface {
	// Load and initialize an adapter
	Load(ctx context.Context, adapterName string, config map[string]interface{}) error

	// Unload an adapter
	Unload(ctx context.Context, adapterName string) error

	// Get all loaded adapters
	GetAdapters() map[string]Runtime

	// Match patterns using all loaded adapters
	Match(ctx context.Context, content string) ([]Match, error)

	// Execute an action using appropriate adapter
	Execute(ctx context.Context, action *Action) error

	// Process event and generate potential actions
	ProcessEvent(ctx context.Context, ev *event.Event) ([]*Action, error)

	// Get manager information
	Info() *ManagerInfo

	// Shutdown all adapters
	Shutdown(ctx context.Context) error
}
```

Manager manages adapter lifecycles and interactions.

## type ManagerInfo

```go
type ManagerInfo struct {
	LoadedAdapters int      `json:"loaded_adapters"`
	ActiveMatchers int      `json:"active_matchers"`
	ActiveActions  int      `json:"active_actions"`
	Capabilities   []string `json:"capabilities"`
}
```

ManagerInfo provides information about the adapter manager.

## type Match

```go
type Match struct {
	PatternID  string                 `json:"pattern_id"`
	Offset     int                    `json:"offset"`
	Length     int                    `json:"length"`
	Groups     map[string]string      `json:"groups,omitempty"`
	Confidence float64                `json:"confidence,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}
```

Match represents a successful pattern match.

## type NoopManager

```go
type NoopManager struct {
	adapters map[string]Runtime
	config   map[string]map[string]interface{}
}
```

NoopManager provides a no-op implementation for Phase 0.

### Functions returning NoopManager

#### NewNoopManager

```go
func NewNoopManager() *NoopManager
```

NewNoopManager creates a new no-op adapter manager.


### Methods

#### NoopManager.Execute

```go
func () Execute(ctx context.Context, action *Action) error
```

Execute implements Manager interface (no-op).

#### NoopManager.GetAdapters

```go
func () GetAdapters() map[string]Runtime
```

GetAdapters implements Manager interface.

#### NoopManager.Info

```go
func () Info() *ManagerInfo
```

Info implements Manager interface.

#### NoopManager.Load

```go
func () Load(ctx context.Context, adapterName string, config map[string]interface{}) error
```

Load implements Manager interface.

#### NoopManager.Match

```go
func () Match(ctx context.Context, content string) ([]Match, error)
```

Match implements Manager interface (no-op).

#### NoopManager.ProcessEvent

```go
func () ProcessEvent(ctx context.Context, ev *event.Event) ([]*Action, error)
```

ProcessEvent implements Manager interface (no-op).

#### NoopManager.Shutdown

```go
func () Shutdown(ctx context.Context) error
```

Shutdown implements Manager interface.

#### NoopManager.Unload

```go
func () Unload(ctx context.Context, adapterName string) error
```

Unload implements Manager interface.


## type PTYSnapshot

```go
type PTYSnapshot struct {
	Content   string `json:"content"`
	Length    int    `json:"length"`
	Timestamp int64  `json:"timestamp"` // Unix timestamp
}
```

PTYSnapshot represents a snapshot of PTY output.

## type Pattern

```go
type Pattern struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "regex", "glob", "semantic"
	Pattern     string                 `json:"pattern"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config,omitempty"`
}
```

Pattern represents a pattern that adapters can match against.

## type ProcessSnapshot

```go
type ProcessSnapshot struct {
	PID        int                    `json:"pid"`
	ParentPID  int                    `json:"parent_pid"`
	Command    string                 `json:"command"`
	Args       []string               `json:"args"`
	Env        map[string]string      `json:"env"`
	WorkingDir string                 `json:"working_dir"`
	Status     string                 `json:"status"` // "running", "stopped", "zombie"
	Children   []*ProcessSnapshot     `json:"children,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}
```

ProcessSnapshot represents a snapshot of process information.

## type Runtime

```go
type Runtime interface {
	// Initialize the adapter runtime with configuration
	Initialize(ctx context.Context, config map[string]interface{}) error

	// Match patterns against provided content
	Match(ctx context.Context, patterns []Pattern, content string) ([]Match, error)

	// Execute an action
	Execute(ctx context.Context, action *Action) error

	// Get adapter information
	Info() *RuntimeInfo

	// Shutdown the adapter runtime
	Shutdown(ctx context.Context) error
}
```

Runtime provides adapter pattern matching and action execution.

## type RuntimeInfo

```go
type RuntimeInfo struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Capabilities []string               `json:"capabilities"` // e.g., "regex", "semantic", "input", "command"
	Config       map[string]interface{} `json:"config,omitempty"`
}
```

RuntimeInfo provides information about the adapter runtime.

