# package inference

`import "github.com/agentflare-ai/amux/internal/inference"`

Package inference provides the local inference engine interface for amux.
This package defines the interface that liquidgen must implement.

Package inference provides the local inference engine interface for amux.

- `ErrModelUnavailable` — ErrModelUnavailable is returned when a known model is unavailable.
- `ErrUnknownModel` — ErrUnknownModel is returned when an unknown model ID is requested.
- `func GetLiquidgenVersion() string` — GetLiquidgenVersion returns the liquidgen version/commit identifier for traceability.
- `type Engine` — Engine is the interface for local inference engines.
- `type Request` — Request represents an inference request.
- `type Stream` — Stream represents a streaming inference response.
- `type liquidgenEngine` — liquidgenEngine integrates with the liquidgen inference engine from third_party/liquidgen.
- `type noopEngine` — noopEngine is a noop implementation for Phase 0.
- `type noopStream`

### Variables

#### ErrModelUnavailable

```go
var ErrModelUnavailable = fmt.Errorf("model unavailable")
```

ErrModelUnavailable is returned when a known model is unavailable.

#### ErrUnknownModel

```go
var ErrUnknownModel = fmt.Errorf("unknown model")
```

ErrUnknownModel is returned when an unknown model ID is requested.


### Functions

#### GetLiquidgenVersion

```go
func GetLiquidgenVersion() string
```

GetLiquidgenVersion returns the liquidgen version/commit identifier for traceability.
Phase 0: Placeholder that returns a placeholder value.


## type Engine

```go
type Engine interface {
	// Generate produces a completion; implementations SHOULD stream tokens.
	Generate(ctx context.Context, req Request) (Stream, error)
}
```

Engine is the interface for local inference engines.

### Functions returning Engine

#### NewEngine

```go
func NewEngine() Engine
```

NewEngine creates a new inference engine.
For Phase 0, this returns a noop implementation.
Phase 0 TODO: Wire in liquidgen from third_party/liquidgen

#### NewLiquidgenEngine

```go
func NewLiquidgenEngine() (Engine, error)
```

NewLiquidgenEngine creates a new liquidgen-based inference engine.
Phase 0: Returns a placeholder that validates model IDs but doesn't actually run inference.


## type Request

```go
type Request struct {
	Model       string // "lfm2.5-thinking" or "lfm2.5-VL" (quantized variant)
	Prompt      string
	MaxTokens   int
	Temperature float64
}
```

Request represents an inference request.

## type Stream

```go
type Stream interface {
	// Next returns the next token chunk; io.EOF indicates end of stream.
	Next() (token string, err error)
	Close() error
}
```

Stream represents a streaming inference response.

## type liquidgenEngine

```go
type liquidgenEngine struct {
}
```

liquidgenEngine integrates with the liquidgen inference engine from third_party/liquidgen.
Phase 0: Placeholder implementation that will be completed when liquidgen integration is finalized.

### Methods

#### liquidgenEngine.Generate

```go
func () Generate(ctx context.Context, req Request) (Stream, error)
```


## type noopEngine

```go
type noopEngine struct{}
```

noopEngine is a noop implementation for Phase 0.

### Methods

#### noopEngine.Generate

```go
func () Generate(ctx context.Context, req Request) (Stream, error)
```


## type noopStream

```go
type noopStream struct{}
```

### Methods

#### noopStream.Close

```go
func () Close() error
```

#### noopStream.Next

```go
func () Next() (string, error)
```


