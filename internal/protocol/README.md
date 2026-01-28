# package protocol

`import "github.com/copilot-claude-sonnet-4/amux/internal/protocol"`

Package protocol provides remote communication transport functionality.
This package transports events generically using NATS without any
agent-specific knowledge.

- `ErrConnectionFailed, ErrPublishFailed, ErrSubscribeFailed` — Common sentinel errors for protocol operations.
- `type TransportConfig` — TransportConfig holds NATS transport configuration.
- `type Transport` — Transport manages NATS-based event communication.

### Variables

#### ErrConnectionFailed, ErrPublishFailed, ErrSubscribeFailed

```go
var (
	// ErrConnectionFailed indicates NATS connection failed.
	ErrConnectionFailed = errors.New("connection failed")

	// ErrPublishFailed indicates event publishing failed.
	ErrPublishFailed = errors.New("publish failed")

	// ErrSubscribeFailed indicates subscription setup failed.
	ErrSubscribeFailed = errors.New("subscribe failed")
)
```

Common sentinel errors for protocol operations.


## type Transport

```go
type Transport struct {
	conn   *nats.Conn
	config TransportConfig
}
```

Transport manages NATS-based event communication.
Transports events generically for both local and remote distribution.

### Functions returning Transport

#### NewTransport

```go
func NewTransport(config TransportConfig) (*Transport, error)
```

NewTransport creates a new NATS transport instance.


### Methods

#### Transport.Close

```go
func () Close()
```

Close closes the NATS connection.

#### Transport.Connect

```go
func () Connect() error
```

Connect establishes NATS connection with configured options.

#### Transport.Publish

```go
func () Publish(subject string, data []byte) error
```

Publish sends an event to the specified NATS subject.

#### Transport.Subscribe

```go
func () Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error)
```

Subscribe sets up a subscription to the specified NATS subject.


## type TransportConfig

```go
type TransportConfig struct {
	URL             string
	ConnectionName  string
	ReconnectDelay  time.Duration
	MaxReconnects   int
	CredentialsFile string
}
```

TransportConfig holds NATS transport configuration.

