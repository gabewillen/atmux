# package remote

`import "github.com/agentflare-ai/amux/internal/remote"`

bootstrap.go implements SSH bootstrap for remote hosts per spec §5.5.2, §5.5.3, §5.5.6.4.
The director provisions per-host NATS creds, copies bootstrap payload, and starts the daemon.

control.go defines control message types and payloads for the remote protocol per spec §5.5.7.2, §5.5.7.3.

director.go implements director-side remote orchestration: hub NATS, handshake handling,
request-reply spawn/kill/replay with timeout and fail-fast semantics per spec §5.5.6.1, §5.5.7, §5.5.7.2.1.

kv.go implements JetStream KV bucket provisioning and required durable state keys per spec §5.5.6.3.

manager.go implements manager-role daemon behaviors: leaf NATS connection, handshake client,
control request handler (spawn/kill/replay), per-session replay buffer, and PTY I/O subjects per spec §5.5.5, §5.5.7, §5.5.8, §5.5.9.

ringbuffer.go implements a per-session replay ring buffer per spec §5.5.7.3.
Buffer is capped at remote.buffer_size; when exceeded, oldest bytes are dropped.

Package remote implements remote agent orchestration: NATS hub/leaf, JetStream KV,
handshake, request-reply control (spawn/kill/replay), PTY I/O subjects, and replay buffering
per spec §5.5, §5.5.6–§5.5.8.

This file defines NATS subject namespaces per spec §5.5.7.1.

- `ControlTypeHandshake, ControlTypePing, ControlTypePong, ControlTypeSpawn, ControlTypeKill, ControlTypeReplay, ControlTypeError` — Control message type constants (spec §5.5.7.2).
- `ErrorCodeNotReady, ErrorCodeSessionConflict, ErrorCodeInvalidRepo` — Error codes (spec §5.5.7.3).
- `func CopyBootstrapZip(ctx context.Context, cfg *BootstrapConfig, localZipPath string) (remotePath string, err error)` — CopyBootstrapZip copies the local bootstrap zip to the remote bootstrap dir (spec §5.5.2 step 3).
- `func CopyFile(ctx context.Context, cfg *BootstrapConfig, localPath, remotePath string) error` — CopyFile copies a local file to the remote path via scp.
- `func DaemonStatus(ctx context.Context, cfg *BootstrapConfig) ([]byte, error)` — DaemonStatus runs "amux-manager status" on the remote host (spec §5.5.2 step 6).
- `func EnsureBootstrapDir(ctx context.Context, cfg *BootstrapConfig) error` — EnsureBootstrapDir creates the remote bootstrap directory (e.g.
- `func EnsureKVBucket(ctx context.Context, js nats.JetStreamContext, bucket string) (nats.KeyValue, error)` — EnsureKVBucket creates the JetStream KV bucket if it does not exist (spec §5.5.6.3).
- `func KVKeyHostHeartbeat(hostID string) string` — KVKeyHostHeartbeat returns the key for last-seen heartbeat: hosts/<host_id>/heartbeat.
- `func KVKeyHostInfo(hostID string) string` — KVKeyHostInfo returns the key for host metadata: hosts/<host_id>/info.
- `func KVKeySession(hostID, sessionID string) string` — KVKeySession returns the key for session metadata: sessions/<host_id>/<session_id>.
- `func ProvisionCreds(ctx context.Context, cfg *BootstrapConfig, credsContent []byte) error` — ProvisionCreds writes credential content to the remote path with permissions 0600 (spec §5.5.6.4).
- `func PutHostHeartbeat(kv nats.KeyValue, hostID string) error` — PutHostHeartbeat writes hosts/<host_id>/heartbeat (RFC 3339 timestamp).
- `func PutHostInfo(kv nats.KeyValue, hostID string, v HostInfoValue) error` — PutHostInfo writes hosts/<host_id>/info as UTF-8 JSON.
- `func PutSession(kv nats.KeyValue, hostID, sessionID string, v SessionKVValue) error` — PutSession writes sessions/<host_id>/<session_id> as UTF-8 JSON.
- `func RunSSH(ctx context.Context, cfg *BootstrapConfig, cmd string) ([]byte, error)` — RunSSH runs a remote command via ssh.
- `func StartDaemon(ctx context.Context, cfg *BootstrapConfig, hostID, natsURL, credsPath string) error` — StartDaemon starts the daemon in manager role so it survives the SSH session (spec §5.5.2 step 7).
- `func SubjectCtl(prefix, hostID string) string` — SubjectCtl returns P.ctl.<host_id> (director → daemon, request-reply).
- `func SubjectEvents(prefix, hostID string) string` — SubjectEvents returns P.events.<host_id> (daemon → director, host events).
- `func SubjectHandshake(prefix, hostID string) string` — SubjectHandshake returns P.handshake.<host_id> (daemon → director, request-reply).
- `func SubjectHandshakeAll(prefix string) string` — SubjectHandshakeAll returns P.handshake.> for subscribing to all handshakes (director).
- `func SubjectPTYIn(prefix, hostID, sessionID string) string` — SubjectPTYIn returns P.pty.<host_id>.<session_id>.in (director → daemon).
- `func SubjectPTYInWildcard(prefix, hostID string) string` — SubjectPTYInWildcard returns P.pty.<host_id>.*.in for subscribing to all sessions on a host.
- `func SubjectPTYOut(prefix, hostID, sessionID string) string` — SubjectPTYOut returns P.pty.<host_id>.<session_id>.out (daemon → director).
- `func SubjectPrefix(prefix string) string` — SubjectPrefix returns the subject prefix P (default "amux").
- `func WaitForConnection(ctx context.Context, cfg *BootstrapConfig, timeout time.Duration) error` — WaitForConnection waits until the daemon reports hub_connected or timeout.
- `func extractHostIDFromSubject(prefix, subject string) string`
- `func parseBufferSize(s string) int` — parseBufferSize parses remote.buffer_size per spec §4.2.8 (byte size: integer or NKB/NMB/NGB, binary).
- `func quoteRemotePath(p string) string`
- `type BootstrapConfig` — BootstrapConfig holds SSH target and paths for bootstrap (spec §5.5.2).
- `type ControlMessage` — ControlMessage is the top-level envelope for control requests and responses (spec §5.5.7.2).
- `type Director` — Director manages the director role: hub connectivity, handshake handling, and control request-reply.
- `type ErrorPayload` — ErrorPayload is the error response payload (spec §5.5.7.3).
- `type HandshakePayload` — HandshakePayload is the handshake request/response payload (spec §5.5.7.3).
- `type HostInfoValue` — HostInfoValue is the UTF-8 JSON value for hosts/<host_id>/info (spec §5.5.6.3).
- `type KillPayloadRequest` — KillPayloadRequest is the kill request payload (spec §5.5.7.3).
- `type KillPayloadResponse` — KillPayloadResponse is the kill response payload (spec §5.5.7.3).
- `type ManagedSession` — ManagedSession holds a remote session's replay buffer and metadata (spec §5.5.9).
- `type Manager` — Manager runs the manager role: leaf connection, handshake, and control/PTY handling.
- `type ReplayPayloadRequest` — ReplayPayloadRequest is the replay request payload (spec §5.5.7.3).
- `type ReplayPayloadResponse` — ReplayPayloadResponse is the replay response payload (spec §5.5.7.3).
- `type RingBuffer` — RingBuffer is a fixed-capacity byte ring buffer for PTY replay (spec §5.5.7.3).
- `type SessionKVValue` — SessionKVValue is the UTF-8 JSON value for sessions/<host_id>/<session_id> (spec §5.5.6.3).
- `type SpawnPayloadRequest` — SpawnPayloadRequest is the spawn request payload (spec §5.5.7.3).
- `type SpawnPayloadResponse` — SpawnPayloadResponse is the spawn response payload (spec §5.5.7.3).

### Constants

#### ControlTypeHandshake, ControlTypePing, ControlTypePong, ControlTypeSpawn, ControlTypeKill, ControlTypeReplay, ControlTypeError

```go
const (
	ControlTypeHandshake = "handshake"
	ControlTypePing      = "ping"
	ControlTypePong      = "pong"
	ControlTypeSpawn     = "spawn"
	ControlTypeKill      = "kill"
	ControlTypeReplay    = "replay"
	ControlTypeError     = "error"
)
```

Control message type constants (spec §5.5.7.2).

#### ErrorCodeNotReady, ErrorCodeSessionConflict, ErrorCodeInvalidRepo

```go
const (
	ErrorCodeNotReady        = "not_ready"
	ErrorCodeSessionConflict = "session_conflict"
	ErrorCodeInvalidRepo     = "invalid_repo"
)
```

Error codes (spec §5.5.7.3).


### Functions

#### CopyBootstrapZip

```go
func CopyBootstrapZip(ctx context.Context, cfg *BootstrapConfig, localZipPath string) (remotePath string, err error)
```

CopyBootstrapZip copies the local bootstrap zip to the remote bootstrap dir (spec §5.5.2 step 3).

#### CopyFile

```go
func CopyFile(ctx context.Context, cfg *BootstrapConfig, localPath, remotePath string) error
```

CopyFile copies a local file to the remote path via scp. Remote path is interpreted on the remote host.

#### DaemonStatus

```go
func DaemonStatus(ctx context.Context, cfg *BootstrapConfig) ([]byte, error)
```

DaemonStatus runs "amux-manager status" on the remote host (spec §5.5.2 step 6).

#### EnsureBootstrapDir

```go
func EnsureBootstrapDir(ctx context.Context, cfg *BootstrapConfig) error
```

EnsureBootstrapDir creates the remote bootstrap directory (e.g. ~/.amux/bootstrap).

#### EnsureKVBucket

```go
func EnsureKVBucket(ctx context.Context, js nats.JetStreamContext, bucket string) (nats.KeyValue, error)
```

EnsureKVBucket creates the JetStream KV bucket if it does not exist (spec §5.5.6.3).
Bucket name is from config (default AMUX_KV). Returns the KV store and any error.

#### KVKeyHostHeartbeat

```go
func KVKeyHostHeartbeat(hostID string) string
```

KVKeyHostHeartbeat returns the key for last-seen heartbeat: hosts/<host_id>/heartbeat.

#### KVKeyHostInfo

```go
func KVKeyHostInfo(hostID string) string
```

KVKeyHostInfo returns the key for host metadata: hosts/<host_id>/info.

#### KVKeySession

```go
func KVKeySession(hostID, sessionID string) string
```

KVKeySession returns the key for session metadata: sessions/<host_id>/<session_id>.

#### ProvisionCreds

```go
func ProvisionCreds(ctx context.Context, cfg *BootstrapConfig, credsContent []byte) error
```

ProvisionCreds writes credential content to the remote path with permissions 0600 (spec §5.5.6.4).
The director MUST ensure file permissions are no more permissive than 0600.

#### PutHostHeartbeat

```go
func PutHostHeartbeat(kv nats.KeyValue, hostID string) error
```

PutHostHeartbeat writes hosts/<host_id>/heartbeat (RFC 3339 timestamp).

#### PutHostInfo

```go
func PutHostInfo(kv nats.KeyValue, hostID string, v HostInfoValue) error
```

PutHostInfo writes hosts/<host_id>/info as UTF-8 JSON.

#### PutSession

```go
func PutSession(kv nats.KeyValue, hostID, sessionID string, v SessionKVValue) error
```

PutSession writes sessions/<host_id>/<session_id> as UTF-8 JSON.

#### RunSSH

```go
func RunSSH(ctx context.Context, cfg *BootstrapConfig, cmd string) ([]byte, error)
```

RunSSH runs a remote command via ssh. cmd is the remote command string (e.g. "amux-manager status").

#### StartDaemon

```go
func StartDaemon(ctx context.Context, cfg *BootstrapConfig, hostID, natsURL, credsPath string) error
```

StartDaemon starts the daemon in manager role so it survives the SSH session (spec §5.5.2 step 7).
It runs amux-manager daemon --role manager --host-id <hostID> --nats-url <url> --nats-creds <path>.

#### SubjectCtl

```go
func SubjectCtl(prefix, hostID string) string
```

SubjectCtl returns P.ctl.<host_id> (director → daemon, request-reply).

#### SubjectEvents

```go
func SubjectEvents(prefix, hostID string) string
```

SubjectEvents returns P.events.<host_id> (daemon → director, host events).

#### SubjectHandshake

```go
func SubjectHandshake(prefix, hostID string) string
```

SubjectHandshake returns P.handshake.<host_id> (daemon → director, request-reply).

#### SubjectHandshakeAll

```go
func SubjectHandshakeAll(prefix string) string
```

SubjectHandshakeAll returns P.handshake.> for subscribing to all handshakes (director).

#### SubjectPTYIn

```go
func SubjectPTYIn(prefix, hostID, sessionID string) string
```

SubjectPTYIn returns P.pty.<host_id>.<session_id>.in (director → daemon).

#### SubjectPTYInWildcard

```go
func SubjectPTYInWildcard(prefix, hostID string) string
```

SubjectPTYInWildcard returns P.pty.<host_id>.*.in for subscribing to all sessions on a host.

#### SubjectPTYOut

```go
func SubjectPTYOut(prefix, hostID, sessionID string) string
```

SubjectPTYOut returns P.pty.<host_id>.<session_id>.out (daemon → director).

#### SubjectPrefix

```go
func SubjectPrefix(prefix string) string
```

SubjectPrefix returns the subject prefix P (default "amux"). All remote subjects use P as literal prefix.

#### WaitForConnection

```go
func WaitForConnection(ctx context.Context, cfg *BootstrapConfig, timeout time.Duration) error
```

WaitForConnection waits until the daemon reports hub_connected or timeout.

#### extractHostIDFromSubject

```go
func extractHostIDFromSubject(prefix, subject string) string
```

#### parseBufferSize

```go
func parseBufferSize(s string) int
```

parseBufferSize parses remote.buffer_size per spec §4.2.8 (byte size: integer or NKB/NMB/NGB, binary).

#### quoteRemotePath

```go
func quoteRemotePath(p string) string
```


## type BootstrapConfig

```go
type BootstrapConfig struct {
	Host         string // SSH host (location.host)
	User         string // SSH user (optional; from location or SSH config)
	Port         int    // SSH port (optional; 0 = default)
	CredsPath    string // Remote path for NATS creds (remote.nats.creds_path)
	BootstrapDir string // Remote dir for bootstrap zip (e.g. ~/.amux/bootstrap)
	DaemonPath   string // Remote path for amux-manager binary (e.g. ~/.local/bin/amux-manager)
}
```

BootstrapConfig holds SSH target and paths for bootstrap (spec §5.5.2).

### Methods

#### BootstrapConfig.SSHArgs

```go
func () SSHArgs() []string
```

SSHArgs returns base SSH args (e.g. ["-p", "22", "user@host"]).

#### BootstrapConfig.SSHTarget

```go
func () SSHTarget() string
```

SSHTarget returns the SSH target string (user@host or host, with optional -p port).


## type ControlMessage

```go
type ControlMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
```

ControlMessage is the top-level envelope for control requests and responses (spec §5.5.7.2).

## type Director

```go
type Director struct {
	cfg    *config.RemoteConfig
	nc     *nats.Conn
	js     nats.JetStreamContext
	kv     nats.KeyValue
	prefix string

	mu         sync.RWMutex
	ready      map[string]struct{} // host_id that have completed handshake
	peerByHost map[string]string   // host_id -> peer_id (base-10)
}
```

Director manages the director role: hub connectivity, handshake handling, and control request-reply.

### Functions returning Director

#### NewDirector

```go
func NewDirector(cfg *config.RemoteConfig) *Director
```

NewDirector creates a director that will use the given remote config.


### Methods

#### Director.Close

```go
func () Close()
```

Close closes the NATS connection.

#### Director.Connect

```go
func () Connect(ctx context.Context, url string) error
```

Connect establishes the NATS connection to the hub and provisions the JetStream KV bucket.

#### Director.IsReady

```go
func () IsReady(hostID string) bool
```

IsReady returns true if the host has completed handshake (spec §5.5.7.2.1 fail-fast).

#### Director.Kill

```go
func () Kill(ctx context.Context, hostID string, sessionID string) (KillPayloadResponse, error)
```

Kill sends a kill request to P.ctl.<host_id> and waits for response.

#### Director.Replay

```go
func () Replay(ctx context.Context, hostID string, sessionID string) (ReplayPayloadResponse, error)
```

Replay sends a replay request to P.ctl.<host_id> and waits for response.

#### Director.RequestTimeout

```go
func () RequestTimeout() time.Duration
```

RequestTimeout returns the configured remote.request_timeout duration.

#### Director.RunHandshakeHandler

```go
func () RunHandshakeHandler(ctx context.Context) error
```

RunHandshakeHandler subscribes to P.handshake.> and handles handshake requests.
On success it records the host as ready and replies with director handshake payload.
Call once after Connect; it runs until the subscription is closed or context is done.

#### Director.Spawn

```go
func () Spawn(ctx context.Context, hostID string, req SpawnPayloadRequest) (SpawnPayloadResponse, error)
```

Spawn sends a spawn request to P.ctl.<host_id> and waits for response with request_timeout.
Fail-fast: if host is not ready, returns error without sending (spec §5.5.7.2.1).


## type ErrorPayload

```go
type ErrorPayload struct {
	RequestType string `json:"request_type"`
	Code        string `json:"code"`
	Message     string `json:"message"`
}
```

ErrorPayload is the error response payload (spec §5.5.7.3).
request_type is one of: "handshake", "spawn", "kill", "replay", "unknown".

## type HandshakePayload

```go
type HandshakePayload struct {
	Protocol int    `json:"protocol"`
	PeerID   string `json:"peer_id"` // base-10 string
	Role     string `json:"role"`    // "daemon" or "director"
	HostID   string `json:"host_id"`
}
```

HandshakePayload is the handshake request/response payload (spec §5.5.7.3).
Daemon sends to P.handshake.<host_id>; director replies with peer_id and role.

## type HostInfoValue

```go
type HostInfoValue struct {
	Version   string `json:"version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	PeerID    string `json:"peer_id"`
	StartedAt string `json:"started_at"` // RFC 3339
}
```

HostInfoValue is the UTF-8 JSON value for hosts/<host_id>/info (spec §5.5.6.3).

## type KillPayloadRequest

```go
type KillPayloadRequest struct {
	SessionID string `json:"session_id"`
}
```

KillPayloadRequest is the kill request payload (spec §5.5.7.3).

## type KillPayloadResponse

```go
type KillPayloadResponse struct {
	SessionID string `json:"session_id"`
	Killed    bool   `json:"killed"`
}
```

KillPayloadResponse is the kill response payload (spec §5.5.7.3).

## type ManagedSession

```go
type ManagedSession struct {
	ID         string
	AgentID    string
	AgentSlug  string
	RepoPath   string
	Buffer     *RingBuffer
	liveGate   sync.Mutex // guards live output until replay request handled
	replayDone bool
}
```

ManagedSession holds a remote session's replay buffer and metadata (spec §5.5.9).

## type Manager

```go
type Manager struct {
	cfg    *config.RemoteConfig
	hostID string
	nc     *nats.Conn
	prefix string
	bufCap int

	mu            sync.RWMutex
	handshakeDone bool
	sessions      map[string]*ManagedSession // session_id -> session
	agentByID     map[string]string          // agent_id -> session_id (for spawn idempotency)
}
```

Manager runs the manager role: leaf connection, handshake, and control/PTY handling.

### Functions returning Manager

#### NewManager

```go
func NewManager(cfg *config.RemoteConfig, hostID string) *Manager
```

NewManager creates a manager for the given host_id and remote config.


### Methods

#### Manager.Close

```go
func () Close()
```

Close closes the NATS connection and clears session state.

#### Manager.Connect

```go
func () Connect(ctx context.Context, url, credsPath string) error
```

Connect establishes the NATS connection to the hub (optionally with creds).

#### Manager.Handshake

```go
func () Handshake(ctx context.Context) error
```

Handshake sends the handshake request to P.handshake.<host_id> and waits for director reply.
Must be called after Connect and before accepting spawn/kill/replay (spec §5.5.7.3).

#### Manager.IsHandshakeDone

```go
func () IsHandshakeDone() bool
```

IsHandshakeDone returns true after Handshake has succeeded.

#### Manager.RunControlHandler

```go
func () RunControlHandler(ctx context.Context) error
```

RunControlHandler subscribes to P.ctl.<host_id> and handles spawn/kill/replay requests.
For spawn: idempotent by agent_id; session_conflict if repo_path or agent_slug differs (spec §5.5.7.3).
For replay: publishes replay buffer to P.pty.<host_id>.<session_id>.out then allows live output.

#### Manager.SubscribePTYIn

```go
func () SubscribePTYIn(ctx context.Context, sessionID string, fn func([]byte)) error
```

SubscribePTYIn subscribes to P.pty.<host_id>.<session_id>.in for a session and calls fn for each message.

#### Manager.WritePTYOut

```go
func () WritePTYOut(sessionID string, p []byte) error
```

WritePTYOut appends bytes to the session's replay buffer and publishes to P.pty.<host_id>.<session_id>.out
only if replay for this session has been handled (spec §5.5.8: no live publish until replay handled).

#### Manager.handleKill

```go
func () handleKill(msg *nats.Msg, payload json.RawMessage)
```

#### Manager.handleReplay

```go
func () handleReplay(msg *nats.Msg, payload json.RawMessage)
```

#### Manager.handleSpawn

```go
func () handleSpawn(msg *nats.Msg, payload json.RawMessage)
```

#### Manager.requestTimeout

```go
func () requestTimeout() time.Duration
```

#### Manager.respondError

```go
func () respondError(msg *nats.Msg, requestType, code, message string)
```

#### Manager.respondSpawn

```go
func () respondSpawn(msg *nats.Msg, resp SpawnPayloadResponse)
```


## type ReplayPayloadRequest

```go
type ReplayPayloadRequest struct {
	SessionID string `json:"session_id"`
}
```

ReplayPayloadRequest is the replay request payload (spec §5.5.7.3).

## type ReplayPayloadResponse

```go
type ReplayPayloadResponse struct {
	SessionID string `json:"session_id"`
	Accepted  bool   `json:"accepted"`
}
```

ReplayPayloadResponse is the replay response payload (spec §5.5.7.3).

## type RingBuffer

```go
type RingBuffer struct {
	mu     sync.Mutex
	buf    []byte
	cap    int
	start  int
	length int
}
```

RingBuffer is a fixed-capacity byte ring buffer for PTY replay (spec §5.5.7.3).
When cap is exceeded, oldest bytes are dropped. Safe for concurrent read/write.

### Functions returning RingBuffer

#### NewRingBuffer

```go
func NewRingBuffer(cap int) *RingBuffer
```

NewRingBuffer creates a ring buffer with the given capacity in bytes.
If cap <= 0, the buffer is disabled (Write is a no-op, Snapshot returns nil).


### Methods

#### RingBuffer.Cap

```go
func () Cap() int
```

Cap returns the buffer capacity in bytes (0 if disabled).

#### RingBuffer.Len

```go
func () Len() int
```

Len returns the current number of bytes in the buffer.

#### RingBuffer.Snapshot

```go
func () Snapshot() []byte
```

Snapshot returns a copy of the buffer contents in oldest-to-newest order.
Used when handling a replay request; caller must not modify the returned slice.

#### RingBuffer.Write

```go
func () Write(p []byte) (n int, err error)
```

Write appends data to the buffer. If the buffer would exceed capacity, the oldest bytes are dropped.


## type SessionKVValue

```go
type SessionKVValue struct {
	AgentID   string `json:"agent_id"`
	AgentSlug string `json:"agent_slug"`
	RepoPath  string `json:"repo_path"`
	State     string `json:"state"`
}
```

SessionKVValue is the UTF-8 JSON value for sessions/<host_id>/<session_id> (spec §5.5.6.3).

## type SpawnPayloadRequest

```go
type SpawnPayloadRequest struct {
	AgentID   string            `json:"agent_id"`
	AgentSlug string            `json:"agent_slug"`
	RepoPath  string            `json:"repo_path"`
	Command   []string          `json:"command"`
	Env       map[string]string `json:"env,omitempty"`
}
```

SpawnPayloadRequest is the spawn request payload (spec §5.5.7.3).

## type SpawnPayloadResponse

```go
type SpawnPayloadResponse struct {
	AgentID   string `json:"agent_id"`
	SessionID string `json:"session_id"`
}
```

SpawnPayloadResponse is the spawn response payload (spec §5.5.7.3).

