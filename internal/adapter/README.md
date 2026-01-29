# package adapter

`import "github.com/agentflare-ai/amux/internal/adapter"`

Package adapter manages the WASM runtime and adapter loading.

- `type Action` — Action represents an action returned by the adapter.
- `type CLIConfig` — CLIConfig defines CLI version requirements.
- `type Manifest` — Manifest represents the adapter configuration.
- `type WasmRuntime` — WasmRuntime executes a WASM adapter.

## type Action

```go
type Action struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
```

Action represents an action returned by the adapter.

## type CLIConfig

```go
type CLIConfig struct {
	MinVersion string `toml:"min_version"`
}
```

CLIConfig defines CLI version requirements.

## type Manifest

```go
type Manifest struct {
	Name        string    `toml:"name"`
	Version     string    `toml:"version"`
	Description string    `toml:"description"`
	CLI         CLIConfig `toml:"cli"`
}
```

Manifest represents the adapter configuration.

### Functions returning Manifest

#### ParseManifest

```go
func ParseManifest(data []byte) (*Manifest, error)
```

ParseManifest parses and validates the adapter manifest.


### Methods

#### Manifest.Validate

```go
func () Validate() error
```

Validate checks for required fields.


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

#### WasmRuntime.FormatInput

```go
func () FormatInput(input any) ([]byte, error)
```

FormatInput invokes the adapter's format_input function.

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


