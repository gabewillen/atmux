# package adapter

`import "github.com/agentflare-ai/amux/internal/adapter"`

- `type Instance` — Instance represents a running adapter instance.
- `type NoopInstance`
- `type NoopRuntime` — NoopRuntime implements a no-op runtime.
- `type Runtime` — Runtime defines the interface for the WASM adapter runtime.

## type Instance

```go
type Instance interface {
	// ExecuteAction checks for patterns and executes actions.
	ExecuteAction(ctx context.Context, input []byte) ([]byte, error)
}
```

Instance represents a running adapter instance.

## type NoopInstance

```go
type NoopInstance struct{}
```

### Methods

#### NoopInstance.ExecuteAction

```go
func () ExecuteAction(ctx context.Context, input []byte) ([]byte, error)
```


## type NoopRuntime

```go
type NoopRuntime struct{}
```

NoopRuntime implements a no-op runtime.

### Functions returning NoopRuntime

#### NewNoopRuntime

```go
func NewNoopRuntime() *NoopRuntime
```


### Methods

#### NoopRuntime.Load

```go
func () Load(ctx context.Context, name string) (Instance, error)
```


## type Runtime

```go
type Runtime interface {
	// Load loads an adapter from the registry.
	Load(ctx context.Context, name string) (Instance, error)
}
```

Runtime defines the interface for the WASM adapter runtime.

