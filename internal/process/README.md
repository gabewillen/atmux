# package process

`import "github.com/agentflare-ai/amux/internal/process"`

- `type EventType` — EventType for process events.
- `type Event` — Event represents a process event.
- `type Gater` — Gater uses an LLM to decide if an event should trigger a notification.
- `type HookMessage` — HookMessage matches the JSON sent by hook.c
- `type MCPNotification` — MCPNotification represents a notification sent to clients.
- `type MCPServer` — MCPServer handles notification subscriptions via a Unix socket.
- `type Process` — Process represents a tracked process.
- `type Tracker` — Tracker manages the process tree.

## type Event

```go
type Event struct {
	Type      EventType   `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}
```

Event represents a process event.

## type EventType

```go
type EventType string
```

EventType for process events.

### Constants

#### EventSpawned, EventExited, EventIO

```go
const (
	EventSpawned EventType = "process.spawned"
	EventExited  EventType = "process.exited"
	EventIO      EventType = "process.io"
)
```


## type Gater

```go
type Gater struct {
	Engine inference.LiquidgenEngine
}
```

Gater uses an LLM to decide if an event should trigger a notification.

### Methods

#### Gater.ShouldNotify

```go
func () ShouldNotify(ctx context.Context, event Event) bool
```

ShouldNotify returns true if the LLM thinks the event is noteworthy.


## type HookMessage

```go
type HookMessage struct {
	PID  int    `json:"pid"`
	PPID int    `json:"ppid"`
	Cmd  string `json:"cmd"`
}
```

HookMessage matches the JSON sent by hook.c

## type MCPNotification

```go
type MCPNotification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}
```

MCPNotification represents a notification sent to clients.

## type MCPServer

```go
type MCPServer struct {
	SocketPath string
	mu         sync.Mutex
	clients    map[net.Conn]struct{}
	listener   net.Listener
}
```

MCPServer handles notification subscriptions via a Unix socket.

### Functions returning MCPServer

#### NewMCPServer

```go
func NewMCPServer(socketPath string) *MCPServer
```

NewMCPServer creates a new MCP server.


### Methods

#### MCPServer.Broadcast

```go
func () Broadcast(method string, params interface{})
```

Broadcast sends a notification to all connected clients.

#### MCPServer.Start

```go
func () Start(ctx context.Context) error
```

Start starts the server.

#### MCPServer.acceptLoop

```go
func () acceptLoop(ctx context.Context)
```

#### MCPServer.addClient

```go
func () addClient(c net.Conn)
```

#### MCPServer.handleClient

```go
func () handleClient(ctx context.Context, conn net.Conn)
```

#### MCPServer.removeClient

```go
func () removeClient(c net.Conn)
```


## type Process

```go
type Process struct {
	PID       int           `json:"pid"`
	AgentID   api.AgentID   `json:"agent_id"`
	ProcessID api.ProcessID `json:"process_id"`
	Command   string        `json:"command"`
	Args      []string      `json:"args"`
	WorkDir   string        `json:"work_dir"`
	ParentPID int           `json:"parent_pid"`
	StartedAt time.Time     `json:"started_at"`
	EndedAt   time.Time     `json:"ended_at,omitempty"`
	ExitCode  int           `json:"exit_code,omitempty"`
	Running   bool          `json:"running"`
}
```

Process represents a tracked process.

## type Tracker

```go
type Tracker struct {
	mu         sync.RWMutex
	processes  map[int]*Process
	Events     chan Event
	Gater      *Gater
	SocketPath string
}
```

Tracker manages the process tree.

### Functions returning Tracker

#### NewTracker

```go
func NewTracker() *Tracker
```

NewTracker creates a new process tracker.


### Methods

#### Tracker.GetProcess

```go
func () GetProcess(pid int) (*Process, bool)
```

GetProcess returns a copy of the process info.

#### Tracker.Start

```go
func () Start(ctx context.Context, socketDir string) error
```

Start initiates the hook server.

#### Tracker.StartHookServer

```go
func () StartHookServer(ctx context.Context, socketDir string) error
```

StartHookServer starts listening on a Unix socket for exec notifications.

#### Tracker.TrackExit

```go
func () TrackExit(pid int, exitCode int) error
```

TrackExit records a process exit.

#### Tracker.TrackSpawn

```go
func () TrackSpawn(proc *Process)
```

TrackSpawn records a new process start.

#### Tracker.acceptLoop

```go
func () acceptLoop(ctx context.Context, l net.Listener)
```

#### Tracker.handleConnection

```go
func () handleConnection(conn net.Conn)
```


