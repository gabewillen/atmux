# package process

`import "github.com/agentflare-ai/amux/internal/process"`

- `func StartMCPServer(ctx context.Context, cfg config.ProcessConfig, tracker *Tracker) error` — StartMCPServer starts the Notification MCP server.
- `type EventType` — EventType for process events.
- `type Event` — Event represents a process event.
- `type MCPServer`
- `type MCPSession`
- `type Process` — Process represents a tracked process.
- `type Tracker` — Tracker manages the process tree.

### Functions

#### StartMCPServer

```go
func StartMCPServer(ctx context.Context, cfg config.ProcessConfig, tracker *Tracker) error
```

StartMCPServer starts the Notification MCP server.


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


## type MCPServer

```go
type MCPServer struct {
	Tracker *Tracker
	mu      sync.Mutex
	Clients map[*MCPSession]struct{}
}
```

### Methods

#### MCPServer.Run

```go
func () Run(ctx context.Context, ln net.Listener)
```

#### MCPServer.addClient

```go
func () addClient(c *MCPSession)
```

#### MCPServer.removeClient

```go
func () removeClient(c *MCPSession)
```


## type MCPSession

```go
type MCPSession struct {
	conn net.Conn
	srv  *MCPServer
}
```

### Methods

#### MCPSession.Serve

```go
func () Serve(ctx context.Context)
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
	mu        sync.RWMutex
	processes map[int]*Process
	Events    chan Event
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
func () Start(ctx context.Context)
```

Start polling/monitoring logic would go here or be driven by hooks.

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


