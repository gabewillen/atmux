# package daemon

`import "github.com/agentflare-ai/amux/internal/daemon"`

Package daemon hosts the JSON-RPC control plane.

- `SpecVersion, AmuxVersion`
- `func decodeParams(raw json.RawMessage, dest any) *rpc.Error`
- `func rpcInternal(err error) *rpc.Error`
- `func rpcInvalidParams(err error) *rpc.Error`
- `type Daemon` — Daemon hosts the JSON-RPC control plane.
- `type agentAddParams`
- `type agentAddResult`
- `type agentListResult`
- `type agentRefParams`
- `type attachResult`
- `type daemonStatusResult`
- `type daemonStopParams`
- `type locationParam`
- `type mergeParams`

### Constants

#### SpecVersion, AmuxVersion

```go
const (
	// SpecVersion is the spec version implemented by the daemon.
	SpecVersion = "v1.22"
	// AmuxVersion is the daemon version string.
	AmuxVersion = "0.0.0-dev"
)
```


### Functions

#### decodeParams

```go
func decodeParams(raw json.RawMessage, dest any) *rpc.Error
```

#### rpcInternal

```go
func rpcInternal(err error) *rpc.Error
```

#### rpcInvalidParams

```go
func rpcInvalidParams(err error) *rpc.Error
```


## type Daemon

```go
type Daemon struct {
	resolver   *paths.Resolver
	cfg        config.Config
	manager    *manager.Manager
	hostMgr    *remote.HostManager
	dispatcher protocol.Dispatcher
	server     *rpc.Server
	listener   net.Listener
	embedded   *protocol.NATSServer
	logger     *log.Logger
	closeMu    sync.Mutex
	closed     bool
}
```

Daemon hosts the JSON-RPC control plane.

### Functions returning Daemon

#### New

```go
func New(ctx context.Context, resolver *paths.Resolver, cfg config.Config, logger *log.Logger) (*Daemon, error)
```

New constructs a daemon instance.


### Methods

#### Daemon.Close

```go
func () Close(ctx context.Context, force bool) error
```

Close shuts down the daemon, optionally forcing termination.

#### Daemon.Serve

```go
func () Serve(ctx context.Context) error
```

Serve starts listening on the daemon socket.

#### Daemon.handleAgentAdd

```go
func () handleAgentAdd(ctx context.Context, raw json.RawMessage) (any, *rpc.Error)
```

#### Daemon.handleAgentAttach

```go
func () handleAgentAttach(ctx context.Context, raw json.RawMessage) (any, *rpc.Error)
```

#### Daemon.handleAgentKill

```go
func () handleAgentKill(ctx context.Context, raw json.RawMessage) (any, *rpc.Error)
```

#### Daemon.handleAgentList

```go
func () handleAgentList(ctx context.Context, raw json.RawMessage) (any, *rpc.Error)
```

#### Daemon.handleAgentRemove

```go
func () handleAgentRemove(ctx context.Context, raw json.RawMessage) (any, *rpc.Error)
```

#### Daemon.handleAgentRestart

```go
func () handleAgentRestart(ctx context.Context, raw json.RawMessage) (any, *rpc.Error)
```

#### Daemon.handleAgentStart

```go
func () handleAgentStart(ctx context.Context, raw json.RawMessage) (any, *rpc.Error)
```

#### Daemon.handleAgentStop

```go
func () handleAgentStop(ctx context.Context, raw json.RawMessage) (any, *rpc.Error)
```

#### Daemon.handleGitMerge

```go
func () handleGitMerge(ctx context.Context, raw json.RawMessage) (any, *rpc.Error)
```

#### Daemon.handlePing

```go
func () handlePing(ctx context.Context, raw json.RawMessage) (any, *rpc.Error)
```

#### Daemon.handleStatus

```go
func () handleStatus(ctx context.Context, raw json.RawMessage) (any, *rpc.Error)
```

#### Daemon.handleStop

```go
func () handleStop(ctx context.Context, raw json.RawMessage) (any, *rpc.Error)
```

#### Daemon.handleVersion

```go
func () handleVersion(ctx context.Context, raw json.RawMessage) (any, *rpc.Error)
```

#### Daemon.registerHandlers

```go
func () registerHandlers()
```

#### Daemon.resolveAgentID

```go
func () resolveAgentID(params agentRefParams) (api.AgentID, error)
```

#### Daemon.startAttachProxy

```go
func () startAttachProxy(ctx context.Context, repoRoot string, agentID api.AgentID, stream io.ReadWriteCloser) (string, error)
```


## type agentAddParams

```go
type agentAddParams struct {
	Name           string        `json:"name"`
	About          string        `json:"about"`
	Adapter        string        `json:"adapter"`
	Location       locationParam `json:"location"`
	Cwd            string        `json:"cwd"`
	ListenChannels []string      `json:"listen_channels"`
}
```

## type agentAddResult

```go
type agentAddResult struct {
	AgentID api.AgentID `json:"agent_id"`
}
```

## type agentListResult

```go
type agentListResult struct {
	Roster []api.RosterEntry `json:"roster"`
}
```

## type agentRefParams

```go
type agentRefParams struct {
	AgentID string `json:"agent_id"`
	Name    string `json:"name"`
}
```

## type attachResult

```go
type attachResult struct {
	SocketPath string `json:"socket_path"`
}
```

## type daemonStatusResult

```go
type daemonStatusResult struct {
	Role         string `json:"role"`
	HubConnected bool   `json:"hub_connected"`
	Ready        bool   `json:"ready"`
	HostID       string `json:"host_id,omitempty"`
}
```

## type daemonStopParams

```go
type daemonStopParams struct {
	Force bool `json:"force"`
}
```

## type locationParam

```go
type locationParam struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	RepoPath string `json:"repo_path"`
}
```

## type mergeParams

```go
type mergeParams struct {
	AgentID      string `json:"agent_id"`
	Name         string `json:"name"`
	Strategy     string `json:"strategy"`
	TargetBranch string `json:"target_branch"`
}
```

