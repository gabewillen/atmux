# package protocol

`import "github.com/agentflare-ai/amux/internal/protocol"`

Package protocol provides the remote communication protocol for amux.

This package implements the NATS-based protocol for communication
between the director and remote host managers. All protocol operations
are agent-agnostic.

Control messages use the ControlMessage envelope with a type discriminator
and a JSON payload. Subject namespaces follow spec §5.5.7.1.

See spec §5.5.7 for the communication protocol and §9.1.3 for wire format.

Package protocol - wire.go provides EventMessage and WireEvent types
for the hsmnet wire format as specified in §9.1.3.

EventMessage envelopes are published on the host events subject
(P.events.<host_id>) and carry structured event data.

- `CodeNotReady, CodeSessionConflict, CodeInvalidRepo, CodeInvalidAgent, CodeSessionNotFound, CodeProtocolError, CodeInternalError, CodeHostIDMismatch, CodePeerCollision, CodeHostCollision` — Error codes used in ErrorPayload.Code.
- `ProtocolVersion` — ProtocolVersion is the current protocol version used in handshake exchanges.
- `TypeHandshake, TypePing, TypePong, TypeSpawn, TypeKill, TypeReplay, TypeError` — Control message type constants per spec §5.5.7.2.
- `func AgentChannelSubject(prefix, hostID, agentID string) string` — AgentChannelSubject returns the agent communication channel subject.
- `func BroadcastChannelSubject(prefix string) string` — BroadcastChannelSubject returns the broadcast communication channel subject.
- `func ControlSubject(prefix, hostID string) string` — ControlSubject returns the control request subject for a host.
- `func DirectorChannelSubject(prefix string) string` — DirectorChannelSubject returns the director communication channel subject.
- `func EventsSubject(prefix, hostID string) string` — EventsSubject returns the host events subject.
- `func HandshakeSubject(prefix, hostID string) string` — HandshakeSubject returns the handshake subject for a host.
- `func ManagerChannelSubject(prefix, hostID string) string` — ManagerChannelSubject returns the host manager communication channel subject.
- `func PTYInputSubject(prefix, hostID, sessionID string) string` — PTYInputSubject returns the PTY input subject for a session.
- `func PTYInputWildcard(prefix, hostID string) string` — PTYInputWildcard returns the wildcard subscription for all PTY input on a host.
- `func PTYOutputSubject(prefix, hostID, sessionID string) string` — PTYOutputSubject returns the PTY output subject for a session.
- `type AgentTerminatedEvent` — AgentTerminatedEvent is the payload for "agent.terminated" events.
- `type ConnectionEstablishedEvent` — ConnectionEstablishedEvent is the payload for "connection.established" events.
- `type ConnectionLostEvent` — ConnectionLostEvent is the payload for "connection.lost" events.
- `type ConnectionRecoveredEvent` — ConnectionRecoveredEvent is the payload for "connection.recovered" events.
- `type ControlMessage` — ControlMessage is the top-level envelope for all control messages exchanged between the director and manager-role daemons over NATS request-reply.
- `type ErrorPayload` — ErrorPayload is the error response payload.
- `type EventMessage` — EventMessage is the wire envelope for events transported over NATS.
- `type HandshakePayload` — HandshakePayload is the handshake request/response payload.
- `type HostInfoPayload` — HostInfoPayload carries host metadata in the handshake.
- `type KillRequest` — KillRequest is the kill request payload.
- `type KillResponse` — KillResponse is the kill response payload.
- `type MessageType` — MessageType represents the event message routing type.
- `type PingPayload` — PingPayload is the ping request payload.
- `type PongPayload` — PongPayload is the pong response payload.
- `type ProcessCompletedEvent` — ProcessCompletedEvent is the payload for "process.completed", "process.failed", and "process.killed" events.
- `type ProcessIOEvent` — ProcessIOEvent is the payload for "process.stdout", "process.stderr", and "process.stdin" events.
- `type ProcessSpawnedEvent` — ProcessSpawnedEvent is the payload for "process.spawned" events.
- `type ReplayRequest` — ReplayRequest is the replay request payload.
- `type ReplayResponse` — ReplayResponse is the replay response payload.
- `type SpawnRequest` — SpawnRequest is the spawn request payload.
- `type SpawnResponse` — SpawnResponse is the spawn response payload.
- `type WireEvent` — WireEvent carries an event name and its opaque JSON data.

### Constants

#### TypeHandshake, TypePing, TypePong, TypeSpawn, TypeKill, TypeReplay, TypeError

```go
const (
	TypeHandshake = "handshake"
	TypePing      = "ping"
	TypePong      = "pong"
	TypeSpawn     = "spawn"
	TypeKill      = "kill"
	TypeReplay    = "replay"
	TypeError     = "error"
)
```

Control message type constants per spec §5.5.7.2.

#### CodeNotReady, CodeSessionConflict, CodeInvalidRepo, CodeInvalidAgent, CodeSessionNotFound, CodeProtocolError, CodeInternalError, CodeHostIDMismatch, CodePeerCollision, CodeHostCollision

```go
const (
	CodeNotReady        = "not_ready"
	CodeSessionConflict = "session_conflict"
	CodeInvalidRepo     = "invalid_repo"
	CodeInvalidAgent    = "invalid_agent"
	CodeSessionNotFound = "session_not_found"
	CodeProtocolError   = "protocol_error"
	CodeInternalError   = "internal_error"
	CodeHostIDMismatch  = "host_id_mismatch"
	CodePeerCollision   = "peer_collision"
	CodeHostCollision   = "host_collision"
)
```

Error codes used in ErrorPayload.Code.

#### ProtocolVersion

```go
const ProtocolVersion = 1
```

ProtocolVersion is the current protocol version used in handshake exchanges.


### Functions

#### AgentChannelSubject

```go
func AgentChannelSubject(prefix, hostID, agentID string) string
```

AgentChannelSubject returns the agent communication channel subject.
Format: P.comm.agent.<host_id>.<agent_id>

#### BroadcastChannelSubject

```go
func BroadcastChannelSubject(prefix string) string
```

BroadcastChannelSubject returns the broadcast communication channel subject.
Format: P.comm.broadcast

#### ControlSubject

```go
func ControlSubject(prefix, hostID string) string
```

ControlSubject returns the control request subject for a host.
Format: P.ctl.<host_id>

#### DirectorChannelSubject

```go
func DirectorChannelSubject(prefix string) string
```

DirectorChannelSubject returns the director communication channel subject.
Format: P.comm.director

#### EventsSubject

```go
func EventsSubject(prefix, hostID string) string
```

EventsSubject returns the host events subject.
Format: P.events.<host_id>

#### HandshakeSubject

```go
func HandshakeSubject(prefix, hostID string) string
```

HandshakeSubject returns the handshake subject for a host.
Format: P.handshake.<host_id>

#### ManagerChannelSubject

```go
func ManagerChannelSubject(prefix, hostID string) string
```

ManagerChannelSubject returns the host manager communication channel subject.
Format: P.comm.manager.<host_id>

#### PTYInputSubject

```go
func PTYInputSubject(prefix, hostID, sessionID string) string
```

PTYInputSubject returns the PTY input subject for a session.
Format: P.pty.<host_id>.<session_id>.in

#### PTYInputWildcard

```go
func PTYInputWildcard(prefix, hostID string) string
```

PTYInputWildcard returns the wildcard subscription for all PTY input on a host.
Format: P.pty.<host_id>.*.in

#### PTYOutputSubject

```go
func PTYOutputSubject(prefix, hostID, sessionID string) string
```

PTYOutputSubject returns the PTY output subject for a session.
Format: P.pty.<host_id>.<session_id>.out


## type AgentTerminatedEvent

```go
type AgentTerminatedEvent struct {
	SessionID string `json:"session_id"`
	AgentID   string `json:"agent_id"`
}
```

AgentTerminatedEvent is the payload for "agent.terminated" events.

## type ConnectionEstablishedEvent

```go
type ConnectionEstablishedEvent struct {
	PeerID    string `json:"peer_id"`
	Timestamp string `json:"timestamp"`
}
```

ConnectionEstablishedEvent is the payload for "connection.established" events.

## type ConnectionLostEvent

```go
type ConnectionLostEvent struct {
	PeerID    string   `json:"peer_id"`
	HostID    string   `json:"host_id,omitempty"`
	Timestamp string   `json:"timestamp"`
	Reason    string   `json:"reason,omitempty"`
	Sessions  []string `json:"sessions,omitempty"`
}
```

ConnectionLostEvent is the payload for "connection.lost" events.

## type ConnectionRecoveredEvent

```go
type ConnectionRecoveredEvent struct {
	PeerID    string `json:"peer_id"`
	Timestamp string `json:"timestamp"`
}
```

ConnectionRecoveredEvent is the payload for "connection.recovered" events.

## type ControlMessage

```go
type ControlMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
```

ControlMessage is the top-level envelope for all control messages exchanged
between the director and manager-role daemons over NATS request-reply.

See spec §5.5.7.2.

### Functions returning ControlMessage

#### NewControlMessage

```go
func NewControlMessage(msgType string, payload any) (*ControlMessage, error)
```

NewControlMessage creates a ControlMessage with the given type and payload.

#### NewErrorMessage

```go
func NewErrorMessage(requestType, code, message string) (*ControlMessage, error)
```

NewErrorMessage creates a ControlMessage of type "error".


### Methods

#### ControlMessage.DecodePayload

```go
func () DecodePayload(target any) error
```

DecodePayload decodes the Payload field into the given target.


## type ErrorPayload

```go
type ErrorPayload struct {
	// RequestType is the type of request that caused the error.
	// One of: "handshake", "spawn", "kill", "replay", "unknown".
	RequestType string `json:"request_type"`

	// Code is a short machine-readable error code.
	Code string `json:"code"`

	// Message is a human-readable diagnostic.
	Message string `json:"message"`
}
```

ErrorPayload is the error response payload.

See spec §5.5.7.3.

## type EventMessage

```go
type EventMessage struct {
	// Type is the routing type (1=broadcast, 2=multicast, 3=unicast).
	Type MessageType `json:"type"`

	// Target is the destination peer ID for unicast (base-10 muid.ID string).
	Target string `json:"target,omitempty"`

	// Targets is the list of destination peer IDs for multicast.
	Targets []string `json:"targets,omitempty"`

	// Event is the wrapped event.
	Event WireEvent `json:"event"`
}
```

EventMessage is the wire envelope for events transported over NATS.

See spec §9.1.3.

### Functions returning EventMessage

#### NewBroadcastEvent

```go
func NewBroadcastEvent(name string, data any) (*EventMessage, error)
```

NewBroadcastEvent creates a broadcast EventMessage.

#### NewMulticastEvent

```go
func NewMulticastEvent(name string, targets []string, data any) (*EventMessage, error)
```

NewMulticastEvent creates a multicast EventMessage routed to specified targets.

#### NewUnicastEvent

```go
func NewUnicastEvent(name string, target string, data any) (*EventMessage, error)
```

NewUnicastEvent creates a unicast EventMessage.


## type HandshakePayload

```go
type HandshakePayload struct {
	// Protocol is the protocol version (must be ProtocolVersion).
	Protocol int `json:"protocol"`

	// PeerID is the unique peer identifier (base-10 unsigned integer string).
	PeerID string `json:"peer_id"`

	// Role is the peer role: "daemon" for manager, "director" for director.
	Role string `json:"role"`

	// HostID is the host identifier.
	HostID string `json:"host_id"`

	// HostInfo contains optional host metadata (version, OS, arch).
	// Populated by the manager during handshake.
	HostInfo *HostInfoPayload `json:"host_info,omitempty"`
}
```

HandshakePayload is the handshake request/response payload.

See spec §5.5.7.3.

## type HostInfoPayload

```go
type HostInfoPayload struct {
	// Version is the daemon version string.
	Version string `json:"version"`

	// OS is the operating system (runtime.GOOS).
	OS string `json:"os"`

	// Arch is the CPU architecture (runtime.GOARCH).
	Arch string `json:"arch"`
}
```

HostInfoPayload carries host metadata in the handshake.

## type KillRequest

```go
type KillRequest struct {
	// SessionID is the session to terminate (base-10 unsigned integer string).
	SessionID string `json:"session_id"`
}
```

KillRequest is the kill request payload.

See spec §5.5.7.3.

## type KillResponse

```go
type KillResponse struct {
	// SessionID is echoed from the request.
	SessionID string `json:"session_id"`

	// Killed is true if a session was found and termination was initiated.
	Killed bool `json:"killed"`
}
```

KillResponse is the kill response payload.

See spec §5.5.7.3.

## type MessageType

```go
type MessageType uint8
```

MessageType represents the event message routing type.
Encoded as a JSON number per spec §9.1.3.1.

### Constants

#### MsgBroadcast, MsgMulticast, MsgUnicast

```go
const (
	// MsgBroadcast routes the event to all peers.
	MsgBroadcast MessageType = 1

	// MsgMulticast routes the event to specified targets.
	MsgMulticast MessageType = 2

	// MsgUnicast routes the event to a single target.
	MsgUnicast MessageType = 3
)
```


## type PingPayload

```go
type PingPayload struct {
	// TSUnixMs is the timestamp in milliseconds since Unix epoch.
	TSUnixMs int64 `json:"ts_unix_ms"`
}
```

PingPayload is the ping request payload.

See spec §5.5.7.3.

### Functions returning PingPayload

#### NewPingPayload

```go
func NewPingPayload() *PingPayload
```

NewPingPayload creates a PingPayload with the current timestamp.


## type PongPayload

```go
type PongPayload struct {
	// TSUnixMs is echoed from the ping request.
	TSUnixMs int64 `json:"ts_unix_ms"`
}
```

PongPayload is the pong response payload.

See spec §5.5.7.3.

## type ProcessCompletedEvent

```go
type ProcessCompletedEvent struct {
	PID       int    `json:"pid"`
	AgentID   string `json:"agent_id"`
	ProcessID string `json:"process_id"`
	Command   string `json:"command"`
	ExitCode  int    `json:"exit_code"`
	Signal    *int   `json:"signal"`
	StartedAt string `json:"started_at"`
	EndedAt   string `json:"ended_at"`
	Duration  string `json:"duration"`
}
```

ProcessCompletedEvent is the payload for "process.completed", "process.failed",
and "process.killed" events.

## type ProcessIOEvent

```go
type ProcessIOEvent struct {
	PID       int    `json:"pid"`
	AgentID   string `json:"agent_id"`
	ProcessID string `json:"process_id"`
	Command   string `json:"command"`
	Stream    string `json:"stream"`
	DataB64   string `json:"data_b64"`
	Timestamp string `json:"timestamp"`
}
```

ProcessIOEvent is the payload for "process.stdout", "process.stderr",
and "process.stdin" events.

## type ProcessSpawnedEvent

```go
type ProcessSpawnedEvent struct {
	PID       int      `json:"pid"`
	AgentID   string   `json:"agent_id"`
	ProcessID string   `json:"process_id"`
	Command   string   `json:"command"`
	Args      []string `json:"args"`
	WorkDir   string   `json:"work_dir"`
	ParentPID int      `json:"parent_pid"`
	StartedAt string   `json:"started_at"`
}
```

ProcessSpawnedEvent is the payload for "process.spawned" events.

## type ReplayRequest

```go
type ReplayRequest struct {
	// SessionID is the session to replay (base-10 unsigned integer string).
	SessionID string `json:"session_id"`
}
```

ReplayRequest is the replay request payload.

See spec §5.5.7.3.

## type ReplayResponse

```go
type ReplayResponse struct {
	// SessionID is echoed from the request.
	SessionID string `json:"session_id"`

	// Accepted is true if the daemon will replay buffered PTY output.
	Accepted bool `json:"accepted"`
}
```

ReplayResponse is the replay response payload.

See spec §5.5.7.3.

## type SpawnRequest

```go
type SpawnRequest struct {
	// AgentID is the agent identifier (base-10 unsigned integer string).
	AgentID string `json:"agent_id"`

	// AgentSlug is the normalized agent slug per §5.3.1.
	AgentSlug string `json:"agent_slug"`

	// RepoPath is the git repository root on the remote host.
	// May begin with ~/ and is expanded by the remote daemon.
	RepoPath string `json:"repo_path"`

	// Command is the argv vector for the agent CLI.
	Command []string `json:"command"`

	// Env is optional environment variables for the spawned process.
	Env map[string]string `json:"env,omitempty"`
}
```

SpawnRequest is the spawn request payload.

See spec §5.5.7.3.

## type SpawnResponse

```go
type SpawnResponse struct {
	// AgentID is echoed from the request (base-10 unsigned integer string).
	AgentID string `json:"agent_id"`

	// SessionID is the session identifier (base-10 unsigned integer string, non-zero).
	SessionID string `json:"session_id"`
}
```

SpawnResponse is the spawn response payload.

See spec §5.5.7.3.

## type WireEvent

```go
type WireEvent struct {
	// Name is the event type identifier (e.g., "process.spawned").
	Name string `json:"name"`

	// Data is the event payload as raw JSON.
	Data json.RawMessage `json:"data"`
}
```

WireEvent carries an event name and its opaque JSON data.

