# package protocol

`import "github.com/agentflare-ai/amux/internal/protocol"`

Package protocol defines the event transport interfaces for amux.

All event dispatch flows through NATS-backed implementations.

- `ErrNATSNotConnected, ErrNATSProtocol`
- `func Subject(parts ...string) string` — Subject joins subject segments for NATS routing.
- `func parseMsgLine(line string) (string, string, int, error)`
- `type Dispatcher` — Dispatcher publishes and subscribes to events over NATS.
- `type Event` — Event is the generic event envelope used for dispatch.
- `type NATSDispatcher` — NATSDispatcher publishes and subscribes to events over NATS.
- `type Subscription` — Subscription represents an active event subscription.
- `type natsSubscription`

### Variables

#### ErrNATSNotConnected, ErrNATSProtocol

```go
var (
	// ErrNATSNotConnected is returned when the dispatcher is not connected.
	ErrNATSNotConnected = errors.New("nats not connected")
	// ErrNATSProtocol is returned for malformed NATS protocol frames.
	ErrNATSProtocol = errors.New("nats protocol error")
)
```


### Functions

#### Subject

```go
func Subject(parts ...string) string
```

Subject joins subject segments for NATS routing.

#### parseMsgLine

```go
func parseMsgLine(line string) (string, string, int, error)
```


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

## type NATSDispatcher

```go
type NATSDispatcher struct {
	conn    net.Conn
	reader  *bufio.Reader
	writer  *bufio.Writer
	mu      sync.Mutex
	subs    map[string]func(Event)
	closed  bool
	nextSID uint64
}
```

NATSDispatcher publishes and subscribes to events over NATS.

### Functions returning NATSDispatcher

#### NewNATSDispatcher

```go
func NewNATSDispatcher(ctx context.Context, rawURL string) (*NATSDispatcher, error)
```

NewNATSDispatcher connects to a NATS server and returns a dispatcher.


### Methods

#### NATSDispatcher.Close

```go
func () Close(ctx context.Context) error
```

Close closes the underlying NATS connection.

#### NATSDispatcher.Publish

```go
func () Publish(ctx context.Context, subject string, event Event) error
```

Publish publishes an event to a subject.

#### NATSDispatcher.Subscribe

```go
func () Subscribe(ctx context.Context, subject string, handler func(Event)) (Subscription, error)
```

Subscribe subscribes to a subject.

#### NATSDispatcher.handlerForSID

```go
func () handlerForSID(sid string) func(Event)
```

#### NATSDispatcher.readInfo

```go
func () readInfo(ctx context.Context) error
```

#### NATSDispatcher.readLoop

```go
func () readLoop()
```

#### NATSDispatcher.sendConnect

```go
func () sendConnect(ctx context.Context) error
```

#### NATSDispatcher.unsubscribe

```go
func () unsubscribe(sid string) error
```

#### NATSDispatcher.writePong

```go
func () writePong()
```


## type Subscription

```go
type Subscription interface {
	Unsubscribe() error
}
```

Subscription represents an active event subscription.

## type natsSubscription

```go
type natsSubscription struct {
	dispatcher *NATSDispatcher
	sid        string
}
```

### Methods

#### natsSubscription.Unsubscribe

```go
func () Unsubscribe() error
```


