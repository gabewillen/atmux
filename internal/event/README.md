# package event

`import "github.com/stateforward/amux/internal/event"`

Package event provides event dispatch interfaces for amux per spec §9.

Phase 0: Provides stable interfaces with local implementations.
Phase 7 will add full hsmnet network-aware dispatch.

- `type BasicEvent` — BasicEvent is a simple Event implementation used for local dispatch.
- `type Dispatcher` — Dispatcher is the interface for event dispatch.
- `type EventFilter` — EventFilter filters events by type or other criteria.
- `type Event` — Event represents a generic event in the system.
- `type TypeFilter` — TypeFilter matches events by type prefix.
- `type localDispatcher` — localDispatcher is an in-memory dispatcher suitable for single-process tests.
- `type subscriber` — subscriber represents a single subscription on the local dispatcher.

## type BasicEvent

```go
type BasicEvent struct {
	EventType string
	Payload   any
}
```

BasicEvent is a simple Event implementation used for local dispatch.

### Methods

#### BasicEvent.Data

```go
func () Data() any
```

Data implements Event.Data.

#### BasicEvent.Type

```go
func () Type() string
```

Type implements Event.Type.


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
This is a Phase 0 implementation that keeps all dispatch in-process.
Phase 7 will replace this with a network-aware dispatcher behind the same interface.


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

## type TypeFilter

```go
type TypeFilter struct {
	// Prefix is matched against the event type using strings.HasPrefix.
	// An empty prefix matches all events.
	Prefix string
}
```

TypeFilter matches events by type prefix.

### Methods

#### TypeFilter.Matches

```go
func () Matches(event Event) bool
```

Matches returns true if the event type starts with the configured prefix.


## type localDispatcher

```go
type localDispatcher struct {
	mu          sync.RWMutex
	subscribers []subscriber
}
```

localDispatcher is an in-memory dispatcher suitable for single-process tests.
It is safe for concurrent use by multiple goroutines.

### Methods

#### localDispatcher.Close

```go
func () Close() error
```

Close closes all subscriber channels and clears internal state.

#### localDispatcher.Dispatch

```go
func () Dispatch(ctx context.Context, event Event) error
```

Dispatch sends an event to all matching subscribers.

#### localDispatcher.Subscribe

```go
func () Subscribe(ctx context.Context, filter EventFilter) (<-chan Event, error)
```

Subscribe registers a new subscriber with the given filter.
The returned channel is buffered to reduce the risk of blocking producers.


## type subscriber

```go
type subscriber struct {
	filter EventFilter
	ch     chan Event
}
```

subscriber represents a single subscription on the local dispatcher.

