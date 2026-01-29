# package remote

`import "github.com/agentflare-ai/amux/internal/remote"`

- `func BootstrapRemote(cfg config.AgentConfig, hostID api.HostID) error` — BootstrapRemote installs/configures the daemon on a remote host via SSH.
- `func ConfigureEmbeddedHub(cfg *config.Config) error` — ConfigureEmbeddedHub configures the embedded NATS server for the director.
- `func ConfigureLeaf(cfg *config.Config, credsPath string) error` — ConfigureLeaf configures the NATS leaf connection for the manager.
- `func GenerateAccountKey() (nkeys.KeyPair, error)` — GenerateAccountKey generates a new NATS Account Key Pair.
- `func GenerateHostCredentials(accountKP nkeys.KeyPair, hostID api.HostID, prefix string) (string, string, error)` — GenerateHostCredentials creates a NATS User JWT and Seed for a host, signed by the provided Account Key.
- `func LoadAmuxAccountKey() (nkeys.KeyPair, error)` — LoadAmuxAccountKey loads the Amux Account KeyPair for signing User creds.
- `func NewSpawnRequest(agentID api.AgentID, slug api.AgentSlug, repoPath string, cmd []string, env map[string]string) protocol.ControlRequest` — Helper to construct spawn request
- `func SendControlRequest(ctx context.Context, nc *nats.Conn, hostID api.HostID, req protocol.ControlRequest) (*protocol.ControlResponse, error)` — SendControlRequest sends a control request to a remote host and waits for a response.
- `func executeSSHBootstrap(cfg SSHConfig, credsContent string, hostID api.HostID) error`
- `func loadKeyPair(path string) (nkeys.KeyPair, error)`
- `func loadOrGenerateKey(path string, createFn func() (nkeys.KeyPair, error)) (nkeys.KeyPair, error)`
- `func setupNATSAuth(baseDir string) error`
- `type Manager` — Manager runs on the remote host and manages agents/sessions via NATS.
- `type RemoteSession`
- `type ReplayBuffer` — ReplayBuffer is a thread-safe ring buffer for PTY output.
- `type SSHConfig` — SSHConfig holds SSH connection parameters.

### Functions

#### BootstrapRemote

```go
func BootstrapRemote(cfg config.AgentConfig, hostID api.HostID) error
```

BootstrapRemote installs/configures the daemon on a remote host via SSH.

#### ConfigureEmbeddedHub

```go
func ConfigureEmbeddedHub(cfg *config.Config) error
```

ConfigureEmbeddedHub configures the embedded NATS server for the director.
It sets up the Operator/System/Account hierarchy if needed and writes the config file.

#### ConfigureLeaf

```go
func ConfigureLeaf(cfg *config.Config, credsPath string) error
```

ConfigureLeaf configures the NATS leaf connection for the manager.

#### GenerateAccountKey

```go
func GenerateAccountKey() (nkeys.KeyPair, error)
```

GenerateAccountKey generates a new NATS Account Key Pair.
The private key (seed) should be persisted by the Director.
The public key is needed by the NATS Server configuration.

#### GenerateHostCredentials

```go
func GenerateHostCredentials(accountKP nkeys.KeyPair, hostID api.HostID, prefix string) (string, string, error)
```

GenerateHostCredentials creates a NATS User JWT and Seed for a host, signed by the provided Account Key.
It enforces the subject permissions specified in the plan.

#### LoadAmuxAccountKey

```go
func LoadAmuxAccountKey() (nkeys.KeyPair, error)
```

LoadAmuxAccountKey loads the Amux Account KeyPair for signing User creds.

#### NewSpawnRequest

```go
func NewSpawnRequest(agentID api.AgentID, slug api.AgentSlug, repoPath string, cmd []string, env map[string]string) protocol.ControlRequest
```

Helper to construct spawn request

#### SendControlRequest

```go
func SendControlRequest(ctx context.Context, nc *nats.Conn, hostID api.HostID, req protocol.ControlRequest) (*protocol.ControlResponse, error)
```

SendControlRequest sends a control request to a remote host and waits for a response.
It enforces the timeout and error handling specified in the plan.

#### executeSSHBootstrap

```go
func executeSSHBootstrap(cfg SSHConfig, credsContent string, hostID api.HostID) error
```

#### loadKeyPair

```go
func loadKeyPair(path string) (nkeys.KeyPair, error)
```

#### loadOrGenerateKey

```go
func loadOrGenerateKey(path string, createFn func() (nkeys.KeyPair, error)) (nkeys.KeyPair, error)
```

#### setupNATSAuth

```go
func setupNATSAuth(baseDir string) error
```


## type Manager

```go
type Manager struct {
	HostID api.HostID
	Config *config.Config
	NC     *nats.Conn
	Bus    *agent.EventBus

	// State
	agents   map[api.AgentID]*agent.Agent
	sessions map[api.SessionID]*RemoteSession
}
```

Manager runs on the remote host and manages agents/sessions via NATS.

### Functions returning Manager

#### NewManager

```go
func NewManager(cfg *config.Config, hostID api.HostID) *Manager
```

NewManager creates a new Manager.


### Methods

#### Manager.Start

```go
func () Start(ctx context.Context, nc *nats.Conn) error
```

Start connects to NATS (if not provided) and starts the control loop.
Note: NC might be provided if we are reusing a connection.

#### Manager.handleControl

```go
func () handleControl(msg *nats.Msg)
```

#### Manager.handlePTYInput

```go
func () handlePTYInput(msg *nats.Msg)
```

#### Manager.handleSpawn

```go
func () handleSpawn(p protocol.SpawnPayload) (json.RawMessage, error)
```

#### Manager.performHandshake

```go
func () performHandshake(ctx context.Context) error
```

#### Manager.replyError

```go
func () replyError(msg *nats.Msg, code, message string)
```


## type RemoteSession

```go
type RemoteSession struct {
	*agent.Session
	Replay  *ReplayBuffer
	Monitor *monitor.Monitor
}
```

## type ReplayBuffer

```go
type ReplayBuffer struct {
	mu       sync.RWMutex
	buffer   []byte
	capacity int
	head     int // Points to the next write position
	full     bool
}
```

ReplayBuffer is a thread-safe ring buffer for PTY output.

### Functions returning ReplayBuffer

#### NewReplayBuffer

```go
func NewReplayBuffer(capacity int) *ReplayBuffer
```

NewReplayBuffer creates a new replay buffer with the given capacity.


### Methods

#### ReplayBuffer.Bytes

```go
func () Bytes() []byte
```

Bytes returns the current content of the buffer ordered from oldest to newest.

#### ReplayBuffer.Write

```go
func () Write(p []byte) (n int, err error)
```

Write appends data to the buffer, overwriting old data if full.


## type SSHConfig

```go
type SSHConfig struct {
	Host     string
	User     string
	Port     int
	KeyPath  string
	Password string
}
```

SSHConfig holds SSH connection parameters.

