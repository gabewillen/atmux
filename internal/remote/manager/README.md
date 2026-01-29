# package manager

`import "github.com/agentflare-ai/amux/internal/remote/manager"`

Package manager implements the manager-role daemon for amux remote agents.

The manager runs on a remote host and manages PTY sessions for agents.
It connects to the director's hub NATS server via a leaf connection,
performs a handshake exchange, and handles control requests (spawn/kill/replay).

Key responsibilities:
  - Own PTYs on the host (one per agent)
  - Stream PTY output to the director over NATS
  - Receive PTY input from the director over NATS
  - Maintain per-session replay buffers
  - Handle connection recovery with replay-before-live semantics

See spec §5.5.4 and §5.5.5 for manager daemon requirements.

Package manager - outbound.go provides buffering for cross-host publications
during hub disconnection.

Per spec §5.5.8: the manager-role node SHOULD buffer outbound publications
while disconnected, up to a maximum queued payload size of remote.buffer_size
bytes total across all buffered publications. Oldest publications are dropped
first when the limit is exceeded. Per-subject publish order MUST be preserved.

Package manager - pty.go provides PTY creation for managed sessions.

- `Version` — Version is the manager daemon version string.
- `func extractSessionIDFromPTYSubject(subject, prefix, hostID string) string` — extractSessionIDFromPTYSubject extracts the session_id from a PTY input subject.
- `func startPTY(cmd *exec.Cmd) (io.ReadWriteCloser, error)` — startPTY starts a command in a new PTY and returns the master file descriptor.
- `type ManagedSession` — ManagedSession represents a PTY session managed by this host manager.
- `type Manager` — Manager implements the manager-role daemon on a remote host.
- `type OutboundBuffer` — OutboundBuffer buffers cross-host NATS publications during hub disconnection.
- `type outboundEntry` — outboundEntry holds a single buffered publication.

### Constants

#### Version

```go
const Version = "0.1.0-dev"
```

Version is the manager daemon version string.


### Functions

#### extractSessionIDFromPTYSubject

```go
func extractSessionIDFromPTYSubject(subject, prefix, hostID string) string
```

extractSessionIDFromPTYSubject extracts the session_id from a PTY input subject.
Subject format: P.pty.<host_id>.<session_id>.in

#### startPTY

```go
func startPTY(cmd *exec.Cmd) (io.ReadWriteCloser, error)
```

startPTY starts a command in a new PTY and returns the master file descriptor.
The master FD is used for reading output and writing input.


## type ManagedSession

```go
type ManagedSession struct {
	mu sync.Mutex

	// SessionID is the unique session identifier (base-10 string).
	SessionID string

	// AgentID is the agent this session belongs to (base-10 string).
	AgentID string

	// AgentSlug is the normalized agent slug.
	AgentSlug string

	// RepoPath is the git repository root on this host.
	RepoPath string

	// ReplayBuf is the per-session PTY output replay buffer.
	ReplayBuf *buffer.Ring

	// cmd is the running process.
	cmd *exec.Cmd

	// ptyMaster is the PTY master file descriptor for I/O.
	ptyMaster io.ReadWriteCloser

	// done is closed when the session exits.
	done chan struct{}

	// startedAt records when the session was created (UTC).
	startedAt time.Time

	// running indicates whether the session is active.
	running bool

	// replayPending indicates whether a replay is in progress.
	// While true, live PTY output MUST NOT be published.
	replayPending bool

	// awaitingReplay indicates whether the session is waiting for a replay
	// request after hub reconnection. While true, live PTY output MUST NOT
	// be published; instead it is buffered in liveBuf.
	// Per spec §5.5.7.3: after reconnect, the manager MUST NOT publish
	// live PTY output until a replay request is received for each session.
	awaitingReplay bool

	// liveBuf holds PTY output produced during a replay operation or
	// while awaiting a replay request after reconnection.
	liveBuf []byte
}
```

ManagedSession represents a PTY session managed by this host manager.

### Methods

#### ManagedSession.stop

```go
func () stop()
```

stop gracefully terminates a managed session.


## type Manager

```go
type Manager struct {
	mu     sync.RWMutex
	conn   *natsconn.Conn
	cfg    *config.Config
	prefix string
	hostID string

	// peerMUID is the manager's runtime ID (non-zero, per spec §3.22).
	peerMUID muid.MUID
	// peerID is peerMUID encoded as a base-10 string for wire use.
	peerID string

	// handshakeComplete indicates whether the handshake exchange is done.
	handshakeComplete bool

	// sessions maps agent_id (base-10 string) to active sessions.
	sessions map[string]*ManagedSession

	// sessionsByID maps session_id (base-10 string) to sessions.
	sessionsByID map[string]*ManagedSession

	dispatcher event.Dispatcher
	resolver   *paths.Resolver
	bufferSize int64

	// hubConnected tracks whether the hub connection is active.
	hubConnected bool

	// outboundBuffer holds cross-host publications buffered during disconnection.
	outboundBuffer *OutboundBuffer

	// subs holds active NATS subscriptions.
	subs []*nats.Subscription

	cancel context.CancelFunc
}
```

Manager implements the manager-role daemon on a remote host.

### Functions returning Manager

#### New

```go
func New(conn *natsconn.Conn, cfg *config.Config, hostID string, dispatcher event.Dispatcher) *Manager
```

New creates a new Manager with the given NATS connection and configuration.


### Methods

#### Manager.SetHubConnected

```go
func () SetHubConnected(connected bool)
```

SetHubConnected updates the hub connection state.
Called by the NATS disconnect/reconnect handlers.

When reconnecting (connected=true), all active sessions are marked as
awaitingReplay. Live PTY output MUST NOT be published for those sessions
until a replay request is received, per spec §5.5.7.3.

#### Manager.Start

```go
func () Start(ctx context.Context) error
```

Start performs the handshake and begins listening for control requests.

Per spec §5.5.7.6: daemon MUST:
1. Connect to NATS
2. Perform handshake on P.handshake.<host_id>
3. Start listening on P.ctl.<host_id> and P.pty.<host_id>.*.in

#### Manager.Stop

```go
func () Stop() error
```

Stop gracefully shuts down the manager.

#### Manager.handleControlRequest

```go
func () handleControlRequest(msg *nats.Msg)
```

handleControlRequest processes control requests from the director.

#### Manager.handleKill

```go
func () handleKill(msg *nats.Msg, ctlMsg *protocol.ControlMessage)
```

handleKill terminates a session.

#### Manager.handlePTYInput

```go
func () handlePTYInput(msg *nats.Msg)
```

handlePTYInput receives PTY input from the director and writes it to the session.

#### Manager.handlePing

```go
func () handlePing(msg *nats.Msg, ctlMsg *protocol.ControlMessage)
```

handlePing responds with a pong.

#### Manager.handleReplay

```go
func () handleReplay(msg *nats.Msg, ctlMsg *protocol.ControlMessage)
```

handleReplay replays buffered PTY output for a session.

Per spec §5.5.7.3: the daemon MUST publish all replay bytes before
any subsequently produced live PTY output bytes.

#### Manager.handleSpawn

```go
func () handleSpawn(msg *nats.Msg, ctlMsg *protocol.ControlMessage)
```

handleSpawn creates a new PTY session for an agent.

Per spec §5.5.7.3: spawn MUST be idempotent for a given agent_id.

#### Manager.heartbeatLoop

```go
func () heartbeatLoop(ctx context.Context)
```

heartbeatLoop sends periodic heartbeat pings to the director at the configured
interval. If a ping fails (timeout), a warning is logged but the loop continues.

Per spec: after handshake, the manager MUST send heartbeat pings at the
configured heartbeat_interval to maintain presence.

#### Manager.performHandshake

```go
func () performHandshake(ctx context.Context) error
```

performHandshake sends a handshake request to the director and waits for a reply.

Per spec §5.5.7.3: the daemon MUST send a handshake request after establishing
a NATS connection and MUST NOT accept spawn/kill/replay until complete.

#### Manager.publishChunked

```go
func () publishChunked(subject string, data []byte)
```

publishChunked publishes data to a NATS subject, splitting into chunks
that don't exceed the maximum NATS payload size.

Per spec §5.5.7.4: "Implementations MUST chunk PTY bytes such that no
single NATS message payload exceeds the maximum supported NATS payload size."

#### Manager.publishEvent

```go
func () publishEvent(name string, data any)
```

publishEvent publishes an EventMessage on the host events subject.

#### Manager.readPTYOutput

```go
func () readPTYOutput(sess *ManagedSession)
```

readPTYOutput continuously reads PTY output and publishes it to NATS.

Per spec §5.5.7.3: the replay buffer MUST be updated for all PTY output
bytes regardless of hub connectivity.

#### Manager.replyControl

```go
func () replyControl(msg *nats.Msg, msgType string, payload any)
```

replyControl sends a ControlMessage reply.

#### Manager.replyError

```go
func () replyError(msg *nats.Msg, requestType, code, message string)
```

replyError sends an error ControlMessage reply.

#### Manager.watchSession

```go
func () watchSession(sess *ManagedSession)
```

watchSession monitors a session and emits events when it exits.


## type OutboundBuffer

```go
type OutboundBuffer struct {
	mu       sync.Mutex
	entries  []outboundEntry
	totalLen int64
	maxLen   int64
}
```

OutboundBuffer buffers cross-host NATS publications during hub disconnection.

The buffer has a maximum total payload size. When exceeded, the oldest
entries are dropped first (FIFO eviction). Per-subject order is preserved
because entries are stored in global FIFO order.

### Functions returning OutboundBuffer

#### NewOutboundBuffer

```go
func NewOutboundBuffer(maxBytes int64) *OutboundBuffer
```

NewOutboundBuffer creates a new OutboundBuffer with the given capacity.


### Methods

#### OutboundBuffer.Enqueue

```go
func () Enqueue(subject string, data []byte)
```

Enqueue adds a publication to the buffer.
If the total size exceeds maxLen, the oldest entries are dropped.

Per spec §5.5.8: "MUST account queued size as the sum of NATS message
payload lengths in bytes (excluding subject names and headers)."

#### OutboundBuffer.FlushTo

```go
func () FlushTo(publish func(subject string, data []byte))
```

FlushTo drains all buffered entries to the given publish function.

Per spec §5.5.8: "Flush MUST be FIFO per subject. New publications
generated while a flush is in progress MUST be appended after older
buffered publications for that same subject."

#### OutboundBuffer.Len

```go
func () Len() int
```

Len returns the number of buffered entries.

#### OutboundBuffer.TotalBytes

```go
func () TotalBytes() int64
```

TotalBytes returns the total buffered payload size.


## type outboundEntry

```go
type outboundEntry struct {
	subject string
	data    []byte
}
```

outboundEntry holds a single buffered publication.

