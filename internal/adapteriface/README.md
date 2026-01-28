# package adapteriface

`import "github.com/stateforward/amux/internal/adapteriface"`

Package adapteriface implements a stable interface for adapter-provided pattern matching and actions
that can be used by other packages during Phase 0 before the full WASM implementation is complete in Phase 8.

- `func ExecuteAction(ctx context.Context, action Action) error` — ExecuteAction is a convenience function to execute an action using the global interface
- `type Action` — Action represents an action to be taken based on a match
- `type Interface` — Interface defines the interface for adapter-provided pattern matching and actions
- `type Manifest` — Manifest describes the adapter's capabilities
- `type Match` — Match represents a pattern match result
- `type NoopInterface` — NoopInterface is a no-op implementation of the Interface

### Functions

#### ExecuteAction

```go
func ExecuteAction(ctx context.Context, action Action) error
```

ExecuteAction is a convenience function to execute an action using the global interface


## type Action

```go
type Action struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}
```

Action represents an action to be taken based on a match

## type Interface

```go
type Interface interface {
	// MatchPatterns attempts to match patterns against the provided input
	MatchPatterns(ctx context.Context, input string) ([]Match, error)

	// ExecuteAction executes an action returned by MatchPatterns
	ExecuteAction(ctx context.Context, action Action) error

	// GetManifest returns the adapter manifest
	GetManifest() Manifest
}
```

Interface defines the interface for adapter-provided pattern matching and actions

### Variables

#### GlobalInterface

```go
var GlobalInterface Interface = NewNoopInterface(Manifest{
	Name:        "noop-adapter",
	Version:     "v0.0.0",
	Description: "No-op adapter for Phase 0",
	Patterns:    []string{},
	Actions:     []string{},
})
```

GlobalInterface is a global instance of the adapter interface that can be used by other packages


## type Manifest

```go
type Manifest struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Patterns    []string `json:"patterns"`
	Actions     []string `json:"actions"`
}
```

Manifest describes the adapter's capabilities

### Functions returning Manifest

#### GetAdapterManifest

```go
func GetAdapterManifest() Manifest
```

GetAdapterManifest is a convenience function to get the adapter manifest


## type Match

```go
type Match struct {
	PatternID string                 `json:"pattern_id"`
	Action    string                 `json:"action"`
	Data      map[string]interface{} `json:"data"`
	Score     float64                `json:"score"` // Confidence score between 0 and 1
}
```

Match represents a pattern match result

### Functions returning Match

#### MatchPatterns

```go
func MatchPatterns(ctx context.Context, input string) ([]Match, error)
```

MatchPatterns is a convenience function to match patterns using the global interface


## type NoopInterface

```go
type NoopInterface struct {
	manifest Manifest
}
```

NoopInterface is a no-op implementation of the Interface

### Functions returning NoopInterface

#### NewNoopInterface

```go
func NewNoopInterface(manifest Manifest) *NoopInterface
```

NewNoopInterface creates a new no-op adapter interface


### Methods

#### NoopInterface.ExecuteAction

```go
func () ExecuteAction(ctx context.Context, action Action) error
```

ExecuteAction implements the Interface

#### NoopInterface.GetManifest

```go
func () GetManifest() Manifest
```

GetManifest implements the Interface

#### NoopInterface.MatchPatterns

```go
func () MatchPatterns(ctx context.Context, input string) ([]Match, error)
```

MatchPatterns implements the Interface


