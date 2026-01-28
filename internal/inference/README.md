# package inference

`import "github.com/agentflare-ai/amux/internal/inference"`

Package inference provides local inference engine integration.

The default implementation is backed by the liquidgen engine in third_party.

- `ErrUnknownModel` — ErrUnknownModel is returned when a model ID is not mapped.
- `func extractVersion(line string) string`
- `func readLiquidgenVersion(root string) (string, error)`
- `type Engine` — Engine defines the local inference interface.
- `type LiquidgenEngine` — LiquidgenEngine implements Engine using the bundled liquidgen runtime.
- `type ModelInfo` — ModelInfo describes a logical model ID and its artifact path.
- `type Request` — Request describes a local inference request.
- `type Response` — Response contains the inference output.

### Variables

#### ErrUnknownModel

```go
var ErrUnknownModel = errors.New("unknown model id")
```

ErrUnknownModel is returned when a model ID is not mapped.


### Functions

#### extractVersion

```go
func extractVersion(line string) string
```

#### readLiquidgenVersion

```go
func readLiquidgenVersion(root string) (string, error)
```


## type Engine

```go
type Engine interface {
	// Version returns the engine version or commit identifier.
	Version() string
	// Models returns the available model mappings.
	Models(ctx context.Context) ([]ModelInfo, error)
	// Infer executes a local inference request.
	Infer(ctx context.Context, req Request) (Response, error)
}
```

Engine defines the local inference interface.

### Functions returning Engine

#### NewDefaultEngine

```go
func NewDefaultEngine(repoRoot string, logger *log.Logger) (Engine, error)
```

NewDefaultEngine constructs the default local inference engine.


## type LiquidgenEngine

```go
type LiquidgenEngine struct {
	root    string
	version string
	models  map[string]string
	logger  *log.Logger
}
```

LiquidgenEngine implements Engine using the bundled liquidgen runtime.

### Functions returning LiquidgenEngine

#### NewLiquidgenEngine

```go
func NewLiquidgenEngine(root string, logger *log.Logger) (*LiquidgenEngine, error)
```

NewLiquidgenEngine loads the liquidgen engine from the provided root.


### Methods

#### LiquidgenEngine.Infer

```go
func () Infer(ctx context.Context, req Request) (Response, error)
```

Infer returns an error for unknown models until liquidgen runtime is wired.

#### LiquidgenEngine.Models

```go
func () Models(ctx context.Context) ([]ModelInfo, error)
```

Models returns the registered model mappings.

#### LiquidgenEngine.RegisterModel

```go
func () RegisterModel(id string, artifactPath string)
```

RegisterModel maps a logical model ID to an artifact path.

#### LiquidgenEngine.Version

```go
func () Version() string
```

Version returns the liquidgen version or commit identifier.


## type ModelInfo

```go
type ModelInfo struct {
	ID           string
	ArtifactPath string
}
```

ModelInfo describes a logical model ID and its artifact path.

## type Request

```go
type Request struct {
	ModelID string
	Prompt  string
}
```

Request describes a local inference request.

## type Response

```go
type Response struct {
	Output string
}
```

Response contains the inference output.

