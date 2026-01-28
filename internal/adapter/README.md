# package adapter

`import "github.com/agentflare-ai/amux/internal/adapter"`

Package adapter provides the WASM adapter runtime for amux.

Adapters are WebAssembly modules that implement agent-specific behavior.
The core amux system loads adapters by name and interacts with them
through a standardized WASM interface.

This package is agent-agnostic; all agent-specific code resides in the
adapter WASM modules themselves.

See spec §10 for the full adapter interface specification.

- `ErrAdapterNotFound, ErrAdapterAlreadyExists, ErrAdapterLoadFailed, ErrAdapterCallFailed` — Adapter errors
- `type Adapter` — Adapter is the interface for WASM adapters.
- `type CLIConstraint` — CLIConstraint defines version constraints for the agent CLI.
- `type Manifest` — Manifest represents an adapter manifest.
- `type NoopAdapter` — NoopAdapter is a no-op adapter for testing.
- `type NoopPatternMatcher` — NoopPatternMatcher is a no-op pattern matcher.
- `type OutputEvent` — OutputEvent represents an event detected from PTY output.
- `type PatternMatch` — PatternMatch represents a pattern match result.
- `type PatternMatcher` — PatternMatcher is the interface for pattern matching.
- `type Registry` — Registry manages adapter discovery and loading.

### Variables

#### ErrAdapterNotFound, ErrAdapterAlreadyExists, ErrAdapterLoadFailed, ErrAdapterCallFailed

```go
var (
	// ErrAdapterNotFound indicates the adapter was not found.
	ErrAdapterNotFound = fmt.Errorf("not found")

	// ErrAdapterAlreadyExists indicates the adapter already exists.
	ErrAdapterAlreadyExists = fmt.Errorf("already exists")

	// ErrAdapterLoadFailed indicates the adapter failed to load.
	ErrAdapterLoadFailed = fmt.Errorf("load failed")

	// ErrAdapterCallFailed indicates an adapter call failed.
	ErrAdapterCallFailed = fmt.Errorf("call failed")
)
```

Adapter errors


## type Adapter

```go
type Adapter interface {
	// Name returns the adapter name.
	Name() string

	// Manifest returns the adapter manifest.
	Manifest() (*Manifest, error)

	// OnOutput processes PTY output and returns detected events.
	OnOutput(ctx context.Context, output []byte) ([]OutputEvent, error)

	// FormatInput formats input for the agent.
	FormatInput(ctx context.Context, input string) (string, error)

	// OnEvent handles an incoming event.
	OnEvent(ctx context.Context, event []byte) error

	// Close releases adapter resources.
	Close() error
}
```

Adapter is the interface for WASM adapters.

## type CLIConstraint

```go
type CLIConstraint struct {
	// Constraint is a semver constraint string.
	Constraint string `json:"constraint"`
}
```

CLIConstraint defines version constraints for the agent CLI.

## type Manifest

```go
type Manifest struct {
	// Name is the adapter name.
	Name string `json:"name"`

	// Version is the adapter version (semver).
	Version string `json:"version"`

	// CLI contains CLI version constraints.
	CLI CLIConstraint `json:"cli"`

	// Patterns contains pattern definitions.
	Patterns map[string]string `json:"patterns,omitempty"`
}
```

Manifest represents an adapter manifest.

## type NoopAdapter

```go
type NoopAdapter struct {
	name string
}
```

NoopAdapter is a no-op adapter for testing.

### Functions returning NoopAdapter

#### NewNoopAdapter

```go
func NewNoopAdapter(name string) *NoopAdapter
```

NewNoopAdapter creates a new no-op adapter.


### Methods

#### NoopAdapter.Close

```go
func () Close() error
```

Close is a no-op.

#### NoopAdapter.FormatInput

```go
func () FormatInput(ctx context.Context, input string) (string, error)
```

FormatInput returns the input unchanged.

#### NoopAdapter.Manifest

```go
func () Manifest() (*Manifest, error)
```

Manifest returns a minimal manifest.

#### NoopAdapter.Name

```go
func () Name() string
```

Name returns the adapter name.

#### NoopAdapter.OnEvent

```go
func () OnEvent(ctx context.Context, event []byte) error
```

OnEvent is a no-op.

#### NoopAdapter.OnOutput

```go
func () OnOutput(ctx context.Context, output []byte) ([]OutputEvent, error)
```

OnOutput returns no events.


## type NoopPatternMatcher

```go
type NoopPatternMatcher struct{}
```

NoopPatternMatcher is a no-op pattern matcher.

### Functions returning NoopPatternMatcher

#### NewNoopPatternMatcher

```go
func NewNoopPatternMatcher() *NoopPatternMatcher
```

NewNoopPatternMatcher creates a new no-op pattern matcher.


### Methods

#### NoopPatternMatcher.Match

```go
func () Match(output []byte) []PatternMatch
```

Match returns no matches.


## type OutputEvent

```go
type OutputEvent struct {
	// Type is the event type.
	Type string `json:"type"`

	// Data is the event-specific data.
	Data any `json:"data,omitempty"`
}
```

OutputEvent represents an event detected from PTY output.

## type PatternMatch

```go
type PatternMatch struct {
	// Pattern is the matched pattern name.
	Pattern string

	// Match is the matched text.
	Match string

	// Index is the byte offset of the match.
	Index int
}
```

PatternMatch represents a pattern match result.

## type PatternMatcher

```go
type PatternMatcher interface {
	// Match checks if output matches any configured patterns.
	Match(output []byte) []PatternMatch
}
```

PatternMatcher is the interface for pattern matching.
During Phase 0, this uses a noop implementation.
Phase 8 will provide the full WASM-backed implementation.

## type Registry

```go
type Registry struct {
	mu       sync.RWMutex
	adapters map[string]Adapter
	runtime  wazero.Runtime
}
```

Registry manages adapter discovery and loading.

### Functions returning Registry

#### NewRegistry

```go
func NewRegistry(ctx context.Context) (*Registry, error)
```

NewRegistry creates a new adapter registry.


### Methods

#### Registry.Close

```go
func () Close(ctx context.Context) error
```

Close closes the registry and all adapters.

#### Registry.List

```go
func () List() []string
```

List returns the names of all registered adapters.

#### Registry.Load

```go
func () Load(name string) (Adapter, error)
```

Load loads an adapter by name.
Returns ErrAdapterNotFound if the adapter is not registered.

#### Registry.Register

```go
func () Register(adapter Adapter) error
```

Register registers an adapter.

#### Registry.Unregister

```go
func () Unregister(name string) error
```

Unregister removes an adapter from the registry.


