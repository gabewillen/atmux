# package event

`import "github.com/agentflare-ai/amux/internal/event"`

Package event provides stable interfaces for event system (to be fully implemented in Phase 7).

- `type Dispatcher` — Dispatcher handles event routing and delivery.
- `type EventType` — EventType represents the type of an event.
- `type Event` — Event represents a system event with typed payload.
- `type HandlerFunc` — HandlerFunc adapts a function to Handler interface.
- `type Handler` — Handler processes incoming events.
- `type NoopDispatcher` — NoopDispatcher provides a no-op implementation for Phase 0.

## type Dispatcher

```go
type Dispatcher interface {
	// Emit publishes an event to the event system
	Emit(ctx context.Context, event *Event) error

	// Subscribe registers for events of specific type(s)
	Subscribe(ctx context.Context, eventTypes []EventType, handler Handler) error

	// Unsubscribe removes an event subscription
	Unsubscribe(handler Handler) error

	// Shutdown gracefully shuts down the dispatcher
	Shutdown(ctx context.Context) error
}
```

Dispatcher handles event routing and delivery.

## type Event

```go
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Source    string                 `json:"source"`
	Timestamp int64                  `json:"timestamp"` // Unix timestamp
	Data      map[string]interface{} `json:"data"`
}
```

Event represents a system event with typed payload.

## type EventType

```go
type EventType string
```

EventType represents the type of an event.

### Constants

#### EventAgentSpawned, EventAgentStarted, EventAgentStopped, EventAgentTerminated, EventAgentErrored, EventConnectionEstablished, EventConnectionLost, EventProcessSpawned, EventProcessExited, EventProcessIO, EventPTYActivity, EventPTYPattern, EventPTYDecoded, EventNotificationBatch, EventNotificationError

```go
const (
	// Agent lifecycle events
	EventAgentSpawned    EventType = "agent.spawned"
	EventAgentStarted    EventType = "agent.started"
	EventAgentStopped    EventType = "agent.stopped"
	EventAgentTerminated EventType = "agent.terminated"
	EventAgentErrored    EventType = "agent.errored"

	// Connection events
	EventConnectionEstablished EventType = "connection.established"
	EventConnectionLost        EventType = "connection.lost"

	// Process events
	EventProcessSpawned EventType = "process.spawned"
	EventProcessExited  EventType = "process.exited"
	EventProcessIO      EventType = "process.io"

	// PTY events
	EventPTYActivity EventType = "pty.activity"
	EventPTYPattern  EventType = "pty.pattern"
	EventPTYDecoded  EventType = "pty.decoded"

	// Notification events
	EventNotificationBatch EventType = "notification.batch"
	EventNotificationError EventType = "notification.error"
)
```


## type Handler

```go
type Handler interface {
	// Handle processes an event
	Handle(ctx context.Context, event *Event) error

	// HandlerID returns unique identifier for this handler
	HandlerID() string
}
```

Handler processes incoming events.

## type HandlerFunc

```go
type HandlerFunc struct {
	id      string
	handler func(context.Context, *Event) error
}
```

HandlerFunc adapts a function to Handler interface.

### Functions returning HandlerFunc

#### NewHandlerFunc

```go
func NewHandlerFunc(id string, fn func(context.Context, *Event) error) *HandlerFunc
```

NewHandlerFunc creates a handler from a function.


### Methods

#### HandlerFunc.Handle

```go
func () Handle(ctx context.Context, event *Event) error
```

Handle implements Handler interface.

#### HandlerFunc.HandlerID

```go
func () HandlerID() string
```

HandlerID implements Handler interface.


## type NoopDispatcher

```go
type NoopDispatcher struct {
	handlers map[string]Handler
}
```

NoopDispatcher provides a no-op implementation for Phase 0.

### Functions returning NoopDispatcher

#### NewNoopDispatcher

```go
func NewNoopDispatcher() *NoopDispatcher
```

NewNoopDispatcher creates a new no-op dispatcher.


### Methods

#### NoopDispatcher.Emit

```go
func () Emit(ctx context.Context, event *Event) error
```

Emit implements Dispatcher interface (no-op).

#### NoopDispatcher.Shutdown

```go
func () Shutdown(ctx context.Context) error
```

Shutdown implements Dispatcher interface.

#### NoopDispatcher.Subscribe

```go
func () Subscribe(ctx context.Context, eventTypes []EventType, handler Handler) error
```

Subscribe implements Dispatcher interface.

#### NoopDispatcher.Unsubscribe

```go
func () Unsubscribe(handler Handler) error
```

Unsubscribe implements Dispatcher interface.


