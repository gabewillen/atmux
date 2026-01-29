# package inference

`import "github.com/agentflare-ai/amux/internal/inference"`

Package inference provides the local inference engine interface for amux.

This package defines the interface for local LLM inference using liquidgen.
The implementation supports the following models:
  - lfm2.5-thinking (text-only reasoning)
  - lfm2.5-VL (vision-language)

See spec §4.2.10 for the full requirements.

Package inference - liquidgen.go provides the liquidgen-backed inference engine.

When the "liquidgen" build tag is set and the liquidgen library is available,
this engine uses the C++ liquidgen runtime for local LLM inference. Otherwise,
the default NoopEngine is used as a fallback.

To enable: go build -tags liquidgen

See spec §4.2.10 for inference engine requirements.

- `ErrModelNotFound, ErrModelLoadFailed, ErrEngineUnavailable, ErrGenerationFailed` — Errors for inference operations.
- `ModelLFM25Thinking, ModelLFM25VL` — Model identifiers as required by the specification.
- `ValidModels` — ValidModels contains all valid model identifiers.
- `func Available() bool` — Available returns whether the default engine is available.
- `func InitDefaultEngine()` — InitDefaultEngine initializes the default inference engine.
- `func IsValidModel(model string) bool` — IsValidModel returns true if the model identifier is valid.
- `func SetDefaultEngine(engine Engine)` — SetDefaultEngine sets the default inference engine.
- `type Engine` — Engine is the interface for local inference engines.
- `type LiquidgenEngine` — LiquidgenEngine wraps the liquidgen C++ inference runtime.
- `type LiquidgenOptions` — LiquidgenOptions configures the liquidgen engine.
- `type NoopEngine` — NoopEngine is a no-op implementation of Engine for testing and environments where liquidgen is not available.
- `type Request` — Request contains the parameters for a generation request.
- `type Stream` — Stream provides access to generated tokens.
- `type StringStream` — StringStream is a simple stream that returns a fixed string.
- `type liquidgenStream` — liquidgenStream reads tokens from the liquidgen process stdout.
- `type noopStream` — noopStream is a stream that immediately returns EOF.

### Constants

#### ModelLFM25Thinking, ModelLFM25VL

```go
const (
	// ModelLFM25Thinking is the text-only reasoning model.
	ModelLFM25Thinking = "lfm2.5-thinking"

	// ModelLFM25VL is the vision-language model.
	ModelLFM25VL = "lfm2.5-VL"
)
```

Model identifiers as required by the specification.


### Variables

#### ErrModelNotFound, ErrModelLoadFailed, ErrEngineUnavailable, ErrGenerationFailed

```go
var (
	// ErrModelNotFound indicates the requested model is not available.
	ErrModelNotFound = errors.New("inference: model not found")

	// ErrModelLoadFailed indicates the model failed to load.
	ErrModelLoadFailed = errors.New("inference: model load failed")

	// ErrEngineUnavailable indicates the engine is not available.
	ErrEngineUnavailable = errors.New("inference: engine unavailable")

	// ErrGenerationFailed indicates generation failed.
	ErrGenerationFailed = errors.New("inference: generation failed")
)
```

Errors for inference operations.

#### ValidModels

```go
var ValidModels = []string{ModelLFM25Thinking, ModelLFM25VL}
```

ValidModels contains all valid model identifiers.


### Functions

#### Available

```go
func Available() bool
```

Available returns whether the default engine is available.

#### InitDefaultEngine

```go
func InitDefaultEngine()
```

InitDefaultEngine initializes the default inference engine.

It attempts to use liquidgen if available, falling back to NoopEngine.
This should be called during daemon initialization.

#### IsValidModel

```go
func IsValidModel(model string) bool
```

IsValidModel returns true if the model identifier is valid.

#### SetDefaultEngine

```go
func SetDefaultEngine(engine Engine)
```

SetDefaultEngine sets the default inference engine.


## type Engine

```go
type Engine interface {
	// Generate produces a completion for the given request.
	// The returned Stream allows consuming tokens as they are generated.
	Generate(ctx context.Context, req Request) (Stream, error)

	// Available returns true if the engine is ready to serve requests.
	Available() bool

	// Close releases any resources held by the engine.
	Close() error
}
```

Engine is the interface for local inference engines.
Implementations SHOULD stream tokens for best user experience.

### Variables

#### defaultEngine, engineMu

```go
var (
	defaultEngine Engine = NewNoopEngine()
	engineMu      sync.RWMutex
)
```

defaultEngine is the default inference engine.
It is set during initialization based on liquidgen availability.


### Functions returning Engine

#### DefaultEngine

```go
func DefaultEngine() Engine
```

DefaultEngine returns the default inference engine.


## type LiquidgenEngine

```go
type LiquidgenEngine struct {
	mu         sync.RWMutex
	binaryPath string
	available  bool
	modelDir   string
}
```

LiquidgenEngine wraps the liquidgen C++ inference runtime.

This engine checks for the presence of the liquidgen binary at
initialization time. If the binary is not available, Available()
returns false and Generate() returns ErrEngineUnavailable.

### Functions returning LiquidgenEngine

#### NewLiquidgenEngine

```go
func NewLiquidgenEngine(opts *LiquidgenOptions) *LiquidgenEngine
```

NewLiquidgenEngine creates a liquidgen-backed inference engine.

The engine probes for the liquidgen binary at creation time. If the
binary is not found, the engine is created in unavailable state and
all Generate() calls return ErrEngineUnavailable.


### Methods

#### LiquidgenEngine.Available

```go
func () Available() bool
```

Available returns true if the liquidgen binary was found.

#### LiquidgenEngine.Close

```go
func () Close() error
```

Close releases resources.

#### LiquidgenEngine.Generate

```go
func () Generate(ctx context.Context, req Request) (Stream, error)
```

Generate produces a completion using the liquidgen runtime.

If the engine is unavailable, returns ErrEngineUnavailable.
If the model is unknown, returns ErrModelNotFound.


## type LiquidgenOptions

```go
type LiquidgenOptions struct {
	// BinaryPath is the path to the liquidgen binary.
	// If empty, searches PATH for "liquidgen".
	BinaryPath string

	// ModelDir is the directory containing model weights.
	// If empty, uses ~/.amux/models/.
	ModelDir string
}
```

LiquidgenOptions configures the liquidgen engine.

## type NoopEngine

```go
type NoopEngine struct {
	mu        sync.RWMutex
	available bool
}
```

NoopEngine is a no-op implementation of Engine for testing and
environments where liquidgen is not available.

### Functions returning NoopEngine

#### NewNoopEngine

```go
func NewNoopEngine() *NoopEngine
```

NewNoopEngine creates a new no-op engine.


### Methods

#### NoopEngine.Available

```go
func () Available() bool
```

Available returns whether the engine is available.

#### NoopEngine.Close

```go
func () Close() error
```

Close is a no-op.

#### NoopEngine.Generate

```go
func () Generate(ctx context.Context, req Request) (Stream, error)
```

Generate returns an empty stream that immediately completes.
Returns ErrModelNotFound for unknown models.

#### NoopEngine.SetAvailable

```go
func () SetAvailable(available bool)
```

SetAvailable sets the availability of the engine (for testing).


## type Request

```go
type Request struct {
	// Model is the model identifier ("lfm2.5-thinking" or "lfm2.5-VL").
	// If an unknown model is specified, Generate MUST return an error.
	Model string

	// Prompt is the input text for generation.
	Prompt string

	// MaxTokens is the maximum number of tokens to generate.
	// If 0, the engine uses its default limit.
	MaxTokens int

	// Temperature controls randomness in generation.
	// 0 means deterministic, higher values increase randomness.
	Temperature float64

	// Images is an optional list of image data for vision-language models.
	// Only used when Model is ModelLFM25VL.
	Images [][]byte
}
```

Request contains the parameters for a generation request.

## type Stream

```go
type Stream interface {
	// Next returns the next token chunk.
	// Returns io.EOF when generation is complete.
	// Returns an error if generation fails.
	Next() (token string, err error)

	// Close releases resources and cancels any ongoing generation.
	Close() error
}
```

Stream provides access to generated tokens.

### Functions returning Stream

#### Generate

```go
func Generate(ctx context.Context, req Request) (Stream, error)
```

Generate uses the default engine to generate a completion.


## type StringStream

```go
type StringStream struct {
	tokens []string
	index  int
	mu     sync.Mutex
}
```

StringStream is a simple stream that returns a fixed string.
Useful for testing.

### Functions returning StringStream

#### NewStringStream

```go
func NewStringStream(tokens ...string) *StringStream
```

NewStringStream creates a stream that returns the given tokens.


### Methods

#### StringStream.Close

```go
func () Close() error
```

Close is a no-op.

#### StringStream.Next

```go
func () Next() (string, error)
```

Next returns the next token or io.EOF.


## type liquidgenStream

```go
type liquidgenStream struct {
	mu     sync.Mutex
	cmd    *exec.Cmd
	stdout io.ReadCloser
	buf    []byte
	done   bool
}
```

liquidgenStream reads tokens from the liquidgen process stdout.

### Methods

#### liquidgenStream.Close

```go
func () Close() error
```

#### liquidgenStream.Next

```go
func () Next() (string, error)
```


## type noopStream

```go
type noopStream struct{}
```

noopStream is a stream that immediately returns EOF.

### Methods

#### noopStream.Close

```go
func () Close() error
```

#### noopStream.Next

```go
func () Next() (string, error)
```


