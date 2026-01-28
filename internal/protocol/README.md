# package protocol

`import "github.com/agentflare-ai/amux/internal/protocol"`

Package protocol provides the remote communication protocol for amux.

This package implements the NATS-based protocol for communication
between the director and remote host managers. All protocol operations
are agent-agnostic.

See spec §5.5 and §9 for remote protocol requirements.

- `func PTYSubject(prefix, hostID, sessionID string, isOutput bool) string` — PTYSubject returns the NATS subject for PTY I/O.
- `func Subject(prefix string, hostID string, msgType MessageType) string` — Subject returns the NATS subject for a message type.
- `type HandshakeRequest` — HandshakeRequest is the handshake request payload.
- `type HandshakeResponse` — HandshakeResponse is the handshake response payload.
- `type KillRequest` — KillRequest is the kill request payload.
- `type KillResponse` — KillResponse is the kill response payload.
- `type MessageType` — MessageType represents a protocol message type.
- `type Message` — Message is the base protocol message envelope.
- `type PTYData` — PTYData is the PTY I/O payload.
- `type ReplayRequest` — ReplayRequest is the replay request payload.
- `type ReplayResponse` — ReplayResponse is the replay response payload.
- `type SpawnRequest` — SpawnRequest is the spawn request payload.
- `type SpawnResponse` — SpawnResponse is the spawn response payload.

### Functions

#### PTYSubject

```go
func PTYSubject(prefix, hostID, sessionID string, isOutput bool) string
```

PTYSubject returns the NATS subject for PTY I/O.

#### Subject

```go
func Subject(prefix string, hostID string, msgType MessageType) string
```

Subject returns the NATS subject for a message type.


## type HandshakeRequest

```go
type HandshakeRequest struct {
	HostID      string `json:"host_id"`
	Version     string `json:"version"`
	SpecVersion string `json:"spec_version"`
}
```

HandshakeRequest is the handshake request payload.

## type HandshakeResponse

```go
type HandshakeResponse struct {
	Accepted    bool   `json:"accepted"`
	Error       string `json:"error,omitempty"`
	Version     string `json:"version"`
	SpecVersion string `json:"spec_version"`
}
```

HandshakeResponse is the handshake response payload.

## type KillRequest

```go
type KillRequest struct {
	SessionID string `json:"session_id"`
	Force     bool   `json:"force,omitempty"`
}
```

KillRequest is the kill request payload.

## type KillResponse

```go
type KillResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
```

KillResponse is the kill response payload.

## type Message

```go
type Message struct {
	// ID is the unique message identifier.
	ID string `json:"id"`

	// Type is the message type.
	Type MessageType `json:"type"`

	// Timestamp is when the message was created (RFC 3339).
	Timestamp string `json:"timestamp"`

	// HostID is the sending host identifier.
	HostID string `json:"host_id,omitempty"`

	// AgentID is the target agent identifier (base-10 string).
	AgentID string `json:"agent_id,omitempty"`

	// SessionID is the session identifier (base-10 string).
	SessionID string `json:"session_id,omitempty"`

	// Data is the message payload.
	Data json.RawMessage `json:"data,omitempty"`

	// TraceID is the optional trace context.
	TraceID string `json:"trace_id,omitempty"`
}
```

Message is the base protocol message envelope.

### Functions returning Message

#### NewMessage

```go
func NewMessage(msgType MessageType, data any) (*Message, error)
```

NewMessage creates a new protocol message.


### Methods

#### Message.WithAgent

```go
func () WithAgent(agentID muid.MUID) *Message
```

WithAgent sets the agent ID.

#### Message.WithHost

```go
func () WithHost(hostID string) *Message
```

WithHost sets the host ID.

#### Message.WithSession

```go
func () WithSession(sessionID muid.MUID) *Message
```

WithSession sets the session ID.

#### Message.WithTrace

```go
func () WithTrace(traceID string) *Message
```

WithTrace sets the trace ID.


## type MessageType

```go
type MessageType string
```

MessageType represents a protocol message type.

### Constants

#### TypeHandshakeRequest, TypeHandshakeResponse, TypeSpawnRequest, TypeSpawnResponse, TypeKillRequest, TypeKillResponse, TypeReplayRequest, TypeReplayResponse, TypePTYInput, TypePTYOutput, TypeEvent

```go
const (
	// Handshake messages
	TypeHandshakeRequest  MessageType = "handshake.request"
	TypeHandshakeResponse MessageType = "handshake.response"

	// Control messages
	TypeSpawnRequest   MessageType = "spawn.request"
	TypeSpawnResponse  MessageType = "spawn.response"
	TypeKillRequest    MessageType = "kill.request"
	TypeKillResponse   MessageType = "kill.response"
	TypeReplayRequest  MessageType = "replay.request"
	TypeReplayResponse MessageType = "replay.response"

	// PTY messages
	TypePTYInput  MessageType = "pty.input"
	TypePTYOutput MessageType = "pty.output"

	// Event messages
	TypeEvent MessageType = "event"
)
```


## type PTYData

```go
type PTYData struct {
	Data string `json:"data_b64"` // Base64-encoded PTY data
}
```

PTYData is the PTY I/O payload.

## type ReplayRequest

```go
type ReplayRequest struct {
	SessionID string `json:"session_id"`
}
```

ReplayRequest is the replay request payload.

## type ReplayResponse

```go
type ReplayResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
```

ReplayResponse is the replay response payload.

## type SpawnRequest

```go
type SpawnRequest struct {
	AgentID   string `json:"agent_id"`
	RepoPath  string `json:"repo_path"`
	AgentSlug string `json:"agent_slug"`
	Adapter   string `json:"adapter"`
}
```

SpawnRequest is the spawn request payload.

## type SpawnResponse

```go
type SpawnResponse struct {
	SessionID string `json:"session_id"`
	Error     string `json:"error,omitempty"`
	Code      string `json:"code,omitempty"`
}
```

SpawnResponse is the spawn response payload.

