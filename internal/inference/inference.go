// Package inference provides the local inference engine interface for amux per spec §4.2.10.
//
// This package integrates the liquidgen inference engine from third_party/liquidgen.
// The liquidgen engine is a C++ application that must be built separately using CMake.
//
// Build instructions:
//   cd third_party/liquidgen
//   mkdir build && cd build
//   cmake ..
//   make
//
// Phase 0: This implementation provides the interface and stub. Full liquidgen integration
// requires CGO bindings or exec-based integration to be completed in subsequent work.
package inference

import (
	"context"

	"github.com/stateforward/amux/internal/errors"
)

// ModelID represents a logical model identifier.
type ModelID string

// Defined model IDs per spec §4.2.10
const (
	ModelLFM25Thinking ModelID = "lfm2.5-thinking"
	ModelLFM25Fast     ModelID = "lfm2.5-fast"
)

// Engine is the interface to the local inference engine.
type Engine interface {
	// Generate generates text using the specified model.
	Generate(ctx context.Context, model ModelID, prompt string) (string, error)
	
	// GenerateStream generates text using streaming.
	GenerateStream(ctx context.Context, model ModelID, prompt string) (<-chan string, <-chan error)
	
	// Close releases resources.
	Close() error
}

// NewEngine creates a new inference engine.
// Phase 0: Returns a stub that references liquidgen from third_party/liquidgen.
// Full integration requires either:
//   1. CGO bindings to liquidgen C++ library
//   2. Exec-based integration with liquidgen CLI
//
// The liquidgen source is available at third_party/liquidgen/ and includes:
//   - src/inference/: Core inference engine
//   - src/orchestrator/: Multi-model orchestration
//   - CMakeLists.txt: Build configuration
//
// Traceability: liquidgen is a git submodule at third_party/liquidgen
func NewEngine() (Engine, error) {
	return &stubEngine{
		liquidgenPath: "third_party/liquidgen",
	}, nil
}

// stubEngine is a placeholder implementation for Phase 0.
type stubEngine struct {
	liquidgenPath string
}

func (s *stubEngine) Generate(ctx context.Context, model ModelID, prompt string) (string, error) {
	return "", errors.ErrNotImplemented
}

func (s *stubEngine) GenerateStream(ctx context.Context, model ModelID, prompt string) (<-chan string, <-chan error) {
	ch := make(chan string)
	errCh := make(chan error, 1)
	close(ch)
	errCh <- errors.ErrNotImplemented
	return ch, errCh
}

func (s *stubEngine) Close() error {
	return nil
}
