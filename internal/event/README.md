# package event

`import "github.com/agentflare-ai/amux/internal/event"`

- `type Dispatcher` — Dispatcher defines the interface for event dispatching.
- `type NoopDispatcher` — NoopDispatcher is a dispatcher that does nothing.

## type Dispatcher

```go
type Dispatcher interface {
	Dispatch(ctx context.Context, event any) error
	Subscribe(pattern string) (<-chan any, func())
}
```

Dispatcher defines the interface for event dispatching.
Phase 0: Interface and Noop implementation.

## type NoopDispatcher

```go
type NoopDispatcher struct{}
```

NoopDispatcher is a dispatcher that does nothing.

### Functions returning NoopDispatcher

#### NewNoopDispatcher

```go
func NewNoopDispatcher() *NoopDispatcher
```


### Methods

#### NoopDispatcher.Dispatch

```go
func () Dispatch(ctx context.Context, event any) error
```

#### NoopDispatcher.Subscribe

```go
func () Subscribe(pattern string) (<-chan any, func())
```


