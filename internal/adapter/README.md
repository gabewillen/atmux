# package adapter

`import "github.com/agentflare-ai/amux/internal/adapter"`

Package adapter provides the adapter interface for pattern matching and actions.
Phase 0 introduces stable interfaces with noop implementations.
Phase 8 will provide the full WASM-backed runtime.

- `type Adapter` — Adapter provides pattern matching and action capabilities.
- `type Match` — Match represents a pattern match result.
- `type Registry` — Registry manages adapter instances.
- `type noopAdapter` — noopAdapter is a Phase 0 noop adapter implementation.
- `type noopRegistry` — noopRegistry is a Phase 0 noop adapter registry.

## type Adapter

```go
type Adapter interface {
	// MatchPatterns checks PTY output for matching patterns.
	MatchPatterns(ctx context.Context, output []byte) ([]Match, error)

	// FormatInput formats input for the agent CLI.
	FormatInput(ctx context.Context, input string) ([]byte, error)

	// OnEvent handles events from the system.
	OnEvent(ctx context.Context, event interface{}) error
}
```

Adapter provides pattern matching and action capabilities.
Phase 0: Noop implementation that returns no matches
Phase 8: WASM-backed implementation

## type Match

```go
type Match struct {
	Pattern string
	Data    interface{}
}
```

Match represents a pattern match result.

## type Registry

```go
type Registry interface {
	// Load loads an adapter by name.
	Load(name string) (Adapter, error)
}
```

Registry manages adapter instances.

### Functions returning Registry

#### NewRegistry

```go
func NewRegistry() Registry
```

NewRegistry creates a new adapter registry.
Phase 0: Returns a noop registry


## type noopAdapter

```go
type noopAdapter struct{}
```

noopAdapter is a Phase 0 noop adapter implementation.

### Methods

#### noopAdapter.FormatInput

```go
func () FormatInput(ctx context.Context, input string) ([]byte, error)
```

#### noopAdapter.MatchPatterns

```go
func () MatchPatterns(ctx context.Context, output []byte) ([]Match, error)
```

#### noopAdapter.OnEvent

```go
func () OnEvent(ctx context.Context, event interface{}) error
```


## type noopRegistry

```go
type noopRegistry struct{}
```

noopRegistry is a Phase 0 noop adapter registry.

### Methods

#### noopRegistry.Load

```go
func () Load(name string) (Adapter, error)
```


