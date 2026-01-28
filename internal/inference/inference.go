// Package inference provides local inference engine integration for amux.
package inference

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
	amuxerrors "github.com/agentflare-ai/amux/internal/errors"
)

// Engine represents a local inference engine.
type Engine interface {
	// Initialize the engine with configuration
	Initialize(ctx context.Context, config *config.ModelConfig) error

	// Generate text completion
	Generate(ctx context.Context, prompt string, options *GenerateOptions) (*GenerateResponse, error)

	// Create embeddings
	Embed(ctx context.Context, texts []string, options *EmbedOptions) (*EmbedResponse, error)

	// Get engine information
	Info() *EngineInfo

	// Shutdown the engine
	Shutdown(ctx context.Context) error
}

// EngineInfo provides information about the inference engine.
type EngineInfo struct {
	Name    string
	Version string
	Type    string // "local" or "remote"
	Models  map[string]*ModelInfo
}

// ModelInfo provides information about available models.
type ModelInfo struct {
	ID          string
	Type        string // "generation" or "embedding"
	Path        string
	Description string
	Parameters  map[string]interface{}
}

// GenerateOptions controls text generation.
type GenerateOptions struct {
	Model       string                 `json:"model,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	StopTokens  []string               `json:"stop_tokens,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// GenerateResponse contains text generation results.
type GenerateResponse struct {
	Text         string            `json:"text"`
	Tokens       int               `json:"tokens"`
	FinishReason string            `json:"finish_reason"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// EmbedOptions controls embedding generation.
type EmbedOptions struct {
	Model      string            `json:"model,omitempty"`
	Dimensions int               `json:"dimensions,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// EmbedResponse contains embedding results.
type EmbedResponse struct {
	Embeddings [][]float64       `json:"embeddings"`
	Dimensions int               `json:"dimensions"`
	Model      string            `json:"model"`
	Tokens     []int             `json:"tokens"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// Manager manages local inference engines.
type Manager struct {
	engines map[string]Engine
	config  *config.InferenceConfig
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewManager creates a new inference manager.
func NewManager(inferenceConfig *config.InferenceConfig) (*Manager, error) {
	if !inferenceConfig.Enabled {
		return &Manager{
			engines: make(map[string]Engine),
			config:  inferenceConfig,
		}, nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	manager := &Manager{
		engines: make(map[string]Engine),
		config:  inferenceConfig,
		ctx:     ctx,
		cancel:  cancel,
	}

	// Initialize engines based on configuration
	if err := manager.initializeEngines(); err != nil {
		cancel()
		return nil, amuxerrors.Wrap("initializing engines", err)
	}

	return manager, nil
}

// initializeEngines sets up inference engines based on configuration.
func (m *Manager) initializeEngines() error {
	switch m.config.Engine {
	case "liquidgen":
		engine, err := NewLiquidgenEngine(m.config.Models)
		if err != nil {
			return amuxerrors.Wrap("creating liquidgen engine", err)
		}
		m.engines["liquidgen"] = engine

	default:
		return amuxerrors.Wrap("unknown inference engine", amuxerrors.ErrInvalidConfig)
	}

	return nil
}

// GetEngine returns an engine by name.
func (m *Manager) GetEngine(name string) (Engine, error) {
	engine, exists := m.engines[name]
	if !exists {
		return nil, amuxerrors.Wrap("getting engine", amuxerrors.ErrAgentNotFound)
	}
	return engine, nil
}

// GetDefaultEngine returns the default inference engine.
func (m *Manager) GetDefaultEngine() (Engine, error) {
	if len(m.engines) == 0 {
		return nil, amuxerrors.Wrap("getting default engine", amuxerrors.ErrNotReady)
	}

	// Return first available engine (liquidgen if configured)
	for _, engine := range m.engines {
		return engine, nil
	}

	return nil, amuxerrors.Wrap("getting default engine", amuxerrors.ErrNotReady)
}

// ListEngines returns information about all available engines.
func (m *Manager) ListEngines() map[string]*EngineInfo {
	result := make(map[string]*EngineInfo)

	for name, engine := range m.engines {
		result[name] = engine.Info()
	}

	return result
}

// Shutdown gracefully shuts down all engines.
func (m *Manager) Shutdown(ctx context.Context) error {
	if m.cancel != nil {
		m.cancel()
	}

	var lastErr error
	for name, engine := range m.engines {
		if err := engine.Shutdown(ctx); err != nil {
			lastErr = amuxerrors.Wrap(fmt.Sprintf("shutting down engine %s", name), err)
		}
	}

	return lastErr
}

// liquidgenEngine implements Engine interface using liquidgen.
type liquidgenEngine struct {
	models map[string]config.ModelConfig
	info   *EngineInfo
	server string // liquidgen server endpoint
	client *http.Client
}

// LiquidGenRequest represents a request to liquidgen server.
type LiquidGenRequest struct {
	Model       string                 `json:"model"`
	Prompt      string                 `json:"prompt"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
	Options     map[string]interface{} `json:"options,omitempty"`
}

// LiquidGenResponse represents a response from liquidgen server.
type LiquidGenResponse struct {
	Text         string            `json:"text"`
	Tokens       int               `json:"tokens"`
	FinishReason string            `json:"finish_reason"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	Error        *string           `json:"error,omitempty"`
}

// NewLiquidgenEngine creates a new liquidgen-based inference engine.
func NewLiquidgenEngine(models map[string]config.ModelConfig) (Engine, error) {
	// Use default liquidgen server endpoint for Phase 0
	server := "http://localhost:8080" // Default from spec
	if envServer := os.Getenv("LIQUIDGEN_SERVER"); envServer != "" {
		server = envServer
	}

	engine := &liquidgenEngine{
		models: models,
		info: &EngineInfo{
			Name:    "liquidgen",
			Version: "1.1.0", // From spec
			Type:    "local",
			Models:  make(map[string]*ModelInfo),
		},
		server: server,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Convert config models to ModelInfo
	for modelID, modelConfig := range models {
		engine.info.Models[modelID] = &ModelInfo{
			ID:          modelID,
			Type:        modelConfig.Type,
			Path:        modelConfig.Path,
			Description: fmt.Sprintf("LiquidGen model: %s", modelID),
			Parameters:  modelConfig.Parameters,
		}
	}

	return engine, nil
}

// Initialize implements Engine interface.
func (e *liquidgenEngine) Initialize(ctx context.Context, config *config.ModelConfig) error {
	// Check if liquidgen server is available
	req, err := http.NewRequestWithContext(ctx, "GET", e.server+"/health", nil)
	if err != nil {
		return amuxerrors.Wrap("creating health check request", err)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return amuxerrors.Wrap("liquidgen server not available", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return amuxerrors.Wrap("liquidgen server health check failed", amuxerrors.ErrNotReady)
	}

	return nil
}

// Generate implements Engine interface.
func (e *liquidgenEngine) Generate(ctx context.Context, prompt string, options *GenerateOptions) (*GenerateResponse, error) {
	// Default model for Phase 0
	model := "lfm2.5-thinking"
	if options != nil && options.Model != "" {
		model = options.Model
	}

	// Prepare request
	req := LiquidGenRequest{
		Model:       model,
		Prompt:      prompt,
		MaxTokens:   1000,
		Temperature: 0.7,
		Stream:      false,
	}

	if options != nil {
		if options.MaxTokens > 0 {
			req.MaxTokens = options.MaxTokens
		}
		if options.Temperature > 0 {
			req.Temperature = options.Temperature
		}
		req.Stream = options.Stream
		req.Options = options.Metadata
	}

	// Send request to liquidgen server
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, amuxerrors.Wrap("marshaling liquidgen request", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", e.server+"/v1/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, amuxerrors.Wrap("creating liquidgen request", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(httpReq)
	if err != nil {
		return nil, amuxerrors.Wrap("sending liquidgen request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, amuxerrors.Wrap(fmt.Sprintf("liquidgen request failed: %s", string(body)), amuxerrors.ErrNotReady)
	}

	// Parse response
	var liquidResp LiquidGenResponse
	if err := json.NewDecoder(resp.Body).Decode(&liquidResp); err != nil {
		return nil, amuxerrors.Wrap("parsing liquidgen response", err)
	}

	if liquidResp.Error != nil {
		return nil, amuxerrors.Wrap(fmt.Sprintf("liquidgen generation error: %s", *liquidResp.Error), amuxerrors.ErrNotReady)
	}

	return &GenerateResponse{
		Text:         liquidResp.Text,
		Tokens:       liquidResp.Tokens,
		FinishReason: liquidResp.FinishReason,
		Metadata:     liquidResp.Metadata,
	}, nil
}

// Embed implements Engine interface.
func (e *liquidgenEngine) Embed(ctx context.Context, texts []string, options *EmbedOptions) (*EmbedResponse, error) {
	// Default model for Phase 0
	model := "lfm2.5-thinking"
	if options != nil && options.Model != "" {
		model = options.Model
	}

	// Prepare request for embeddings
	req := map[string]interface{}{
		"model":   model,
		"texts":   texts,
		"options": map[string]interface{}{},
	}

	if options != nil {
		if options.Dimensions > 0 {
			req["options"].(map[string]interface{})["dimensions"] = options.Dimensions
		}
		req["options"].(map[string]interface{})["metadata"] = options.Metadata
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, amuxerrors.Wrap("marshaling embed request", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", e.server+"/v1/embed", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, amuxerrors.Wrap("creating embed request", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(httpReq)
	if err != nil {
		return nil, amuxerrors.Wrap("sending embed request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, amuxerrors.Wrap(fmt.Sprintf("embed request failed: %s", string(body)), amuxerrors.ErrNotReady)
	}

	// Parse response
	var embedResp struct {
		Embeddings [][]float64 `json:"embeddings"`
		Dimensions int         `json:"dimensions"`
		Model      string      `json:"model"`
		Tokens     []int       `json:"tokens"`
		Error      *string     `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, amuxerrors.Wrap("parsing embed response", err)
	}

	if embedResp.Error != nil {
		return nil, amuxerrors.Wrap(fmt.Sprintf("embed error: %s", *embedResp.Error), amuxerrors.ErrNotReady)
	}

	return &EmbedResponse{
		Embeddings: embedResp.Embeddings,
		Dimensions: embedResp.Dimensions,
		Model:      embedResp.Model,
		Tokens:     embedResp.Tokens,
		Metadata:   options.Metadata,
	}, nil
}

// Info implements Engine interface.
func (e *liquidgenEngine) Info() *EngineInfo {
	return e.info
}

// Shutdown implements Engine interface.
func (e *liquidgenEngine) Shutdown(ctx context.Context) error {
	// No shutdown needed for HTTP-based client
	return nil
}
