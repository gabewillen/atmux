# package pty

`import "github.com/copilot-claude-sonnet-4/amux/internal/pty"`

Package pty provides pseudo-terminal creation and I/O operations.
This package handles raw PTY operations without any agent-specific logic.

Uses creack/pty for cross-platform PTY management on Linux and macOS
with non-blocking I/O via standard Go interfaces.

- `ErrPTYCreateFailed, ErrInvalidSize, ErrPTYClosed` — Common sentinel errors for PTY operations.
- `type Session` — Session represents a PTY session with master and slave file descriptors.

### Variables

#### ErrPTYCreateFailed, ErrInvalidSize, ErrPTYClosed

```go
var (
	// ErrPTYCreateFailed indicates PTY creation failed.
	ErrPTYCreateFailed = errors.New("PTY creation failed")

	// ErrInvalidSize indicates an invalid PTY window size.
	ErrInvalidSize = errors.New("invalid PTY size")

	// ErrPTYClosed indicates the PTY has been closed.
	ErrPTYClosed = errors.New("PTY closed")
)
```

Common sentinel errors for PTY operations.


## type Session

```go
type Session struct {
	master *os.File
	slave  *os.File
	size   *pty.Winsize
}
```

Session represents a PTY session with master and slave file descriptors.

### Functions returning Session

#### NewSession

```go
func NewSession() (*Session, error)
```

NewSession creates a new PTY session.


### Methods

#### Session.Close

```go
func () Close() error
```

Close closes both master and slave file descriptors.

#### Session.Master

```go
func () Master() *os.File
```

Master returns the master file descriptor for writing to the PTY.

#### Session.SetSize

```go
func () SetSize(rows, cols uint16) error
```

SetSize updates the PTY window size.

#### Session.Slave

```go
func () Slave() *os.File
```

Slave returns the slave file descriptor for the child process.


