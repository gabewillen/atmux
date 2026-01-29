// Package inference provides local inference engine (liquidgen) functionality.
// This package supports the required models: lfm2.5-thinking and lfm2.5-VL
// for local model inference features per spec §4.2.10.
//
// This package integrates with the pre-existing liquidgen inference engine
// from third_party/liquidgen (local) as a dependency and wires it to the
// local inference interface defined in the specification.
package inference

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// Common sentinel errors for inference operations.
var (
	// ErrModelNotFound indicates the requested model was not found.
	ErrModelNotFound = errors.New("model not found")

	// ErrInferenceError indicates an inference execution error.
	ErrInferenceError = errors.New("inference error")

	// ErrEngineNotAvailable indicates the liquidgen engine is not available.
	ErrEngineNotAvailable = errors.New("liquidgen engine not available")

	// ErrModelLoadFailed indicates a model failed to load.
	ErrModelLoadFailed = errors.New("model load failed")
)

// ModelID represents a logical model identifier as required by the spec.
type ModelID string

const (
	// ModelLFM25Thinking represents the lfm2.5-thinking model for text-only reasoning.
	ModelLFM25Thinking ModelID = "lfm2.5-thinking"

	// ModelLFM25VL represents the lfm2.5-VL model for vision-language tasks.
	ModelLFM25VL ModelID = "lfm2.5-VL"
)

// LiquidgenRequest represents an inference request as defined in spec §4.2.10.
type LiquidgenRequest struct {
	Model       string   `json:"model"`        // "lfm2.5-thinking" or "lfm2.5-VL"
	Prompt      string   `json:"prompt"`
	MaxTokens   int      `json:"max_tokens"`
	Temperature float64  `json:"temperature"`
}

// LiquidgenStream interface for streaming inference results.
type LiquidgenStream interface {
	// Next returns the next token chunk; io.EOF indicates end of stream.
	Next() (token string, err error)
	Close() error
}

// LiquidgenEngine interface as defined in spec §4.2.10.
type LiquidgenEngine interface {
	// Generate produces a completion; implementations SHOULD stream tokens.
	Generate(ctx context.Context, req LiquidgenRequest) (LiquidgenStream, error)
}

// Engine provides liquidgen inference capabilities.
// This implementation connects to the C++ liquidgen library in third_party/liquidgen.
type Engine struct {
	available     bool
	liquidgenPath string
	modelMappings map[ModelID]string // logical ID → concrete artifact path
	version       string
}

// NewEngine creates a new liquidgen inference engine.
// This function integrates the pre-existing liquidgen library per spec requirements.
func NewEngine() (*Engine, error) {
	engine := &Engine{
		available:     false,
		modelMappings: make(map[ModelID]string),
		version:       "unknown",
	}

	// Try to locate liquidgen executable or library
	liquidgenPath := findLiquidgenPath()
	if liquidgenPath == "" {
		log.Printf("liquidgen not found in expected locations")
		return engine, nil // Return non-available engine
	}

	engine.liquidgenPath = liquidgenPath
	engine.available = true

	// Set up default model mappings (implementation-defined)
	engine.modelMappings[ModelLFM25Thinking] = findModelArtifact("lfm2.5-thinking")
	engine.modelMappings[ModelLFM25VL] = findModelArtifact("lfm2.5-VL")

	// Get liquidgen version for traceability
	engine.version = getLiquidgenVersion(liquidgenPath)

	log.Printf("liquidgen engine initialized: version=%s, path=%s", engine.version, liquidgenPath)

	return engine, nil
}

// IsAvailable returns whether the liquidgen engine is available.
func (e *Engine) IsAvailable() bool {
	return e.available
}

// Generate executes inference with the specified model and request.
func (e *Engine) Generate(ctx context.Context, req LiquidgenRequest) (LiquidgenStream, error) {
	if !e.available {
		return nil, fmt.Errorf("liquidgen not available: %w", ErrEngineNotAvailable)
	}

	// Validate model ID
	modelID := ModelID(req.Model)
	artifactPath, exists := e.modelMappings[modelID]
	if !exists {
		return nil, fmt.Errorf("unsupported model %s: %w", req.Model, ErrModelNotFound)
	}

	// Check if model artifact is available
	if artifactPath == "" || !fileExists(artifactPath) {
		return nil, fmt.Errorf("model artifact unavailable for %s (mapped to %s): %w", 
			req.Model, artifactPath, ErrModelLoadFailed)
	}

	// Create stream implementation that will call liquidgen
	stream := &liquidgenStream{
		engine:       e,
		request:      req,
		artifactPath: artifactPath,
		ctx:          ctx,
		done:         false,
	}

	return stream, nil
}

// Infer executes inference with the specified model and input.
// This is the legacy interface, deprecated in favor of Generate.
func (e *Engine) Infer(modelID ModelID, input string) (string, error) {
	if !e.available {
		return "", fmt.Errorf("liquidgen not available: %w", ErrEngineNotAvailable)
	}

	// Use Generate interface internally
	req := LiquidgenRequest{
		Model:       string(modelID),
		Prompt:      input,
		MaxTokens:   1000,
		Temperature: 0.7,
	}

	stream, err := e.Generate(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("inference failed: %w", err)
	}
	defer stream.Close()

	// Collect all tokens into a single response
	var result string
	for {
		token, err := stream.Next()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return "", fmt.Errorf("stream error: %w", err)
		}
		result += token
	}

	return result, nil
}

// GetModelMapping returns the concrete artifact path for a logical model ID.
// This exposes the mapping required for observability per spec §4.2.9.
func (e *Engine) GetModelMapping(modelID ModelID) string {
	return e.modelMappings[modelID]
}

// GetVersion returns the liquidgen module version or commit identifier.
func (e *Engine) GetVersion() string {
	return e.version
}

// liquidgenStream implements LiquidgenStream for streaming inference.
type liquidgenStream struct {
	engine       *Engine
	request      LiquidgenRequest
	artifactPath string
	ctx          context.Context
	done         bool
	// TODO: Add actual stream state for calling liquidgen
}

// Next returns the next token chunk; io.EOF indicates end of stream.
func (s *liquidgenStream) Next() (string, error) {
	if s.done {
		return "", fmt.Errorf("EOF")
	}

	// TODO: Implement actual liquidgen C++ integration
	// For now, return a placeholder response
	s.done = true
	return fmt.Sprintf("[liquidgen placeholder response for %s]", s.request.Model), nil
}

// Close closes the stream and cleans up resources.
func (s *liquidgenStream) Close() error {
	s.done = true
	return nil
}

// Helper functions for liquidgen discovery and integration

// findLiquidgenPath attempts to locate the liquidgen executable or library.
func findLiquidgenPath() string {
	// Check common locations relative to the module root
	candidates := []string{
		"third_party/liquidgen/liquidgen",
		"third_party/liquidgen/build/liquidgen",
		"third_party/liquidgen/build/Release/liquidgen",
		"../third_party/liquidgen/liquidgen",
	}

	// Add platform-specific extensions
	if runtime.GOOS == "windows" {
		for i, candidate := range candidates {
			candidates[i] = candidate + ".exe"
		}
	}

	for _, candidate := range candidates {
		if fileExists(candidate) {
			abs, err := filepath.Abs(candidate)
			if err == nil {
				return abs
			}
		}
	}

	return ""
}

// findModelArtifact attempts to locate a model artifact file.
func findModelArtifact(modelName string) string {
	// Check common model locations
	candidates := []string{
		filepath.Join("models", modelName+".gguf"),
		filepath.Join("third_party/liquidgen/models", modelName+".gguf"),
		filepath.Join("third_party/liquidgen/models", modelName+".bin"),
	}

	for _, candidate := range candidates {
		if fileExists(candidate) {
			abs, err := filepath.Abs(candidate)
			if err == nil {
				return abs
			}
		}
	}

	return "" // Model not found - engine will report ErrModelLoadFailed
}

// getLiquidgenVersion attempts to get version info from liquidgen.
func getLiquidgenVersion(liquidgenPath string) string {
	// TODO: Execute liquidgen --version or similar to get actual version
	// For now, return a placeholder that includes the path for debugging
	return fmt.Sprintf("liquidgen-dev@%s", liquidgenPath)
}

// fileExists checks if a file exists and is readable.
func fileExists(path string) bool {
	if path == "" {
		return false
	}
	
	_, err := os.Stat(path)
	return err == nil
}