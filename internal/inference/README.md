# package inference

`import "github.com/stateforward/amux/internal/inference"`

Package inference provides the local inference engine interface for amux per spec §4.2.10.

This package integrates the liquidgen inference engine from third_party/liquidgen
via an HTTP-compatible gateway ("liquid-server"). The engine is discovered at
runtime using the AMUX_LIQUIDGEN_ROOT and AMUX_LIQUIDGEN_ADDR environment
variables and supports dynamic model ID routing and structured error
propagation.

- `type Engine` — Engine is the interface to the local inference engine.
- `type ModelID` — ModelID represents a logical model identifier.
- `type stubEngine` — stubEngine is a placeholder implementation for Phase 0.

## type Engine

```go
type Engine interface {
	// Generate generates text using the specified model.
	Generate(ctx context.Context, model ModelID, prompt string) (string, error)

	// GenerateStream generates text using streaming.
	GenerateStream(ctx context.Context, model ModelID, prompt string) (<-chan string, <-chan error)

	// Close releases resources.
	Close() error
}
```

Engine is the interface to the local inference engine.

### Functions returning Engine

#### NewEngine

```go
func NewEngine() (Engine, error)
```

NewEngine creates a new inference engine.
Phase 0: Returns a stub that references liquidgen from third_party/liquidgen.
Full integration requires either:
  1. CGO bindings to liquidgen C++ library
  2. Exec-based integration with liquidgen CLI

The liquidgen source is available at third_party/liquidgen/ and includes:
  - src/inference/: Core inference engine
  - src/orchestrator/: Multi-model orchestration
  - CMakeLists.txt: Build configuration

Traceability: liquidgen is a git submodule at third_party/liquidgen


## type ModelID

```go
type ModelID string
```

ModelID represents a logical model identifier.

### Constants

#### ModelLFM25Thinking, ModelLFM25VL

```go
const (
	ModelLFM25Thinking ModelID = "lfm2.5-thinking"
	ModelLFM25VL       ModelID = "lfm2.5-VL"
)
```

Defined model IDs per spec §4.2.10


## type stubEngine

```go
type stubEngine struct {
	liquidgenPath string
	binaryPath    string
	commit        string
	addr          string
}
```

stubEngine is a placeholder implementation for Phase 0.

### Methods

#### stubEngine.Close

```go
func () Close() error
```

#### stubEngine.Generate

```go
func () Generate(ctx context.Context, model ModelID, prompt string) (string, error)
```

#### stubEngine.GenerateStream

```go
func () GenerateStream(ctx context.Context, model ModelID, prompt string) (<-chan string, <-chan error)
```


