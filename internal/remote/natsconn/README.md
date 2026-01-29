# package natsconn

`import "github.com/agentflare-ai/amux/internal/remote/natsconn"`

Package natsconn provides NATS connection management for amux.

This package handles connecting to NATS as either a director (hub)
or manager (leaf) role, with support for credential-based authentication,
JetStream initialization, and reconnection handling.

See spec §5.5.6 for NATS connectivity requirements.

Package natsconn - kv.go provides JetStream Key-Value operations
for durable remote control-plane state.

See spec §5.5.6.3 for KV bucket requirements.

- `func NkeyOptionFromSeed(seed string) (nats.Option, error)` — NkeyOptionFromSeed is a helper that creates a NATS NKey option from a seed.
- `type Conn` — Conn wraps a NATS connection with amux-specific configuration.
- `type HostHeartbeat` — HostHeartbeat holds the last-seen heartbeat timestamp.
- `type HostInfo` — HostInfo holds host metadata stored in JetStream KV.
- `type KVStore` — KVStore wraps a JetStream KV bucket for amux remote state.
- `type Options` — Options configures a NATS connection.
- `type SessionMeta` — SessionMeta holds session metadata sufficient for reconnection.

### Functions

#### NkeyOptionFromSeed

```go
func NkeyOptionFromSeed(seed string) (nats.Option, error)
```

NkeyOptionFromSeed is a helper that creates a NATS NKey option from a seed.


## type Conn

```go
type Conn struct {
	nc     *nats.Conn
	js     jetstream.JetStream
	hostID string
	role   string
}
```

Conn wraps a NATS connection with amux-specific configuration.

### Functions returning Conn

#### Connect

```go
func Connect(ctx context.Context, opts *Options) (*Conn, error)
```

Connect establishes a NATS connection.


### Methods

#### Conn.Close

```go
func () Close() error
```

Close gracefully drains and closes the NATS connection.

#### Conn.Flush

```go
func () Flush() error
```

Flush flushes the connection buffer.

#### Conn.HostID

```go
func () HostID() string
```

HostID returns the host identifier for this connection.

#### Conn.IsConnected

```go
func () IsConnected() bool
```

IsConnected returns true if the NATS connection is currently active.

#### Conn.JetStream

```go
func () JetStream() (jetstream.JetStream, error)
```

JetStream returns the JetStream context, initializing it on first call.

#### Conn.NC

```go
func () NC() *nats.Conn
```

NC returns the underlying NATS connection.

#### Conn.Publish

```go
func () Publish(subject string, data []byte) error
```

Publish publishes a message to a NATS subject.

#### Conn.Request

```go
func () Request(subject string, data []byte, timeout time.Duration) (*nats.Msg, error)
```

Request sends a request and waits for a reply.

#### Conn.Role

```go
func () Role() string
```

Role returns the role for this connection.

#### Conn.Subscribe

```go
func () Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error)
```

Subscribe subscribes to a NATS subject.


## type HostHeartbeat

```go
type HostHeartbeat struct {
	Timestamp string `json:"timestamp"`
}
```

HostHeartbeat holds the last-seen heartbeat timestamp.

## type HostInfo

```go
type HostInfo struct {
	Version   string `json:"version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	PeerID    string `json:"peer_id"`
	StartedAt string `json:"started_at"`
}
```

HostInfo holds host metadata stored in JetStream KV.

## type KVStore

```go
type KVStore struct {
	kv     jetstream.KeyValue
	bucket string
}
```

KVStore wraps a JetStream KV bucket for amux remote state.

### Functions returning KVStore

#### GetKV

```go
func GetKV(ctx context.Context, conn *Conn, bucket string) (*KVStore, error)
```

GetKV connects to an existing JetStream KV bucket.

#### InitKV

```go
func InitKV(ctx context.Context, conn *Conn, bucket string) (*KVStore, error)
```

InitKV initializes the JetStream KV bucket, creating it if it does not exist.

Per spec §5.5.6.3: "The director MUST create the bucket if it does not exist."


### Methods

#### KVStore.Bucket

```go
func () Bucket() string
```

Bucket returns the bucket name.

#### KVStore.DeleteSessionMeta

```go
func () DeleteSessionMeta(ctx context.Context, hostID, sessionID string) error
```

DeleteSessionMeta removes session metadata from KV.

#### KVStore.GetHeartbeat

```go
func () GetHeartbeat(ctx context.Context, hostID string) (*HostHeartbeat, error)
```

GetHeartbeat reads the last heartbeat timestamp from KV.

#### KVStore.GetHostInfo

```go
func () GetHostInfo(ctx context.Context, hostID string) (*HostInfo, error)
```

GetHostInfo reads host metadata from KV.

#### KVStore.GetSessionMeta

```go
func () GetSessionMeta(ctx context.Context, hostID, sessionID string) (*SessionMeta, error)
```

GetSessionMeta reads session metadata from KV.

#### KVStore.PutHeartbeat

```go
func () PutHeartbeat(ctx context.Context, hostID string) error
```

PutHeartbeat writes a heartbeat timestamp to KV.
Key: hosts/<host_id>/heartbeat

#### KVStore.PutHostInfo

```go
func () PutHostInfo(ctx context.Context, hostID string, info *HostInfo) error
```

PutHostInfo writes host metadata to KV.
Key: hosts/<host_id>/info

#### KVStore.PutSessionMeta

```go
func () PutSessionMeta(ctx context.Context, hostID, sessionID string, meta *SessionMeta) error
```

PutSessionMeta writes session metadata to KV.
Key: sessions.<host_id>.<session_id>


## type Options

```go
type Options struct {
	// URL is the NATS server URL to connect to.
	URL string

	// CredsFile is the path to a NATS credentials file (NKey seed).
	CredsFile string

	// NKeySeed is the raw NKey seed bytes (alternative to CredsFile).
	NKeySeed []byte

	// HostID is this node's host identifier.
	HostID string

	// Role is "director" or "manager".
	Role string

	// Name is a connection name for debugging.
	Name string

	// ReconnectWait is the time to wait between reconnect attempts.
	ReconnectWait time.Duration

	// MaxReconnects is the maximum number of reconnect attempts.
	// -1 means unlimited.
	MaxReconnects int

	// DisconnectHandler is called when the connection is lost.
	DisconnectHandler func(*nats.Conn, error)

	// ReconnectHandler is called when the connection is restored.
	ReconnectHandler func(*nats.Conn)

	// ClosedHandler is called when the connection is permanently closed.
	ClosedHandler func(*nats.Conn)
}
```

Options configures a NATS connection.

### Functions returning Options

#### OptionsFromConfig

```go
func OptionsFromConfig(cfg *config.Config, hostID string) *Options
```

OptionsFromConfig creates Options from the amux configuration.


## type SessionMeta

```go
type SessionMeta struct {
	AgentID   string `json:"agent_id"`
	AgentSlug string `json:"agent_slug"`
	RepoPath  string `json:"repo_path"`
	State     string `json:"state"`
}
```

SessionMeta holds session metadata sufficient for reconnection.

