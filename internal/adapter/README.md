# package adapter

`import "github.com/agentflare-ai/amux/internal/adapter"`

Package adapter defines the WASM adapter runtime interface.

The core loads adapters by string name via the WASM registry.

- `ErrAdapterInvalid, ErrAdapterMissingExport, ErrAdapterManifestMismatch, ErrAdapterExecutionFailed`
- `ErrAdapterNotFound` — ErrAdapterNotFound is returned when a named adapter cannot be loaded.
- `type ActionFormatter` — ActionFormatter converts a high-level action into agent input.
- `type Adapter` — Adapter is the runtime-facing interface to a loaded adapter.
- `type Manifest` — Manifest describes the minimal adapter manifest fields needed by the runtime.
- `type NoopAdapter` — NoopAdapter returns no matches and echoes input.
- `type NoopFormatter` — NoopFormatter returns the input unchanged.
- `type NoopMatcher` — NoopMatcher returns no matches.
- `type PatternMatch` — PatternMatch describes a detected pattern match.
- `type PatternMatcher` — PatternMatcher scans output and returns matches.
- `type Registry` — Registry loads adapters by name.
- `type WazeroRegistry` — WazeroRegistry loads adapters from WASM modules using wazero.
- `type wasmAdapter`
- `type wasmFormatter`
- `type wasmMatcher`
- `wasmMemoryLimitPages`

### Constants

#### wasmMemoryLimitPages

```go
const wasmMemoryLimitPages = 4096
```


### Variables

#### ErrAdapterInvalid, ErrAdapterMissingExport, ErrAdapterManifestMismatch, ErrAdapterExecutionFailed

```go
var (
	// ErrAdapterInvalid is returned when adapter inputs are invalid.
	ErrAdapterInvalid = errors.New("adapter invalid")
	// ErrAdapterMissingExport is returned when required exports are missing.
	ErrAdapterMissingExport = errors.New("adapter missing export")
	// ErrAdapterManifestMismatch is returned when the manifest name mismatches the requested name.
	ErrAdapterManifestMismatch = errors.New("adapter manifest mismatch")
	// ErrAdapterExecutionFailed is returned when a WASM call fails.
	ErrAdapterExecutionFailed = errors.New("adapter execution failed")
)
```

#### ErrAdapterNotFound

```go
var ErrAdapterNotFound = errors.New("adapter not found")
```

ErrAdapterNotFound is returned when a named adapter cannot be loaded.


## type ActionFormatter

```go
type ActionFormatter interface {
	Format(ctx context.Context, input string) (string, error)
}
```

ActionFormatter converts a high-level action into agent input.

## type Adapter

```go
type Adapter interface {
	Name() string
	Matcher() PatternMatcher
	Formatter() ActionFormatter
}
```

Adapter is the runtime-facing interface to a loaded adapter.

## type Manifest

```go
type Manifest struct {
	Name string `json:"name"`
}
```

Manifest describes the minimal adapter manifest fields needed by the runtime.

## type NoopAdapter

```go
type NoopAdapter struct {
	name string
}
```

NoopAdapter returns no matches and echoes input.

### Functions returning NoopAdapter

#### NewNoopAdapter

```go
func NewNoopAdapter(name string) *NoopAdapter
```

NewNoopAdapter constructs a noop adapter.


### Methods

#### NoopAdapter.Formatter

```go
func () Formatter() ActionFormatter
```

Formatter returns a noop formatter.

#### NoopAdapter.Matcher

```go
func () Matcher() PatternMatcher
```

Matcher returns a noop matcher.

#### NoopAdapter.Name

```go
func () Name() string
```

Name returns the adapter name.


## type NoopFormatter

```go
type NoopFormatter struct{}
```

NoopFormatter returns the input unchanged.

### Methods

#### NoopFormatter.Format

```go
func () Format(ctx context.Context, input string) (string, error)
```

Format returns the input unchanged.


## type NoopMatcher

```go
type NoopMatcher struct{}
```

NoopMatcher returns no matches.

### Methods

#### NoopMatcher.Match

```go
func () Match(ctx context.Context, output []byte) ([]PatternMatch, error)
```

Match returns no matches.


## type PatternMatch

```go
type PatternMatch struct {
	Pattern string
	Text    string
}
```

PatternMatch describes a detected pattern match.

## type PatternMatcher

```go
type PatternMatcher interface {
	Match(ctx context.Context, output []byte) ([]PatternMatch, error)
}
```

PatternMatcher scans output and returns matches.

## type Registry

```go
type Registry interface {
	Load(ctx context.Context, name string) (Adapter, error)
}
```

Registry loads adapters by name.

## type WazeroRegistry

```go
type WazeroRegistry struct {
	resolver *paths.Resolver
	runtime  wazero.Runtime
	mu       sync.Mutex
	compiled map[string]wazero.CompiledModule
}
```

WazeroRegistry loads adapters from WASM modules using wazero.

### Functions returning WazeroRegistry

#### NewWazeroRegistry

```go
func NewWazeroRegistry(ctx context.Context, resolver *paths.Resolver) (*WazeroRegistry, error)
```

NewWazeroRegistry constructs a registry with a wazero runtime.


### Methods

#### WazeroRegistry.Close

```go
func () Close(ctx context.Context) error
```

Close releases the wazero runtime.

#### WazeroRegistry.Load

```go
func () Load(ctx context.Context, name string) (Adapter, error)
```

Load loads an adapter by name from the WASM registry.

#### WazeroRegistry.compile

```go
func () compile(ctx context.Context, path string, wasmBytes []byte) (wazero.CompiledModule, error)
```

#### WazeroRegistry.findModule

```go
func () findModule(name string) (string, []byte, error)
```


## type wasmAdapter

```go
type wasmAdapter struct {
	name       string
	module     api.Module
	memory     api.Memory
	alloc      api.Function
	free       api.Function
	manifestFn api.Function
	onOutputFn api.Function
	formatFn   api.Function
	onEventFn  api.Function
	mu         sync.Mutex
}
```

### Functions returning wasmAdapter

#### newWasmAdapter

```go
func newWasmAdapter(name string, module api.Module) (*wasmAdapter, error)
```


### Methods

#### wasmAdapter.Formatter

```go
func () Formatter() ActionFormatter
```

#### wasmAdapter.Matcher

```go
func () Matcher() PatternMatcher
```

#### wasmAdapter.Name

```go
func () Name() string
```

#### wasmAdapter.callNoInput

```go
func () callNoInput(ctx context.Context, fn api.Function) ([]byte, error)
```

#### wasmAdapter.callWithInput

```go
func () callWithInput(ctx context.Context, fn api.Function, input []byte) ([]byte, error)
```

#### wasmAdapter.freeBuffer

```go
func () freeBuffer(ctx context.Context, ptr uint32, length uint32) error
```

#### wasmAdapter.manifest

```go
func () manifest(ctx context.Context) (Manifest, error)
```

#### wasmAdapter.readPacked

```go
func () readPacked(ctx context.Context, results []uint64) ([]byte, error)
```

#### wasmAdapter.writeInput

```go
func () writeInput(ctx context.Context, input []byte) (uint32, uint32, error)
```


## type wasmFormatter

```go
type wasmFormatter struct {
	adapter *wasmAdapter
}
```

### Methods

#### wasmFormatter.Format

```go
func () Format(ctx context.Context, input string) (string, error)
```


## type wasmMatcher

```go
type wasmMatcher struct {
	adapter *wasmAdapter
}
```

### Methods

#### wasmMatcher.Match

```go
func () Match(ctx context.Context, output []byte) ([]PatternMatch, error)
```


