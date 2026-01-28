# package inference

`import "github.com/copilot-claude-sonnet-4/amux/internal/inference"`

Package inference provides local inference engine (liquidgen) functionality.
This package supports the required models: lfm2.5-thinking and lfm2.5-VL
for local model inference features.

- `ErrModelNotFound, ErrInferenceError, ErrEngineNotAvailable` — Common sentinel errors for inference operations.
- `type Engine` — Engine provides liquidgen inference capabilities.
- `type ModelID` — ModelID represents a logical model identifier as required by the spec.

### Variables

#### ErrModelNotFound, ErrInferenceError, ErrEngineNotAvailable

```go
var (
	// ErrModelNotFound indicates the requested model was not found.
	ErrModelNotFound = errors.New("model not found")

	// ErrInferenceError indicates an inference execution error.
	ErrInferenceError = errors.New("inference error")

	// ErrEngineNotAvailable indicates the liquidgen engine is not available.
	ErrEngineNotAvailable = errors.New("liquidgen engine not available")
)
```

Common sentinel errors for inference operations.


## type Engine

```go
type Engine struct {
	available bool
}
```

Engine provides liquidgen inference capabilities.
Implementation deferred to Phase 0 completion and later phases.

### Functions returning Engine

#### NewEngine

```go
func NewEngine() (*Engine, error)
```

NewEngine creates a new liquidgen inference engine.


### Methods

#### Engine.Infer

```go
func () Infer(modelID ModelID, input string) (string, error)
```

Infer executes inference with the specified model and input.

#### Engine.IsAvailable

```go
func () IsAvailable() bool
```

IsAvailable returns whether the liquidgen engine is available.


## type ModelID

```go
type ModelID string
```

ModelID represents a logical model identifier as required by the spec.

### Constants

#### ModelLFM25Thinking, ModelLFM25VL

```go
const (
	// ModelLFM25Thinking represents the lfm2.5-thinking model for text-only reasoning.
	ModelLFM25Thinking ModelID = "lfm2.5-thinking"

	// ModelLFM25VL represents the lfm2.5-VL model for vision-language tasks.
	ModelLFM25VL ModelID = "lfm2.5-VL"
)
```


