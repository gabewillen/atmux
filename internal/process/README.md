# package process

`import "github.com/agentflare-ai/amux/internal/process"`

Package process provides process tracking and interception for amux.

The process tracker monitors child processes spawned within agent PTYs,
capturing spawn/exit events and optionally intercepting I/O.

See spec §8 for process tracking requirements.

- `type CaptureMode` — CaptureMode represents the I/O capture mode.
- `type HookMode` — HookMode represents the process hook mode.
- `type Process` — Process represents a tracked process.
- `type Tracker` — Tracker tracks processes for an agent.

## type CaptureMode

```go
type CaptureMode string
```

CaptureMode represents the I/O capture mode.

### Constants

#### CaptureModeNone, CaptureModeStdout, CaptureModeStderr, CaptureModeStdin, CaptureModeAll

```go
const (
	// CaptureModeNone disables I/O capture.
	CaptureModeNone CaptureMode = "none"

	// CaptureModeStdout captures stdout only.
	CaptureModeStdout CaptureMode = "stdout"

	// CaptureModeStderr captures stderr only.
	CaptureModeStderr CaptureMode = "stderr"

	// CaptureModeStdin captures stdin only.
	CaptureModeStdin CaptureMode = "stdin"

	// CaptureModeAll captures all streams.
	CaptureModeAll CaptureMode = "all"
)
```


## type HookMode

```go
type HookMode string
```

HookMode represents the process hook mode.

### Constants

#### HookModeAuto, HookModePreload, HookModePolling, HookModeDisabled

```go
const (
	// HookModeAuto automatically selects preload or polling.
	HookModeAuto HookMode = "auto"

	// HookModePreload uses LD_PRELOAD/DYLD_INSERT_LIBRARIES.
	HookModePreload HookMode = "preload"

	// HookModePolling uses periodic polling.
	HookModePolling HookMode = "polling"

	// HookModeDisabled disables process tracking.
	HookModeDisabled HookMode = "disabled"
)
```


## type Process

```go
type Process struct {
	// PID is the process ID.
	PID int

	// PPID is the parent process ID.
	PPID int

	// Command is the command that started the process.
	Command string

	// Args are the command arguments.
	Args []string

	// AgentID is the agent that spawned the process.
	AgentID muid.MUID

	// StartedAt is when the process started.
	StartedAt time.Time

	// ExitCode is the exit code (nil if still running).
	ExitCode *int

	// ExitedAt is when the process exited (zero if still running).
	ExitedAt time.Time
}
```

Process represents a tracked process.

## type Tracker

```go
type Tracker struct {
	mu         sync.RWMutex
	agentID    muid.MUID
	processes  map[int]*Process
	dispatcher event.Dispatcher
}
```

Tracker tracks processes for an agent.

### Functions returning Tracker

#### NewTracker

```go
func NewTracker(agentID muid.MUID, dispatcher event.Dispatcher) *Tracker
```

NewTracker creates a new process tracker.


### Methods

#### Tracker.Add

```go
func () Add(ctx context.Context, p *Process)
```

Add adds a process to tracking.

#### Tracker.Clear

```go
func () Clear()
```

Clear removes all tracked processes.

#### Tracker.Count

```go
func () Count() int
```

Count returns the number of tracked processes.

#### Tracker.Get

```go
func () Get(pid int) *Process
```

Get returns a process by PID.

#### Tracker.List

```go
func () List() []*Process
```

List returns all tracked processes.

#### Tracker.Remove

```go
func () Remove(ctx context.Context, pid int, exitCode int)
```

Remove removes a process from tracking.


