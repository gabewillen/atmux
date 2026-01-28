// Package inference provides local inference engine (liquidgen) functionality.
// This package supports the required models: lfm2.5-thinking and lfm2.5-VL
// for local model inference features.
package inference

import (
	"errors"
	"fmt"
)

// Common sentinel errors for inference operations.
var (
	// ErrModelNotFound indicates the requested model was not found.
	ErrModelNotFound = errors.New("model not found")

	// ErrInferenceError indicates an inference execution error.
	ErrInferenceError = errors.New("inference error")

	// ErrEngineNotAvailable indicates the liquidgen engine is not available.
	ErrEngineNotAvailable = errors.New("liquidgen engine not available")
)

// ModelID represents a logical model identifier as required by the spec.
type ModelID string

const (
	// ModelLFM25Thinking represents the lfm2.5-thinking model for text-only reasoning.
	ModelLFM25Thinking ModelID = "lfm2.5-thinking"

	// ModelLFM25VL represents the lfm2.5-VL model for vision-language tasks.
	ModelLFM25VL ModelID = "lfm2.5-VL"
)

// Engine provides liquidgen inference capabilities.
// Implementation deferred to Phase 0 completion and later phases.
type Engine struct {
	available bool
}

// NewEngine creates a new liquidgen inference engine.
func NewEngine() (*Engine, error) {
	return &Engine{
		available: false, // Will be implemented in later phases
	}, nil
}

// IsAvailable returns whether the liquidgen engine is available.
func (e *Engine) IsAvailable() bool {
	return e.available
}

// Infer executes inference with the specified model and input.
func (e *Engine) Infer(modelID ModelID, input string) (string, error) {
	if !e.available {
		return "", fmt.Errorf("liquidgen not available: %w", ErrEngineNotAvailable)
	}

	// Validate model ID
	switch modelID {
	case ModelLFM25Thinking, ModelLFM25VL:
		// Valid model IDs
	default:
		return "", fmt.Errorf("unsupported model %s: %w", modelID, ErrModelNotFound)
	}

	// Implementation deferred to later phases
	return "", fmt.Errorf("inference not implemented: %w", ErrInferenceError)
}