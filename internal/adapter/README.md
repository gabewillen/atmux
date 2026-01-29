# package adapter

`import "github.com/copilot-claude-sonnet-4/amux/internal/adapter"`

Package adapter provides WASM adapter runtime and loading functionality.
This package loads conforming WASM adapters without any knowledge of
specific agent implementations.

The adapter system provides a pluggable interface enabling any coding agent
to be integrated through a WASM adapter that implements the required ABI.

- `ErrAdapterNotFound, ErrInvalidABI, ErrRuntimeFailed` — Common sentinel errors for adapter operations.
- `type AdapterCommands` — AdapterCommands defines commands to interact with the agent
- `type AdapterInstance` — AdapterInstance represents a loaded WASM adapter instance
- `type AdapterManifest` — AdapterManifest represents adapter metadata per spec §10.2
- `type AdapterPatterns` — AdapterPatterns defines output patterns for monitoring
- `type CLIRequirement` — CLIRequirement defines CLI version constraints per spec §10.3
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


## type AdapterCommands

```go
type AdapterCommands struct {
	Start       []string `json:"start"`
	SendMessage string   `json:"send_message"`
}
```

AdapterCommands defines commands to interact with the agent

## type AdapterInstance

```go
type AdapterInstance struct {
	module   api.Module
	manifest AdapterManifest
	name     string
}
```

AdapterInstance represents a loaded WASM adapter instance

### Methods

#### AdapterInstance.FormatInput

```go
func () FormatInput(ctx context.Context, input []byte) ([]byte, error)
```

FormatInput formats input through the adapter's format_input function

#### AdapterInstance.GetManifest

```go
func () GetManifest() AdapterManifest
```

GetManifest returns the adapter's manifest

#### AdapterInstance.GetName

```go
func () GetName() string
```

GetName returns the adapter's name

#### AdapterInstance.ProcessOutput

```go
func () ProcessOutput(ctx context.Context, output []byte) ([]byte, error)
```

ProcessOutput processes PTY output through the adapter's on_output function


## type AdapterManifest

```go
type AdapterManifest struct {
	Name        string          `json:"name"`
	Version     string          `json:"version"`
	Description string          `json:"description,omitempty"`
	CLI         CLIRequirement  `json:"cli"`
	Patterns    AdapterPatterns `json:"patterns"`
	Commands    AdapterCommands `json:"commands"`
}
```

AdapterManifest represents adapter metadata per spec §10.2

## type AdapterPatterns

```go
type AdapterPatterns struct {
	Ready    string `json:"ready"`
	Error    string `json:"error"`
	Complete string `json:"complete"`
}
```

AdapterPatterns defines output patterns for monitoring

## type CLIRequirement

```go
type CLIRequirement struct {
	Binary     string `json:"binary"`
	VersionCmd string `json:"version_cmd"`
	VersionRe  string `json:"version_re"`
	Constraint string `json:"constraint"`
}
```

CLIRequirement defines CLI version constraints per spec §10.3

## type Runtime

```go
type Runtime struct {
	ctx       context.Context
	engine    wazero.Runtime
	instances map[string]*AdapterInstance
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

#### Runtime.GetInstance

```go
func () GetInstance(name string) (*AdapterInstance, error)
```

GetInstance returns a loaded adapter instance by name

#### Runtime.ListInstances

```go
func () ListInstances() []*AdapterInstance
```

ListInstances returns all loaded adapter instances

#### Runtime.LoadAdapter

```go
func () LoadAdapter(name, path string) (*AdapterInstance, error)
```

LoadAdapter loads a WASM adapter from the given path.
Returns an error if the adapter doesn't implement required exports:
amux_alloc, amux_free, manifest, on_output, format_input, on_event

#### Runtime.LoadAdaptersFromDirectory

```go
func () LoadAdaptersFromDirectory(dir string) error
```

LoadAdaptersFromDirectory loads all WASM adapters from the given directory


