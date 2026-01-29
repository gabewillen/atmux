# package conn

`import "github.com/agentflare-ai/amux/internal/remote/conn"`

- `func Connect(opts Options) (*nats.Conn, error)` — Connect establishes a connection to NATS.
- `func JetStream(nc *nats.Conn) (jetstream.JetStream, error)` — JetStream returns a JetStream context from a NATS connection.
- `type Options` — Options configuration for NATS connection.

### Functions

#### Connect

```go
func Connect(opts Options) (*nats.Conn, error)
```

Connect establishes a connection to NATS.

#### JetStream

```go
func JetStream(nc *nats.Conn) (jetstream.JetStream, error)
```

JetStream returns a JetStream context from a NATS connection.


## type Options

```go
type Options struct {
	URL           string
	Name          string
	CredsPath     string
	ReconnectWait time.Duration
	MaxReconnects int
	OnDisconnect  func(*nats.Conn, error)
	OnReconnect   func(*nats.Conn)
	OnClosed      func(*nats.Conn)
}
```

Options configuration for NATS connection.

