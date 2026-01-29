# package monitor

`import "github.com/agentflare-ai/amux/internal/monitor"`

- `type Hook` — Hook allows injecting logic on data read (e.g.
- `type Monitor` — Monitor observes PTY output.

## type Hook

```go
type Hook func(data []byte)
```

Hook allows injecting logic on data read (e.g. pattern matching).

## type Monitor

```go
type Monitor struct {
	AgentID api.AgentID
	Bus     *agent.EventBus
	Input   io.Reader

	// Configuration
	ActivityTimeout time.Duration
	CheckInterval   time.Duration

	// State
	lastActivity time.Time
}
```

Monitor observes PTY output.

### Functions returning Monitor

#### NewMonitor

```go
func NewMonitor(agentID api.AgentID, bus *agent.EventBus, input io.Reader) *Monitor
```

NewMonitor creates a new monitor.


### Methods

#### Monitor.Start

```go
func () Start(ctx context.Context, hooks ...Hook)
```

Start runs the monitoring loop.


