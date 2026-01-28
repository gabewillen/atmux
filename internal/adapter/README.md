# package adapter

`import "github.com/agentflare-ai/amux/internal/adapter"`

Package adapter defines the WASM adapter runtime interface.

The core loads adapters by string name via the WASM registry.

- `ErrAdapterNotFound` — ErrAdapterNotFound is returned when a named adapter cannot be loaded.
- `type ActionFormatter` — ActionFormatter converts a high-level action into agent input.
- `type Adapter` — Adapter is the runtime-facing interface to a loaded adapter.
- `type NoopAdapter` — NoopAdapter returns no matches and echoes input.
- `type NoopFormatter` — NoopFormatter returns the input unchanged.
- `type NoopMatcher` — NoopMatcher returns no matches.
- `type PatternMatch` — PatternMatch describes a detected pattern match.
- `type PatternMatcher` — PatternMatcher scans output and returns matches.
- `type Registry` — Registry loads adapters by name.

### Variables

#### ErrAdapterNotFound

```go
var ErrAdapterNotFound = errors.New("adapter not found")
```

ErrAdapterNotFound is returned when a named adapter cannot be loaded.


## type ActionFormatter

```go
type ActionFormatter interface {
	Format(ctx context.Context, input string) (string, error)
}
```

ActionFormatter converts a high-level action into agent input.

## type Adapter

```go
type Adapter interface {
	Name() string
	Matcher() PatternMatcher
	Formatter() ActionFormatter
}
```

Adapter is the runtime-facing interface to a loaded adapter.

## type NoopAdapter

```go
type NoopAdapter struct {
	name string
}
```

NoopAdapter returns no matches and echoes input.

### Functions returning NoopAdapter

#### NewNoopAdapter

```go
func NewNoopAdapter(name string) *NoopAdapter
```

NewNoopAdapter constructs a noop adapter.


### Methods

#### NoopAdapter.Formatter

```go
func () Formatter() ActionFormatter
```

Formatter returns a noop formatter.

#### NoopAdapter.Matcher

```go
func () Matcher() PatternMatcher
```

Matcher returns a noop matcher.

#### NoopAdapter.Name

```go
func () Name() string
```

Name returns the adapter name.


## type NoopFormatter

```go
type NoopFormatter struct{}
```

NoopFormatter returns the input unchanged.

### Methods

#### NoopFormatter.Format

```go
func () Format(ctx context.Context, input string) (string, error)
```

Format returns the input unchanged.


## type NoopMatcher

```go
type NoopMatcher struct{}
```

NoopMatcher returns no matches.

### Methods

#### NoopMatcher.Match

```go
func () Match(ctx context.Context, output []byte) ([]PatternMatch, error)
```

Match returns no matches.


## type PatternMatch

```go
type PatternMatch struct {
	Pattern string
	Text    string
}
```

PatternMatch describes a detected pattern match.

## type PatternMatcher

```go
type PatternMatcher interface {
	Match(ctx context.Context, output []byte) ([]PatternMatch, error)
}
```

PatternMatcher scans output and returns matches.

## type Registry

```go
type Registry interface {
	Load(ctx context.Context, name string) (Adapter, error)
}
```

Registry loads adapters by name.

