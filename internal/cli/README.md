# package cli

`import "github.com/agentflare-ai/amux/internal/cli"`

Package cli implements the amux CLI client command handling.

The CLI communicates with the amux daemon over JSON-RPC 2.0 via a Unix
socket. This package provides the command parsing, dispatching, and
output formatting for all CLI commands.

See spec §12 for the full CLI specification.

Package cli - rpc.go provides JSON-RPC 2.0 client for daemon communication.

The CLI communicates with the amux daemon (amuxd) over a Unix socket
using JSON-RPC 2.0 per spec §12.

- `Version` — Version is the CLI version string.
- `func GetFlag(flags map[string]string, names []string, defaultValue string) string` — GetFlag gets a flag value with a default.
- `func HasFlag(flags map[string]string, names ...string) bool` — HasFlag checks if any of the flag names are set.
- `func ParseFlags(args []string) (flags map[string]string, positional []string)` — ParseFlags is a simple flag parser for CLI commands.
- `func Run(ctx context.Context, args []string) error` — Run executes the CLI with the given arguments.
- `func Stderr() *os.File` — Stderr returns the standard error writer.
- `func Stdout() *os.File` — Stdout returns the standard output writer.
- `func agentAdd(ctx context.Context, args []string) error`
- `func agentList(ctx context.Context, args []string) error`
- `func agentRemove(ctx context.Context, args []string) error`
- `func agentStart(ctx context.Context, args []string) error`
- `func agentStop(ctx context.Context, args []string) error`
- `func pluginInstall(ctx context.Context, args []string) error`
- `func pluginList(ctx context.Context, args []string) error`
- `func pluginRemove(ctx context.Context, args []string) error`
- `func runAgent(ctx context.Context, args []string) error`
- `func runPlugin(ctx context.Context, args []string) error`
- `func showAgentHelp() error`
- `func showHelp() error`
- `func showPluginHelp() error`
- `func showVersion() error`
- `type RPCClient` — RPCClient communicates with the amux daemon over Unix socket.
- `type rpcError` — rpcError is a JSON-RPC 2.0 error object.
- `type rpcRequest` — rpcRequest is a JSON-RPC 2.0 request.
- `type rpcResponse` — rpcResponse is a JSON-RPC 2.0 response.

### Constants

#### Version

```go
const Version = "0.1.0-dev"
```

Version is the CLI version string.


### Functions

#### GetFlag

```go
func GetFlag(flags map[string]string, names []string, defaultValue string) string
```

GetFlag gets a flag value with a default.

#### HasFlag

```go
func HasFlag(flags map[string]string, names ...string) bool
```

HasFlag checks if any of the flag names are set.

#### ParseFlags

```go
func ParseFlags(args []string) (flags map[string]string, positional []string)
```

ParseFlags is a simple flag parser for CLI commands.

#### Run

```go
func Run(ctx context.Context, args []string) error
```

Run executes the CLI with the given arguments.

#### Stderr

```go
func Stderr() *os.File
```

Stderr returns the standard error writer.

#### Stdout

```go
func Stdout() *os.File
```

Stdout returns the standard output writer.

#### agentAdd

```go
func agentAdd(ctx context.Context, args []string) error
```

#### agentList

```go
func agentList(ctx context.Context, args []string) error
```

#### agentRemove

```go
func agentRemove(ctx context.Context, args []string) error
```

#### agentStart

```go
func agentStart(ctx context.Context, args []string) error
```

#### agentStop

```go
func agentStop(ctx context.Context, args []string) error
```

#### pluginInstall

```go
func pluginInstall(ctx context.Context, args []string) error
```

#### pluginList

```go
func pluginList(ctx context.Context, args []string) error
```

#### pluginRemove

```go
func pluginRemove(ctx context.Context, args []string) error
```

#### runAgent

```go
func runAgent(ctx context.Context, args []string) error
```

#### runPlugin

```go
func runPlugin(ctx context.Context, args []string) error
```

#### showAgentHelp

```go
func showAgentHelp() error
```

#### showHelp

```go
func showHelp() error
```

#### showPluginHelp

```go
func showPluginHelp() error
```

#### showVersion

```go
func showVersion() error
```


## type RPCClient

```go
type RPCClient struct {
	socketPath string
	nextID     atomic.Int64
}
```

RPCClient communicates with the amux daemon over Unix socket.

### Functions returning RPCClient

#### NewRPCClient

```go
func NewRPCClient() *RPCClient
```

NewRPCClient creates a new JSON-RPC client connected to the daemon socket.


### Methods

#### RPCClient.Call

```go
func () Call(method string, params any, result any) error
```

Call sends a JSON-RPC request and returns the result.


## type rpcError

```go
type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}
```

rpcError is a JSON-RPC 2.0 error object.

### Methods

#### rpcError.Error

```go
func () Error() string
```


## type rpcRequest

```go
type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}
```

rpcRequest is a JSON-RPC 2.0 request.

## type rpcResponse

```go
type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}
```

rpcResponse is a JSON-RPC 2.0 response.

