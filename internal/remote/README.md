# package remote

`import "github.com/stateforward/amux/internal/remote"`

Package remote implements Phase 3 remote agent orchestration.
This file implements SSH bootstrap per spec §5.5.2.

Package remote implements Phase 3 remote agent orchestration.
This file implements the director role per spec §5.5.4, §5.5.6, §5.5.7.

Package remote implements Phase 3 remote agent orchestration.
This file implements the manager role per spec §5.5.5.

Package remote implements Phase 3 remote agent orchestration via NATS + JetStream.

This package provides:
- Director role: hub NATS server + orchestration of remote hosts
- Manager role: leaf NATS connection + local PTY session management
- NATS protocol subjects, control messages, and handshake per spec §5.5.7
- SSH bootstrap for remote daemon installation per spec §5.5.2
- JetStream KV state for durable control-plane metadata per spec §5.5.6.3
- Per-host credentials and subject authorization per spec §5.5.6.4

- `func BootstrapRemoteHost(ctx context.Context, opts BootstrapOptions) error` — BootstrapRemoteHost performs SSH bootstrap per spec §5.5.2.
- `func FormatID(id muid.MUID) string` — FormatID formats a muid.MUID as a base-10 string for JSON encoding.
- `func MarshalControlMessage(msgType string, payload any) ([]byte, error)` — MarshalControlMessage marshals a control message with the given type and payload.
- `func ParseID(s string) (muid.MUID, error)` — ParseID parses a base-10 string into a muid.MUID.
- `func UnmarshalControlMessage(data []byte, payload any) (string, error)` — UnmarshalControlMessage unmarshals a control message and extracts the payload.
- `func checkRemoteDaemonStatus(ctx context.Context, host, user, port, identity string) (bool, error)` — checkRemoteDaemonStatus checks if the remote daemon is running per spec §5.5.2 step 6.
- `func copyToRemote(ctx context.Context, host, user, port, identity string, data []byte, remotePath string) error` — copyToRemote copies data to a remote path via SSH/SCP.
- `func createBootstrapZIP(ctx context.Context, adapterNames []string) ([]byte, error)` — createBootstrapZIP creates a bootstrap ZIP containing the daemon binary and adapter WASMs per spec §5.5.2.
- `func extractHostIDFromSubject(subject, prefix string) string` — extractHostIDFromSubject extracts the host_id from a NATS subject.
- `func generateNATSCredentials(hostID string) ([]byte, error)` — generateNATSCredentials generates per-host NATS credentials per spec §5.5.6.4.
- `func resolveSSHConfig(host string) (hostname, user, port, identityFile string, err error)` — resolveSSHConfig resolves SSH configuration using ~/.ssh/config per spec §5.2.
- `func runRemoteCommand(ctx context.Context, host, user, port, identity, command string) error` — runRemoteCommand runs a shell command on the remote host via SSH.
- `func runRemoteCommandOutput(ctx context.Context, host, user, port, identity, command string) ([]byte, error)` — runRemoteCommandOutput runs a shell command on the remote host and returns output.
- `func setRemoteFilePerms(ctx context.Context, host, user, port, identity, path, perms string) error` — setRemoteFilePerms sets file permissions on a remote file per spec §5.5.6.4.
- `func startRemoteDaemon(ctx context.Context, host, user, port, identity, hostID, hubURL, credsPath string) error` — startRemoteDaemon starts the remote daemon per spec §5.5.2 step 7.
- `func unpackRemoteBootstrap(ctx context.Context, host, user, port, identity, zipPath string) error` — unpackRemoteBootstrap unpacks the bootstrap ZIP on the remote host.
- `type BootstrapOptions` — BootstrapOptions holds options for SSH bootstrap per spec §5.5.2.
- `type ControlMessage` — ControlMessage is the top-level envelope for NATS control messages.
- `type Director` — Director implements the director role per spec §5.5.6, §5.5.7.
- `type ErrorPayload` — ErrorPayload represents an error response to a control request.
- `type HandshakePayload` — HandshakePayload represents a handshake request or response.
- `type HostState` — HostState tracks the state of a connected manager-role host.
- `type KillRequestPayload` — KillRequestPayload represents a kill control request from director to manager.
- `type KillResponsePayload` — KillResponsePayload represents the manager's response to a kill request.
- `type Manager` — Manager implements the manager role for a remote host per spec §5.5.5.
- `type PingPayload` — PingPayload represents a ping message.
- `type PongPayload` — PongPayload represents a pong message.
- `type RemoteSession` — RemoteSession represents a single agent PTY session managed by the manager role.
- `type ReplayRequestPayload` — ReplayRequestPayload represents a replay control request from director to manager.
- `type ReplayResponsePayload` — ReplayResponsePayload represents the manager's response to a replay request.
- `type RingBuffer` — RingBuffer implements a ring buffer for PTY output replay per spec §5.5.7.3.
- `type SpawnRequestPayload` — SpawnRequestPayload represents a spawn control request from director to manager.
- `type SpawnResponsePayload` — SpawnResponsePayload represents the manager's response to a spawn request.
- `type SubjectBuilder` — SubjectBuilder constructs NATS subjects using the configured prefix.

### Functions

#### BootstrapRemoteHost

```go
func BootstrapRemoteHost(ctx context.Context, opts BootstrapOptions) error
```

BootstrapRemoteHost performs SSH bootstrap per spec §5.5.2.

Steps:
1. Resolve SSH target host using location.host and ~/.ssh/config
2. Construct bootstrap ZIP (daemon binary + adapter WASMs)
3. Copy bootstrap ZIP to remote host
4. Unpack ZIP and install daemon + adapters
5. Provision leaf→hub connection material (NATS credentials)
6. Check if daemon is running
7. Start daemon if not running
8. Verify daemon has connected to hub

#### FormatID

```go
func FormatID(id muid.MUID) string
```

FormatID formats a muid.MUID as a base-10 string for JSON encoding.
Per spec §9.1.3.1, IDs are encoded as base-10 unsigned integer strings.

#### MarshalControlMessage

```go
func MarshalControlMessage(msgType string, payload any) ([]byte, error)
```

MarshalControlMessage marshals a control message with the given type and payload.

#### ParseID

```go
func ParseID(s string) (muid.MUID, error)
```

ParseID parses a base-10 string into a muid.MUID.

#### UnmarshalControlMessage

```go
func UnmarshalControlMessage(data []byte, payload any) (string, error)
```

UnmarshalControlMessage unmarshals a control message and extracts the payload.

#### checkRemoteDaemonStatus

```go
func checkRemoteDaemonStatus(ctx context.Context, host, user, port, identity string) (bool, error)
```

checkRemoteDaemonStatus checks if the remote daemon is running per spec §5.5.2 step 6.

#### copyToRemote

```go
func copyToRemote(ctx context.Context, host, user, port, identity string, data []byte, remotePath string) error
```

copyToRemote copies data to a remote path via SSH/SCP.

#### createBootstrapZIP

```go
func createBootstrapZIP(ctx context.Context, adapterNames []string) ([]byte, error)
```

createBootstrapZIP creates a bootstrap ZIP containing the daemon binary and adapter WASMs per spec §5.5.2.

#### extractHostIDFromSubject

```go
func extractHostIDFromSubject(subject, prefix string) string
```

extractHostIDFromSubject extracts the host_id from a NATS subject.
For P.handshake.<host_id>, returns <host_id>.

#### generateNATSCredentials

```go
func generateNATSCredentials(hostID string) ([]byte, error)
```

generateNATSCredentials generates per-host NATS credentials per spec §5.5.6.4.

For Phase 3, we generate a placeholder credential.
In a production implementation, this would generate NKey + JWT using NATS nkeys.

#### resolveSSHConfig

```go
func resolveSSHConfig(host string) (hostname, user, port, identityFile string, err error)
```

resolveSSHConfig resolves SSH configuration using ~/.ssh/config per spec §5.2.

#### runRemoteCommand

```go
func runRemoteCommand(ctx context.Context, host, user, port, identity, command string) error
```

runRemoteCommand runs a shell command on the remote host via SSH.

#### runRemoteCommandOutput

```go
func runRemoteCommandOutput(ctx context.Context, host, user, port, identity, command string) ([]byte, error)
```

runRemoteCommandOutput runs a shell command on the remote host and returns output.

#### setRemoteFilePerms

```go
func setRemoteFilePerms(ctx context.Context, host, user, port, identity, path, perms string) error
```

setRemoteFilePerms sets file permissions on a remote file per spec §5.5.6.4.

#### startRemoteDaemon

```go
func startRemoteDaemon(ctx context.Context, host, user, port, identity, hostID, hubURL, credsPath string) error
```

startRemoteDaemon starts the remote daemon per spec §5.5.2 step 7.

#### unpackRemoteBootstrap

```go
func unpackRemoteBootstrap(ctx context.Context, host, user, port, identity, zipPath string) error
```

unpackRemoteBootstrap unpacks the bootstrap ZIP on the remote host.


## type BootstrapOptions

```go
type BootstrapOptions struct {
	Host         string   // SSH host (supports ~/.ssh/config aliases)
	HostID       string   // Unique host identifier
	HubURL       string   // NATS hub URL to advertise
	CredsPath    string   // Remote path for NATS credentials
	AdapterNames []string // Required adapter WASM modules
}
```

BootstrapOptions holds options for SSH bootstrap per spec §5.5.2.

## type ControlMessage

```go
type ControlMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
```

ControlMessage is the top-level envelope for NATS control messages.
Per spec §5.5.7.2, all control requests and responses use this shape.

## type Director

```go
type Director struct {
	cfg      *config.Config
	peerID   muid.MUID
	nc       *nats.Conn
	js       jetstream.JetStream
	kv       jetstream.KeyValue
	subjects SubjectBuilder

	hostsMu sync.RWMutex
	hosts   map[string]*HostState // hostID -> state
}
```

Director implements the director role per spec §5.5.6, §5.5.7.

A director:
- Runs (or connects to) a hub-mode NATS server with JetStream enabled
- Provisions JetStream KV bucket for durable remote control-plane state
- Accepts handshake requests from manager-role nodes
- Sends spawn/kill/replay control requests to managers
- Subscribes to PTY output and host events from managers

### Functions returning Director

#### NewDirector

```go
func NewDirector(ctx context.Context, cfg *config.Config, peerID muid.MUID) (*Director, error)
```

NewDirector creates a new director instance.

The director connects to the hub NATS server at cfg.NATS.Listen (or cfg.Remote.NATS.URL)
and provisions the required JetStream KV bucket per spec §5.5.6.3.


### Methods

#### Director.Close

```go
func () Close() error
```

Close closes the director and releases resources.

#### Director.Kill

```go
func () Kill(ctx context.Context, hostID string, req KillRequestPayload) (*KillResponsePayload, error)
```

Kill sends a kill control request to the target host per spec §5.5.7.2, §5.5.7.3.

#### Director.PublishPTYInput

```go
func () PublishPTYInput(ctx context.Context, hostID string, sessionID muid.MUID, data []byte) error
```

PublishPTYInput publishes PTY input for a specific session per spec §5.5.7.4.

#### Director.Replay

```go
func () Replay(ctx context.Context, hostID string, req ReplayRequestPayload) (*ReplayResponsePayload, error)
```

Replay sends a replay control request to the target host per spec §5.5.7.2, §5.5.7.3.

#### Director.Spawn

```go
func () Spawn(ctx context.Context, hostID string, req SpawnRequestPayload) (*SpawnResponsePayload, error)
```

Spawn sends a spawn control request to the target host per spec §5.5.7.2, §5.5.7.3.

#### Director.SubscribePTYOutput

```go
func () SubscribePTYOutput(ctx context.Context, hostID string, sessionID muid.MUID, handler func([]byte)) error
```

SubscribePTYOutput subscribes to PTY output for a specific session per spec §5.5.7.4.

#### Director.handleHandshake

```go
func () handleHandshake(ctx context.Context, msg *nats.Msg)
```

handleHandshake handles a handshake request from a manager per spec §5.5.7.3.

#### Director.provisionKV

```go
func () provisionKV(ctx context.Context) error
```

provisionKV provisions the JetStream KV bucket per spec §5.5.6.3.

#### Director.replyHandshakeError

```go
func () replyHandshakeError(msg *nats.Msg, code, message string) error
```

replyHandshakeError sends an error response to a handshake request.

#### Director.subscribeHandshake

```go
func () subscribeHandshake(ctx context.Context) error
```

subscribeHandshake subscribes to handshake requests on P.handshake.* per spec §5.5.7.3.


## type ErrorPayload

```go
type ErrorPayload struct {
	RequestType string `json:"request_type"` // "handshake", "spawn", "kill", "replay", "unknown"
	Code        string `json:"code"`
	Message     string `json:"message"`
}
```

ErrorPayload represents an error response to a control request.
Per spec §5.5.7.3.

## type HandshakePayload

```go
type HandshakePayload struct {
	Protocol int    `json:"protocol"`
	PeerID   string `json:"peer_id"` // base-10 MUID
	Role     string `json:"role"`    // "director" or "daemon"
	HostID   string `json:"host_id"`
}
```

HandshakePayload represents a handshake request or response.
Per spec §5.5.7.3, both daemon→director and director→daemon use this shape.

## type HostState

```go
type HostState struct {
	HostID     string
	PeerID     muid.MUID
	Connected  bool
	Handshaken bool
}
```

HostState tracks the state of a connected manager-role host.

## type KillRequestPayload

```go
type KillRequestPayload struct {
	SessionID string `json:"session_id"` // base-10 MUID
}
```

KillRequestPayload represents a kill control request from director to manager.
Per spec §5.5.7.3.

## type KillResponsePayload

```go
type KillResponsePayload struct {
	SessionID string `json:"session_id"` // base-10 MUID
	Killed    bool   `json:"killed"`
}
```

KillResponsePayload represents the manager's response to a kill request.
Per spec §5.5.7.3.

## type Manager

```go
type Manager struct {
	cfg        *config.Config
	hostID     string
	peerID     muid.MUID
	nc         *nats.Conn
	subjects   SubjectBuilder
	handshaken bool

	sessionsMu sync.RWMutex
	sessions   map[muid.MUID]*RemoteSession // sessionID -> session
}
```

Manager implements the manager role for a remote host per spec §5.5.5.

A manager:
- Connects to the director's hub NATS server via a leaf-mode connection
- Owns PTY sessions for agents on this host
- Handles spawn/kill/replay control requests from the director
- Streams PTY output to the director
- Maintains per-session replay buffers

### Functions returning Manager

#### NewManager

```go
func NewManager(ctx context.Context, cfg *config.Config, hostID string, peerID muid.MUID) (*Manager, error)
```

NewManager creates a new manager instance.

The manager connects to the hub using cfg.Remote.NATS.URL and cfg.Remote.NATS.CredsPath.
The hostID MUST be unique among concurrently connected hosts.


### Methods

#### Manager.Close

```go
func () Close() error
```

Close closes the manager and all managed sessions.

#### Manager.expandHomeDir

```go
func () expandHomeDir(path string) string
```

expandHomeDir expands ~/ to the user's home directory.

#### Manager.handleControlRequest

```go
func () handleControlRequest(ctx context.Context, msg *nats.Msg)
```

handleControlRequest handles incoming control requests from the director.

#### Manager.handleKill

```go
func () handleKill(ctx context.Context, msg *nats.Msg, payloadRaw json.RawMessage)
```

handleKill handles a kill control request per spec §5.5.7.3.

#### Manager.handlePing

```go
func () handlePing(ctx context.Context, msg *nats.Msg, payloadRaw json.RawMessage)
```

handlePing handles a ping control request per spec §5.5.7.3.

#### Manager.handleReplay

```go
func () handleReplay(ctx context.Context, msg *nats.Msg, payloadRaw json.RawMessage)
```

handleReplay handles a replay control request per spec §5.5.7.3.

#### Manager.handleSpawn

```go
func () handleSpawn(ctx context.Context, msg *nats.Msg, payloadRaw json.RawMessage)
```

handleSpawn handles a spawn control request per spec §5.5.7.3.

#### Manager.handshake

```go
func () handshake(ctx context.Context) error
```

handshake performs the initial handshake with the director per spec §5.5.7.3.

#### Manager.publishChunked

```go
func () publishChunked(subject string, data []byte, maxChunkSize int)
```

publishChunked publishes data in chunks not exceeding maxChunkSize per spec §5.5.7.4.

#### Manager.replyError

```go
func () replyError(msg *nats.Msg, requestType, code, message string) error
```

replyError sends an error response.

#### Manager.streamPTYOutput

```go
func () streamPTYOutput(ctx context.Context, sess *RemoteSession)
```

streamPTYOutput streams PTY output to the director per spec §5.5.7.4.

#### Manager.subscribeControl

```go
func () subscribeControl(ctx context.Context) error
```

subscribeControl subscribes to control requests on P.ctl.<host_id> per spec §5.5.7.2.


## type PingPayload

```go
type PingPayload struct {
	TsUnixMs int64 `json:"ts_unix_ms"`
}
```

PingPayload represents a ping message.
Per spec §5.5.7.3.

## type PongPayload

```go
type PongPayload struct {
	TsUnixMs int64 `json:"ts_unix_ms"`
}
```

PongPayload represents a pong message.
Per spec §5.5.7.3.

## type RemoteSession

```go
type RemoteSession struct {
	SessionID muid.MUID
	AgentID   muid.MUID
	AgentSlug string
	RepoPath  string
	Cmd       []string
	Env       map[string]string

	PTY       *os.File
	LocalSess *agent.LocalSession

	// Replay buffer per spec §5.5.7.3
	replayMu     sync.Mutex
	replayBuffer *RingBuffer
	replayActive bool // true while replaying, blocks live output
}
```

RemoteSession represents a single agent PTY session managed by the manager role.

## type ReplayRequestPayload

```go
type ReplayRequestPayload struct {
	SessionID string `json:"session_id"` // base-10 MUID
}
```

ReplayRequestPayload represents a replay control request from director to manager.
Per spec §5.5.7.3.

## type ReplayResponsePayload

```go
type ReplayResponsePayload struct {
	SessionID string `json:"session_id"` // base-10 MUID
	Accepted  bool   `json:"accepted"`
}
```

ReplayResponsePayload represents the manager's response to a replay request.
Per spec §5.5.7.3.

## type RingBuffer

```go
type RingBuffer struct {
	data []byte
	cap  int
	head int
	tail int
	full bool
}
```

RingBuffer implements a ring buffer for PTY output replay per spec §5.5.7.3.

### Functions returning RingBuffer

#### NewRingBuffer

```go
func NewRingBuffer(capacity int) *RingBuffer
```

NewRingBuffer creates a new ring buffer with the given capacity.


### Methods

#### RingBuffer.Snapshot

```go
func () Snapshot() []byte
```

Snapshot returns a snapshot of the current buffer contents in oldest-to-newest order.

#### RingBuffer.Write

```go
func () Write(p []byte) (n int, err error)
```

Write appends data to the ring buffer, overwriting oldest data if at capacity.


## type SpawnRequestPayload

```go
type SpawnRequestPayload struct {
	AgentID   string            `json:"agent_id"` // base-10 MUID
	AgentSlug string            `json:"agent_slug"`
	RepoPath  string            `json:"repo_path"`
	Command   []string          `json:"command"`
	Env       map[string]string `json:"env,omitempty"`
}
```

SpawnRequestPayload represents a spawn control request from director to manager.
Per spec §5.5.7.3.

## type SpawnResponsePayload

```go
type SpawnResponsePayload struct {
	AgentID   string `json:"agent_id"`   // base-10 MUID (echoed from request)
	SessionID string `json:"session_id"` // base-10 MUID
}
```

SpawnResponsePayload represents the manager's response to a spawn request.
Per spec §5.5.7.3.

## type SubjectBuilder

```go
type SubjectBuilder struct {
	Prefix string
}
```

SubjectBuilder constructs NATS subjects using the configured prefix.
Per spec §5.5.7.1, the subject prefix is configurable (default "amux").

### Methods

#### SubjectBuilder.CommAgent

```go
func () CommAgent(hostID string, agentID muid.MUID) string
```

CommAgent returns the agent communication channel subject for the given host_id and agent_id.
Subject: P.comm.agent.<host_id>.<agent_id>

#### SubjectBuilder.CommBroadcast

```go
func () CommBroadcast() string
```

CommBroadcast returns the broadcast communication channel subject.
Subject: P.comm.broadcast

#### SubjectBuilder.CommDirector

```go
func () CommDirector() string
```

CommDirector returns the director communication channel subject.
Subject: P.comm.director

#### SubjectBuilder.CommManager

```go
func () CommManager(hostID string) string
```

CommManager returns the manager communication channel subject for the given host_id.
Subject: P.comm.manager.<host_id>

#### SubjectBuilder.Control

```go
func () Control(hostID string) string
```

Control returns the control request subject for the given host_id.
Subject: P.ctl.<host_id>

#### SubjectBuilder.Events

```go
func () Events(hostID string) string
```

Events returns the host events subject for the given host_id.
Subject: P.events.<host_id>

#### SubjectBuilder.Handshake

```go
func () Handshake(hostID string) string
```

Handshake returns the handshake request subject for the given host_id.
Subject: P.handshake.<host_id>

#### SubjectBuilder.PTYIn

```go
func () PTYIn(hostID string, sessionID muid.MUID) string
```

PTYIn returns the PTY input subject for the given host_id and session_id.
Subject: P.pty.<host_id>.<session_id>.in

#### SubjectBuilder.PTYOut

```go
func () PTYOut(hostID string, sessionID muid.MUID) string
```

PTYOut returns the PTY output subject for the given host_id and session_id.
Subject: P.pty.<host_id>.<session_id>.out


