# package protocol

`import "github.com/agentflare-ai/amux/internal/protocol"`

Package protocol provides event dispatch and routing interfaces.
Phase 0 introduces stable interfaces with noop/local implementations.
Phase 7 will provide full network-aware routing via NATS.

- `func matchFilter(filter EventFilter, event Event) bool`
- `type Dispatcher` — Dispatcher provides event dispatch functionality.
- `type EventFilter` — EventFilter defines criteria for event subscription.
- `type Event` — Event represents a generic event in the system.
- `type MessageRouter` — MessageRouter routes inter-agent messages (spec §6.4).
- `type localDispatcher` — localDispatcher is a Phase 0 local in-memory event dispatcher (spec §6.2, §6.3).
- `type messageRouter` — messageRouter is the Phase 4 local implementation: dispatches message.inbound (spec §6.4.2).

### Functions

#### matchFilter

```go
func matchFilter(filter EventFilter, event Event) bool
```


## type Dispatcher

```go
type Dispatcher interface {
	// Dispatch dispatches an event to all subscribers whose filter matches.
	Dispatch(ctx context.Context, event Event) error

	// Subscribe subscribes to events matching the given filter.
	Subscribe(filter EventFilter) (<-chan Event, func())
}
```

Dispatcher provides event dispatch functionality.
Phase 0: Local in-memory implementation
Phase 7: Full network-aware routing

### Functions returning Dispatcher

#### NewDispatcher

```go
func NewDispatcher() Dispatcher
```

NewDispatcher creates a new event dispatcher.
Phase 0: Returns a local in-memory dispatcher that delivers to subscribers.


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

## type MessageRouter

```go
type MessageRouter interface {
	// Route delivers the message to the recipient(s); for local Phase 4 implementation
	// this dispatches a message.inbound event so subscribers can deliver to PTYs.
	Route(ctx context.Context, msg interface{}) error
}
```

MessageRouter routes inter-agent messages (spec §6.4).
Phase 4: Local delivery via Dispatch message.inbound.
Phase 7: NATS P.comm.* subjects.

### Functions returning MessageRouter

#### NewMessageRouter

```go
func NewMessageRouter(d Dispatcher) MessageRouter
```

NewMessageRouter creates a local message router that dispatches message.inbound (spec §6.4).


## type localDispatcher

```go
type localDispatcher struct {
	mu   sync.RWMutex
	subs map[chan Event]EventFilter
}
```

localDispatcher is a Phase 0 local in-memory event dispatcher (spec §6.2, §6.3).

### Methods

#### localDispatcher.Dispatch

```go
func () Dispatch(ctx context.Context, event Event) error
```

Dispatch delivers the event to all subscribers whose filter matches (spec §6.2 presence.changed, roster.updated).

#### localDispatcher.Subscribe

```go
func () Subscribe(filter EventFilter) (<-chan Event, func())
```


## type messageRouter

```go
type messageRouter struct {
	Dispatcher Dispatcher
}
```

messageRouter is the Phase 4 local implementation: dispatches message.inbound (spec §6.4.2).

### Methods

#### messageRouter.Route

```go
func () Route(ctx context.Context, msg interface{}) error
```


