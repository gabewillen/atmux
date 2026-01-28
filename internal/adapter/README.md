# package adapter

`import "github.com/stateforward/amux/internal/adapter"`

Package adapter provides the WASM adapter runtime interface per spec §10.

Phase 0: Provides stable interfaces with noop implementations.
Phase 8 will add full WASM runtime with wazero.

- `type Action` — Action represents an action to be taken by the core.
- `type Adapter` — Adapter is the interface for WASM adapters.
- `type Pattern` — Pattern represents a matched pattern in PTY output.
- `type Runtime` — Runtime manages adapter loading and lifecycle.
- `type stubAdapter` — stubAdapter is a placeholder adapter for Phase 0.
- `type stubRuntime` — stubRuntime is a placeholder implementation for Phase 0.

## type Action

```go
type Action struct {
	Type string // Action type (e.g., "send_input", "notify")
	Data any    // Action-specific data
}
```

Action represents an action to be taken by the core.

## type Adapter

```go
type Adapter interface {
	// Name returns the adapter name.
	Name() string

	// OnOutput processes PTY output and returns pattern matches.
	OnOutput(ctx context.Context, output []byte) ([]Pattern, error)

	// FormatInput formats input for the agent.
	FormatInput(ctx context.Context, input string) (string, error)

	// OnEvent processes an event and returns actions.
	OnEvent(ctx context.Context, event any) ([]Action, error)

	// Close releases adapter resources.
	Close() error
}
```

Adapter is the interface for WASM adapters.

## type Pattern

```go
type Pattern struct {
	Name    string // Pattern name (e.g., "prompt", "error")
	Matched string // Matched text
}
```

Pattern represents a matched pattern in PTY output.

## type Runtime

```go
type Runtime interface {
	// LoadAdapter loads an adapter by name.
	LoadAdapter(ctx context.Context, name string) (Adapter, error)

	// Close releases all adapter resources.
	Close() error
}
```

Runtime manages adapter loading and lifecycle.

### Functions returning Runtime

#### NewRuntime

```go
func NewRuntime() Runtime
```

NewRuntime creates a new adapter runtime.
Phase 0: Returns a stub that will be implemented with wazero in Phase 8.


## type stubAdapter

```go
type stubAdapter struct {
	name string
}
```

stubAdapter is a placeholder adapter for Phase 0.

### Methods

#### stubAdapter.Close

```go
func () Close() error
```

#### stubAdapter.FormatInput

```go
func () FormatInput(ctx context.Context, input string) (string, error)
```

#### stubAdapter.Name

```go
func () Name() string
```

#### stubAdapter.OnEvent

```go
func () OnEvent(ctx context.Context, event any) ([]Action, error)
```

#### stubAdapter.OnOutput

```go
func () OnOutput(ctx context.Context, output []byte) ([]Pattern, error)
```


## type stubRuntime

```go
type stubRuntime struct{}
```

stubRuntime is a placeholder implementation for Phase 0.

### Methods

#### stubRuntime.Close

```go
func () Close() error
```

#### stubRuntime.LoadAdapter

```go
func () LoadAdapter(ctx context.Context, name string) (Adapter, error)
```


