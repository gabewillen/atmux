# package events

`import "github.com/copilot-claude-sonnet-4/amux/internal/events"`

Package events provides event dispatch functionality.
This package provides stable interfaces for event emission and subscription
with noop implementations to unblock phased work.

- `ErrDispatcherNotConnected, ErrSubscriptionFailed, ErrEventEmitFailed` — Common sentinel errors for event operations.
- `type Dispatcher` — Dispatcher provides event dispatch functionality.
- `type Handler` — Handler represents an event handler function.

### Variables

#### ErrDispatcherNotConnected, ErrSubscriptionFailed, ErrEventEmitFailed

```go
var (
	// ErrDispatcherNotConnected indicates the dispatcher is not connected.
	ErrDispatcherNotConnected = errors.New("dispatcher not connected")

	// ErrSubscriptionFailed indicates subscription setup failed.
	ErrSubscriptionFailed = errors.New("subscription failed")

	// ErrEventEmitFailed indicates event emission failed.
	ErrEventEmitFailed = errors.New("event emit failed")
)
```

Common sentinel errors for event operations.


## type Dispatcher

```go
type Dispatcher struct {
	connected bool
	handlers  map[string][]Handler
}
```

Dispatcher provides event dispatch functionality.
Phase 0 provides a noop implementation to unblock later phases.

### Functions returning Dispatcher

#### NewDispatcher

```go
func NewDispatcher() *Dispatcher
```

NewDispatcher creates a new event dispatcher.


### Methods

#### Dispatcher.Close

```go
func () Close() error
```

Close shuts down the dispatcher.

#### Dispatcher.Connect

```go
func () Connect(ctx context.Context) error
```

Connect establishes the dispatcher connection.
Phase 0: noop implementation.

#### Dispatcher.Emit

```go
func () Emit(ctx context.Context, event api.Event) error
```

Emit sends an event through the dispatch system.
Phase 0: noop implementation that accepts but doesn't distribute events.

#### Dispatcher.Subscribe

```go
func () Subscribe(eventType string, handler Handler) error
```

Subscribe registers a handler for events of the given type.
Phase 0: noop implementation that accepts but doesn't invoke handlers.


## type Handler

```go
type Handler func(ctx context.Context, event api.Event) error
```

Handler represents an event handler function.

