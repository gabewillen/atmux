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
- `type agentSummary`
- `type attachResult`
- `type locationParam`
- `type mergeParams`

### Constants

#### SpecVersion, AmuxVersion

```go
const (
	// SpecVersion is the spec version implemented by the daemon.
	SpecVersion = "v1.22"
	// AmuxVersion is the daemon version string.
	AmuxVersion = "dev"
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
	manager    *manager.LocalManager
	dispatcher protocol.Dispatcher
	server     *rpc.Server
	listener   net.Listener
	embedded   *protocol.EmbeddedServer
	logger     *log.Logger
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
func () Close(ctx context.Context) error
```

Close shuts down the daemon.

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
func () startAttachProxy(ctx context.Context, repoRoot string, agentID api.AgentID, ptyFile *os.File) (string, error)
```


## type agentAddParams

```go
type agentAddParams struct {
	Name     string        `json:"name"`
	About    string        `json:"about"`
	Adapter  string        `json:"adapter"`
	Location locationParam `json:"location"`
	Cwd      string        `json:"cwd"`
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
	Agents []agentSummary `json:"agents"`
}
```

## type agentRefParams

```go
type agentRefParams struct {
	AgentID string `json:"agent_id"`
	Name    string `json:"name"`
}
```

## type agentSummary

```go
type agentSummary struct {
	AgentID  api.AgentID `json:"agent_id"`
	Name     string      `json:"name"`
	Adapter  string      `json:"adapter"`
	Presence string      `json:"presence"`
	RepoRoot string      `json:"repo_root"`
}
```

## type attachResult

```go
type attachResult struct {
	SocketPath string `json:"socket_path"`
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

