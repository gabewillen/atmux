# package pty

`import "github.com/copilot-claude-sonnet-4/amux/internal/pty"`

Package pty provides PTY session management for agents.
This file implements the owned PTY session model per spec requirements.

Package pty provides pseudo-terminal creation and I/O operations.
This package handles raw PTY operations without any agent-specific logic.

Uses creack/pty for cross-platform PTY management on Linux and macOS
with non-blocking I/O via standard Go interfaces.

- `ErrManagedSessionNotFound, ErrManagedSessionClosed` — Additional errors for the session manager
- `ErrPTYCreateFailed, ErrInvalidSize, ErrPTYClosed` — Common sentinel errors for PTY operations.
- `type ManagedSession` — ManagedSession represents an owned PTY session for an agent.
- `type SessionInfo` — SessionInfo provides read-only session information.
- `type SessionManager` — SessionManager manages multiple PTY sessions.
- `type SessionState` — SessionState represents the state of a PTY session.
- `type Session` — Session represents a PTY session with master and slave file descriptors.

### Variables

#### ErrManagedSessionNotFound, ErrManagedSessionClosed

```go
var (
	ErrManagedSessionNotFound = errors.New("managed session not found")
	ErrManagedSessionClosed   = errors.New("managed session closed")
)
```

Additional errors for the session manager

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


## type ManagedSession

```go
type ManagedSession struct {
	ID         muid.MUID
	AgentID    muid.MUID
	Command    []string
	WorkingDir string
	State      SessionState
	CreatedAt  time.Time
	StartedAt  *time.Time
	EndedAt    *time.Time
	ExitCode   *int

	// Internal fields
	pty       *os.File
	process   *os.Process
	cmd       *exec.Cmd
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex
	output    chan []byte
	closed    bool
	waitGroup sync.WaitGroup
}
```

ManagedSession represents an owned PTY session for an agent.

### Methods

#### ManagedSession.GetInfo

```go
func () GetInfo() SessionInfo
```

GetInfo returns session information safely.

#### ManagedSession.GetState

```go
func () GetState() SessionState
```

GetState returns the current session state safely.

#### ManagedSession.Kill

```go
func () Kill() error
```

Kill forcefully kills the PTY session.

#### ManagedSession.ReadOutput

```go
func () ReadOutput() <-chan []byte
```

ReadOutput reads buffered output from the PTY session.

#### ManagedSession.Terminate

```go
func () Terminate() error
```

Terminate gracefully terminates the PTY session.

#### ManagedSession.Write

```go
func () Write(data []byte) error
```

Write sends input to the PTY session.

#### ManagedSession.monitorOutput

```go
func () monitorOutput()
```

monitorOutput continuously reads from PTY and buffers output

#### ManagedSession.waitForCompletion

```go
func () waitForCompletion()
```

waitForCompletion waits for the process to complete and updates state


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


## type SessionInfo

```go
type SessionInfo struct {
	ID         muid.MUID
	AgentID    muid.MUID
	Command    []string
	WorkingDir string
	State      SessionState
	CreatedAt  time.Time
	StartedAt  *time.Time
	EndedAt    *time.Time
	ExitCode   *int
}
```

SessionInfo provides read-only session information.

## type SessionManager

```go
type SessionManager struct {
	sessions map[muid.MUID]*ManagedSession
	mu       sync.RWMutex
}
```

SessionManager manages multiple PTY sessions.

### Functions returning SessionManager

#### NewSessionManager

```go
func NewSessionManager() *SessionManager
```

NewSessionManager creates a new PTY session manager.


### Methods

#### SessionManager.CreateSession

```go
func () CreateSession(agentID muid.MUID, command []string, workingDir string) (*ManagedSession, error)
```

CreateSession creates a new PTY session for the given agent.

#### SessionManager.GetSession

```go
func () GetSession(sessionID muid.MUID) (*ManagedSession, error)
```

GetSession returns a PTY session by ID.

#### SessionManager.GetSessionByAgent

```go
func () GetSessionByAgent(agentID muid.MUID) (*ManagedSession, error)
```

GetSessionByAgent returns the active PTY session for an agent.

#### SessionManager.ListSessions

```go
func () ListSessions() []*ManagedSession
```

ListSessions returns all PTY sessions.

#### SessionManager.StartSession

```go
func () StartSession(sessionID muid.MUID) error
```

StartSession starts the PTY session and begins command execution.

#### SessionManager.TerminateSession

```go
func () TerminateSession(sessionID muid.MUID) error
```

TerminateSession terminates a PTY session.


## type SessionState

```go
type SessionState string
```

SessionState represents the state of a PTY session.

### Constants

#### SessionStateStarting, SessionStateRunning, SessionStateTerminated, SessionStateErrored

```go
const (
	// SessionStateStarting indicates the session is starting up.
	SessionStateStarting SessionState = "starting"

	// SessionStateRunning indicates the session is active.
	SessionStateRunning SessionState = "running"

	// SessionStateTerminated indicates the session has ended normally.
	SessionStateTerminated SessionState = "terminated"

	// SessionStateErrored indicates the session ended with an error.
	SessionStateErrored SessionState = "errored"
)
```


