# package manager

`import "github.com/agentflare-ai/amux/internal/remote/manager"`

- `type Manager` — Manager manages remote sessions on this host.
- `type Options` — Options for Manager.
- `type ReplayBuffer` — ReplayBuffer is a thread-safe ring buffer for PTY output replay.
- `type Session` — Session represents a running remote session.

## type Manager

```go
type Manager struct {
	opts     Options
	nc       *nats.Conn
	js       jetstream.JetStream
	kv       jetstream.KeyValue
	sessions map[string]*Session
	mu       sync.RWMutex
	ctx      context.Context
	cancel   func()
}
```

Manager manages remote sessions on this host.

### Functions returning Manager

#### New

```go
func New(opts Options) *Manager
```

New creates a new Manager.


### Methods

#### Manager.Start

```go
func () Start(ctx context.Context) error
```

Start connects to NATS and begins listening for control messages.

#### Manager.Stop

```go
func () Stop()
```

Stop shuts down the manager and all sessions.

#### Manager.handleControl

```go
func () handleControl(msg *nats.Msg)
```

#### Manager.handleHandshake

```go
func () handleHandshake(msg *nats.Msg, req protocol.ControlRequest)
```

#### Manager.handleReplay

```go
func () handleReplay(msg *nats.Msg, req protocol.ControlRequest)
```

#### Manager.handleResize

```go
func () handleResize(msg *nats.Msg, req protocol.ControlRequest)
```

#### Manager.handleSignal

```go
func () handleSignal(msg *nats.Msg, req protocol.ControlRequest)
```

#### Manager.handleSpawn

```go
func () handleSpawn(msg *nats.Msg, req protocol.ControlRequest)
```

#### Manager.heartbeatLoop

```go
func () heartbeatLoop()
```

#### Manager.replyError

```go
func () replyError(msg *nats.Msg, reqID string, code, message string)
```

#### Manager.streamPTY

```go
func () streamPTY(s *Session)
```


## type Options

```go
type Options struct {
	Config   config.Config
	HostID   string
	Worktree *worktree.Manager
}
```

Options for Manager.

## type ReplayBuffer

```go
type ReplayBuffer struct {
	capacity int64
	current  int64
	items    []*protocol.PTYIO
	mu       sync.RWMutex
	nextSeq  uint64
}
```

ReplayBuffer is a thread-safe ring buffer for PTY output replay.

### Functions returning ReplayBuffer

#### NewReplayBuffer

```go
func NewReplayBuffer(capacityBytes int64) *ReplayBuffer
```

NewReplayBuffer creates a new buffer with the given capacity in bytes.


### Methods

#### ReplayBuffer.Append

```go
func () Append(sessionID string, data []byte) *protocol.PTYIO
```

Append adds data to the buffer, dropping oldest items if needed.

#### ReplayBuffer.Replay

```go
func () Replay(sinceSeq uint64) []*protocol.PTYIO
```

Replay returns all items since the given sequence number (exclusive).


## type Session

```go
type Session struct {
	ID        string
	Agent     *agent.AgentActor
	Buffer    *ReplayBuffer
	Cancel    func()
	CreatedAt time.Time
}
```

Session represents a running remote session.

