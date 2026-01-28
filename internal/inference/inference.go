// Package inference provides the local inference engine interface for amux.
// This package defines the interface that liquidgen must implement.
package inference

import (
	"context"
	"fmt"
	"io"
)

// Engine is the interface for local inference engines.
type Engine interface {
	// Generate produces a completion; implementations SHOULD stream tokens.
	Generate(ctx context.Context, req Request) (Stream, error)
}

// Request represents an inference request.
type Request struct {
	Model       string   // "lfm2.5-thinking" or "lfm2.5-VL" (quantized variant)
	Prompt      string
	MaxTokens   int
	Temperature float64
}

// Stream represents a streaming inference response.
type Stream interface {
	// Next returns the next token chunk; io.EOF indicates end of stream.
	Next() (token string, err error)
	Close() error
}

// ErrUnknownModel is returned when an unknown model ID is requested.
var ErrUnknownModel = fmt.Errorf("unknown model")

// ErrModelUnavailable is returned when a known model is unavailable.
var ErrModelUnavailable = fmt.Errorf("model unavailable")

// NewEngine creates a new inference engine.
// For Phase 0, this returns a noop implementation.
// Phase 0 TODO: Wire in liquidgen from third_party/liquidgen
func NewEngine() Engine {
	return &noopEngine{}
}

// noopEngine is a noop implementation for Phase 0.
type noopEngine struct{}

func (e *noopEngine) Generate(ctx context.Context, req Request) (Stream, error) {
	if req.Model != "lfm2.5-thinking" && req.Model != "lfm2.5-VL" {
		return nil, fmt.Errorf("%w: %s", ErrUnknownModel, req.Model)
	}
	return &noopStream{}, nil
}

type noopStream struct{}

func (s *noopStream) Next() (string, error) {
	return "", io.EOF
}

func (s *noopStream) Close() error {
	return nil
}
