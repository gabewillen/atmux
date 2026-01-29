# package remote

`import "github.com/agentflare-ai/amux/internal/remote"`

Package remote implements NATS-based remote orchestration for amux.

- `ErrInvalidSubject, ErrInvalidMessage, ErrHostDisconnected, ErrNotReady, ErrSessionConflict, ErrSessionNotFound, ErrReplayDisabled, ErrBootstrapFailed`
- `func AgentCommSubject(prefix string, hostID api.HostID, agentID api.AgentID) string` — AgentCommSubject returns the subject for an agent channel.
- `func BroadcastCommSubject(prefix string) string` — BroadcastCommSubject returns the broadcast channel subject.
- `func ControlSubject(prefix string, hostID api.HostID) string` — ControlSubject returns the subject for control requests.
- `func DecodePayload(msg ControlMessage, dest any) error` — DecodePayload decodes a control payload into the provided struct.
- `func DirectorCommSubject(prefix string) string` — DirectorCommSubject returns the subject for the director channel.
- `func EncodeControlMessage(msg ControlMessage) ([]byte, error)` — EncodeControlMessage marshals a control message to JSON.
- `func EncodeEventMessageJSON(msg EventMessage) ([]byte, error)` — EncodeEventMessageJSON marshals an event envelope to JSON.
- `func EventsSubject(prefix string, hostID api.HostID) string` — EventsSubject returns the subject for host events.
- `func HandshakeSubject(prefix string, hostID api.HostID) string` — HandshakeSubject returns the subject for handshake requests.
- `func HostIDFromLocation(location api.Location) (api.HostID, error)` — HostIDFromLocation derives host_id from location.
- `func HostPermissions(prefix string, hostID api.HostID) protocol.Permissions` — HostPermissions returns the per-host subject permissions.
- `func ManagerCommSubject(prefix string, hostID api.HostID) string` — ManagerCommSubject returns the subject for a manager channel.
- `func NewPTYConn(ctx context.Context, dispatcher protocol.Dispatcher, prefix string, hostID api.HostID, sessionID api.SessionID) (net.Conn, error)` — NewPTYConn returns a net.Conn that bridges PTY I/O over NATS.
- `func NowRFC3339() string` — NowRFC3339 returns the current time in RFC3339 UTC format.
- `func ParseHandshakeSubject(prefix string, subject string) (api.HostID, error)` — ParseHandshakeSubject extracts the host_id from a handshake subject.
- `func ParseSessionSubject(prefix string, subject string) (api.HostID, api.SessionID, string, error)` — ParseSessionSubject extracts the session_id from a PTY subject.
- `func PtyInSubject(prefix string, hostID api.HostID, sessionID api.SessionID) string` — PtyInSubject returns the subject for PTY input.
- `func PtyOutSubject(prefix string, hostID api.HostID, sessionID api.SessionID) string` — PtyOutSubject returns the subject for PTY output.
- `func SubjectPrefix(prefix string) string` — SubjectPrefix normalizes the configured subject prefix.
- `func bootstrapConfig(req BootstrapRequest) ([]byte, error)`
- `func chunkBytes(maxPayload int, data []byte) [][]byte`
- `func ensureRepo(repoRoot string) error`
- `func hostnameFallback() string`
- `func hubURL(cfg config.Config) string`
- `func reconnectDelay(cfg config.Config, attempt int) time.Duration`
- `func shellEscape(raw string) string`
- `func sshOptions(location api.Location) []string`
- `func sshTarget(location api.Location) string`
- `type BootstrapRequest` — BootstrapRequest describes a remote bootstrap request.
- `type Bootstrapper` — Bootstrapper provisions remote credentials and configuration.
- `type ConnectionEstablishedPayload` — ConnectionEstablishedPayload is the payload for connection.established.
- `type ConnectionLostPayload` — ConnectionLostPayload is the payload for connection.lost.
- `type ConnectionRecoveredPayload` — ConnectionRecoveredPayload is the payload for connection.recovered.
- `type ControlMessage` — ControlMessage is the envelope for remote control requests.
- `type CredentialStore` — CredentialStore persists host credentials on disk.
- `type Credential` — Credential holds the per-host auth token.
- `type DirectorOptions` — DirectorOptions configures the director runtime.
- `type Director` — Director orchestrates remote hosts via NATS.
- `type ErrorPayload` — ErrorPayload describes a control error response.
- `type EventMessage` — EventMessage wraps a remote event in a wire envelope.
- `type ExecSSHRunner` — ExecSSHRunner executes SSH commands using the system ssh binary.
- `type HandshakePayload` — HandshakePayload is the handshake request/response payload.
- `type HostManager` — HostManager runs sessions and responds to remote control requests.
- `type KVStore` — KVStore provides a simple bucketed key-value store.
- `type KillRequest` — KillRequest describes a kill request payload.
- `type KillResponse` — KillResponse describes a kill response payload.
- `type MessageType` — MessageType describes the remote event envelope type.
- `type Outbox` — Outbox buffers outbound publications while disconnected.
- `type PingPayload` — PingPayload describes ping/pong payloads.
- `type ReplayBuffer` — ReplayBuffer stores a bounded history of PTY output.
- `type ReplayRequest` — ReplayRequest describes a replay request payload.
- `type ReplayResponse` — ReplayResponse describes a replay response payload.
- `type SSHRunner` — SSHRunner executes SSH commands.
- `type SpawnRequest` — SpawnRequest describes a spawn request payload.
- `type SpawnResponse` — SpawnResponse describes a spawn response payload.
- `type WireEvent` — WireEvent describes an event payload.
- `type hostState`
- `type queuedMessage`
- `type remoteSession`

### Variables

#### ErrInvalidSubject, ErrInvalidMessage, ErrHostDisconnected, ErrNotReady, ErrSessionConflict, ErrSessionNotFound, ErrReplayDisabled, ErrBootstrapFailed

```go
var (
	// ErrInvalidSubject is returned for malformed NATS subjects.
	ErrInvalidSubject = errors.New("invalid subject")
	// ErrInvalidMessage is returned for malformed protocol messages.
	ErrInvalidMessage = errors.New("invalid message")
	// ErrHostDisconnected is returned when a host is offline.
	ErrHostDisconnected = errors.New("host disconnected")
	// ErrNotReady is returned when the remote daemon has not completed handshake.
	ErrNotReady = errors.New("remote not ready")
	// ErrSessionConflict is returned when spawn conflicts with existing session metadata.
	ErrSessionConflict = errors.New("session conflict")
	// ErrSessionNotFound is returned when a session is missing.
	ErrSessionNotFound = errors.New("session not found")
	// ErrReplayDisabled is returned when replay buffering is disabled.
	ErrReplayDisabled = errors.New("replay disabled")
	// ErrBootstrapFailed is returned when SSH bootstrap fails.
	ErrBootstrapFailed = errors.New("bootstrap failed")
)
```


### Functions

#### AgentCommSubject

```go
func AgentCommSubject(prefix string, hostID api.HostID, agentID api.AgentID) string
```

AgentCommSubject returns the subject for an agent channel.

#### BroadcastCommSubject

```go
func BroadcastCommSubject(prefix string) string
```

BroadcastCommSubject returns the broadcast channel subject.

#### ControlSubject

```go
func ControlSubject(prefix string, hostID api.HostID) string
```

ControlSubject returns the subject for control requests.

#### DecodePayload

```go
func DecodePayload(msg ControlMessage, dest any) error
```

DecodePayload decodes a control payload into the provided struct.

#### DirectorCommSubject

```go
func DirectorCommSubject(prefix string) string
```

DirectorCommSubject returns the subject for the director channel.

#### EncodeControlMessage

```go
func EncodeControlMessage(msg ControlMessage) ([]byte, error)
```

EncodeControlMessage marshals a control message to JSON.

#### EncodeEventMessageJSON

```go
func EncodeEventMessageJSON(msg EventMessage) ([]byte, error)
```

EncodeEventMessageJSON marshals an event envelope to JSON.

#### EventsSubject

```go
func EventsSubject(prefix string, hostID api.HostID) string
```

EventsSubject returns the subject for host events.

#### HandshakeSubject

```go
func HandshakeSubject(prefix string, hostID api.HostID) string
```

HandshakeSubject returns the subject for handshake requests.

#### HostIDFromLocation

```go
func HostIDFromLocation(location api.Location) (api.HostID, error)
```

HostIDFromLocation derives host_id from location.

#### HostPermissions

```go
func HostPermissions(prefix string, hostID api.HostID) protocol.Permissions
```

HostPermissions returns the per-host subject permissions.

#### ManagerCommSubject

```go
func ManagerCommSubject(prefix string, hostID api.HostID) string
```

ManagerCommSubject returns the subject for a manager channel.

#### NewPTYConn

```go
func NewPTYConn(ctx context.Context, dispatcher protocol.Dispatcher, prefix string, hostID api.HostID, sessionID api.SessionID) (net.Conn, error)
```

NewPTYConn returns a net.Conn that bridges PTY I/O over NATS.

#### NowRFC3339

```go
func NowRFC3339() string
```

NowRFC3339 returns the current time in RFC3339 UTC format.

#### ParseHandshakeSubject

```go
func ParseHandshakeSubject(prefix string, subject string) (api.HostID, error)
```

ParseHandshakeSubject extracts the host_id from a handshake subject.

#### ParseSessionSubject

```go
func ParseSessionSubject(prefix string, subject string) (api.HostID, api.SessionID, string, error)
```

ParseSessionSubject extracts the session_id from a PTY subject.

#### PtyInSubject

```go
func PtyInSubject(prefix string, hostID api.HostID, sessionID api.SessionID) string
```

PtyInSubject returns the subject for PTY input.

#### PtyOutSubject

```go
func PtyOutSubject(prefix string, hostID api.HostID, sessionID api.SessionID) string
```

PtyOutSubject returns the subject for PTY output.

#### SubjectPrefix

```go
func SubjectPrefix(prefix string) string
```

SubjectPrefix normalizes the configured subject prefix.

#### bootstrapConfig

```go
func bootstrapConfig(req BootstrapRequest) ([]byte, error)
```

#### chunkBytes

```go
func chunkBytes(maxPayload int, data []byte) [][]byte
```

#### ensureRepo

```go
func ensureRepo(repoRoot string) error
```

#### hostnameFallback

```go
func hostnameFallback() string
```

#### hubURL

```go
func hubURL(cfg config.Config) string
```

#### reconnectDelay

```go
func reconnectDelay(cfg config.Config, attempt int) time.Duration
```

#### shellEscape

```go
func shellEscape(raw string) string
```

#### sshOptions

```go
func sshOptions(location api.Location) []string
```

#### sshTarget

```go
func sshTarget(location api.Location) string
```


## type BootstrapRequest

```go
type BootstrapRequest struct {
	HostID        api.HostID
	Location      api.Location
	HubURL        string
	CredsPath     string
	SubjectPrefix string
	KVBucket      string
	ManagerModel  string
}
```

BootstrapRequest describes a remote bootstrap request.

## type Bootstrapper

```go
type Bootstrapper struct {
	Runner SSHRunner
}
```

Bootstrapper provisions remote credentials and configuration.

### Methods

#### Bootstrapper.Bootstrap

```go
func () Bootstrap(ctx context.Context, req BootstrapRequest, cred Credential) error
```

Bootstrap performs SSH bootstrap for a remote host.


## type ConnectionEstablishedPayload

```go
type ConnectionEstablishedPayload struct {
	PeerID    string `json:"peer_id"`
	Timestamp string `json:"timestamp"`
}
```

ConnectionEstablishedPayload is the payload for connection.established.

## type ConnectionLostPayload

```go
type ConnectionLostPayload struct {
	PeerID    string `json:"peer_id"`
	Timestamp string `json:"timestamp"`
	Reason    string `json:"reason"`
}
```

ConnectionLostPayload is the payload for connection.lost.

## type ConnectionRecoveredPayload

```go
type ConnectionRecoveredPayload struct {
	PeerID    string `json:"peer_id"`
	Timestamp string `json:"timestamp"`
}
```

ConnectionRecoveredPayload is the payload for connection.recovered.

## type ControlMessage

```go
type ControlMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
```

ControlMessage is the envelope for remote control requests.

### Functions returning ControlMessage

#### DecodeControlMessage

```go
func DecodeControlMessage(data []byte) (ControlMessage, error)
```

DecodeControlMessage decodes a control message from JSON.

#### EncodePayload

```go
func EncodePayload(msgType string, payload any) (ControlMessage, error)
```

EncodePayload marshals a payload into a control message.

#### NewErrorMessage

```go
func NewErrorMessage(requestType, code, message string) (ControlMessage, error)
```

NewErrorMessage constructs a control error response.


## type Credential

```go
type Credential struct {
	Token string `json:"token"`
}
```

Credential holds the per-host auth token.

### Functions returning Credential

#### LoadCredential

```go
func LoadCredential(path string) (Credential, error)
```

LoadCredential reads a credential from disk.

#### NewCredential

```go
func NewCredential() (Credential, error)
```

NewCredential generates a new credential.

#### ParseCredential

```go
func ParseCredential(data []byte) (Credential, error)
```

ParseCredential decodes a credential from JSON.


### Methods

#### Credential.Marshal

```go
func () Marshal() ([]byte, error)
```

Marshal serializes the credential to JSON.


## type CredentialStore

```go
type CredentialStore struct {
	baseDir string
}
```

CredentialStore persists host credentials on disk.

### Functions returning CredentialStore

#### NewCredentialStore

```go
func NewCredentialStore(baseDir string) (*CredentialStore, error)
```

NewCredentialStore constructs a credential store rooted at baseDir.


### Methods

#### CredentialStore.GetOrCreate

```go
func () GetOrCreate(hostID string) (Credential, error)
```

GetOrCreate returns a credential for the host, creating one if missing.


## type Director

```go
type Director struct {
	cfg            config.Config
	dispatcher     protocol.Dispatcher
	subjectPrefix  string
	requestTimeout time.Duration
	kv             *KVStore
	creds          *CredentialStore
	bootstrapper   *Bootstrapper
	hostID         api.HostID
	peerID         api.PeerID
	version        string
	mu             sync.Mutex
	hosts          map[api.HostID]*hostState
	peerIndex      map[string]api.HostID
}
```

Director orchestrates remote hosts via NATS.

### Functions returning Director

#### NewDirector

```go
func NewDirector(cfg config.Config, dispatcher protocol.Dispatcher, options DirectorOptions) (*Director, error)
```

NewDirector constructs a director orchestrator.


### Methods

#### Director.AttachPTY

```go
func () AttachPTY(ctx context.Context, hostID api.HostID, sessionID api.SessionID) (net.Conn, error)
```

AttachPTY opens a PTY connection via NATS subjects.

#### Director.EnsureHost

```go
func () EnsureHost(ctx context.Context, location api.Location) (api.HostID, Credential, error)
```

EnsureHost bootstraps the remote host and returns its host ID.

#### Director.HostReady

```go
func () HostReady(hostID api.HostID) bool
```

HostReady reports whether the host is ready for control requests.

#### Director.Kill

```go
func () Kill(ctx context.Context, hostID api.HostID, req KillRequest) (KillResponse, error)
```

Kill requests a remote kill for the host.

#### Director.Replay

```go
func () Replay(ctx context.Context, hostID api.HostID, req ReplayRequest) (ReplayResponse, error)
```

Replay requests a replay for the host.

#### Director.Spawn

```go
func () Spawn(ctx context.Context, hostID api.HostID, req SpawnRequest) (SpawnResponse, error)
```

Spawn requests a remote spawn for the host.

#### Director.Start

```go
func () Start(ctx context.Context) error
```

Start subscribes to handshake and host events subjects.

#### Director.ensureConnected

```go
func () ensureConnected(hostID api.HostID) error
```

#### Director.handleHandshake

```go
func () handleHandshake(msg protocol.Message)
```

#### Director.handleHostEvent

```go
func () handleHostEvent(msg protocol.Message)
```

#### Director.replyError

```go
func () replyError(reply, requestType, code, message string) error
```

#### Director.sendControl

```go
func () sendControl(ctx context.Context, hostID api.HostID, msg ControlMessage) (ControlMessage, error)
```

#### Director.setConnected

```go
func () setConnected(hostID api.HostID, connected bool)
```

#### Director.setReady

```go
func () setReady(hostID api.HostID, ready bool)
```

#### Director.writeHostKV

```go
func () writeHostKV(ctx context.Context, hostID api.HostID, peerID api.PeerID) error
```

#### Director.writeSessionKV

```go
func () writeSessionKV(ctx context.Context, hostID api.HostID, sessionID string, req SpawnRequest) error
```


## type DirectorOptions

```go
type DirectorOptions struct {
	Version      string
	HostID       api.HostID
	Bootstrapper *Bootstrapper
}
```

DirectorOptions configures the director runtime.

## type ErrorPayload

```go
type ErrorPayload struct {
	RequestType string `json:"request_type"`
	Code        string `json:"code"`
	Message     string `json:"message"`
}
```

ErrorPayload describes a control error response.

## type EventMessage

```go
type EventMessage struct {
	Type    MessageType `json:"type"`
	Target  string      `json:"target,omitempty"`
	Targets []string    `json:"targets,omitempty"`
	Event   WireEvent   `json:"event"`
}
```

EventMessage wraps a remote event in a wire envelope.

### Functions returning EventMessage

#### EncodeEventMessage

```go
func EncodeEventMessage(name string, payload any) (EventMessage, error)
```

EncodeEventMessage builds a broadcast event envelope.


## type ExecSSHRunner

```go
type ExecSSHRunner struct{}
```

ExecSSHRunner executes SSH commands using the system ssh binary.

### Methods

#### ExecSSHRunner.Run

```go
func () Run(ctx context.Context, target string, options []string, command string, stdin []byte) error
```

Run executes an SSH command.


## type HandshakePayload

```go
type HandshakePayload struct {
	Protocol int    `json:"protocol"`
	PeerID   string `json:"peer_id"`
	Role     string `json:"role"`
	HostID   string `json:"host_id"`
}
```

HandshakePayload is the handshake request/response payload.

## type HostManager

```go
type HostManager struct {
	cfg           config.Config
	resolver      *paths.Resolver
	dispatcher    protocol.Dispatcher
	subjectPrefix string
	hostID        api.HostID
	peerID        api.PeerID
	bufferSize    int
	outbox        *Outbox
	mu            sync.Mutex
	sessions      map[api.SessionID]*remoteSession
	agentIndex    map[api.AgentID]*remoteSession
	ready         bool
	connected     bool
	everConnected bool
}
```

HostManager runs sessions and responds to remote control requests.

### Functions returning HostManager

#### NewHostManager

```go
func NewHostManager(cfg config.Config, resolver *paths.Resolver) (*HostManager, error)
```

NewHostManager constructs a host manager.


### Methods

#### HostManager.Start

```go
func () Start(ctx context.Context) error
```

Start connects to NATS and begins serving control requests.

#### HostManager.connect

```go
func () connect(ctx context.Context) error
```

#### HostManager.expandPath

```go
func () expandPath(path string) string
```

#### HostManager.flushOutbox

```go
func () flushOutbox()
```

#### HostManager.handleControl

```go
func () handleControl(msg protocol.Message)
```

#### HostManager.handleKill

```go
func () handleKill(reply string, control ControlMessage)
```

#### HostManager.handleOutput

```go
func () handleOutput(session *remoteSession, chunk []byte)
```

#### HostManager.handlePTYInput

```go
func () handlePTYInput(msg protocol.Message)
```

#### HostManager.handlePing

```go
func () handlePing(reply string, control ControlMessage)
```

#### HostManager.handleReplay

```go
func () handleReplay(reply string, control ControlMessage)
```

#### HostManager.handleSpawn

```go
func () handleSpawn(reply string, control ControlMessage)
```

#### HostManager.isReady

```go
func () isReady() bool
```

#### HostManager.markDisconnected

```go
func () markDisconnected(reason string)
```

#### HostManager.observeSession

```go
func () observeSession(session *remoteSession)
```

#### HostManager.performHandshake

```go
func () performHandshake(ctx context.Context, recovered bool) error
```

#### HostManager.publishConnectionEvent

```go
func () publishConnectionEvent(ctx context.Context, name string, payload any)
```

#### HostManager.publishEvent

```go
func () publishEvent(ctx context.Context, payload []byte)
```

#### HostManager.publishHostEvent

```go
func () publishHostEvent(ctx context.Context, name string, payload any)
```

#### HostManager.publishPTY

```go
func () publishPTY(sessionID api.SessionID, data []byte)
```

#### HostManager.replaySession

```go
func () replaySession(session *remoteSession)
```

#### HostManager.replyError

```go
func () replyError(reply, requestType, code, message string) error
```

#### HostManager.replySpawn

```go
func () replySpawn(reply string, agentID api.AgentID, sessionID api.SessionID)
```

#### HostManager.subscribeControl

```go
func () subscribeControl(ctx context.Context) error
```


## type KVStore

```go
type KVStore struct {
	baseDir string
	bucket  string
}
```

KVStore provides a simple bucketed key-value store.

### Functions returning KVStore

#### NewKVStore

```go
func NewKVStore(baseDir string, bucket string) (*KVStore, error)
```

NewKVStore ensures the KV bucket directory exists.


### Methods

#### KVStore.Get

```go
func () Get(ctx context.Context, key string) ([]byte, error)
```

Get loads a key value if it exists.

#### KVStore.Put

```go
func () Put(ctx context.Context, key string, value []byte) error
```

Put writes a key-value entry as UTF-8 bytes.

#### KVStore.keyPath

```go
func () keyPath(key string) (string, error)
```


## type KillRequest

```go
type KillRequest struct {
	SessionID string `json:"session_id"`
}
```

KillRequest describes a kill request payload.

## type KillResponse

```go
type KillResponse struct {
	SessionID string `json:"session_id"`
	Killed    bool   `json:"killed"`
}
```

KillResponse describes a kill response payload.

## type MessageType

```go
type MessageType uint8
```

MessageType describes the remote event envelope type.

### Constants

#### MsgBroadcast, MsgMulticast, MsgUnicast

```go
const (
	// MsgBroadcast is a broadcast message type.
	MsgBroadcast MessageType = 1
	// MsgMulticast is a multicast message type.
	MsgMulticast MessageType = 2
	// MsgUnicast is a unicast message type.
	MsgUnicast MessageType = 3
)
```


## type Outbox

```go
type Outbox struct {
	mu       sync.Mutex
	maxBytes int
	total    int
	queue    []queuedMessage
}
```

Outbox buffers outbound publications while disconnected.

### Functions returning Outbox

#### NewOutbox

```go
func NewOutbox(maxBytes int) *Outbox
```

NewOutbox constructs an outbox with a max payload size.


### Methods

#### Outbox.Drain

```go
func () Drain() []queuedMessage
```

Drain returns buffered messages in enqueue order and clears the outbox.

#### Outbox.Enqueue

```go
func () Enqueue(subject string, payload []byte)
```

Enqueue buffers a subject payload pair, dropping oldest entries when full.


## type PingPayload

```go
type PingPayload struct {
	UnixMS int64 `json:"ts_unix_ms"`
}
```

PingPayload describes ping/pong payloads.

## type ReplayBuffer

```go
type ReplayBuffer struct {
	mu   sync.Mutex
	max  int
	data []byte
}
```

ReplayBuffer stores a bounded history of PTY output.

### Functions returning ReplayBuffer

#### NewReplayBuffer

```go
func NewReplayBuffer(maxBytes int) *ReplayBuffer
```

NewReplayBuffer constructs a replay buffer with a maximum size.


### Methods

#### ReplayBuffer.Add

```go
func () Add(chunk []byte)
```

Add appends bytes to the replay buffer with ring semantics.

#### ReplayBuffer.Enabled

```go
func () Enabled() bool
```

Enabled reports whether replay buffering is enabled.

#### ReplayBuffer.Snapshot

```go
func () Snapshot() []byte
```

Snapshot returns a copy of the buffered bytes.


## type ReplayRequest

```go
type ReplayRequest struct {
	SessionID string `json:"session_id"`
}
```

ReplayRequest describes a replay request payload.

## type ReplayResponse

```go
type ReplayResponse struct {
	SessionID string `json:"session_id"`
	Accepted  bool   `json:"accepted"`
}
```

ReplayResponse describes a replay response payload.

## type SSHRunner

```go
type SSHRunner interface {
	Run(ctx context.Context, target string, options []string, command string, stdin []byte) error
}
```

SSHRunner executes SSH commands.

## type SpawnRequest

```go
type SpawnRequest struct {
	AgentID   string            `json:"agent_id"`
	AgentSlug string            `json:"agent_slug"`
	RepoPath  string            `json:"repo_path"`
	Command   []string          `json:"command"`
	Env       map[string]string `json:"env,omitempty"`
}
```

SpawnRequest describes a spawn request payload.

## type SpawnResponse

```go
type SpawnResponse struct {
	AgentID   string `json:"agent_id"`
	SessionID string `json:"session_id"`
}
```

SpawnResponse describes a spawn response payload.

## type WireEvent

```go
type WireEvent struct {
	Name string          `json:"name"`
	Data json.RawMessage `json:"data"`
}
```

WireEvent describes an event payload.

## type hostState

```go
type hostState struct {
	hostID    api.HostID
	peerID    api.PeerID
	connected bool
	ready     bool
}
```

## type queuedMessage

```go
type queuedMessage struct {
	subject string
	payload []byte
}
```

## type remoteSession

```go
type remoteSession struct {
	agentID    api.AgentID
	sessionID  api.SessionID
	slug       string
	repoPath   string
	worktree   string
	runtime    *session.LocalSession
	buffer     *ReplayBuffer
	replayGate bool
	replaying  bool
	pending    [][]byte
	mu         sync.Mutex
}
```

