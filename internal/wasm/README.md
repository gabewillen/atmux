# package wasm

`import "github.com/agentflare-ai/amux/internal/wasm"`

Package wasm provides WASM runtime management for adapters.

- `func Demo(ctx context.Context) error` — Demo demonstrates WASM functionality for Phase 0.
- `type Config` — Config contains WASM runtime configuration.
- `type Runtime` — Runtime represents a WASM runtime for adapters.

### Functions

#### Demo

```go
func Demo(ctx context.Context) error
```

Demo demonstrates WASM functionality for Phase 0.


## type Config

```go
type Config struct {
	// Path to WASM module
	ModulePath string `json:"module_path"`

	// Memory limit in bytes
	MemoryLimit uint64 `json:"memory_limit"`

	// Enable debugging
	Debug bool `json:"debug"`
}
```

Config contains WASM runtime configuration.

## type Runtime

```go
type Runtime struct {
	ctx     context.Context
	runtime wazero.Runtime
	module  api.Module
	config  *Config
}
```

Runtime represents a WASM runtime for adapters.

### Functions returning Runtime

#### New

```go
func New(ctx context.Context, config *Config) (*Runtime, error)
```

New creates a new WASM runtime.


### Methods

#### Runtime.CallFunction

```go
func () CallFunction(name string, args ...uint64) (uint64, error)
```

CallFunction calls a function in the loaded WASM module.

#### Runtime.Close

```go
func () Close() error
```

Close closes the WASM runtime.

#### Runtime.LoadModule

```go
func () LoadModule() error
```

LoadModule loads a WASM module.

#### Runtime.Memory

```go
func () Memory() api.Memory
```

Memory returns the memory exports from the WASM module.


