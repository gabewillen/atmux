# package inference

`import "github.com/stateforward/amux/internal/inference"`

Package inference provides the local inference engine interface for amux per spec §4.2.10.

This package integrates the liquidgen inference engine from third_party/liquidgen.
The liquidgen engine is a C++ application that must be built separately using CMake.

Build instructions:
  cd third_party/liquidgen
  mkdir build && cd build
  cmake ..
  make

Phase 0: This implementation provides the interface and stub. Full liquidgen integration
requires CGO bindings or exec-based integration to be completed in subsequent work.

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

#### ModelLFM25Thinking, ModelLFM25Fast

```go
const (
	ModelLFM25Thinking ModelID = "lfm2.5-thinking"
	ModelLFM25Fast     ModelID = "lfm2.5-fast"
)
```

Defined model IDs per spec §4.2.10


## type stubEngine

```go
type stubEngine struct {
	liquidgenPath string
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


