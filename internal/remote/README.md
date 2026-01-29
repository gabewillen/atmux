# package remote

`import "github.com/agentflare-ai/amux/internal/remote"`

- `func BootstrapRemote(cfg config.AgentConfig, hostID api.HostID) error` — BootstrapRemote installs/configures the daemon on a remote host via SSH.
- `func ConfigureEmbeddedHub(cfg *config.Config) error` — ConfigureEmbeddedHub configures the embedded NATS server for the director.
- `func ConfigureLeaf(cfg *config.Config, credsPath string) error` — ConfigureLeaf configures the NATS leaf connection for the manager.
- `func GenerateHostCredentials(hostID api.HostID, prefix string) (string, string, error)` — GenerateHostCredentials creates a NATS User JWT and Seed for a host.
- `func NewSpawnRequest(agentID api.AgentID, slug api.AgentSlug, repoPath string, cmd []string, env map[string]string) protocol.ControlRequest` — Helper to construct spawn request
- `func SendControlRequest(ctx context.Context, nc *nats.Conn, hostID api.HostID, req protocol.ControlRequest) (*protocol.ControlResponse, error)` — SendControlRequest sends a control request to a remote host and waits for a response.
- `func executeSSHBootstrap(cfg SSHConfig, credsContent string, hostID api.HostID) error`
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

#### ConfigureLeaf

```go
func ConfigureLeaf(cfg *config.Config, credsPath string) error
```

ConfigureLeaf configures the NATS leaf connection for the manager.
This is used by the daemon to connect to the hub.

#### GenerateHostCredentials

```go
func GenerateHostCredentials(hostID api.HostID, prefix string) (string, string, error)
```

GenerateHostCredentials creates a NATS User JWT and Seed for a host.
It enforces the subject permissions specified in the plan.

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

