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
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/stateforward/amux/internal/errors"
)

// ModelID represents a logical model identifier.
type ModelID string

// Defined model IDs per spec §4.2.10
const (
	ModelLFM25Thinking ModelID = "lfm2.5-thinking"
	ModelLFM25VL       ModelID = "lfm2.5-VL"
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
	root := "third_party/liquidgen"
	if env := os.Getenv("AMUX_LIQUIDGEN_ROOT"); env != "" {
		root = env
	}

	bin := filepath.Join(root, "build", "liquid-server")
	commit := ""
	if out, err := exec.Command("git", "-C", root, "rev-parse", "--short", "HEAD").Output(); err == nil {
		commit = strings.TrimSpace(string(out))
	}

	addr := os.Getenv("AMUX_LIQUIDGEN_ADDR")
	if addr == "" {
		addr = "http://127.0.0.1:8080"
	}

	return &stubEngine{
		liquidgenPath: root,
		binaryPath:    bin,
		commit:        commit,
		addr:          addr,
	}, nil
}

// stubEngine is a placeholder implementation for Phase 0.
type stubEngine struct {
	liquidgenPath string
	binaryPath    string
	commit        string
	addr          string
}

func (s *stubEngine) Generate(ctx context.Context, model ModelID, prompt string) (string, error) {
	if model != ModelLFM25Thinking && model != ModelLFM25VL {
		return "", errors.Wrapf(errors.ErrInvalidInput, "unknown model ID: %s", model)
	}

	reqBody := struct {
		Model    string `json:"model,omitempty"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
		Stream bool `json:"stream,omitempty"`
	}{
		Model: string(model),
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{{
			Role:    "user",
			Content: prompt,
		}},
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", errors.Wrap(err, "marshal liquidgen request")
	}

	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.addr+"/v1/chat/completions", bytes.NewReader(data))
	if err != nil {
		return "", errors.Wrap(err, "create liquidgen request")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "call liquidgen server")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "read liquidgen response")
	}

	// Error mapping: propagate server-side errors when present.
	type chatResponse struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	var cr chatResponse
	if err := json.Unmarshal(body, &cr); err != nil {
		return "", errors.Wrapf(err, "unmarshal liquidgen response (status=%d)", resp.StatusCode)
	}

	if resp.StatusCode >= 400 {
		msg := cr.Error.Message
		if msg == "" {
			msg = string(body)
		}
		return "", errors.Wrapf(errors.ErrInvalidInput, "liquidgen HTTP %d: %s", resp.StatusCode, msg)
	}

	if len(cr.Choices) == 0 {
		return "", errors.Wrap(errors.ErrNotImplemented, "liquidgen response missing choices")
	}

	return cr.Choices[0].Message.Content, nil
}

func (s *stubEngine) GenerateStream(ctx context.Context, model ModelID, prompt string) (<-chan string, <-chan error) {
	ch := make(chan string)
	errCh := make(chan error, 1)

	go func() {
		defer close(ch)
		defer close(errCh)

		if model != ModelLFM25Thinking && model != ModelLFM25VL {
			errCh <- errors.Wrapf(errors.ErrInvalidInput, "unknown model ID: %s", model)
			return
		}

		reqBody := struct {
			Model    string `json:"model,omitempty"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
			Stream bool `json:"stream,omitempty"`
		}{
			Model: string(model),
			Messages: []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			}{{
				Role:    "user",
				Content: prompt,
			}},
			Stream: true,
		}

		data, err := json.Marshal(reqBody)
		if err != nil {
			errCh <- errors.Wrap(err, "marshal liquidgen request")
			return
		}

		client := &http.Client{Timeout: 0} // streaming; rely on ctx
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.addr+"/v1/chat/completions", bytes.NewReader(data))
		if err != nil {
			errCh <- errors.Wrap(err, "create liquidgen request")
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "text/event-stream")

		resp, err := client.Do(req)
		if err != nil {
			errCh <- errors.Wrap(err, "call liquidgen server")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			errCh <- errors.Wrapf(errors.ErrInvalidInput, "liquidgen HTTP %d: %s", resp.StatusCode, string(body))
			return
		}

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					return
				}
				errCh <- errors.Wrap(err, "read liquidgen stream")
				return
			}

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
			if data == "[DONE]" {
				return
			}

			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}

			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				errCh <- errors.Wrap(err, "unmarshal liquidgen stream chunk")
				return
			}

			for _, choice := range chunk.Choices {
				if choice.Delta.Content == "" {
					continue
				}
				select {
				case ch <- choice.Delta.Content:
				case <-ctx.Done():
					errCh <- ctx.Err()
					return
				}
			}
		}
	}()

	return ch, errCh
}

func (s *stubEngine) Close() error {
	return nil
}
