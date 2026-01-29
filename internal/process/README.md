# package process

`import "github.com/stateforward/amux/internal/process"`

Package process implements process tracking and interception (generic)

- `func splitEnvVar(envVar string) []string` — splitEnvVar splits an environment variable string into key and value
- `type ProcessStatus` — ProcessStatus represents the status of a process
- `type Process` — Process represents a tracked process
- `type Tracker` — Tracker tracks processes and their relationships

### Functions

#### splitEnvVar

```go
func splitEnvVar(envVar string) []string
```

splitEnvVar splits an environment variable string into key and value


## type Process

```go
type Process struct {
	ID          muid.MUID         `json:"id"`
	PID         int               `json:"pid"`
	Command     string            `json:"command"`
	Args        []string          `json:"args"`
	StartTime   time.Time         `json:"start_time"`
	EndTime     *time.Time        `json:"end_time,omitempty"`
	Status      ProcessStatus     `json:"status"`
	ExitCode    *int              `json:"exit_code,omitempty"`
	ParentID    *muid.MUID        `json:"parent_id,omitempty"`
	Children    []muid.MUID       `json:"children,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
}
```

Process represents a tracked process

## type ProcessStatus

```go
type ProcessStatus string
```

ProcessStatus represents the status of a process

### Constants

#### ProcessRunning, ProcessCompleted, ProcessErrored, ProcessKilled

```go
const (
	ProcessRunning   ProcessStatus = "running"
	ProcessCompleted ProcessStatus = "completed"
	ProcessErrored   ProcessStatus = "errored"
	ProcessKilled    ProcessStatus = "killed"
)
```


## type Tracker

```go
type Tracker struct {
	mu        sync.RWMutex
	processes map[muid.MUID]*Process
	commands  map[int]muid.MUID // PID -> Process ID mapping
	ctx       context.Context
	cancel    context.CancelFunc
}
```

Tracker tracks processes and their relationships

### Functions returning Tracker

#### NewTracker

```go
func NewTracker() *Tracker
```

NewTracker creates a new process tracker


### Methods

#### Tracker.GetProcess

```go
func () GetProcess(id muid.MUID) (*Process, error)
```

GetProcess retrieves a process by ID

#### Tracker.GetProcessByPID

```go
func () GetProcessByPID(pid int) (*Process, error)
```

GetProcessByPID retrieves a process by PID

#### Tracker.ListProcesses

```go
func () ListProcesses() []*Process
```

ListProcesses returns a list of all tracked processes

#### Tracker.RecordProcessExit

```go
func () RecordProcessExit(id muid.MUID, exitCode int) error
```

RecordProcessExit records the exit of a process

#### Tracker.Stop

```go
func () Stop()
```

Stop stops the tracker and cleans up resources

#### Tracker.TrackCommand

```go
func () TrackCommand(cmd *exec.Cmd) (muid.MUID, error)
```

TrackCommand starts tracking a command

#### Tracker.UpdateProcessStatus

```go
func () UpdateProcessStatus(id muid.MUID, status ProcessStatus) error
```

UpdateProcessStatus updates the status of a process

#### Tracker.cleanupCompletedProcesses

```go
func () cleanupCompletedProcesses()
```

cleanupCompletedProcesses removes completed processes older than 10 minutes

#### Tracker.cleanupRoutine

```go
func () cleanupRoutine()
```

cleanupRoutine periodically cleans up completed processes


