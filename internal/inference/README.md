# package inference

`import "github.com/stateforward/amux/internal/inference"`

Package inference implements the local inference integration interface using liquidgen

Package inference implements local inference integration (liquidgen)

Package inference implements the local inference integration using liquidgen

- `ErrInference` — ErrInference is returned when inference operations fail
- `func GetLiquidGenVersion() string` — GetLiquidGenVersion returns the version of the liquidgen library
- `type Engine` — Engine defines the interface for the local inference engine
- `type InferenceRequest` — InferenceRequest represents a request to the inference engine
- `type InferenceResponse` — InferenceResponse represents a response from the inference engine
- `type LiquidGenEngine` — LiquidGenEngine implements the Engine interface using liquidgen
- `type LiquidGenWrapper` — LiquidGenWrapper wraps the liquidgen engine
- `type ModelID` — ModelID represents a logical identifier for an inference model
- `type ModelInfo` — ModelInfo contains information about a model
- `type Stats` — Stats contains statistics about an inference operation

### Variables

#### ErrInference

```go
var ErrInference = errors.New("inference operation failed")
```

ErrInference is returned when inference operations fail


### Functions

#### GetLiquidGenVersion

```go
func GetLiquidGenVersion() string
```

GetLiquidGenVersion returns the version of the liquidgen library


## type Engine

```go
type Engine interface {
	// Infer performs inference using the specified model
	Infer(ctx context.Context, req InferenceRequest) (InferenceResponse, error)

	// ListModels returns a list of available models
	ListModels() []ModelID

	// GetModelInfo returns information about a specific model
	GetModelInfo(modelID ModelID) (ModelInfo, error)

	// IsModelLoaded checks if a model is currently loaded in memory
	IsModelLoaded(modelID ModelID) bool

	// LoadModel loads a model into memory
	LoadModel(ctx context.Context, modelID ModelID) error

	// UnloadModel removes a model from memory
	UnloadModel(modelID ModelID) error
}
```

Engine defines the interface for the local inference engine

## type InferenceRequest

```go
type InferenceRequest struct {
	ModelID ModelID                `json:"model_id"`
	Prompt  string                 `json:"prompt"`
	Options map[string]interface{} `json:"options,omitempty"`
}
```

InferenceRequest represents a request to the inference engine

## type InferenceResponse

```go
type InferenceResponse struct {
	ModelID ModelID `json:"model_id"`
	Output  string  `json:"output"`
	Stats   Stats   `json:"stats,omitempty"`
	Error   string  `json:"error,omitempty"`
}
```

InferenceResponse represents a response from the inference engine

## type LiquidGenEngine

```go
type LiquidGenEngine struct {
	models map[ModelID]*ModelInfo
}
```

LiquidGenEngine implements the Engine interface using liquidgen

### Functions returning LiquidGenEngine

#### NewLiquidGenEngine

```go
func NewLiquidGenEngine() *LiquidGenEngine
```

NewLiquidGenEngine creates a new LiquidGenEngine instance


### Methods

#### LiquidGenEngine.GetModelInfo

```go
func () GetModelInfo(modelID ModelID) (ModelInfo, error)
```

GetModelInfo returns information about a specific model

#### LiquidGenEngine.Infer

```go
func () Infer(ctx context.Context, req InferenceRequest) (InferenceResponse, error)
```

Infer performs inference using the liquidgen engine

#### LiquidGenEngine.IsModelLoaded

```go
func () IsModelLoaded(modelID ModelID) bool
```

IsModelLoaded checks if a model is currently loaded in memory

#### LiquidGenEngine.ListModels

```go
func () ListModels() []ModelID
```

ListModels returns a list of available models

#### LiquidGenEngine.LoadModel

```go
func () LoadModel(ctx context.Context, modelID ModelID) error
```

LoadModel loads a model into memory

#### LiquidGenEngine.RegisterModel

```go
func () RegisterModel(info ModelInfo)
```

RegisterModel registers a new model with the engine

#### LiquidGenEngine.UnloadModel

```go
func () UnloadModel(modelID ModelID) error
```

UnloadModel removes a model from memory


## type LiquidGenWrapper

```go
type LiquidGenWrapper struct {
	modelPath string
	loaded    bool
	version   string
}
```

LiquidGenWrapper wraps the liquidgen engine

### Functions returning LiquidGenWrapper

#### NewLiquidGenWrapper

```go
func NewLiquidGenWrapper(modelPath string) (*LiquidGenWrapper, error)
```

NewLiquidGenWrapper creates a new wrapper for the liquidgen engine


### Methods

#### LiquidGenWrapper.Close

```go
func () Close()
```

Close releases resources held by the liquidgen wrapper

#### LiquidGenWrapper.Infer

```go
func () Infer(ctx context.Context, req InferenceRequest) (InferenceResponse, error)
```

Infer performs inference using the liquidgen engine


## type ModelID

```go
type ModelID string
```

ModelID represents a logical identifier for an inference model

## type ModelInfo

```go
type ModelInfo struct {
	ID          ModelID `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	SizeBytes   int64   `json:"size_bytes"`
	Loaded      bool    `json:"loaded"`
	Version     string  `json:"version"`
	URL         string  `json:"url"`
}
```

ModelInfo contains information about a model

## type Stats

```go
type Stats struct {
	DurationMS float64 `json:"duration_ms"`
	TokenCount int     `json:"token_count"`
	ModelSize  int64   `json:"model_size_bytes"`
}
```

Stats contains statistics about an inference operation

