# package daemon

`import "github.com/agentflare-ai/amux/internal/daemon"`

Package daemon implements the amux daemon (amuxd) and manager (amux-manager).

The daemon serves a JSON-RPC 2.0 control plane over a Unix socket,
manages agent lifecycles, and coordinates with remote hosts.

The role (director vs manager) is determined by the node.role configuration:
  - director: Runs the amux director with hub-mode NATS
  - manager: Runs as a host manager with leaf-mode NATS

See spec §12 for the full daemon specification.

Package daemon - rpcserver.go provides a JSON-RPC 2.0 server over Unix socket.

The daemon serves a control plane via JSON-RPC 2.0 over a Unix socket
at ~/.amux/amuxd.sock, per spec §12. All CLI commands are routed
through this server.

Agent methods delegate to the director or manager. Plugin methods
return stubs (full plugin system is Phase 4+).

- `Version` — Version is the daemon version string.
- `func Run(ctx context.Context, args []string) error` — Run starts the daemon with the given arguments.
- `func getHostID() string` — getHostID returns the host identifier from env or generates one.
- `func parseArgs(cfg *config.Config, args []string)` — parseArgs overrides config values from command-line arguments.
- `func runDirector(ctx context.Context, cfg *config.Config) error`
- `func runManager(ctx context.Context, cfg *config.Config) error`
- `func showHelp() error`
- `func showVersion() error`
- `rpcCodeParseError, rpcCodeInvalidRequest, rpcCodeMethodNotFound, rpcCodeInvalidParams, rpcCodeInternalError, rpcCodePermissionDenied` — JSON-RPC error codes per spec §12.
- `type RPCServer` — RPCServer serves JSON-RPC 2.0 over a Unix socket.
- `type rpcErrorObj` — rpcErrorObj is a JSON-RPC 2.0 error object.
- `type rpcRequest` — rpcRequest is a JSON-RPC 2.0 request.
- `type rpcResponse` — rpcResponse is a JSON-RPC 2.0 response.

### Constants

#### rpcCodeParseError, rpcCodeInvalidRequest, rpcCodeMethodNotFound, rpcCodeInvalidParams, rpcCodeInternalError, rpcCodePermissionDenied

```go
const (
	// Standard JSON-RPC errors.
	rpcCodeParseError     = -32700
	rpcCodeInvalidRequest = -32600
	rpcCodeMethodNotFound = -32601
	rpcCodeInvalidParams  = -32602
	rpcCodeInternalError  = -32603

	// Application-defined error codes per spec.
	rpcCodePermissionDenied = -32001
)
```

JSON-RPC error codes per spec §12.

#### Version

```go
const Version = "0.1.0-dev"
```

Version is the daemon version string.


### Functions

#### Run

```go
func Run(ctx context.Context, args []string) error
```

Run starts the daemon with the given arguments.

#### getHostID

```go
func getHostID() string
```

getHostID returns the host identifier from env or generates one.

#### parseArgs

```go
func parseArgs(cfg *config.Config, args []string)
```

parseArgs overrides config values from command-line arguments.

#### runDirector

```go
func runDirector(ctx context.Context, cfg *config.Config) error
```

#### runManager

```go
func runManager(ctx context.Context, cfg *config.Config) error
```

#### showHelp

```go
func showHelp() error
```

#### showVersion

```go
func showVersion() error
```


## type RPCServer

```go
type RPCServer struct {
	mu         sync.Mutex
	socketPath string
	listener   net.Listener
	dir        *director.Director
	mgr        *manager.Manager
	agentMgr   *agent.Manager
	done       chan struct{}
}
```

RPCServer serves JSON-RPC 2.0 over a Unix socket.

### Functions returning RPCServer

#### NewRPCServer

```go
func NewRPCServer(dir *director.Director, mgr *manager.Manager, agentMgr *agent.Manager) *RPCServer
```

NewRPCServer creates a new JSON-RPC server. Pass the director and/or
manager depending on the daemon role. Either may be nil.
agentMgr is the local agent manager for handling agent RPC methods.


### Methods

#### RPCServer.Start

```go
func () Start() error
```

Start begins listening on the Unix socket.

#### RPCServer.Stop

```go
func () Stop()
```

Stop gracefully shuts down the server.

#### RPCServer.acceptLoop

```go
func () acceptLoop()
```

#### RPCServer.dispatch

```go
func () dispatch(method string, params json.RawMessage) (any, *rpcErrorObj)
```

#### RPCServer.handleAgentAdd

```go
func () handleAgentAdd(params json.RawMessage) (any, *rpcErrorObj)
```

#### RPCServer.handleAgentList

```go
func () handleAgentList(_ json.RawMessage) (any, *rpcErrorObj)
```

#### RPCServer.handleAgentRemove

```go
func () handleAgentRemove(params json.RawMessage) (any, *rpcErrorObj)
```

#### RPCServer.handleAgentStart

```go
func () handleAgentStart(params json.RawMessage) (any, *rpcErrorObj)
```

#### RPCServer.handleAgentStop

```go
func () handleAgentStop(params json.RawMessage) (any, *rpcErrorObj)
```

#### RPCServer.handleConn

```go
func () handleConn(conn net.Conn)
```

#### RPCServer.handleEventsSubscribe

```go
func () handleEventsSubscribe(params json.RawMessage) (any, *rpcErrorObj)
```

handleEventsSubscribe registers a subscription for event types.
Per spec §12, this returns a subscription ID that can be used to receive events.
Note: Full streaming event delivery requires WebSocket/SSE which is Phase 4+.
For now, this returns a subscription ID but events are delivered via NATS.

#### RPCServer.handlePing

```go
func () handlePing(_ json.RawMessage) (any, *rpcErrorObj)
```

#### RPCServer.handlePluginInstall

```go
func () handlePluginInstall(_ json.RawMessage) (any, *rpcErrorObj)
```

handlePluginInstall rejects with permission denied per spec plugin permission model.
Full plugin system is Phase 4+.

#### RPCServer.handlePluginList

```go
func () handlePluginList(_ json.RawMessage) (any, *rpcErrorObj)
```

handlePluginList returns an empty list since no plugins are installed.

#### RPCServer.handlePluginRemove

```go
func () handlePluginRemove(_ json.RawMessage) (any, *rpcErrorObj)
```

handlePluginRemove returns success (no-op when no plugins exist).

#### RPCServer.handleSystemUpdate

```go
func () handleSystemUpdate(_ json.RawMessage) (any, *rpcErrorObj)
```

handleSystemUpdate is a stub for the system.update RPC method.
Full system update functionality is Phase 4+.

#### RPCServer.handleVersionRPC

```go
func () handleVersionRPC(_ json.RawMessage) (any, *rpcErrorObj)
```


## type rpcErrorObj

```go
type rpcErrorObj struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}
```

rpcErrorObj is a JSON-RPC 2.0 error object.

## type rpcRequest

```go
type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}
```

rpcRequest is a JSON-RPC 2.0 request.

## type rpcResponse

```go
type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcErrorObj    `json:"error,omitempty"`
}
```

rpcResponse is a JSON-RPC 2.0 response.

