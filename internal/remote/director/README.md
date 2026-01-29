# package director

`import "github.com/agentflare-ai/amux/internal/remote/director"`

Package director implements the director-side remote orchestration for amux.

The director is responsible for:
  - Managing remote host connections via NATS hub
  - Tracking connected hosts and their state
  - Processing handshake exchanges with manager-role daemons
  - Routing control operations (spawn/kill/replay) to remote hosts
  - Subscribing to PTY output and publishing PTY input
  - Handling connection recovery and replay-before-live semantics

See spec §5.5 for remote agent architecture.

- `func extractHostIDFromSubject(subject, prefix string) string` — extractHostIDFromSubject extracts the host_id suffix from a NATS subject.
- `type Director` — Director manages remote host orchestration from the director-role node.
- `type HostState` — HostState tracks the state of a connected remote host.
- `type RemoteSession` — RemoteSession tracks a remote PTY session.

### Functions

#### extractHostIDFromSubject

```go
func extractHostIDFromSubject(subject, prefix string) string
```

extractHostIDFromSubject extracts the host_id suffix from a NATS subject.


## type Director

```go
type Director struct {
	mu     sync.RWMutex
	conn   *natsconn.Conn
	kv     *natsconn.KVStore
	cfg    *config.Config
	prefix string
	peerID string

	// hosts tracks connected remote hosts by host_id.
	hosts map[string]*HostState

	// sessions tracks active remote sessions by session_id.
	sessions map[string]*RemoteSession

	dispatcher event.Dispatcher

	// subs holds active NATS subscriptions for cleanup.
	subs []*nats.Subscription

	cancel context.CancelFunc
}
```

Director manages remote host orchestration from the director-role node.

### Functions returning Director

#### New

```go
func New(conn *natsconn.Conn, cfg *config.Config, dispatcher event.Dispatcher) *Director
```

New creates a new Director with the given NATS connection and configuration.


### Methods

#### Director.ConnectedHosts

```go
func () ConnectedHosts() []string
```

ConnectedHosts returns the list of currently connected host IDs.

#### Director.HostConnected

```go
func () HostConnected(hostID string) bool
```

HostConnected returns true if the given host has completed its handshake.

#### Director.Kill

```go
func () Kill(ctx context.Context, hostID string, sessionID string) (*protocol.KillResponse, error)
```

Kill sends a kill request to a remote host.

Per spec §5.5.7.2.1: the director MUST fail fast if the host is disconnected.

#### Director.PublishPTYInput

```go
func () PublishPTYInput(hostID, sessionID string, data []byte) error
```

PublishPTYInput sends PTY input to a session on a remote host.

#### Director.Replay

```go
func () Replay(ctx context.Context, hostID string, sessionID string) (*protocol.ReplayResponse, error)
```

Replay sends a replay request to a remote host.

Per spec §5.5.7.2.1: the director MUST fail fast if the host is disconnected.

#### Director.SendPing

```go
func () SendPing(hostID string) (*protocol.PongPayload, error)
```

SendPing sends a ping to a remote host via the control subject.

#### Director.SessionsForHost

```go
func () SessionsForHost(hostID string) []string
```

SessionsForHost returns the session IDs for a given host.

#### Director.SetHostDisconnected

```go
func () SetHostDisconnected(hostID string)
```

SetHostDisconnected marks a host as disconnected.
Called when the NATS connection to a host is lost.

#### Director.Spawn

```go
func () Spawn(ctx context.Context, hostID string, req *protocol.SpawnRequest) (*protocol.SpawnResponse, error)
```

Spawn sends a spawn request to a remote host.

Per spec §5.5.7.2.1: the director MUST fail fast if the host is disconnected.

#### Director.Start

```go
func () Start(ctx context.Context) error
```

Start begins listening for handshake requests and host events.

#### Director.Stop

```go
func () Stop() error
```

Stop gracefully shuts down the director.

#### Director.SubscribePTYOutput

```go
func () SubscribePTYOutput(hostID, sessionID string, handler func(data []byte)) error
```

SubscribePTYOutput subscribes to PTY output for a session on a remote host.
The handler receives raw PTY output bytes.

#### Director.handleHandshake

```go
func () handleHandshake(msg *nats.Msg)
```

handleHandshake processes a handshake request from a manager-role daemon.

Per spec §5.5.7.3: the director MUST treat the <host_id> token in the
request subject as canonical. If the handshake payload contains a different
host_id, the director MUST reject the handshake.

#### Director.handleHostEvent

```go
func () handleHostEvent(msg *nats.Msg)
```

handleHostEvent processes an event from a remote host.

#### Director.replyControl

```go
func () replyControl(msg *nats.Msg, msgType string, payload any)
```

replyControl sends a ControlMessage reply.

#### Director.replyError

```go
func () replyError(msg *nats.Msg, requestType, code, message string)
```

replyError sends an error ControlMessage reply.


## type HostState

```go
type HostState struct {
	// HostID is the unique host identifier.
	HostID string

	// PeerID is the remote daemon's peer identifier.
	PeerID string

	// Connected indicates whether the host is currently connected.
	Connected bool

	// HandshakeComplete indicates whether the handshake exchange is done.
	HandshakeComplete bool

	// ConnectedAt is when the host last connected.
	ConnectedAt time.Time

	// Sessions is the set of session IDs running on this host.
	Sessions map[string]bool
}
```

HostState tracks the state of a connected remote host.

## type RemoteSession

```go
type RemoteSession struct {
	SessionID string
	AgentID   string
	HostID    string
	AgentSlug string
	RepoPath  string

	// ptyOutSub is the subscription for PTY output from this session.
	ptyOutSub *nats.Subscription
}
```

RemoteSession tracks a remote PTY session.

