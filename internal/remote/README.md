# package remote

`import "github.com/copilot-claude-sonnet-4/amux/internal/remote"`

Package remote implements remote host management, SSH bootstrap, NATS connectivity
and control plane operations for distributed agent orchestration.

Package remote implements control operations for remote agent management.

Package remote implements the main remote host manager and director functionality.

Package remote implements NATS-based communication for remote agents.

Package remote implements PTY streaming and buffering for remote sessions.

- `ErrBootstrapFailed, ErrSSHFailed, ErrCredentialsFailed, ErrHostUnreachable` — Common sentinel errors for remote operations.
- `ErrNATSConnectionFailed, ErrJetStreamFailed, ErrHandshakeFailed, ErrNotReady` — Common sentinel errors for NATS operations.
- `func Bootstrap(ctx context.Context, hostID string, config BootstrapConfig) error` — Bootstrap performs SSH bootstrap for a remote host.
- `func GenerateHostID() string` — GenerateHostID creates a new unique host identifier.
- `func addFileToZip(w *zip.Writer, srcPath, zipPath string) error` — addFileToZip adds a file to a ZIP archive.
- `func copyToRemote(ctx context.Context, sshHost string, data []byte, remotePath string, timeout time.Duration) error` — copyToRemote copies data to a remote path via SSH.
- `func createBootstrapZip(config BootstrapConfig) ([]byte, error)` — createBootstrapZip creates a ZIP file containing the binary and adapter modules.
- `func ensureDaemonRunning(ctx context.Context, hostID string, config BootstrapConfig) error` — ensureDaemonRunning starts the daemon on the remote host if not already running.
- `func extractBootstrapOnRemote(ctx context.Context, sshHost, remotePath string, timeout time.Duration) error` — extractBootstrapOnRemote extracts the bootstrap ZIP on the remote host.
- `func provisionCredentials(ctx context.Context, sshHost, localPath, remotePath string, timeout time.Duration) error` — provisionCredentials copies NATS credentials to the remote host with proper permissions.
- `type BootstrapConfig` — BootstrapConfig holds configuration for SSH bootstrap operations.
- `type ControlMessage` — ControlMessage represents a control protocol message.
- `type ControlOperations` — ControlOperations provides remote control operations for director role.
- `type DirectorOperations` — DirectorOperations provides control operations for director role.
- `type ErrorPayload` — ErrorPayload represents error message payload.
- `type HandshakePayload` — HandshakePayload represents handshake message payload.
- `type NATSConfig` — NATSConfig holds NATS connection configuration.
- `type NATSManager` — NATSManager manages NATS connectivity and protocol operations.
- `type PTYData` — PTYData represents PTY input/output data.
- `type PTYStreamer` — PTYStreamer manages PTY streaming for remote sessions.
- `type RemoteConfig` — RemoteConfig holds configuration for remote operations.
- `type RemoteManager` — RemoteManager coordinates all remote operations.
- `type RingBuffer` — RingBuffer implements a thread-safe ring buffer for PTY data replay.
- `type SpawnPayload` — SpawnPayload represents spawn request/response payload.

### Variables

#### ErrBootstrapFailed, ErrSSHFailed, ErrCredentialsFailed, ErrHostUnreachable

```go
var (
	// ErrBootstrapFailed indicates SSH bootstrap failed.
	ErrBootstrapFailed = error(fmt.Errorf("bootstrap failed"))

	// ErrSSHFailed indicates SSH command execution failed.
	ErrSSHFailed = error(fmt.Errorf("ssh failed"))

	// ErrCredentialsFailed indicates credential provisioning failed.
	ErrCredentialsFailed = error(fmt.Errorf("credentials failed"))

	// ErrHostUnreachable indicates remote host is not reachable.
	ErrHostUnreachable = error(fmt.Errorf("host unreachable"))
)
```

Common sentinel errors for remote operations.

#### ErrNATSConnectionFailed, ErrJetStreamFailed, ErrHandshakeFailed, ErrNotReady

```go
var (
	// ErrNATSConnectionFailed indicates NATS connection failed.
	ErrNATSConnectionFailed = fmt.Errorf("nats connection failed")

	// ErrJetStreamFailed indicates JetStream operation failed.
	ErrJetStreamFailed = fmt.Errorf("jetstream failed")

	// ErrHandshakeFailed indicates handshake exchange failed.
	ErrHandshakeFailed = fmt.Errorf("handshake failed")

	// ErrNotReady indicates daemon is not ready for operations.
	ErrNotReady = fmt.Errorf("not ready")
)
```

Common sentinel errors for NATS operations.


### Functions

#### Bootstrap

```go
func Bootstrap(ctx context.Context, hostID string, config BootstrapConfig) error
```

Bootstrap performs SSH bootstrap for a remote host.
Implements §5.5.2 daemon bootstrap requirements.

#### GenerateHostID

```go
func GenerateHostID() string
```

GenerateHostID creates a new unique host identifier.

#### addFileToZip

```go
func addFileToZip(w *zip.Writer, srcPath, zipPath string) error
```

addFileToZip adds a file to a ZIP archive.

#### copyToRemote

```go
func copyToRemote(ctx context.Context, sshHost string, data []byte, remotePath string, timeout time.Duration) error
```

copyToRemote copies data to a remote path via SSH.

#### createBootstrapZip

```go
func createBootstrapZip(config BootstrapConfig) ([]byte, error)
```

createBootstrapZip creates a ZIP file containing the binary and adapter modules.

#### ensureDaemonRunning

```go
func ensureDaemonRunning(ctx context.Context, hostID string, config BootstrapConfig) error
```

ensureDaemonRunning starts the daemon on the remote host if not already running.

#### extractBootstrapOnRemote

```go
func extractBootstrapOnRemote(ctx context.Context, sshHost, remotePath string, timeout time.Duration) error
```

extractBootstrapOnRemote extracts the bootstrap ZIP on the remote host.

#### provisionCredentials

```go
func provisionCredentials(ctx context.Context, sshHost, localPath, remotePath string, timeout time.Duration) error
```

provisionCredentials copies NATS credentials to the remote host with proper permissions.


## type BootstrapConfig

```go
type BootstrapConfig struct {
	SSHHost         string        // SSH target (e.g., "user@host")
	BinaryPath      string        // Path to amux binary for target arch
	AdapterPaths    []string      // Paths to required adapter WASM modules
	CredsPath       string        // Local path to NATS credentials file
	RemoteCredsPath string        // Remote path for credentials (e.g., ~/.amux/nats.creds)
	HubURL          string        // NATS hub URL to configure
	Timeout         time.Duration // SSH operation timeout
}
```

BootstrapConfig holds configuration for SSH bootstrap operations.

## type ControlMessage

```go
type ControlMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
```

ControlMessage represents a control protocol message.

## type ControlOperations

```go
type ControlOperations struct {
	nm    *NATSManager
	subs  map[string]*nats.Subscription
	mutex sync.RWMutex
}
```

ControlOperations provides remote control operations for director role.

### Functions returning ControlOperations

#### NewControlOperations

```go
func NewControlOperations(nm *NATSManager) *ControlOperations
```

NewControlOperations creates a new control operations manager.


### Methods

#### ControlOperations.Close

```go
func () Close()
```

Close cleans up control operations.

#### ControlOperations.StartControlSubscriptions

```go
func () StartControlSubscriptions() error
```

StartControlSubscriptions starts control request subscriptions for manager role.

#### ControlOperations.getExistingSession

```go
func () getExistingSession(agentID string) string
```

getExistingSession checks if a session already exists for the given agent ID.

#### ControlOperations.handleControlRequest

```go
func () handleControlRequest(msg *nats.Msg)
```

handleControlRequest handles incoming control requests (manager role).

#### ControlOperations.handleKillRequest

```go
func () handleKillRequest(replyTo string, payload json.RawMessage)
```

handleKillRequest handles agent kill requests.

#### ControlOperations.handlePingRequest

```go
func () handlePingRequest(replyTo string, payload json.RawMessage)
```

handlePingRequest handles ping requests.

#### ControlOperations.handleReplayRequest

```go
func () handleReplayRequest(replyTo string, payload json.RawMessage)
```

handleReplayRequest handles PTY replay requests.

#### ControlOperations.handleSpawnRequest

```go
func () handleSpawnRequest(replyTo string, payload json.RawMessage)
```

handleSpawnRequest handles agent spawn requests.


## type DirectorOperations

```go
type DirectorOperations struct {
	nm         *NATSManager
	connStates map[string]bool // hostID -> connected
	mutex      sync.RWMutex
}
```

DirectorOperations provides control operations for director role.

### Functions returning DirectorOperations

#### NewDirectorOperations

```go
func NewDirectorOperations(nm *NATSManager) *DirectorOperations
```

NewDirectorOperations creates a new director operations manager.


### Methods

#### DirectorOperations.KillAgent

```go
func () KillAgent(ctx context.Context, hostID, agentID string) error
```

KillAgent kills an agent on the specified remote host.

#### DirectorOperations.MarkHostConnected

```go
func () MarkHostConnected(hostID string)
```

MarkHostConnected marks a host as connected.

#### DirectorOperations.MarkHostDisconnected

```go
func () MarkHostDisconnected(hostID string)
```

MarkHostDisconnected marks a host as disconnected.

#### DirectorOperations.PingHost

```go
func () PingHost(ctx context.Context, hostID string) error
```

PingHost sends a ping to verify host connectivity.

#### DirectorOperations.ReplayPTY

```go
func () ReplayPTY(ctx context.Context, hostID, sessionID string) error
```

ReplayPTY requests PTY replay for a session.

#### DirectorOperations.SpawnAgent

```go
func () SpawnAgent(ctx context.Context, hostID string, req SpawnPayload) (*SpawnPayload, error)
```

SpawnAgent spawns an agent on the specified remote host.

#### DirectorOperations.isHostConnected

```go
func () isHostConnected(hostID string) bool
```

isHostConnected checks if a host is currently connected.


## type ErrorPayload

```go
type ErrorPayload struct {
	RequestType string `json:"request_type"`
	Code        string `json:"code"`
	Message     string `json:"message"`
}
```

ErrorPayload represents error message payload.

## type HandshakePayload

```go
type HandshakePayload struct {
	Protocol int    `json:"protocol"`
	PeerID   string `json:"peer_id"` // base-10 string
	Role     string `json:"role"`
	HostID   string `json:"host_id"`
}
```

HandshakePayload represents handshake message payload.

## type NATSConfig

```go
type NATSConfig struct {
	URL           string        // NATS server URL
	CredsFile     string        // Path to NATS credentials file
	SubjectPrefix string        // Subject prefix (default "amux")
	KVBucket      string        // JetStream KV bucket name
	Timeout       time.Duration // Request timeout
}
```

NATSConfig holds NATS connection configuration.

## type NATSManager

```go
type NATSManager struct {
	conn   *nats.Conn
	js     nats.JetStreamContext
	kv     nats.KeyValue
	config NATSConfig
	hostID string
	peerID muid.MUID
	role   string // "director" or "manager"
	ready  bool   // handshake completed
}
```

NATSManager manages NATS connectivity and protocol operations.

### Functions returning NATSManager

#### NewNATSManager

```go
func NewNATSManager(hostID, role string, config NATSConfig) (*NATSManager, error)
```

NewNATSManager creates a new NATS manager instance.


### Methods

#### NATSManager.Close

```go
func () Close()
```

Close closes the NATS connection.

#### NATSManager.Connect

```go
func () Connect() error
```

Connect establishes NATS connection and sets up JetStream.

#### NATSManager.GetHostID

```go
func () GetHostID() string
```

GetHostID returns the host ID.

#### NATSManager.GetPeerID

```go
func () GetPeerID() muid.MUID
```

GetPeerID returns the peer ID.

#### NATSManager.Handshake

```go
func () Handshake() error
```

Handshake performs the initial handshake exchange.

#### NATSManager.IsReady

```go
func () IsReady() bool
```

IsReady returns true if handshake has been completed.

#### NATSManager.Subject

```go
func () Subject(parts ...string) string
```

Subject returns a fully qualified subject name.

#### NATSManager.ensureKVBucket

```go
func () ensureKVBucket() error
```

ensureKVBucket creates the KV bucket if it doesn't exist.

#### NATSManager.getKVBucket

```go
func () getKVBucket() error
```

getKVBucket gets the KV bucket (manager role).

#### NATSManager.handleHandshakeRequest

```go
func () handleHandshakeRequest(msg *nats.Msg)
```

handleHandshakeRequest handles incoming handshake requests (director role).

#### NATSManager.performManagerHandshake

```go
func () performManagerHandshake() error
```

performManagerHandshake sends a handshake request to the director.

#### NATSManager.sendControlResponse

```go
func () sendControlResponse(replyTo, msgType string, payload interface{})
```

sendControlResponse sends a successful control response.

#### NATSManager.sendErrorResponse

```go
func () sendErrorResponse(replyTo, requestType, code, message string)
```

sendErrorResponse sends an error response message.

#### NATSManager.subscribeToHandshakes

```go
func () subscribeToHandshakes() error
```

subscribeToHandshakes sets up handshake subscription for director role.

#### NATSManager.updateHostInfo

```go
func () updateHostInfo(hostID, peerID string) error
```

updateHostInfo updates host information in the KV store.


## type PTYData

```go
type PTYData struct {
	SessionID string    `json:"session_id"`
	Data      []byte    `json:"data"`
	Timestamp time.Time `json:"timestamp"`
	Sequence  uint64    `json:"sequence"`
}
```

PTYData represents PTY input/output data.

## type PTYStreamer

```go
type PTYStreamer struct {
	nm         *NATSManager
	hostID     string
	buffers    map[string]*RingBuffer // sessionID -> buffer
	sequences  map[string]uint64      // sessionID -> sequence counter
	subs       map[string]*nats.Subscription
	mutex      sync.RWMutex
	bufferSize int
}
```

PTYStreamer manages PTY streaming for remote sessions.

### Functions returning PTYStreamer

#### NewPTYStreamer

```go
func NewPTYStreamer(nm *NATSManager, bufferSize int) *PTYStreamer
```

NewPTYStreamer creates a new PTY streamer.


### Methods

#### PTYStreamer.Close

```go
func () Close()
```

Close cleans up PTY streaming resources.

#### PTYStreamer.PublishPTYOutput

```go
func () PublishPTYOutput(sessionID string, data []byte) error
```

PublishPTYOutput publishes PTY output data to NATS.

#### PTYStreamer.ReplayPTYOutput

```go
func () ReplayPTYOutput(sessionID string, handler func(PTYData)) error
```

ReplayPTYOutput replays buffered PTY output for a session.

#### PTYStreamer.SendPTYInput

```go
func () SendPTYInput(hostID, sessionID string, data []byte) error
```

SendPTYInput sends input to a remote PTY session (director role).

#### PTYStreamer.StartPTYStreaming

```go
func () StartPTYStreaming(sessionID string) error
```

StartPTYStreaming starts PTY streaming for a session.

#### PTYStreamer.StopPTYStreaming

```go
func () StopPTYStreaming(sessionID string)
```

StopPTYStreaming stops PTY streaming for a session.

#### PTYStreamer.SubscribeToPTYOutput

```go
func () SubscribeToPTYOutput(hostID, sessionID string, handler func(PTYData)) error
```

SubscribeToPTYOutput subscribes to PTY output from a remote session (director role).

#### PTYStreamer.handlePTYInput

```go
func () handlePTYInput(msg *nats.Msg)
```

handlePTYInput handles incoming PTY input from director.

#### PTYStreamer.publishChunked

```go
func () publishChunked(sessionID string, data []byte, baseSeq uint64) error
```

publishChunked publishes large PTY data in chunks.


## type RemoteConfig

```go
type RemoteConfig struct {
	Role           string        // "director" or "manager"
	HostID         string        // Host identifier
	NATSURL        string        // NATS server URL
	CredsPath      string        // NATS credentials file
	SubjectPrefix  string        // NATS subject prefix
	KVBucket       string        // JetStream KV bucket
	RequestTimeout time.Duration // Request timeout
	BufferSize     int           // PTY buffer size
}
```

RemoteConfig holds configuration for remote operations.

## type RemoteManager

```go
type RemoteManager struct {
	config      *RemoteConfig
	nats        *NATSManager
	control     *ControlOperations
	director    *DirectorOperations
	ptyStreamer *PTYStreamer

	// State
	role   string // "director" or "manager"
	hostID string

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mutex  sync.RWMutex
}
```

RemoteManager coordinates all remote operations.

### Functions returning RemoteManager

#### NewRemoteManager

```go
func NewRemoteManager(config *RemoteConfig) (*RemoteManager, error)
```

NewRemoteManager creates a new remote manager.


### Methods

#### RemoteManager.BootstrapRemoteHost

```go
func () BootstrapRemoteHost(ctx context.Context, hostID string, config BootstrapConfig) error
```

BootstrapRemoteHost bootstraps a remote host via SSH.

#### RemoteManager.GetHostID

```go
func () GetHostID() string
```

GetHostID returns the host ID.

#### RemoteManager.GetRole

```go
func () GetRole() string
```

GetRole returns the role of this manager.

#### RemoteManager.IsReady

```go
func () IsReady() bool
```

IsReady returns true if the manager is ready for operations.

#### RemoteManager.KillRemoteAgent

```go
func () KillRemoteAgent(ctx context.Context, hostID, agentID string) error
```

KillRemoteAgent kills an agent on a remote host.

#### RemoteManager.PublishPTYOutput

```go
func () PublishPTYOutput(sessionID string, data []byte) error
```

PublishPTYOutput publishes PTY output to the director.

#### RemoteManager.SendRemotePTYInput

```go
func () SendRemotePTYInput(hostID, sessionID string, data []byte) error
```

SendRemotePTYInput sends input to a remote PTY session.

#### RemoteManager.SpawnRemoteAgent

```go
func () SpawnRemoteAgent(ctx context.Context, hostID string, req SpawnPayload) (*SpawnPayload, error)
```

SpawnRemoteAgent spawns an agent on a remote host.

#### RemoteManager.Start

```go
func () Start() error
```

Start starts the remote manager.

#### RemoteManager.StartPTYSession

```go
func () StartPTYSession(sessionID string) error
```

StartPTYSession starts PTY streaming for a local session.

#### RemoteManager.Stop

```go
func () Stop() error
```

Stop stops the remote manager.

#### RemoteManager.StopPTYSession

```go
func () StopPTYSession(sessionID string)
```

StopPTYSession stops PTY streaming for a local session.

#### RemoteManager.SubscribeToRemotePTY

```go
func () SubscribeToRemotePTY(hostID, sessionID string, handler func(PTYData)) error
```

SubscribeToRemotePTY subscribes to PTY output from a remote session.

#### RemoteManager.heartbeatLoop

```go
func () heartbeatLoop()
```

heartbeatLoop sends periodic heartbeats to maintain host presence.

#### RemoteManager.sendHeartbeat

```go
func () sendHeartbeat()
```

sendHeartbeat sends a heartbeat to the KV store.


## type RingBuffer

```go
type RingBuffer struct {
	buffer  []interface{}
	head    int
	tail    int
	size    int
	maxSize int
	mutex   sync.RWMutex
}
```

RingBuffer implements a thread-safe ring buffer for PTY data replay.

### Functions returning RingBuffer

#### NewRingBuffer

```go
func NewRingBuffer(capacity int) *RingBuffer
```

NewRingBuffer creates a new ring buffer with the specified capacity.


### Methods

#### RingBuffer.Add

```go
func () Add(item interface{})
```

Add adds an item to the ring buffer.

#### RingBuffer.Clear

```go
func () Clear()
```

Clear empties the ring buffer.

#### RingBuffer.ForEach

```go
func () ForEach(fn func(interface{}))
```

ForEach iterates over all items in the buffer from oldest to newest.

#### RingBuffer.Size

```go
func () Size() int
```

Size returns the current number of items in the buffer.


## type SpawnPayload

```go
type SpawnPayload struct {
	AgentID   string            `json:"agent_id"` // base-10 string
	AgentSlug string            `json:"agent_slug,omitempty"`
	RepoPath  string            `json:"repo_path,omitempty"`
	Command   []string          `json:"command,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	SessionID string            `json:"session_id,omitempty"` // base-10 string
}
```

SpawnPayload represents spawn request/response payload.

