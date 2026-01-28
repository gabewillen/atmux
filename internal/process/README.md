# package process

`import "github.com/copilot-claude-sonnet-4/amux/internal/process"`

Package process provides generic child process tracking functionality.
This package observes processes generically without any agent-specific logic,
supporting both hook-based interception and polling fallback.

- `ErrProcessNotFound, ErrTrackingFailed, ErrInterceptionUnavailable` — Common sentinel errors for process operations.
- `type Tracker` — Tracker manages child process monitoring.

### Variables

#### ErrProcessNotFound, ErrTrackingFailed, ErrInterceptionUnavailable

```go
var (
	// ErrProcessNotFound indicates a process was not found.
	ErrProcessNotFound = errors.New("process not found")

	// ErrTrackingFailed indicates process tracking initialization failed.
	ErrTrackingFailed = errors.New("tracking failed")

	// ErrInterceptionUnavailable indicates process interception is not available.
	ErrInterceptionUnavailable = errors.New("interception unavailable")
)
```

Common sentinel errors for process operations.


## type Tracker

```go
type Tracker struct {
	hookAvailable bool
}
```

Tracker manages child process monitoring.
Uses LD_PRELOAD/DYLD_INSERT_LIBRARIES with polling fallback.

### Functions returning Tracker

#### NewTracker

```go
func NewTracker() (*Tracker, error)
```

NewTracker creates a new process tracker.


### Methods

#### Tracker.IsHookAvailable

```go
func () IsHookAvailable() bool
```

IsHookAvailable returns whether process interception hooks are available.

#### Tracker.Track

```go
func () Track(pid int) error
```

Track begins monitoring a process by PID.


