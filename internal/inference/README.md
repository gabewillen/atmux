# package inference

`import "github.com/agentflare-ai/amux/internal/inference"`

Package inference provides the local inference engine interface for amux.

This package defines the interface for local LLM inference using liquidgen.
The implementation supports the following models:
  - lfm2.5-thinking (text-only reasoning)
  - lfm2.5-VL (vision-language)

See spec §4.2.10 for the full requirements.

- `ErrModelNotFound, ErrModelLoadFailed, ErrEngineUnavailable, ErrGenerationFailed` — Errors for inference operations.
- `ModelLFM25Thinking, ModelLFM25VL` — Model identifiers as required by the specification.
- `ValidModels` — ValidModels contains all valid model identifiers.
- `func Available() bool` — Available returns whether the default engine is available.
- `func IsValidModel(model string) bool` — IsValidModel returns true if the model identifier is valid.
- `func SetDefaultEngine(engine Engine)` — SetDefaultEngine sets the default inference engine.
- `type Engine` — Engine is the interface for local inference engines.
- `type NoopEngine` — NoopEngine is a no-op implementation of Engine for testing and environments where liquidgen is not available.
- `type Request` — Request contains the parameters for a generation request.
- `type Stream` — Stream provides access to generated tokens.
- `type StringStream` — StringStream is a simple stream that returns a fixed string.
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


