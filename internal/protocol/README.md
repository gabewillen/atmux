# package protocol

`import "github.com/agentflare-ai/amux/internal/protocol"`

Package protocol provides event dispatch and routing interfaces.
Phase 0 introduces stable interfaces with noop/local implementations.
Phase 7 will provide full network-aware routing via NATS.

- `type Dispatcher` — Dispatcher provides event dispatch functionality.
- `type EventFilter` — EventFilter defines criteria for event subscription.
- `type Event` — Event represents a generic event in the system.
- `type localDispatcher` — localDispatcher is a Phase 0 local in-memory event dispatcher.

## type Dispatcher

```go
type Dispatcher interface {
	// Dispatch dispatches an event to all subscribers.
	Dispatch(ctx context.Context, event Event) error

	// Subscribe subscribes to events matching the given filter.
	Subscribe(filter EventFilter) (<-chan Event, func())
}
```

Dispatcher provides event dispatch functionality.
Phase 0: Local/noop implementation
Phase 7: Full network-aware routing

### Functions returning Dispatcher

#### NewDispatcher

```go
func NewDispatcher() Dispatcher
```

NewDispatcher creates a new event dispatcher.
Phase 0: Returns a local in-memory dispatcher


## type Event

```go
type Event struct {
	Type string
	Data interface{}
}
```

Event represents a generic event in the system.

## type EventFilter

```go
type EventFilter struct {
	Types []string // Event types to subscribe to (empty = all)
}
```

EventFilter defines criteria for event subscription.

## type localDispatcher

```go
type localDispatcher struct {
	subs map[chan Event]EventFilter
}
```

localDispatcher is a Phase 0 local in-memory event dispatcher.

### Methods

#### localDispatcher.Dispatch

```go
func () Dispatch(ctx context.Context, event Event) error
```

#### localDispatcher.Subscribe

```go
func () Subscribe(filter EventFilter) (<-chan Event, func())
```


