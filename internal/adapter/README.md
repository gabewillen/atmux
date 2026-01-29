# package adapter

`import "github.com/agentflare-ai/amux/internal/adapter"`

Package adapter manages the WASM runtime and adapter loading.

- `type Action` — Action represents an action returned by an adapter.
- `type CLIReq` — CLIReq defines CLI version requirements.
- `type Manifest` — Manifest describes an adapter.
- `type Matcher` — Matcher is the interface for pattern matching.
- `type Runtime` — Runtime manages adapter instances.
- `type WasmRuntime` — WasmRuntime executes a WASM adapter.

## type Action

```go
type Action struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload"`
}
```

Action represents an action returned by an adapter.

## type CLIReq

```go
type CLIReq struct {
	MinVersion string `toml:"min_version"`
	MaxVersion string `toml:"max_version"` // Optional
}
```

CLIReq defines CLI version requirements.

## type Manifest

```go
type Manifest struct {
	Name        string `toml:"name"`
	Version     string `toml:"version"`
	Description string `toml:"description"`
	CLI         CLIReq `toml:"cli"`
}
```

Manifest describes an adapter.

### Functions returning Manifest

#### ParseManifest

```go
func ParseManifest(data []byte) (*Manifest, error)
```

ParseManifest parses a TOML manifest.


### Methods

#### Manifest.Validate

```go
func () Validate() error
```

Validate checks required fields.


## type Matcher

```go
type Matcher interface {
	// Match returns actions for the given input.
	Match(input []byte) ([]Action, error)
}
```

Matcher is the interface for pattern matching.

## type Runtime

```go
type Runtime interface {
	Start() error
	Stop() error
}
```

Runtime manages adapter instances.

## type WasmRuntime

```go
type WasmRuntime struct {
	runtime wazero.Runtime
	module  api.Module
}
```

WasmRuntime executes a WASM adapter.

### Functions returning WasmRuntime

#### NewWasmRuntime

```go
func NewWasmRuntime(ctx context.Context, wasmBytes []byte) (*WasmRuntime, error)
```

NewWasmRuntime creates a new runtime for the given WASM binary.


### Methods

#### WasmRuntime.Match

```go
func () Match(input []byte) ([]Action, error)
```

Match invokes the adapter's on_output function.

#### WasmRuntime.Start

```go
func () Start() error
```

#### WasmRuntime.Stop

```go
func () Stop() error
```


