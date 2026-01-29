// Package inference - liquidgen.go provides the liquidgen-backed inference engine.
//
// When the "liquidgen" build tag is set and the liquidgen library is available,
// this engine uses the C++ liquidgen runtime for local LLM inference. Otherwise,
// the default NoopEngine is used as a fallback.
//
// To enable: go build -tags liquidgen
//
// See spec §4.2.10 for inference engine requirements.
package inference

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

// LiquidgenEngine wraps the liquidgen C++ inference runtime.
//
// This engine checks for the presence of the liquidgen binary at
// initialization time. If the binary is not available, Available()
// returns false and Generate() returns ErrEngineUnavailable.
type LiquidgenEngine struct {
	mu         sync.RWMutex
	binaryPath string
	available  bool
	modelDir   string
}

// LiquidgenOptions configures the liquidgen engine.
type LiquidgenOptions struct {
	// BinaryPath is the path to the liquidgen binary.
	// If empty, searches PATH for "liquidgen".
	BinaryPath string

	// ModelDir is the directory containing model weights.
	// If empty, uses ~/.amux/models/.
	ModelDir string
}

// NewLiquidgenEngine creates a liquidgen-backed inference engine.
//
// The engine probes for the liquidgen binary at creation time. If the
// binary is not found, the engine is created in unavailable state and
// all Generate() calls return ErrEngineUnavailable.
func NewLiquidgenEngine(opts *LiquidgenOptions) *LiquidgenEngine {
	if opts == nil {
		opts = &LiquidgenOptions{}
	}

	e := &LiquidgenEngine{
		binaryPath: opts.BinaryPath,
		modelDir:   opts.ModelDir,
	}

	// Probe for liquidgen binary
	if e.binaryPath == "" {
		path, err := exec.LookPath("liquidgen")
		if err == nil {
			e.binaryPath = path
		}
	}

	if e.modelDir == "" {
		home, _ := os.UserHomeDir()
		if home != "" {
			e.modelDir = home + "/.amux/models"
		}
	}

	// Check if the binary exists and is executable
	if e.binaryPath != "" {
		if info, err := os.Stat(e.binaryPath); err == nil && !info.IsDir() {
			e.available = true
		}
	}

	return e
}

// Generate produces a completion using the liquidgen runtime.
//
// If the engine is unavailable, returns ErrEngineUnavailable.
// If the model is unknown, returns ErrModelNotFound.
func (e *LiquidgenEngine) Generate(ctx context.Context, req Request) (Stream, error) {
	e.mu.RLock()
	avail := e.available
	e.mu.RUnlock()

	if !avail {
		return nil, ErrEngineUnavailable
	}

	if !IsValidModel(req.Model) {
		return nil, ErrModelNotFound
	}

	// Build liquidgen command arguments
	args := []string{
		"--model", req.Model,
		"--prompt", req.Prompt,
	}
	if req.MaxTokens > 0 {
		args = append(args, "--max-tokens", fmt.Sprintf("%d", req.MaxTokens))
	}
	if req.Temperature > 0 {
		args = append(args, "--temperature", fmt.Sprintf("%.2f", req.Temperature))
	}
	if e.modelDir != "" {
		args = append(args, "--model-dir", e.modelDir)
	}

	cmd := exec.CommandContext(ctx, e.binaryPath, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("%w: stdout pipe: %v", ErrGenerationFailed, err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("%w: start: %v", ErrGenerationFailed, err)
	}

	return &liquidgenStream{
		cmd:    cmd,
		stdout: stdout,
		buf:    make([]byte, 4096),
	}, nil
}

// Available returns true if the liquidgen binary was found.
func (e *LiquidgenEngine) Available() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.available
}

// Close releases resources.
func (e *LiquidgenEngine) Close() error {
	return nil
}

// liquidgenStream reads tokens from the liquidgen process stdout.
type liquidgenStream struct {
	mu     sync.Mutex
	cmd    *exec.Cmd
	stdout io.ReadCloser
	buf    []byte
	done   bool
}

func (s *liquidgenStream) Next() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.done {
		return "", io.EOF
	}

	n, err := s.stdout.Read(s.buf)
	if n > 0 {
		return string(s.buf[:n]), nil
	}
	if err != nil {
		s.done = true
		_ = s.cmd.Wait()
		return "", io.EOF
	}
	return "", io.EOF
}

func (s *liquidgenStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.done = true
	if s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
	}
	return s.cmd.Wait()
}

// InitDefaultEngine initializes the default inference engine.
//
// It attempts to use liquidgen if available, falling back to NoopEngine.
// This should be called during daemon initialization.
func InitDefaultEngine() {
	engine := NewLiquidgenEngine(nil)
	if engine.Available() {
		SetDefaultEngine(engine)
		fmt.Fprintln(os.Stderr, "Inference: using liquidgen engine")
	} else {
		fmt.Fprintln(os.Stderr, "Inference: liquidgen not available, using noop engine")
	}
}
