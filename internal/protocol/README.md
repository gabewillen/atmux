# package protocol

`import "github.com/agentflare-ai/amux/internal/protocol"`

Package protocol defines the remote communication protocol.

- `SubjectHandshake, SubjectCtl, SubjectEvents, SubjectPTY` — Subject prefixes
- `func SubjectForCtl(prefix string, hostID api.HostID) string` — SubjectForCtl returns the control subject for a host.
- `func SubjectForEvents(prefix string, hostID api.HostID) string` — SubjectForEvents returns the events subject for a host.
- `func SubjectForHandshake(prefix string, hostID api.HostID) string` — SubjectHandshake returns the handshake subject for a host.
- `func SubjectForPTYIn(prefix string, hostID api.HostID, sessionID api.SessionID) string` — SubjectForPTYIn returns the PTY input subject.
- `func SubjectForPTYOut(prefix string, hostID api.HostID, sessionID api.SessionID) string` — SubjectForPTYOut returns the PTY output subject.
- `type ControlRequest` — ControlRequest is the payload for control operations (spawn/kill/replay).
- `type ControlResponse` — ControlResponse is the payload for control operation responses.
- `type Error` — Error represents a protocol error.
- `type EventMessage` — EventMessage represents an event envelope over NATS.
- `type HandshakeRequest` — HandshakeRequest is the payload for the handshake request.
- `type HandshakeResponse` — HandshakeResponse is the payload for the handshake response.
- `type HsmNetDispatcher` — HsmNetDispatcher manages event routing.
- `type LocalBus` — LocalBus is an interface for the local event bus (e.g., internal/agent/bus.go).
- `type SpawnPayload` — SpawnPayload is the payload for the spawn command.
- `type SpawnResponsePayload` — SpawnResponsePayload is the payload for the spawn response.

### Constants

#### SubjectHandshake, SubjectCtl, SubjectEvents, SubjectPTY

```go
const (
	SubjectHandshake = "handshake"
	SubjectCtl       = "ctl"
	SubjectEvents    = "events"
	SubjectPTY       = "pty"
)
```

Subject prefixes


### Functions

#### SubjectForCtl

```go
func SubjectForCtl(prefix string, hostID api.HostID) string
```

SubjectForCtl returns the control subject for a host.
P.ctl.<host_id>

#### SubjectForEvents

```go
func SubjectForEvents(prefix string, hostID api.HostID) string
```

SubjectForEvents returns the events subject for a host.
P.events.<host_id>

#### SubjectForHandshake

```go
func SubjectForHandshake(prefix string, hostID api.HostID) string
```

SubjectHandshake returns the handshake subject for a host.
P.handshake.<host_id>

#### SubjectForPTYIn

```go
func SubjectForPTYIn(prefix string, hostID api.HostID, sessionID api.SessionID) string
```

SubjectForPTYIn returns the PTY input subject.
P.pty.<host_id>.<session_id>.in

#### SubjectForPTYOut

```go
func SubjectForPTYOut(prefix string, hostID api.HostID, sessionID api.SessionID) string
```

SubjectForPTYOut returns the PTY output subject.
P.pty.<host_id>.<session_id>.out


## type ControlRequest

```go
type ControlRequest struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
```

ControlRequest is the payload for control operations (spawn/kill/replay).

## type ControlResponse

```go
type ControlResponse struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	Error   *Error          `json:"error,omitempty"`
}
```

ControlResponse is the payload for control operation responses.

## type Error

```go
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
```

Error represents a protocol error.

## type EventMessage

```go
type EventMessage struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Source    string      `json:"source"` // peer_id
	Payload   interface{} `json:"payload"`
}
```

EventMessage represents an event envelope over NATS.
It matches the spec for hsmnet wire format.

## type HandshakeRequest

```go
type HandshakeRequest struct {
	Protocol int        `json:"protocol"`
	PeerID   api.PeerID `json:"peer_id"`
	Role     string     `json:"role"`
	HostID   api.HostID `json:"host_id"`
}
```

HandshakeRequest is the payload for the handshake request.

## type HandshakeResponse

```go
type HandshakeResponse struct {
	Protocol int        `json:"protocol"`
	PeerID   api.PeerID `json:"peer_id"`
	Role     string     `json:"role"`
	HostID   api.HostID `json:"host_id"`
	Error    *Error     `json:"error,omitempty"`
}
```

HandshakeResponse is the payload for the handshake response.

## type HsmNetDispatcher

```go
type HsmNetDispatcher struct {
	mu            sync.RWMutex
	localBus      LocalBus // Simplified interface for local bus
	natsConn      *nats.Conn
	peerID        api.PeerID
	hostID        api.HostID
	subjectPrefix string
}
```

HsmNetDispatcher manages event routing.

### Functions returning HsmNetDispatcher

#### NewDispatcher

```go
func NewDispatcher(peerID api.PeerID, hostID api.HostID, nc *nats.Conn) *HsmNetDispatcher
```

NewDispatcher creates a new dispatcher.


### Methods

#### HsmNetDispatcher.Dispatch

```go
func () Dispatch(ctx context.Context, event EventMessage) error
```

Dispatch sends an event.

#### HsmNetDispatcher.SetLocalBus

```go
func () SetLocalBus(bus LocalBus)
```

SetLocalBus attaches the local bus.


## type LocalBus

```go
type LocalBus interface {
	Publish(event interface{})
}
```

LocalBus is an interface for the local event bus (e.g., internal/agent/bus.go).

## type SpawnPayload

```go
type SpawnPayload struct {
	AgentID  api.AgentID       `json:"agent_id"`
	Slug     api.AgentSlug     `json:"agent_slug"`
	RepoPath string            `json:"repo_path"`
	Command  []string          `json:"command,omitempty"`
	Env      map[string]string `json:"env,omitempty"`
}
```

SpawnPayload is the payload for the spawn command.

## type SpawnResponsePayload

```go
type SpawnResponsePayload struct {
	AgentID   api.AgentID   `json:"agent_id"`
	SessionID api.SessionID `json:"session_id"`
}
```

SpawnResponsePayload is the payload for the spawn response.

