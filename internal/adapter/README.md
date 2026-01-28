# package adapter

`import "github.com/copilot-claude-sonnet-4/amux/internal/adapter"`

Package adapter provides WASM adapter runtime and loading functionality.
This package loads conforming WASM adapters without any knowledge of
specific agent implementations.

The adapter system provides a pluggable interface enabling any coding agent
to be integrated through a WASM adapter that implements the required ABI.

- `ErrAdapterNotFound, ErrInvalidABI, ErrRuntimeFailed` — Common sentinel errors for adapter operations.
- `type Runtime` — Runtime manages WASM adapter instances using wazero.

### Variables

#### ErrAdapterNotFound, ErrInvalidABI, ErrRuntimeFailed

```go
var (
	// ErrAdapterNotFound indicates the requested adapter was not found.
	ErrAdapterNotFound = errors.New("adapter not found")

	// ErrInvalidABI indicates the adapter does not implement the required ABI.
	ErrInvalidABI = errors.New("invalid adapter ABI")

	// ErrRuntimeFailed indicates a WASM runtime failure.
	ErrRuntimeFailed = errors.New("WASM runtime failure")
)
```

Common sentinel errors for adapter operations.


## type Runtime

```go
type Runtime struct {
	ctx    context.Context
	engine wazero.Runtime
}
```

Runtime manages WASM adapter instances using wazero.
One WASM instance per agent with 256MB memory cap.

### Functions returning Runtime

#### NewRuntime

```go
func NewRuntime(ctx context.Context) (*Runtime, error)
```

NewRuntime creates a new WASM adapter runtime.


### Methods

#### Runtime.Close

```go
func () Close() error
```

Close releases runtime resources.

#### Runtime.LoadAdapter

```go
func () LoadAdapter(path string) (api.Module, error)
```

LoadAdapter loads a WASM adapter from the given path.
Returns an error if the adapter doesn't implement required exports:
amux_alloc, amux_free, manifest, on_output, format_input, on_event


