# package rpc

`import "github.com/agentflare-ai/amux/internal/rpc"`

Package rpc implements JSON-RPC 2.0 transport over Unix sockets.

- `CodeParseError, CodeInvalidRequest, CodeMethodNotFound, CodeInvalidParams, CodeInternalError`
- `func writeJSON(writer *bufio.Writer, value any) error`
- `type Client` — Client issues JSON-RPC requests over a Unix socket.
- `type Error` — Error describes a JSON-RPC error.
- `type Handler` — Handler processes a JSON-RPC request.
- `type Request` — Request is a JSON-RPC request or notification.
- `type Response` — Response is a JSON-RPC response.
- `type Server` — Server hosts JSON-RPC handlers over a stream transport.

### Constants

#### CodeParseError, CodeInvalidRequest, CodeMethodNotFound, CodeInvalidParams, CodeInternalError

```go
const (
	// CodeParseError indicates invalid JSON.
	CodeParseError = -32700
	// CodeInvalidRequest indicates invalid JSON-RPC.
	CodeInvalidRequest = -32600
	// CodeMethodNotFound indicates missing method.
	CodeMethodNotFound = -32601
	// CodeInvalidParams indicates invalid parameters.
	CodeInvalidParams = -32602
	// CodeInternalError indicates a server error.
	CodeInternalError = -32603
)
```


### Functions

#### writeJSON

```go
func writeJSON(writer *bufio.Writer, value any) error
```


## type Client

```go
type Client struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	mu     sync.Mutex
	nextID uint64
}
```

Client issues JSON-RPC requests over a Unix socket.

### Functions returning Client

#### Dial

```go
func Dial(ctx context.Context, socketPath string) (*Client, error)
```

Dial connects to a JSON-RPC socket.


### Methods

#### Client.Call

```go
func () Call(ctx context.Context, method string, params any, result any) error
```

Call sends a request and decodes the response.

#### Client.Close

```go
func () Close() error
```

Close closes the underlying connection.


## type Error

```go
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}
```

Error describes a JSON-RPC error.

## type Handler

```go
type Handler func(context.Context, json.RawMessage) (any, *Error)
```

Handler processes a JSON-RPC request.

## type Request

```go
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}
```

Request is a JSON-RPC request or notification.

## type Response

```go
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}
```

Response is a JSON-RPC response.

## type Server

```go
type Server struct {
	mu       sync.RWMutex
	handlers map[string]Handler
	logger   *log.Logger
}
```

Server hosts JSON-RPC handlers over a stream transport.

### Functions returning Server

#### NewServer

```go
func NewServer(logger *log.Logger) *Server
```

NewServer constructs a JSON-RPC server.


### Methods

#### Server.Register

```go
func () Register(method string, handler Handler)
```

Register registers a handler for a method.

#### Server.Serve

```go
func () Serve(ctx context.Context, listener net.Listener) error
```

Serve accepts connections and handles requests.

#### Server.handler

```go
func () handler(method string) Handler
```

#### Server.serveConn

```go
func () serveConn(ctx context.Context, conn net.Conn)
```

#### Server.writeError

```go
func () writeError(writer *bufio.Writer, id json.RawMessage, code int, message string)
```


