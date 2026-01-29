# package protocol

`import "github.com/agentflare-ai/amux/internal/remote/protocol"`

- `DefaultSubjectPrefix, ControlSubjectTemplate, EventsSubjectTemplate, PTYInSubjectTemplate, PTYOutSubjectTemplate, KVBucketDefault, KVHostInfoTemplate, KVHostHeartbeatTemplate, KVSessionTemplate` — Subject prefixes and templates.
- `OpHandshake, OpSpawn, OpSignal, OpResize, OpReplay` — Control Operations.
- `type ControlRequest` — ControlRequest is the generic envelope for control requests.
- `type ControlResponse` — ControlResponse is the generic envelope for control responses.
- `type Error` — Error represents a protocol error.
- `type HandshakePayload` — HandshakePayload used for OpHandshake.
- `type Heartbeat` — Heartbeat represents dynamic host status.
- `type HostInfo` — HostInfo represents static host metadata stored in KV.
- `type PTYIO` — PTYIO represents a chunk of PTY data.
- `type ReplayPayload` — ReplayPayload used for OpReplay.
- `type ResizePayload` — ResizePayload used for OpResize.
- `type SessionInfo` — SessionInfo represents session metadata in KV.
- `type SignalPayload` — SignalPayload used for OpSignal.
- `type SpawnPayload` — SpawnPayload used for OpSpawn.
- `type SpawnResponsePayload` — SpawnResponsePayload (inside ControlResponse.Payload).

### Constants

#### DefaultSubjectPrefix, ControlSubjectTemplate, EventsSubjectTemplate, PTYInSubjectTemplate, PTYOutSubjectTemplate, KVBucketDefault, KVHostInfoTemplate, KVHostHeartbeatTemplate, KVSessionTemplate

```go
const (
	// DefaultSubjectPrefix is the default prefix for all Amux NATS subjects.
	DefaultSubjectPrefix = "amux"

	// ControlSubjectTemplate is the template for request/reply control messages.
	// Format: <prefix>.control.<host_id>.<op>
	ControlSubjectTemplate = "%s.control.%s.%s"

	// EventsSubjectTemplate is the template for event publishing.
	// Format: <prefix>.events.<host_id>
	EventsSubjectTemplate = "%s.events.%s"

	// PTYInSubjectTemplate is the template for PTY input streaming.
	// Format: <prefix>.pty.<host_id>.<session_id>.in
	PTYInSubjectTemplate = "%s.pty.%s.%s.in"

	// PTYOutSubjectTemplate is the template for PTY output streaming.
	// Format: <prefix>.pty.<host_id>.<session_id>.out
	PTYOutSubjectTemplate = "%s.pty.%s.%s.out"

	// KVBucketDefault is the default name for the JetStream KV bucket.
	KVBucketDefault = "AMUX_KV"

	// KVHostInfoTemplate is the key for host metadata.
	// Format: hosts/<host_id>/info
	KVHostInfoTemplate = "hosts/%s/info"

	// KVHostHeartbeatTemplate is the key for host heartbeats.
	// Format: hosts/<host_id>/heartbeat
	KVHostHeartbeatTemplate = "hosts/%s/heartbeat"

	// KVSessionTemplate is the key for session metadata.
	// Format: sessions/<host_id>/<session_id>
	KVSessionTemplate = "sessions/%s/%s"
)
```

Subject prefixes and templates.

#### OpHandshake, OpSpawn, OpSignal, OpResize, OpReplay

```go
const (
	OpHandshake = "handshake"
	OpSpawn     = "spawn"
	OpSignal    = "signal" // kill/stop
	OpResize    = "resize"
	OpReplay    = "replay"
)
```

Control Operations.


## type ControlRequest

```go
type ControlRequest struct {
	Op        string    `json:"op"`
	RequestID string    `json:"req_id"`
	Payload   []byte    `json:"payload"` // JSON encoded specific payload
	CreatedAt time.Time `json:"created_at"`
}
```

ControlRequest is the generic envelope for control requests.

## type ControlResponse

```go
type ControlResponse struct {
	RequestID string    `json:"req_id"`
	Status    string    `json:"status"` // "ok", "error"
	Error     *Error    `json:"error,omitempty"`
	Payload   []byte    `json:"payload,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
```

ControlResponse is the generic envelope for control responses.

## type Error

```go
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
```

Error represents a protocol error.

## type HandshakePayload

```go
type HandshakePayload struct {
	HostID       string   `json:"host_id"`
	Version      string   `json:"version"`
	Capabilities []string `json:"capabilities"`
}
```

HandshakePayload used for OpHandshake.

## type Heartbeat

```go
type Heartbeat struct {
	HostID    string    `json:"host_id"`
	Timestamp time.Time `json:"ts"`
	Load      float64   `json:"load"`
	MemUsage  uint64    `json:"mem_usage"`
	Sessions  int       `json:"sessions"`
}
```

Heartbeat represents dynamic host status.

## type HostInfo

```go
type HostInfo struct {
	ID           string    `json:"id"`
	Hostname     string    `json:"hostname"`
	Platform     string    `json:"platform"`
	Arch         string    `json:"arch"`
	Version      string    `json:"version"`
	Capabilities []string  `json:"capabilities"`
	FirstSeenAt  time.Time `json:"first_seen_at"`
}
```

HostInfo represents static host metadata stored in KV.

## type PTYIO

```go
type PTYIO struct {
	SessionID string    `json:"sid"`
	Data      []byte    `json:"d"`             // base64 encoded by json/standard, but we use []byte here
	Seq       uint64    `json:"seq,omitempty"` // For output ordering
	Timestamp time.Time `json:"ts"`
}
```

PTYIO represents a chunk of PTY data.

## type ReplayPayload

```go
type ReplayPayload struct {
	SessionID string `json:"session_id"`
	// If 0, replay all buffered.
	SinceSequence uint64 `json:"since_seq,omitempty"`
}
```

ReplayPayload used for OpReplay.

## type ResizePayload

```go
type ResizePayload struct {
	SessionID string `json:"session_id"`
	Rows      int    `json:"rows"`
	Cols      int    `json:"cols"`
}
```

ResizePayload used for OpResize.

## type SessionInfo

```go
type SessionInfo struct {
	SessionID string    `json:"session_id"`
	AgentID   string    `json:"agent_id"`
	HostID    string    `json:"host_id"`
	State     string    `json:"state"` // "running", "ended"
	CreatedAt time.Time `json:"created_at"`
}
```

SessionInfo represents session metadata in KV.

## type SignalPayload

```go
type SignalPayload struct {
	SessionID string `json:"session_id"`
	Signal    string `json:"signal"` // "kill", "term", "int" or syscall name
}
```

SignalPayload used for OpSignal.

## type SpawnPayload

```go
type SpawnPayload struct {
	AgentID     string            `json:"agent_id"` // muid
	AgentSlug   string            `json:"agent_slug"`
	RepoPath    string            `json:"repo_path"`
	Env         map[string]string `json:"env,omitempty"`
	Command     []string          `json:"command,omitempty"`
	Interactive bool              `json:"interactive"`
}
```

SpawnPayload used for OpSpawn.

## type SpawnResponsePayload

```go
type SpawnResponsePayload struct {
	SessionID string `json:"session_id"` // muid
}
```

SpawnResponsePayload (inside ControlResponse.Payload).

