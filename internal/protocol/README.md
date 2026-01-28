# package protocol

`import "github.com/agentflare-ai/amux/internal/protocol"`

Package protocol defines the event transport interfaces for amux.

All event dispatch flows through NATS-backed implementations.

- `ErrNoopDispatcher` — ErrNoopDispatcher is returned when using a noop dispatcher.
- `func Subject(parts ...string) string` — Subject joins subject segments for NATS routing.
- `type Dispatcher` — Dispatcher publishes and subscribes to events over NATS.
- `type Event` — Event is the generic event envelope used for dispatch.
- `type NoopDispatcher` — NoopDispatcher is a placeholder dispatcher that drops all events.
- `type Subscription` — Subscription represents an active event subscription.
- `type noopSub`

### Variables

#### ErrNoopDispatcher

```go
var ErrNoopDispatcher = errors.New("noop dispatcher")
```

ErrNoopDispatcher is returned when using a noop dispatcher.


### Functions

#### Subject

```go
func Subject(parts ...string) string
```

Subject joins subject segments for NATS routing.


## type Dispatcher

```go
type Dispatcher interface {
	Publish(ctx context.Context, subject string, event Event) error
	Subscribe(ctx context.Context, subject string, handler func(Event)) (Subscription, error)
}
```

Dispatcher publishes and subscribes to events over NATS.

## type Event

```go
type Event struct {
	Name       string
	Payload    any
	OccurredAt time.Time
}
```

Event is the generic event envelope used for dispatch.

## type NoopDispatcher

```go
type NoopDispatcher struct{}
```

NoopDispatcher is a placeholder dispatcher that drops all events.

### Methods

#### NoopDispatcher.Publish

```go
func () Publish(ctx context.Context, subject string, event Event) error
```

Publish drops the event.

#### NoopDispatcher.Subscribe

```go
func () Subscribe(ctx context.Context, subject string, handler func(Event)) (Subscription, error)
```

Subscribe returns a noop subscription.


## type Subscription

```go
type Subscription interface {
	Unsubscribe() error
}
```

Subscription represents an active event subscription.

## type noopSub

```go
type noopSub struct{}
```

### Methods

#### noopSub.Unsubscribe

```go
func () Unsubscribe() error
```


