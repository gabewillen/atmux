# package protocol

`import "github.com/agentflare-ai/amux/internal/protocol"`

Package protocol defines the event transport interfaces for amux.

All event dispatch flows through NATS-backed implementations.

- `func Subject(parts ...string) string` — Subject joins subject segments for NATS routing.
- `func allocatePort(host string) (int, error)`
- `func buildLeafURL(host string, port int, advertise string, tlsEnabled bool) string`
- `func closeOnce(ch chan struct{})`
- `func parsePort(raw string) (int, error)`
- `func resolveLeafListen(cfg HubServerConfig, listenHost string, listenPort int) (string, int, error)`
- `func splitHostPort(addr string) (string, int, error)`
- `type Dispatcher` — Dispatcher publishes and subscribes to events over NATS.
- `type Event` — Event is the generic event envelope used for dispatch.
- `type HubServerConfig` — HubServerConfig configures the embedded hub-mode NATS server.
- `type LeafServerConfig` — LeafServerConfig configures the embedded leaf-mode NATS server.
- `type Message` — Message is a raw NATS message payload with optional reply subject.
- `type NATSDispatcher` — NATSDispatcher publishes and subscribes to events over NATS.
- `type NATSOptions` — NATSOptions configures NATS connection metadata and auth.
- `type NATSServer` — NATSServer wraps a running NATS server instance.
- `type Subscription` — Subscription represents an active event subscription.
- `type natsSubscription`

### Functions

#### Subject

```go
func Subject(parts ...string) string
```

Subject joins subject segments for NATS routing.

#### allocatePort

```go
func allocatePort(host string) (int, error)
```

#### buildLeafURL

```go
func buildLeafURL(host string, port int, advertise string, tlsEnabled bool) string
```

#### closeOnce

```go
func closeOnce(ch chan struct{})
```

#### parsePort

```go
func parsePort(raw string) (int, error)
```

#### resolveLeafListen

```go
func resolveLeafListen(cfg HubServerConfig, listenHost string, listenPort int) (string, int, error)
```

#### splitHostPort

```go
func splitHostPort(addr string) (string, int, error)
```


## type Dispatcher

```go
type Dispatcher interface {
	Publish(ctx context.Context, subject string, event Event) error
	Subscribe(ctx context.Context, subject string, handler func(Event)) (Subscription, error)
	PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error
	SubscribeRaw(ctx context.Context, subject string, handler func(Message)) (Subscription, error)
	Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (Message, error)
	MaxPayload() int
	JetStream() nats.JetStreamContext
	Closed() <-chan struct{}
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

## type HubServerConfig

```go
type HubServerConfig struct {
	Listen    string
	Advertise string
	// LeafListen is the leaf node listen address.
	LeafListen string
	// LeafAdvertiseURL is the advertised leaf node URL.
	LeafAdvertiseURL  string
	JetStreamDir      string
	OperatorPublicKey string
	SystemAccountKey  string
	SystemAccountJWT  string
	AccountPublicKey  string
	AccountJWT        string
}
```

HubServerConfig configures the embedded hub-mode NATS server.

## type LeafServerConfig

```go
type LeafServerConfig struct {
	Listen    string
	HubURL    string
	CredsPath string
}
```

LeafServerConfig configures the embedded leaf-mode NATS server.

## type Message

```go
type Message struct {
	Subject string
	Reply   string
	Data    []byte
}
```

Message is a raw NATS message payload with optional reply subject.

## type NATSDispatcher

```go
type NATSDispatcher struct {
	conn     *nats.Conn
	js       nats.JetStreamContext
	closedCh chan struct{}
}
```

NATSDispatcher publishes and subscribes to events over NATS.

### Functions returning NATSDispatcher

#### NewNATSDispatcher

```go
func NewNATSDispatcher(ctx context.Context, rawURL string, options NATSOptions) (*NATSDispatcher, error)
```

NewNATSDispatcher connects to a NATS server and returns a dispatcher.


### Methods

#### NATSDispatcher.Close

```go
func () Close(ctx context.Context) error
```

Close closes the underlying NATS connection.

#### NATSDispatcher.Closed

```go
func () Closed() <-chan struct{}
```

Closed returns a channel that closes when the connection closes.

#### NATSDispatcher.JetStream

```go
func () JetStream() nats.JetStreamContext
```

JetStream returns the JetStream context.

#### NATSDispatcher.MaxPayload

```go
func () MaxPayload() int
```

MaxPayload returns the maximum payload size for the connection.

#### NATSDispatcher.Publish

```go
func () Publish(ctx context.Context, subject string, event Event) error
```

Publish publishes an event to a subject.

#### NATSDispatcher.PublishRaw

```go
func () PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error
```

PublishRaw publishes a raw payload to a subject with optional reply.

#### NATSDispatcher.Request

```go
func () Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (Message, error)
```

Request sends a request and waits for a response.

#### NATSDispatcher.Subscribe

```go
func () Subscribe(ctx context.Context, subject string, handler func(Event)) (Subscription, error)
```

Subscribe subscribes to a subject.

#### NATSDispatcher.SubscribeRaw

```go
func () SubscribeRaw(ctx context.Context, subject string, handler func(Message)) (Subscription, error)
```

SubscribeRaw subscribes to a subject and receives raw NATS messages.


## type NATSOptions

```go
type NATSOptions struct {
	// Name sets the client name.
	Name string
	// User sets the username for auth.
	User string
	// Password sets the password for auth.
	Password string
	// Token sets the auth token.
	Token string
	// CredsPath sets the path to a NATS .creds file.
	CredsPath string
	// AllowNoJetStream permits connections without JetStream enabled.
	AllowNoJetStream bool
}
```

NATSOptions configures NATS connection metadata and auth.

## type NATSServer

```go
type NATSServer struct {
	server  *server.Server
	leafURL string
}
```

NATSServer wraps a running NATS server instance.

### Functions returning NATSServer

#### StartHubServer

```go
func StartHubServer(ctx context.Context, cfg HubServerConfig) (*NATSServer, error)
```

StartHubServer starts a hub-mode NATS server with JetStream enabled.

#### StartLeafServer

```go
func StartLeafServer(ctx context.Context, cfg LeafServerConfig) (*NATSServer, error)
```

StartLeafServer starts a leaf-mode NATS server connected to the hub.


### Methods

#### NATSServer.Close

```go
func () Close() error
```

Close stops the server.

#### NATSServer.LeafCount

```go
func () LeafCount() int
```

LeafCount reports the number of leaf connections.

#### NATSServer.LeafURL

```go
func () LeafURL() string
```

LeafURL returns the leaf connection URL for the server.

#### NATSServer.Shutdown

```go
func () Shutdown()
```

Shutdown stops the server.

#### NATSServer.URL

```go
func () URL() string
```

URL returns the client connection URL for the server.


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
	sub *nats.Subscription
}
```

### Methods

#### natsSubscription.Unsubscribe

```go
func () Unsubscribe() error
```


