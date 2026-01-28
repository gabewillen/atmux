# package inference

`import "github.com/agentflare-ai/amux/internal/inference"`

Package inference provides local inference engine integration for amux.

- `type EmbedOptions` — EmbedOptions controls embedding generation.
- `type EmbedResponse` — EmbedResponse contains embedding results.
- `type EngineInfo` — EngineInfo provides information about the inference engine.
- `type Engine` — Engine represents a local inference engine.
- `type GenerateOptions` — GenerateOptions controls text generation.
- `type GenerateResponse` — GenerateResponse contains text generation results.
- `type LiquidGenRequest` — LiquidGenRequest represents a request to liquidgen server.
- `type LiquidGenResponse` — LiquidGenResponse represents a response from liquidgen server.
- `type Manager` — Manager manages local inference engines.
- `type ModelInfo` — ModelInfo provides information about available models.
- `type liquidgenEngine` — liquidgenEngine implements Engine interface using liquidgen.

## type EmbedOptions

```go
type EmbedOptions struct {
	Model      string            `json:"model,omitempty"`
	Dimensions int               `json:"dimensions,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}
```

EmbedOptions controls embedding generation.

## type EmbedResponse

```go
type EmbedResponse struct {
	Embeddings [][]float64       `json:"embeddings"`
	Dimensions int               `json:"dimensions"`
	Model      string            `json:"model"`
	Tokens     []int             `json:"tokens"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}
```

EmbedResponse contains embedding results.

## type Engine

```go
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
```

Engine represents a local inference engine.

### Functions returning Engine

#### NewLiquidgenEngine

```go
func NewLiquidgenEngine(models map[string]config.ModelConfig) (Engine, error)
```

NewLiquidgenEngine creates a new liquidgen-based inference engine.


## type EngineInfo

```go
type EngineInfo struct {
	Name    string
	Version string
	Type    string // "local" or "remote"
	Models  map[string]*ModelInfo
}
```

EngineInfo provides information about the inference engine.

## type GenerateOptions

```go
type GenerateOptions struct {
	Model       string                 `json:"model,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	StopTokens  []string               `json:"stop_tokens,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}
```

GenerateOptions controls text generation.

## type GenerateResponse

```go
type GenerateResponse struct {
	Text         string            `json:"text"`
	Tokens       int               `json:"tokens"`
	FinishReason string            `json:"finish_reason"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}
```

GenerateResponse contains text generation results.

## type LiquidGenRequest

```go
type LiquidGenRequest struct {
	Model       string                 `json:"model"`
	Prompt      string                 `json:"prompt"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
	Options     map[string]interface{} `json:"options,omitempty"`
}
```

LiquidGenRequest represents a request to liquidgen server.

## type LiquidGenResponse

```go
type LiquidGenResponse struct {
	Text         string            `json:"text"`
	Tokens       int               `json:"tokens"`
	FinishReason string            `json:"finish_reason"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	Error        *string           `json:"error,omitempty"`
}
```

LiquidGenResponse represents a response from liquidgen server.

## type Manager

```go
type Manager struct {
	engines map[string]Engine
	config  *config.InferenceConfig
	ctx     context.Context
	cancel  context.CancelFunc
}
```

Manager manages local inference engines.

### Functions returning Manager

#### NewManager

```go
func NewManager(inferenceConfig *config.InferenceConfig) (*Manager, error)
```

NewManager creates a new inference manager.


### Methods

#### Manager.GetDefaultEngine

```go
func () GetDefaultEngine() (Engine, error)
```

GetDefaultEngine returns the default inference engine.

#### Manager.GetEngine

```go
func () GetEngine(name string) (Engine, error)
```

GetEngine returns an engine by name.

#### Manager.ListEngines

```go
func () ListEngines() map[string]*EngineInfo
```

ListEngines returns information about all available engines.

#### Manager.Shutdown

```go
func () Shutdown(ctx context.Context) error
```

Shutdown gracefully shuts down all engines.

#### Manager.initializeEngines

```go
func () initializeEngines() error
```

initializeEngines sets up inference engines based on configuration.


## type ModelInfo

```go
type ModelInfo struct {
	ID          string
	Type        string // "generation" or "embedding"
	Path        string
	Description string
	Parameters  map[string]interface{}
}
```

ModelInfo provides information about available models.

## type liquidgenEngine

```go
type liquidgenEngine struct {
	models map[string]config.ModelConfig
	info   *EngineInfo
	server string // liquidgen server endpoint
	client *http.Client
}
```

liquidgenEngine implements Engine interface using liquidgen.

### Methods

#### liquidgenEngine.Embed

```go
func () Embed(ctx context.Context, texts []string, options *EmbedOptions) (*EmbedResponse, error)
```

Embed implements Engine interface.

#### liquidgenEngine.Generate

```go
func () Generate(ctx context.Context, prompt string, options *GenerateOptions) (*GenerateResponse, error)
```

Generate implements Engine interface.

#### liquidgenEngine.Info

```go
func () Info() *EngineInfo
```

Info implements Engine interface.

#### liquidgenEngine.Initialize

```go
func () Initialize(ctx context.Context, config *config.ModelConfig) error
```

Initialize implements Engine interface.

#### liquidgenEngine.Shutdown

```go
func () Shutdown(ctx context.Context) error
```

Shutdown implements Engine interface.


