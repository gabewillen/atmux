# package coordination

`import "github.com/agentflare-ai/amux/internal/coordination"`

- `type Action` — Action represents a coordination action to be executed.
- `type Config` — Config holds coordination settings.
- `type Coordinator` — Coordinator manages the loop.
- `type Mode` — Mode defines the coordination mode.
- `type ObservationLoop` — ObservationLoop runs the coordination cycle.
- `type Snapshot` — Snapshot represents the state of the system at a point in time.

## type Action

```go
type Action struct {
	Type    string            `json:"type"`
	Target  api.AgentID       `json:"target"`
	Payload map[string]string `json:"payload"`
}
```

Action represents a coordination action to be executed.

## type Config

```go
type Config struct {
	Mode     Mode
	Interval time.Duration
}
```

Config holds coordination settings.

## type Coordinator

```go
type Coordinator interface {
	Start() error
	Stop() error
}
```

Coordinator manages the loop.

## type Mode

```go
type Mode string
```

Mode defines the coordination mode.

### Constants

#### ModeAuto, ModeManual

```go
const (
	ModeAuto   Mode = "auto"
	ModeManual Mode = "manual"
)
```


## type ObservationLoop

```go
type ObservationLoop struct {
	Interval time.Duration
	Agent    *agent.Agent // Focus on single agent for now
	cancel   func()
}
```

ObservationLoop runs the coordination cycle.

### Functions returning ObservationLoop

#### NewObservationLoop

```go
func NewObservationLoop(agent *agent.Agent, interval time.Duration) *ObservationLoop
```

NewObservationLoop creates a new loop.


### Methods

#### ObservationLoop.Start

```go
func () Start(ctx context.Context)
```

Start begins the loop.

#### ObservationLoop.Stop

```go
func () Stop()
```

Stop stops the loop.

#### ObservationLoop.captureSnapshot

```go
func () captureSnapshot() (*Snapshot, error)
```

#### ObservationLoop.tick

```go
func () tick(ctx context.Context) error
```


## type Snapshot

```go
type Snapshot struct {
	Timestamp time.Time         `json:"timestamp"`
	AgentID   api.AgentID       `json:"agent_id"`
	TUI       string            `json:"tui_xml"` // XML representation
	Processes []process.Process `json:"processes"`
}
```

Snapshot represents the state of the system at a point in time.

