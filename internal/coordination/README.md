# package coordination

`import "github.com/agentflare-ai/amux/internal/coordination"`

- `type Action` — Action represents a coordination action to be executed.
- `type Config` — Config holds coordination settings.
- `type Coordinator` — Coordinator manages the loop.
- `type Executor` — Executor handles action execution.
- `type Mode` — Mode defines the coordination mode.
- `type ObservationLoop` — ObservationLoop runs the coordination cycle.
- `type Scheduler` — Scheduler manages the periodic execution of the coordination loop.
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

## type Executor

```go
type Executor struct {
	Agent *agent.Agent
}
```

Executor handles action execution.

### Functions returning Executor

#### NewExecutor

```go
func NewExecutor(agent *agent.Agent) *Executor
```

NewExecutor creates a new executor.


### Methods

#### Executor.Execute

```go
func () Execute(ctx context.Context, action Action) error
```

Execute performs the action.

#### Executor.injectInput

```go
func () injectInput(text string) error
```


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


## type Scheduler

```go
type Scheduler struct {
	Config Config
	Agent  *agent.Agent
	Loop   *ObservationLoop
}
```

Scheduler manages the periodic execution of the coordination loop.

### Functions returning Scheduler

#### NewScheduler

```go
func NewScheduler(cfg Config, agent *agent.Agent) *Scheduler
```

NewScheduler creates a new scheduler.


### Methods

#### Scheduler.SetMode

```go
func () SetMode(mode Mode)
```

SetMode updates the coordination mode.

#### Scheduler.Start

```go
func () Start(ctx context.Context)
```

Start starts the scheduler.

#### Scheduler.Stop

```go
func () Stop()
```

Stop stops the scheduler.


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

