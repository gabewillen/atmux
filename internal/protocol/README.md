# package protocol

`import "github.com/agentflare-ai/amux/internal/protocol"`

Package protocol defines the event transport interfaces for amux.

All event dispatch flows through NATS-backed implementations.

- `ErrNATSNotConnected, ErrNATSProtocol`
- `errAuthRequired`
- `func Subject(parts ...string) string` — Subject joins subject segments for NATS routing.
- `func decodeConnect(payload string, dest *map[string]any) error`
- `func jsonUnmarshalStrict(data []byte, dest *map[string]any) error`
- `func matchSubject(subject string, pattern string) bool`
- `func parseLength(raw string) (int, error)`
- `func parseMsgLine(line string) (string, string, int, error)`
- `func parseReplyFromHeader(line string) string`
- `func subjectAllowed(subject string, patterns []string) bool`
- `type AuthConfig` — AuthConfig maps auth credentials to subject permissions.
- `type ConnectInfo` — ConnectInfo captures CONNECT payload fields used for auth.
- `type Dispatcher` — Dispatcher publishes and subscribes to events over NATS.
- `type EmbeddedServerConfig` — EmbeddedServerConfig configures the embedded NATS-compatible server.
- `type EmbeddedServer` — EmbeddedServer provides a minimal NATS-compatible server for local use.
- `type Event` — Event is the generic event envelope used for dispatch.
- `type Message` — Message is a raw NATS message payload with optional reply subject.
- `type NATSDispatcher` — NATSDispatcher publishes and subscribes to events over NATS.
- `type NATSOptions` — NATSOptions configures NATS connection metadata and auth.
- `type Permissions` — Permissions defines publish and subscribe authorizations.
- `type Subscription` — Subscription represents an active event subscription.
- `type UserAuth` — UserAuth defines a username/password and permissions pair.
- `type connState`
- `type natsSubscription`
- `type subscriptionHandler`
- `type subscription`

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

#### errAuthRequired

```go
var errAuthRequired = errors.New("auth required")
```


### Functions

#### Subject

```go
func Subject(parts ...string) string
```

Subject joins subject segments for NATS routing.

#### decodeConnect

```go
func decodeConnect(payload string, dest *map[string]any) error
```

#### jsonUnmarshalStrict

```go
func jsonUnmarshalStrict(data []byte, dest *map[string]any) error
```

#### matchSubject

```go
func matchSubject(subject string, pattern string) bool
```

#### parseLength

```go
func parseLength(raw string) (int, error)
```

#### parseMsgLine

```go
func parseMsgLine(line string) (string, string, int, error)
```

#### parseReplyFromHeader

```go
func parseReplyFromHeader(line string) string
```

#### subjectAllowed

```go
func subjectAllowed(subject string, patterns []string) bool
```


## type AuthConfig

```go
type AuthConfig struct {
	Tokens map[string]Permissions
	Users  map[string]UserAuth
}
```

AuthConfig maps auth credentials to subject permissions.

### Methods

#### AuthConfig.Authorize

```go
func () Authorize(info ConnectInfo) (Permissions, error)
```


## type ConnectInfo

```go
type ConnectInfo struct {
	Token    string
	User     string
	Password string
	Name     string
}
```

ConnectInfo captures CONNECT payload fields used for auth.

## type Dispatcher

```go
type Dispatcher interface {
	Publish(ctx context.Context, subject string, event Event) error
	Subscribe(ctx context.Context, subject string, handler func(Event)) (Subscription, error)
	PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error
	SubscribeRaw(ctx context.Context, subject string, handler func(Message)) (Subscription, error)
	Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (Message, error)
	MaxPayload() int
	Closed() <-chan struct{}
}
```

Dispatcher publishes and subscribes to events over NATS.

## type EmbeddedServer

```go
type EmbeddedServer struct {
	listener   net.Listener
	mu         sync.Mutex
	closed     bool
	subs       map[*connState]map[string]*subscription
	maxPayload int
	auth       AuthConfig
}
```

EmbeddedServer provides a minimal NATS-compatible server for local use.

### Functions returning EmbeddedServer

#### StartEmbeddedServer

```go
func StartEmbeddedServer(ctx context.Context, addr string, cfg EmbeddedServerConfig) (*EmbeddedServer, error)
```

StartEmbeddedServer starts a local NATS-compatible server.


### Methods

#### EmbeddedServer.Close

```go
func () Close() error
```

Close stops the embedded server.

#### EmbeddedServer.URL

```go
func () URL() string
```

URL returns the nats:// URL for the embedded server.

#### EmbeddedServer.acceptLoop

```go
func () acceptLoop(ctx context.Context)
```

#### EmbeddedServer.handleConn

```go
func () handleConn(ctx context.Context, state *connState)
```

#### EmbeddedServer.handleConnect

```go
func () handleConnect(state *connState, line string)
```

#### EmbeddedServer.handlePub

```go
func () handlePub(state *connState, line string)
```

#### EmbeddedServer.handleSub

```go
func () handleSub(state *connState, line string)
```

#### EmbeddedServer.handleUnsub

```go
func () handleUnsub(state *connState, line string)
```

#### EmbeddedServer.publish

```go
func () publish(subject, reply string, payload []byte)
```

#### EmbeddedServer.removeConn

```go
func () removeConn(state *connState)
```


## type EmbeddedServerConfig

```go
type EmbeddedServerConfig struct {
	// MaxPayload sets the advertised max payload size.
	MaxPayload int
	// Auth configures per-connection authorization.
	Auth AuthConfig
}
```

EmbeddedServerConfig configures the embedded NATS-compatible server.

## type Event

```go
type Event struct {
	Name       string
	Payload    any
	OccurredAt time.Time
}
```

Event is the generic event envelope used for dispatch.

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
	conn       net.Conn
	reader     *bufio.Reader
	writer     *bufio.Writer
	mu         sync.Mutex
	subs       map[string]subscriptionHandler
	closed     bool
	closedCh   chan struct{}
	nextSID    uint64
	nextInbox  uint64
	maxPayload int
	options    NATSOptions
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

Closed returns a channel closed when the dispatcher connection ends.

#### NATSDispatcher.MaxPayload

```go
func () MaxPayload() int
```

MaxPayload returns the server-advertised maximum payload size.

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

Request sends a request and waits for a single reply.

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

#### NATSDispatcher.handlerForSID

```go
func () handlerForSID(sid string) subscriptionHandler
```

#### NATSDispatcher.nextInboxSubject

```go
func () nextInboxSubject() string
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
func () writePong() error
```


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
}
```

NATSOptions configures NATS connection metadata and auth.

## type Permissions

```go
type Permissions struct {
	Publish   []string
	Subscribe []string
}
```

Permissions defines publish and subscribe authorizations.

## type Subscription

```go
type Subscription interface {
	Unsubscribe() error
}
```

Subscription represents an active event subscription.

## type UserAuth

```go
type UserAuth struct {
	Password    string
	Permissions Permissions
}
```

UserAuth defines a username/password and permissions pair.

## type connState

```go
type connState struct {
	conn       net.Conn
	reader     *bufio.Reader
	writer     *bufio.Writer
	mu         sync.Mutex
	perms      Permissions
	authorized bool
}
```

### Methods

#### connState.sendMessage

```go
func () sendMessage(subject, sid, reply string, payload []byte)
```

#### connState.writeLine

```go
func () writeLine(line string) error
```


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


## type subscription

```go
type subscription struct {
	subject string
	sid     string
}
```

## type subscriptionHandler

```go
type subscriptionHandler struct {
	onEvent func(Event)
	onRaw   func(Message)
}
```

