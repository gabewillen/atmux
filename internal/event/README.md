# package event

`import "github.com/stateforward/amux/internal/event"`

Package event implements a basic event dispatcher that can be used by other packages
This implementation will eventually be replaced with a full NATS-based implementation

- `func EmitEvent(ctx context.Context, eventType string, data map[string]interface{}) error` — EmitEvent is a convenience function to emit an event using the global dispatcher
- `func SubscribeToEvent(eventType string, handler EventHandler) error` — SubscribeToEvent is a convenience function to subscribe to an event type
- `func UnsubscribeFromEvent(eventType string, handler EventHandler) error` — UnsubscribeFromEvent is a convenience function to unsubscribe from an event type
- `type BasicDispatcher` — BasicDispatcher is a simple in-memory event dispatcher
- `type Dispatcher` — Dispatcher defines the interface for dispatching events
- `type EventHandler` — EventHandler is a function that handles an event
- `type Event` — Event represents a system event
- `type NoopDispatcher` — NoopDispatcher is a no-op implementation of the Dispatcher interface

### Functions

#### EmitEvent

```go
func EmitEvent(ctx context.Context, eventType string, data map[string]interface{}) error
```

EmitEvent is a convenience function to emit an event using the global dispatcher

#### SubscribeToEvent

```go
func SubscribeToEvent(eventType string, handler EventHandler) error
```

SubscribeToEvent is a convenience function to subscribe to an event type

#### UnsubscribeFromEvent

```go
func UnsubscribeFromEvent(eventType string, handler EventHandler) error
```

UnsubscribeFromEvent is a convenience function to unsubscribe from an event type


## type BasicDispatcher

```go
type BasicDispatcher struct {
	handlers map[string][]EventHandler
	mutex    sync.RWMutex
}
```

BasicDispatcher is a simple in-memory event dispatcher

### Functions returning BasicDispatcher

#### NewBasicDispatcher

```go
func NewBasicDispatcher() *BasicDispatcher
```

NewBasicDispatcher creates a new basic dispatcher


### Methods

#### BasicDispatcher.Emit

```go
func () Emit(ctx context.Context, event Event) error
```

Emit implements the Dispatcher interface

#### BasicDispatcher.Subscribe

```go
func () Subscribe(eventType string, handler EventHandler) error
```

Subscribe implements the Dispatcher interface

#### BasicDispatcher.Unsubscribe

```go
func () Unsubscribe(eventType string, handler EventHandler) error
```

Unsubscribe implements the Dispatcher interface


## type Dispatcher

```go
type Dispatcher interface {
	// Emit emits an event to all interested listeners
	Emit(ctx context.Context, event Event) error

	// Subscribe registers a listener for events of a specific type
	Subscribe(eventType string, handler EventHandler) error

	// Unsubscribe removes a listener for events of a specific type
	Unsubscribe(eventType string, handler EventHandler) error
}
```

Dispatcher defines the interface for dispatching events

### Variables

#### GlobalDispatcher

```go
var GlobalDispatcher Dispatcher = NewBasicDispatcher()
```

GlobalDispatcher is a global instance of the dispatcher that can be used by other packages


## type Event

```go
type Event struct {
	Type      string                 `json:"type"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}
```

Event represents a system event

## type EventHandler

```go
type EventHandler func(ctx context.Context, event Event) error
```

EventHandler is a function that handles an event

## type NoopDispatcher

```go
type NoopDispatcher struct{}
```

NoopDispatcher is a no-op implementation of the Dispatcher interface

### Functions returning NoopDispatcher

#### NewNoopDispatcher

```go
func NewNoopDispatcher() *NoopDispatcher
```

NewNoopDispatcher creates a new no-op dispatcher


### Methods

#### NoopDispatcher.Emit

```go
func () Emit(ctx context.Context, event Event) error
```

Emit implements the Dispatcher interface

#### NoopDispatcher.Subscribe

```go
func () Subscribe(eventType string, handler EventHandler) error
```

Subscribe implements the Dispatcher interface

#### NoopDispatcher.Unsubscribe

```go
func () Unsubscribe(eventType string, handler EventHandler) error
```

Unsubscribe implements the Dispatcher interface


