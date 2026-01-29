# package inference

`import "github.com/agentflare-ai/amux/internal/inference"`

Package inference provides the local inference engine interface for amux.
This package defines the interface that liquidgen must implement.

Package inference provides the local inference engine interface for amux.

- `ErrModelUnavailable` — ErrModelUnavailable is returned when a known model is unavailable.
- `ErrUnknownModel` — ErrUnknownModel is returned when an unknown model ID is requested.
- `func GetLiquidgenVersion() string` — GetLiquidgenVersion returns the liquidgen version/commit identifier for traceability.
- `func findLiquidgenBinary(liquidgenDir string) string` — findLiquidgenBinary looks for the liquidgen binary in common build locations.
- `func findModuleRoot() (string, error)` — findModuleRoot finds the Go module root directory.
- `func getLiquidgenVersion(liquidgenDir string) string` — getLiquidgenVersion extracts the git commit hash from third_party/liquidgen.
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

#### findLiquidgenBinary

```go
func findLiquidgenBinary(liquidgenDir string) string
```

findLiquidgenBinary looks for the liquidgen binary in common build locations.

#### findModuleRoot

```go
func findModuleRoot() (string, error)
```

findModuleRoot finds the Go module root directory.

#### getLiquidgenVersion

```go
func getLiquidgenVersion(liquidgenDir string) string
```

getLiquidgenVersion extracts the git commit hash from third_party/liquidgen.


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
It locates the liquidgen binary and extracts version information.


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
	binaryPath string
	version    string
}
```

liquidgenEngine integrates with the liquidgen inference engine from third_party/liquidgen.
Phase 0: Basic integration that validates models and extracts version info.

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


