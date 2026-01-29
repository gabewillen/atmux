# package inference

`import "github.com/copilot-claude-sonnet-4/amux/internal/inference"`

Package inference provides local inference engine (liquidgen) functionality.
This package supports the required models: lfm2.5-thinking and lfm2.5-VL
for local model inference features per spec §4.2.10.

This package integrates with the pre-existing liquidgen inference engine
from third_party/liquidgen (local) as a dependency and wires it to the
local inference interface defined in the specification.

- `ErrModelNotFound, ErrInferenceError, ErrEngineNotAvailable, ErrModelLoadFailed` — Common sentinel errors for inference operations.
- `func fileExists(path string) bool` — fileExists checks if a file exists and is readable.
- `func findLiquidgenPath() string` — findLiquidgenPath attempts to locate the liquidgen executable or library.
- `func findModelArtifact(modelName string) string` — findModelArtifact attempts to locate a model artifact file.
- `func getLiquidgenVersion(liquidgenPath string) string` — getLiquidgenVersion attempts to get version info from liquidgen.
- `type Engine` — Engine provides liquidgen inference capabilities.
- `type LiquidgenEngine` — LiquidgenEngine interface as defined in spec §4.2.10.
- `type LiquidgenRequest` — LiquidgenRequest represents an inference request as defined in spec §4.2.10.
- `type LiquidgenStream` — LiquidgenStream interface for streaming inference results.
- `type ModelID` — ModelID represents a logical model identifier as required by the spec.
- `type liquidgenStream` — liquidgenStream implements LiquidgenStream for streaming inference.

### Variables

#### ErrModelNotFound, ErrInferenceError, ErrEngineNotAvailable, ErrModelLoadFailed

```go
var (
	// ErrModelNotFound indicates the requested model was not found.
	ErrModelNotFound = errors.New("model not found")

	// ErrInferenceError indicates an inference execution error.
	ErrInferenceError = errors.New("inference error")

	// ErrEngineNotAvailable indicates the liquidgen engine is not available.
	ErrEngineNotAvailable = errors.New("liquidgen engine not available")

	// ErrModelLoadFailed indicates a model failed to load.
	ErrModelLoadFailed = errors.New("model load failed")
)
```

Common sentinel errors for inference operations.


### Functions

#### fileExists

```go
func fileExists(path string) bool
```

fileExists checks if a file exists and is readable.

#### findLiquidgenPath

```go
func findLiquidgenPath() string
```

findLiquidgenPath attempts to locate the liquidgen executable or library.

#### findModelArtifact

```go
func findModelArtifact(modelName string) string
```

findModelArtifact attempts to locate a model artifact file.

#### getLiquidgenVersion

```go
func getLiquidgenVersion(liquidgenPath string) string
```

getLiquidgenVersion attempts to get version info from liquidgen.


## type Engine

```go
type Engine struct {
	available     bool
	liquidgenPath string
	modelMappings map[ModelID]string // logical ID → concrete artifact path
	version       string
}
```

Engine provides liquidgen inference capabilities.
This implementation connects to the C++ liquidgen library in third_party/liquidgen.

### Functions returning Engine

#### NewEngine

```go
func NewEngine() (*Engine, error)
```

NewEngine creates a new liquidgen inference engine.
This function integrates the pre-existing liquidgen library per spec requirements.


### Methods

#### Engine.Generate

```go
func () Generate(ctx context.Context, req LiquidgenRequest) (LiquidgenStream, error)
```

Generate executes inference with the specified model and request.

#### Engine.GetModelMapping

```go
func () GetModelMapping(modelID ModelID) string
```

GetModelMapping returns the concrete artifact path for a logical model ID.
This exposes the mapping required for observability per spec §4.2.9.

#### Engine.GetVersion

```go
func () GetVersion() string
```

GetVersion returns the liquidgen module version or commit identifier.

#### Engine.Infer

```go
func () Infer(modelID ModelID, input string) (string, error)
```

Infer executes inference with the specified model and input.
This is the legacy interface, deprecated in favor of Generate.

#### Engine.IsAvailable

```go
func () IsAvailable() bool
```

IsAvailable returns whether the liquidgen engine is available.


## type LiquidgenEngine

```go
type LiquidgenEngine interface {
	// Generate produces a completion; implementations SHOULD stream tokens.
	Generate(ctx context.Context, req LiquidgenRequest) (LiquidgenStream, error)
}
```

LiquidgenEngine interface as defined in spec §4.2.10.

## type LiquidgenRequest

```go
type LiquidgenRequest struct {
	Model       string  `json:"model"` // "lfm2.5-thinking" or "lfm2.5-VL"
	Prompt      string  `json:"prompt"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}
```

LiquidgenRequest represents an inference request as defined in spec §4.2.10.

## type LiquidgenStream

```go
type LiquidgenStream interface {
	// Next returns the next token chunk; io.EOF indicates end of stream.
	Next() (token string, err error)
	Close() error
}
```

LiquidgenStream interface for streaming inference results.

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


## type liquidgenStream

```go
type liquidgenStream struct {
	engine       *Engine
	request      LiquidgenRequest
	artifactPath string
	ctx          context.Context
	done         bool
}
```

liquidgenStream implements LiquidgenStream for streaming inference.

### Methods

#### liquidgenStream.Close

```go
func () Close() error
```

Close closes the stream and cleans up resources.

#### liquidgenStream.Next

```go
func () Next() (string, error)
```

Next returns the next token chunk; io.EOF indicates end of stream.


