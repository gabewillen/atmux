# package event

`import "github.com/stateforward/amux/internal/event"`

Package event provides event dispatch interfaces for amux per spec §9.

Phase 0: Provides stable interfaces with local/noop implementations.
Phase 7 will add full hsmnet network-aware dispatch.

- `type Dispatcher` — Dispatcher is the interface for event dispatch.
- `type EventFilter` — EventFilter filters events by type or other criteria.
- `type Event` — Event represents a generic event in the system.
- `type noopDispatcher` — noopDispatcher is a no-op implementation for Phase 0.

## type Dispatcher

```go
type Dispatcher interface {
	// Dispatch sends an event to the event bus.
	Dispatch(ctx context.Context, event Event) error

	// Subscribe subscribes to events matching the given filter.
	Subscribe(ctx context.Context, filter EventFilter) (<-chan Event, error)

	// Close closes the dispatcher.
	Close() error
}
```

Dispatcher is the interface for event dispatch.

### Functions returning Dispatcher

#### NewLocalDispatcher

```go
func NewLocalDispatcher() Dispatcher
```

NewLocalDispatcher creates a new local-only dispatcher.
This is a Phase 0 stub that will be enhanced in Phase 7.


## type Event

```go
type Event interface {
	// Type returns the event type (e.g., "agent.started", "process.spawned").
	Type() string

	// Data returns the event payload.
	Data() any
}
```

Event represents a generic event in the system.

## type EventFilter

```go
type EventFilter interface {
	// Matches returns true if the event matches this filter.
	Matches(event Event) bool
}
```

EventFilter filters events by type or other criteria.

## type noopDispatcher

```go
type noopDispatcher struct{}
```

noopDispatcher is a no-op implementation for Phase 0.

### Methods

#### noopDispatcher.Close

```go
func () Close() error
```

#### noopDispatcher.Dispatch

```go
func () Dispatch(ctx context.Context, event Event) error
```

#### noopDispatcher.Subscribe

```go
func () Subscribe(ctx context.Context, filter EventFilter) (<-chan Event, error)
```


