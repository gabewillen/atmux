# package daemon

`import "github.com/agentflare-ai/amux/internal/daemon"`

- `type ClientConn` — ClientConn represents a connected client.
- `type Server` — Server is the JSON-RPC daemon server.
- `type jsonRPCError`
- `type jsonRPCRequest` — Minimal JSON-RPC types
- `type jsonRPCResponse`

## type ClientConn

```go
type ClientConn struct {
	conn net.Conn
	srv  *Server
}
```

ClientConn represents a connected client.

### Methods

#### ClientConn.handle

```go
func () handle(ctx context.Context)
```

#### ClientConn.processRequest

```go
func () processRequest(req jsonRPCRequest) jsonRPCResponse
```


## type Server

```go
type Server struct {
	Config   config.DaemonConfig
	Listener net.Listener
	mu       sync.Mutex
	clients  map[*ClientConn]struct{}
}
```

Server is the JSON-RPC daemon server.

### Functions returning Server

#### NewServer

```go
func NewServer(cfg config.DaemonConfig) *Server
```

NewServer creates a new daemon server.


### Methods

#### Server.Start

```go
func () Start(ctx context.Context) error
```

Start starts the server.

#### Server.Stop

```go
func () Stop() error
```

#### Server.addClient

```go
func () addClient(c *ClientConn)
```

#### Server.removeClient

```go
func () removeClient(c *ClientConn)
```

#### Server.serve

```go
func () serve(ctx context.Context)
```


## type jsonRPCError

```go
type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
```

## type jsonRPCRequest

```go
type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
}
```

Minimal JSON-RPC types

## type jsonRPCResponse

```go
type jsonRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *jsonRPCError `json:"error,omitempty"`
	ID      interface{}   `json:"id,omitempty"`
}
```

