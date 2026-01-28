// Package inference provides the local inference engine interface for amux.
//
// This package defines the interface for local LLM inference using liquidgen.
// The implementation supports the following models:
//   - lfm2.5-thinking (text-only reasoning)
//   - lfm2.5-VL (vision-language)
//
// See spec §4.2.10 for the full requirements.
package inference

import (
	"context"
	"errors"
	"io"
	"sync"
)

// Model identifiers as required by the specification.
const (
	// ModelLFM25Thinking is the text-only reasoning model.
	ModelLFM25Thinking = "lfm2.5-thinking"

	// ModelLFM25VL is the vision-language model.
	ModelLFM25VL = "lfm2.5-VL"
)

// ValidModels contains all valid model identifiers.
var ValidModels = []string{ModelLFM25Thinking, ModelLFM25VL}

// Engine is the interface for local inference engines.
// Implementations SHOULD stream tokens for best user experience.
type Engine interface {
	// Generate produces a completion for the given request.
	// The returned Stream allows consuming tokens as they are generated.
	Generate(ctx context.Context, req Request) (Stream, error)

	// Available returns true if the engine is ready to serve requests.
	Available() bool

	// Close releases any resources held by the engine.
	Close() error
}

// Request contains the parameters for a generation request.
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

// Stream provides access to generated tokens.
type Stream interface {
	// Next returns the next token chunk.
	// Returns io.EOF when generation is complete.
	// Returns an error if generation fails.
	Next() (token string, err error)

	// Close releases resources and cancels any ongoing generation.
	Close() error
}

// Errors for inference operations.
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

// IsValidModel returns true if the model identifier is valid.
func IsValidModel(model string) bool {
	for _, m := range ValidModels {
		if m == model {
			return true
		}
	}
	return false
}

// NoopEngine is a no-op implementation of Engine for testing and
// environments where liquidgen is not available.
type NoopEngine struct {
	mu        sync.RWMutex
	available bool
}

// NewNoopEngine creates a new no-op engine.
func NewNoopEngine() *NoopEngine {
	return &NoopEngine{available: true}
}

// Generate returns an empty stream that immediately completes.
// Returns ErrModelNotFound for unknown models.
func (e *NoopEngine) Generate(ctx context.Context, req Request) (Stream, error) {
	if !e.Available() {
		return nil, ErrEngineUnavailable
	}

	if !IsValidModel(req.Model) {
		return nil, ErrModelNotFound
	}

	return &noopStream{}, nil
}

// Available returns whether the engine is available.
func (e *NoopEngine) Available() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.available
}

// SetAvailable sets the availability of the engine (for testing).
func (e *NoopEngine) SetAvailable(available bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.available = available
}

// Close is a no-op.
func (e *NoopEngine) Close() error {
	return nil
}

// noopStream is a stream that immediately returns EOF.
type noopStream struct{}

func (s *noopStream) Next() (string, error) {
	return "", io.EOF
}

func (s *noopStream) Close() error {
	return nil
}

// StringStream is a simple stream that returns a fixed string.
// Useful for testing.
type StringStream struct {
	tokens []string
	index  int
	mu     sync.Mutex
}

// NewStringStream creates a stream that returns the given tokens.
func NewStringStream(tokens ...string) *StringStream {
	return &StringStream{tokens: tokens}
}

// Next returns the next token or io.EOF.
func (s *StringStream) Next() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.index >= len(s.tokens) {
		return "", io.EOF
	}

	token := s.tokens[s.index]
	s.index++
	return token, nil
}

// Close is a no-op.
func (s *StringStream) Close() error {
	return nil
}

// defaultEngine is the default inference engine.
// It is set during initialization based on liquidgen availability.
var (
	defaultEngine Engine = NewNoopEngine()
	engineMu      sync.RWMutex
)

// SetDefaultEngine sets the default inference engine.
func SetDefaultEngine(engine Engine) {
	engineMu.Lock()
	defer engineMu.Unlock()
	defaultEngine = engine
}

// DefaultEngine returns the default inference engine.
func DefaultEngine() Engine {
	engineMu.RLock()
	defer engineMu.RUnlock()
	return defaultEngine
}

// Generate uses the default engine to generate a completion.
func Generate(ctx context.Context, req Request) (Stream, error) {
	return DefaultEngine().Generate(ctx, req)
}

// Available returns whether the default engine is available.
func Available() bool {
	return DefaultEngine().Available()
}
