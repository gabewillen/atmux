# package protocol

`import "github.com/stateforward/amux/internal/protocol"`

Package protocol implements remote communication protocol (transports events)

Package protocol implements remote communication protocol (transports events)

Package protocol implements remote communication protocol (transports events)

Package protocol implements remote communication protocol (transports events)

Package protocol implements remote communication protocol (transports events)

Package protocol implements remote communication protocol (transports events)

Package protocol implements remote communication protocol (transports events)

Package protocol implements remote communication protocol (transports events)

Package protocol implements remote communication protocol (transports events)

- `ErrProtocol` — ErrProtocol is returned when protocol operations fail
- `func expandHomeDir(path string) string` — expandHomeDir expands the ~ symbol to the user's home directory
- `func generateJWTPlaceholder(hostID string) string` — generateJWTPlaceholder creates a placeholder JWT for demonstration purposes In a real implementation, this would generate a proper signed JWT with permissions
- `func mustMarshal(v interface{}) json.RawMessage` — mustMarshal marshals JSON and panics on error (for use in cases where error is unexpected)
- `func subjectMatchesPattern(subject, pattern string) bool` — subjectMatchesPattern checks if a subject matches a pattern that may contain wildcards
- `type BufferedPublication` — BufferedPublication represents a buffered NATS publication
- `type ControlMessageType` — ControlMessageType represents the type of control message
- `type ControlMessage` — ControlMessage represents a control message exchanged between director and daemon
- `type ControlOperations` — ControlOperations handles request-reply control operations (spawn/kill/replay)
- `type ErrorPayload` — ErrorPayload represents the payload for error messages
- `type HandshakeHandler` — HandshakeHandler handles the handshake protocol between director and daemon
- `type HandshakePayload` — HandshakePayload represents the payload for handshake messages
- `type HostInfo` — HostInfo represents host metadata stored in KV
- `type KVStore` — KVStore provides access to NATS JetStream Key-Value store
- `type KillPayload` — KillPayload represents the payload for kill messages
- `type KillResponsePayload` — KillResponsePayload represents the response payload for kill messages
- `type Location` — Location represents the location configuration for an agent
- `type MessageEnvelope` — MessageEnvelope wraps messages sent over NATS
- `type NATSAuth` — NATSAuth provides NATS authentication and authorization functionality
- `type NATSServer` — NATSServer manages NATS server configuration for both hub (director) and leaf (manager) modes
- `type PeerInfo` — PeerInfo holds information about a connected peer
- `type PingPayload` — PingPayload represents the payload for ping messages
- `type PongPayload` — PongPayload represents the payload for pong messages
- `type ReconnectionManager` — ReconnectionManager handles reconnection logic
- `type RemoteProtocol` — RemoteProtocol manages the complete remote protocol implementation
- `type ReplayBuffer` — ReplayBuffer manages the replay buffer for PTY output
- `type ReplayPayload` — ReplayPayload represents the payload for replay messages
- `type ReplayResponsePayload` — ReplayResponsePayload represents the response payload for replay messages
- `type SSHBootstrap` — SSHBootstrap performs the SSH bootstrap for remote hosts
- `type SessionMetadata` — SessionMetadata represents session metadata stored in KV
- `type SpawnPayload` — SpawnPayload represents the payload for spawn messages
- `type SpawnResponsePayload` — SpawnResponsePayload represents the response payload for spawn messages
- `type SubjectBuilder` — SubjectBuilder builds NATS subjects according to the specification

### Variables

#### ErrProtocol

```go
var ErrProtocol = errors.New("protocol operation failed")
```

ErrProtocol is returned when protocol operations fail


### Functions

#### expandHomeDir

```go
func expandHomeDir(path string) string
```

expandHomeDir expands the ~ symbol to the user's home directory

#### generateJWTPlaceholder

```go
func generateJWTPlaceholder(hostID string) string
```

generateJWTPlaceholder creates a placeholder JWT for demonstration purposes
In a real implementation, this would generate a proper signed JWT with permissions

#### mustMarshal

```go
func mustMarshal(v interface{}) json.RawMessage
```

mustMarshal marshals JSON and panics on error (for use in cases where error is unexpected)

#### subjectMatchesPattern

```go
func subjectMatchesPattern(subject, pattern string) bool
```

subjectMatchesPattern checks if a subject matches a pattern that may contain wildcards


## type BufferedPublication

```go
type BufferedPublication struct {
	Subject string
	Data    []byte
	Size    int64
}
```

BufferedPublication represents a buffered NATS publication

## type ControlMessage

```go
type ControlMessage struct {
	Type    ControlMessageType `json:"type"`
	Payload json.RawMessage    `json:"payload"`
}
```

ControlMessage represents a control message exchanged between director and daemon

## type ControlMessageType

```go
type ControlMessageType string
```

ControlMessageType represents the type of control message

### Constants

#### HandshakeType, PingType, PongType, SpawnType, KillType, ReplayType, ErrorType

```go
const (
	HandshakeType ControlMessageType = "handshake"
	PingType      ControlMessageType = "ping"
	PongType      ControlMessageType = "pong"
	SpawnType     ControlMessageType = "spawn"
	KillType      ControlMessageType = "kill"
	ReplayType    ControlMessageType = "replay"
	ErrorType     ControlMessageType = "error"
)
```


## type ControlOperations

```go
type ControlOperations struct {
	nc             *nats.Conn
	subjectBuilder *SubjectBuilder
	requestTimeout time.Duration
}
```

ControlOperations handles request-reply control operations (spawn/kill/replay)

### Functions returning ControlOperations

#### NewControlOperations

```go
func NewControlOperations(nc *nats.Conn, subjectBuilder *SubjectBuilder, requestTimeout time.Duration) *ControlOperations
```

NewControlOperations creates a new ControlOperations instance


### Methods

#### ControlOperations.FailFastCheck

```go
func () FailFastCheck(hostID string, agentLifecycleState string) error
```

FailFastCheck checks if the host is considered disconnected before attempting operations

#### ControlOperations.IsNotReadyError

```go
func () IsNotReadyError(err error) bool
```

IsNotReadyError checks if the daemon replied with an error whose code is "not_ready"

#### ControlOperations.Kill

```go
func () Kill(ctx context.Context, hostID string, payload KillPayload) (*KillResponsePayload, error)
```

Kill sends a kill request to the daemon

#### ControlOperations.Replay

```go
func () Replay(ctx context.Context, hostID string, payload ReplayPayload) (*ReplayResponsePayload, error)
```

Replay sends a replay request to the daemon

#### ControlOperations.Spawn

```go
func () Spawn(ctx context.Context, hostID string, payload SpawnPayload) (*SpawnResponsePayload, error)
```

Spawn sends a spawn request to the daemon


## type ErrorPayload

```go
type ErrorPayload struct {
	RequestType string `json:"request_type"` // The type of request that caused the error
	Code        string `json:"code"`         // Short machine-readable error code
	Message     string `json:"message"`      // Human-readable error message
}
```

ErrorPayload represents the payload for error messages

## type HandshakeHandler

```go
type HandshakeHandler struct {
	nc             *nats.Conn
	subjectBuilder *SubjectBuilder
	connectedHosts map[string]*PeerInfo
}
```

HandshakeHandler handles the handshake protocol between director and daemon

### Functions returning HandshakeHandler

#### NewHandshakeHandler

```go
func NewHandshakeHandler(nc *nats.Conn, subjectBuilder *SubjectBuilder) *HandshakeHandler
```

NewHandshakeHandler creates a new HandshakeHandler


### Methods

#### HandshakeHandler.GetConnectedHosts

```go
func () GetConnectedHosts() []string
```

GetConnectedHosts returns all connected hosts

#### HandshakeHandler.IsConnected

```go
func () IsConnected(hostID string) bool
```

IsConnected checks if a host is connected

#### HandshakeHandler.PerformHandshake

```go
func () PerformHandshake(hostID, peerID string) error
```

PerformHandshake performs the handshake from the daemon side

#### HandshakeHandler.StartListening

```go
func () StartListening(hostID string) error
```

StartListening starts listening for handshake requests

#### HandshakeHandler.handleHandshakeRequest

```go
func () handleHandshakeRequest(msg *nats.Msg, expectedHostID string)
```

handleHandshakeRequest handles an incoming handshake request from a daemon

#### HandshakeHandler.sendHandshakeError

```go
func () sendHandshakeError(replySubject, requestType, code, message string)
```

sendHandshakeError sends an error response for handshake


## type HandshakePayload

```go
type HandshakePayload struct {
	Protocol int    `json:"protocol"`
	PeerID   string `json:"peer_id"`
	Role     string `json:"role"`    // "director" or "daemon"
	HostID   string `json:"host_id"` // The host ID
	Version  string `json:"version"` // Version of the daemon
}
```

HandshakePayload represents the payload for handshake messages

## type HostInfo

```go
type HostInfo struct {
	Version     string    `json:"version"`
	OS          string    `json:"os"`
	Arch        string    `json:"arch"`
	PeerID      string    `json:"peer_id"`
	StartupTime time.Time `json:"startup_time"`
}
```

HostInfo represents host metadata stored in KV

## type KVStore

```go
type KVStore struct {
	kv jetstream.KeyValue
}
```

KVStore provides access to NATS JetStream Key-Value store

### Functions returning KVStore

#### NewKVStore

```go
func NewKVStore(nc *nats.Conn, bucketName string) (*KVStore, error)
```

NewKVStore creates a new KV store instance


### Methods

#### KVStore.DeleteSessionMetadata

```go
func () DeleteSessionMetadata(ctx context.Context, hostID, sessionID string) error
```

DeleteSessionMetadata removes session metadata from the KV store

#### KVStore.GetHeartbeat

```go
func () GetHeartbeat(ctx context.Context, hostID string) (*time.Time, error)
```

GetHeartbeat retrieves the last heartbeat timestamp for a host

#### KVStore.GetHostInfo

```go
func () GetHostInfo(ctx context.Context, hostID string) (*HostInfo, error)
```

GetHostInfo retrieves host information from the KV store

#### KVStore.GetSessionMetadata

```go
func () GetSessionMetadata(ctx context.Context, hostID, sessionID string) (*SessionMetadata, error)
```

GetSessionMetadata retrieves session metadata from the KV store

#### KVStore.PutHeartbeat

```go
func () PutHeartbeat(ctx context.Context, hostID string) error
```

PutHeartbeat stores a heartbeat timestamp for a host

#### KVStore.PutHostInfo

```go
func () PutHostInfo(ctx context.Context, hostID string, info HostInfo) error
```

PutHostInfo stores host information in the KV store

#### KVStore.PutSessionMetadata

```go
func () PutSessionMetadata(ctx context.Context, hostID, sessionID string, metadata SessionMetadata) error
```

PutSessionMetadata stores session metadata in the KV store


## type KillPayload

```go
type KillPayload struct {
	SessionID string `json:"session_id"`
}
```

KillPayload represents the payload for kill messages

## type KillResponsePayload

```go
type KillResponsePayload struct {
	SessionID string `json:"session_id"`
	Killed    bool   `json:"killed"`
}
```

KillResponsePayload represents the response payload for kill messages

## type Location

```go
type Location struct {
	Type     string // "local" or "ssh"
	Host     string // For SSH locations
	RepoPath string // Path to git repo on the host
	HostID   string // Unique identifier for the host
}
```

Location represents the location configuration for an agent

## type MessageEnvelope

```go
type MessageEnvelope struct {
	ID        string      `json:"id"`
	Timestamp time.Time   `json:"timestamp"`
	Source    string      `json:"source"`
	Payload   interface{} `json:"payload"`
}
```

MessageEnvelope wraps messages sent over NATS

## type NATSAuth

```go
type NATSAuth struct {
	nc *nats.Conn
}
```

NATSAuth provides NATS authentication and authorization functionality

### Functions returning NATSAuth

#### NewNATSAuth

```go
func NewNATSAuth(nc *nats.Conn) *NATSAuth
```

NewNATSAuth creates a new NATSAuth instance


### Methods

#### NATSAuth.EnforceSubjectAuthorization

```go
func () EnforceSubjectAuthorization(hostID string, subject string, operation string) error
```

EnforceSubjectAuthorization enforces per-host subject authorization

#### NATSAuth.GenerateHostCredentials

```go
func () GenerateHostCredentials(hostID string) (string, error)
```

GenerateHostCredentials generates unique NATS credentials for a specific host

#### NATSAuth.ValidateHostPermissions

```go
func () ValidateHostPermissions(hostID string, subject string) bool
```

ValidateHostPermissions validates that a host has appropriate permissions for its subjects


## type NATSServer

```go
type NATSServer struct {
	cfg    *config.NATSConfig
	server *server.Server
	nc     *nats.Conn
}
```

NATSServer manages NATS server configuration for both hub (director) and leaf (manager) modes

### Functions returning NATSServer

#### NewNATSServer

```go
func NewNATSServer(cfg *config.NATSConfig) *NATSServer
```

NewNATSServer creates a new NATS server instance


### Methods

#### NATSServer.GetClient

```go
func () GetClient() *nats.Conn
```

GetClient returns the NATS connection

#### NATSServer.StartHubServer

```go
func () StartHubServer(ctx context.Context) error
```

StartHubServer starts a NATS server in hub mode (for director role)

#### NATSServer.StartLeafServer

```go
func () StartLeafServer(ctx context.Context, hubURL, credsPath string) error
```

StartLeafServer starts a NATS server in leaf mode (for manager role)

#### NATSServer.Stop

```go
func () Stop()
```

Stop stops the NATS server

#### NATSServer.WaitForShutdown

```go
func () WaitForShutdown()
```

WaitForShutdown waits for a signal to shut down the server


## type PeerInfo

```go
type PeerInfo struct {
	HostID   string
	PeerID   string
	Role     string
	Version  string
	LastSeen time.Time
}
```

PeerInfo holds information about a connected peer

## type PingPayload

```go
type PingPayload struct {
	TimestampUnixMs int64 `json:"ts_unix_ms"`
}
```

PingPayload represents the payload for ping messages

## type PongPayload

```go
type PongPayload struct {
	TimestampUnixMs int64 `json:"ts_unix_ms"`
}
```

PongPayload represents the payload for pong messages

## type ReconnectionManager

```go
type ReconnectionManager struct {
	replayBuffers  map[string]*ReplayBuffer // Map of session_id -> replay buffer
	natsConn       *nats.Conn
	subjectBuilder *SubjectBuilder
	controlOps     *ControlOperations
	mutex          sync.RWMutex
}
```

ReconnectionManager handles reconnection logic

### Functions returning ReconnectionManager

#### NewReconnectionManager

```go
func NewReconnectionManager(nc *nats.Conn, sb *SubjectBuilder, co *ControlOperations) *ReconnectionManager
```

NewReconnectionManager creates a new ReconnectionManager


### Methods

#### ReconnectionManager.AddSession

```go
func () AddSession(sessionID string, bufferSize int64)
```

AddSession adds a session to be managed for replay

#### ReconnectionManager.BufferPublicationDuringDisconnection

```go
func () BufferPublicationDuringDisconnection(sessionID, subject string, data []byte) error
```

BufferPublicationDuringDisconnection buffers a publication if disconnected

#### ReconnectionManager.GetReplayBuffer

```go
func () GetReplayBuffer(sessionID string) *ReplayBuffer
```

GetReplayBuffer returns the replay buffer for a session

#### ReconnectionManager.HandleReconnection

```go
func () HandleReconnection(ctx context.Context, hostID string, activeSessions []string) error
```

HandleReconnection handles the reconnection process after hub connectivity is restored

#### ReconnectionManager.OnDisconnection

```go
func () OnDisconnection()
```

OnDisconnection marks all sessions as disconnected

#### ReconnectionManager.OnReconnection

```go
func () OnReconnection()
```

OnReconnection marks all sessions as reconnected

#### ReconnectionManager.RemoveSession

```go
func () RemoveSession(sessionID string)
```

RemoveSession removes a session from replay management


## type RemoteProtocol

```go
type RemoteProtocol struct {
	cfg              *config.Config
	sshBootstrap     *SSHBootstrap
	natsServer       *NATSServer
	kvStore          *KVStore
	natsAuth         *NATSAuth
	subjectBuilder   *SubjectBuilder
	controlOps       *ControlOperations
	handshakeHandler *HandshakeHandler
	reconnectionMgr  *ReconnectionManager
}
```

RemoteProtocol manages the complete remote protocol implementation

### Functions returning RemoteProtocol

#### NewRemoteProtocol

```go
func NewRemoteProtocol(cfg *config.Config) *RemoteProtocol
```

NewRemoteProtocol creates a new RemoteProtocol instance


### Methods

#### RemoteProtocol.AddSessionToReconnectionManager

```go
func () AddSessionToReconnectionManager(sessionID string)
```

AddSessionToReconnectionManager adds a session to the reconnection manager

#### RemoteProtocol.BootstrapRemoteHost

```go
func () BootstrapRemoteHost(ctx context.Context, location Location) error
```

BootstrapRemoteHost performs the complete SSH bootstrap for a remote host

#### RemoteProtocol.GetNATSConnection

```go
func () GetNATSConnection() *nats.Conn
```

GetNATSConnection returns the NATS connection

#### RemoteProtocol.Initialize

```go
func () Initialize(ctx context.Context) error
```

Initialize initializes the remote protocol components

#### RemoteProtocol.InitializeLeafNode

```go
func () InitializeLeafNode(ctx context.Context, hubURL, credsPath string) error
```

InitializeLeafNode initializes the protocol for a leaf node (manager role)

#### RemoteProtocol.Kill

```go
func () Kill(ctx context.Context, hostID string, payload KillPayload) (*KillResponsePayload, error)
```

Kill terminates an agent session on a remote host

#### RemoteProtocol.PerformHandshake

```go
func () PerformHandshake(hostID, peerID string) error
```

PerformHandshake performs the handshake with a remote host

#### RemoteProtocol.Replay

```go
func () Replay(ctx context.Context, hostID string, payload ReplayPayload) (*ReplayResponsePayload, error)
```

Replay requests replay of PTY output for a session

#### RemoteProtocol.Spawn

```go
func () Spawn(ctx context.Context, hostID string, payload SpawnPayload) (*SpawnResponsePayload, error)
```

Spawn starts a new agent session on a remote host

#### RemoteProtocol.Stop

```go
func () Stop()
```

Stop shuts down the remote protocol


## type ReplayBuffer

```go
type ReplayBuffer struct {
	buffer               *bytes.Buffer
	maxSize              int64
	mutex                sync.RWMutex
	disconnected         bool
	bufferedPublications map[string][]*BufferedPublication
	maxBufferSize        int64
}
```

ReplayBuffer manages the replay buffer for PTY output

### Functions returning ReplayBuffer

#### NewReplayBuffer

```go
func NewReplayBuffer(maxSize int64) *ReplayBuffer
```

NewReplayBuffer creates a new ReplayBuffer


### Methods

#### ReplayBuffer.AddOutput

```go
func () AddOutput(data []byte)
```

AddOutput adds PTY output to the replay buffer

#### ReplayBuffer.BufferPublication

```go
func () BufferPublication(subject string, data []byte) error
```

BufferPublication buffers a NATS publication during disconnection

#### ReplayBuffer.Clear

```go
func () Clear()
```

Clear clears the replay buffer

#### ReplayBuffer.FlushBuffers

```go
func () FlushBuffers(nc *nats.Conn) error
```

FlushBuffers flushes all buffered publications

#### ReplayBuffer.GetReplayData

```go
func () GetReplayData() []byte
```

GetReplayData returns the current replay buffer content

#### ReplayBuffer.IsDisconnected

```go
func () IsDisconnected() bool
```

IsDisconnected returns the disconnection state

#### ReplayBuffer.SetDisconnected

```go
func () SetDisconnected(disconnected bool)
```

SetDisconnected sets the disconnection state

#### ReplayBuffer.dropOldestPublications

```go
func () dropOldestPublications(requiredSpace int64)
```

dropOldestPublications drops the oldest publications to make room for new ones

#### ReplayBuffer.getTotalBufferSize

```go
func () getTotalBufferSize() int64
```

getTotalBufferSize calculates the total size of buffered publications


## type ReplayPayload

```go
type ReplayPayload struct {
	SessionID string `json:"session_id"`
}
```

ReplayPayload represents the payload for replay messages

## type ReplayResponsePayload

```go
type ReplayResponsePayload struct {
	SessionID string `json:"session_id"`
	Accepted  bool   `json:"accepted"`
}
```

ReplayResponsePayload represents the response payload for replay messages

## type SSHBootstrap

```go
type SSHBootstrap struct {
	cfg *config.Config
}
```

SSHBootstrap performs the SSH bootstrap for remote hosts

### Functions returning SSHBootstrap

#### NewSSHBootstrap

```go
func NewSSHBootstrap(cfg *config.Config) *SSHBootstrap
```

NewSSHBootstrap creates a new SSHBootstrap instance


### Methods

#### SSHBootstrap.Bootstrap

```go
func () Bootstrap(ctx context.Context, location Location) error
```

Bootstrap performs the complete SSH bootstrap process for a remote host

#### SSHBootstrap.addFileToZip

```go
func () addFileToZip(zipWriter *zip.Writer, filePath, zipPath string) error
```

addFileToZip adds a file to the zip archive

#### SSHBootstrap.copyFileToRemote

```go
func () copyFileToRemote(ctx context.Context, localPath, host, remotePath string) error
```

copyFileToRemote copies a file to the remote host via SCP

#### SSHBootstrap.createBootstrapZip

```go
func () createBootstrapZip(ctx context.Context) (string, error)
```

createBootstrapZip creates a bootstrap ZIP file containing the daemon binary and adapters

#### SSHBootstrap.generateNATSCredentials

```go
func () generateNATSCredentials(hostID string) (string, error)
```

generateNATSCredentials generates NATS credentials for a specific host

#### SSHBootstrap.getAdapterPath

```go
func () getAdapterPath() string
```

getAdapterPath returns the path where adapters are stored

#### SSHBootstrap.isDaemonRunning

```go
func () isDaemonRunning(ctx context.Context, host string) (bool, error)
```

isDaemonRunning checks if the daemon is running on the remote host

#### SSHBootstrap.provisionNATSCredentials

```go
func () provisionNATSCredentials(ctx context.Context, host, hostID string) error
```

provisionNATSCredentials generates and provisions NATS credentials for the host

#### SSHBootstrap.startDaemon

```go
func () startDaemon(ctx context.Context, host string) error
```

startDaemon starts the daemon on the remote host

#### SSHBootstrap.unpackAndInstall

```go
func () unpackAndInstall(ctx context.Context, host, remoteZipPath string) error
```

unpackAndInstall unpacks the bootstrap ZIP and installs the components

#### SSHBootstrap.validateSSHConnection

```go
func () validateSSHConnection(ctx context.Context, host string) error
```

validateSSHConnection validates that we can connect to the SSH host

#### SSHBootstrap.verifyConnection

```go
func () verifyConnection(ctx context.Context, host string) error
```

verifyConnection verifies that the node has connected to the hub


## type SessionMetadata

```go
type SessionMetadata struct {
	AgentID   string `json:"agent_id"`
	AgentSlug string `json:"agent_slug"`
	RepoPath  string `json:"repo_path"`
	State     string `json:"state"`
}
```

SessionMetadata represents session metadata stored in KV

## type SpawnPayload

```go
type SpawnPayload struct {
	AgentID   string            `json:"agent_id"`
	AgentSlug string            `json:"agent_slug"`
	RepoPath  string            `json:"repo_path"`
	Command   []string          `json:"command"`
	Env       map[string]string `json:"env"`
}
```

SpawnPayload represents the payload for spawn messages

## type SpawnResponsePayload

```go
type SpawnResponsePayload struct {
	AgentID   string `json:"agent_id"`
	SessionID string `json:"session_id"`
}
```

SpawnResponsePayload represents the response payload for spawn messages

## type SubjectBuilder

```go
type SubjectBuilder struct {
	prefix string
}
```

SubjectBuilder builds NATS subjects according to the specification

### Functions returning SubjectBuilder

#### NewSubjectBuilder

```go
func NewSubjectBuilder(prefix string) *SubjectBuilder
```

NewSubjectBuilder creates a new SubjectBuilder


### Methods

#### SubjectBuilder.AgentCommSubject

```go
func () AgentCommSubject(hostID, agentID string) string
```

AgentCommSubject returns the subject for agent communication

#### SubjectBuilder.BroadcastCommSubject

```go
func () BroadcastCommSubject() string
```

BroadcastCommSubject returns the subject for broadcast communication

#### SubjectBuilder.ControlSubject

```go
func () ControlSubject(hostID string) string
```

ControlSubject returns the subject for control messages

#### SubjectBuilder.DirectorCommSubject

```go
func () DirectorCommSubject() string
```

DirectorCommSubject returns the subject for director communication

#### SubjectBuilder.EventsSubject

```go
func () EventsSubject(hostID string) string
```

EventsSubject returns the subject for host events

#### SubjectBuilder.HandshakeSubject

```go
func () HandshakeSubject(hostID string) string
```

HandshakeSubject returns the subject for handshake messages

#### SubjectBuilder.ManagerCommSubject

```go
func () ManagerCommSubject(hostID string) string
```

ManagerCommSubject returns the subject for manager communication

#### SubjectBuilder.PTYInputSubject

```go
func () PTYInputSubject(hostID, sessionID string) string
```

PTYInputSubject returns the subject for PTY input

#### SubjectBuilder.PTYOutputSubject

```go
func () PTYOutputSubject(hostID, sessionID string) string
```

PTYOutputSubject returns the subject for PTY output

#### SubjectBuilder.Publish

```go
func () Publish(nc *nats.Conn, subject string, msg interface{}) error
```

Publish publishes a message to a NATS subject

#### SubjectBuilder.PublishControlMessage

```go
func () PublishControlMessage(nc *nats.Conn, subject string, msgType ControlMessageType, payload interface{}) error
```

PublishControlMessage publishes a control message to a NATS subject

#### SubjectBuilder.Request

```go
func () Request(nc *nats.Conn, subject string, msg interface{}, timeout time.Duration) (*nats.Msg, error)
```

Request sends a request-reply message to NATS

#### SubjectBuilder.RequestControlMessage

```go
func () RequestControlMessage(nc *nats.Conn, subject string, msgType ControlMessageType, payload interface{}, timeout time.Duration) (*nats.Msg, error)
```

RequestControlMessage sends a control message using request-reply


