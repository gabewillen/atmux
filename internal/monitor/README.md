# package monitor

`import "github.com/copilot-claude-sonnet-4/amux/internal/monitor"`

Package monitor provides PTY observation and timeout detection.
This package observes PTY sessions generically and delegates pattern
matching to adapters, maintaining zero agent-specific knowledge.

- `ErrMonitorStopped, ErrInvalidTimeout, ErrObservationFailed` — Common sentinel errors for monitoring operations.
- `type Observer` — Observer monitors PTY sessions without agent-specific knowledge.

### Variables

#### ErrMonitorStopped, ErrInvalidTimeout, ErrObservationFailed

```go
var (
	// ErrMonitorStopped indicates the monitor has been stopped.
	ErrMonitorStopped = errors.New("monitor stopped")

	// ErrInvalidTimeout indicates an invalid timeout configuration.
	ErrInvalidTimeout = errors.New("invalid timeout")

	// ErrObservationFailed indicates a failure in PTY observation.
	ErrObservationFailed = errors.New("observation failed")
)
```

Common sentinel errors for monitoring operations.


## type Observer

```go
type Observer struct {
	timeout time.Duration
	stopped bool
}
```

Observer monitors PTY sessions without agent-specific knowledge.
Pattern matching and activity detection is delegated to adapters.

### Functions returning Observer

#### NewObserver

```go
func NewObserver(timeout time.Duration) (*Observer, error)
```

NewObserver creates a new PTY observer with the given timeout.


### Methods

#### Observer.Start

```go
func () Start() error
```

Start begins monitoring a PTY session.
The actual monitoring implementation is deferred to Phase 3.

#### Observer.Stop

```go
func () Stop() error
```

Stop halts PTY monitoring.


