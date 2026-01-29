# package session

`import "github.com/agentflare-ai/amux/internal/session"`

Package session provides local agent session management with owned PTYs.

A session represents a running agent PTY instance. Each agent has at most
one active session. Sessions own the PTY and are responsible for starting
the agent shell, managing I/O, and cleaning up on shutdown.

The session package integrates with the agent lifecycle HSM to drive
state transitions (Pending → Starting → Running → Terminated/Errored).

See spec §5.4, §5.6, and §B.5 for lifecycle, shutdown, and PTY ownership.

- `type Manager` — Manager manages active sessions for agents.
- `type Session` — Session represents a running agent PTY session.
- `type State` — State represents the session state.

## type Manager

```go
type Manager struct {
	mu         sync.RWMutex
	sessions   map[muid.MUID]*Session // agent ID -> session
	dispatcher event.Dispatcher
}
```

Manager manages active sessions for agents.

### Functions returning Manager

#### NewManager

```go
func NewManager(dispatcher event.Dispatcher) *Manager
```

NewManager creates a new session manager.


### Methods

#### Manager.Get

```go
func () Get(agentID muid.MUID) *Session
```

Get returns the session for an agent, or nil if none exists.

#### Manager.Kill

```go
func () Kill(ctx context.Context, agentID muid.MUID) error
```

Kill forcefully terminates a session.

#### Manager.KillAll

```go
func () KillAll()
```

KillAll forcefully terminates all sessions.

#### Manager.List

```go
func () List() []*Session
```

List returns all active sessions.

#### Manager.Remove

```go
func () Remove(agentID muid.MUID)
```

Remove removes a session from the manager.

#### Manager.Spawn

```go
func () Spawn(ctx context.Context, ag *agent.Agent, shell string, args ...string) (*Session, error)
```

Spawn creates and starts a new PTY session for an agent.

The shell command is executed in the agent's worktree directory.
The session takes ownership of the PTY and monitors the process.

See spec §5.4 (lifecycle) and §B.5 (owned PTY).

#### Manager.Stop

```go
func () Stop(ctx context.Context, agentID muid.MUID) error
```

Stop stops a session by closing the PTY and waiting for the process to exit.

If the process does not exit within context deadline, it is killed.
See spec §5.6.3 for agent shutdown behavior.

#### Manager.StopAll

```go
func () StopAll()
```

StopAll stops all sessions gracefully.


## type Session

```go
type Session struct {
	mu sync.RWMutex

	// ID is the unique session identifier.
	ID muid.MUID

	// AgentID is the ID of the agent this session belongs to.
	AgentID muid.MUID

	// Agent is the managed agent instance.
	Agent *agent.Agent

	// PTY is the owned pseudo-terminal for this session.
	PTY *amuxpty.PTY

	// cmd is the shell command running in the PTY.
	cmd *exec.Cmd

	// state tracks whether the session is running.
	state State

	// dispatcher is used to emit session events.
	dispatcher event.Dispatcher

	// done is closed when the session exits.
	done chan struct{}

	// exitErr holds the process exit error, if any.
	exitErr error
}
```

Session represents a running agent PTY session.

Each session owns a PTY and manages the agent's shell process.
The session is the single owner of the agent's terminal I/O.

### Methods

#### Session.Done

```go
func () Done() <-chan struct{}
```

Done returns a channel that is closed when the session exits.

#### Session.ExitErr

```go
func () ExitErr() error
```

ExitErr returns the process exit error, or nil if still running or exited cleanly.

#### Session.Kill

```go
func () Kill() error
```

Kill forcefully terminates the session.

#### Session.Read

```go
func () Read(buf []byte) (int, error)
```

Read reads data from the PTY (output from the agent).

#### Session.Resize

```go
func () Resize(rows, cols uint16) error
```

Resize changes the PTY window size.

#### Session.SessionState

```go
func () SessionState() State
```

State returns the session state.

#### Session.Stop

```go
func () Stop() error
```

Stop gracefully stops the session by closing the PTY (sends EOF to shell).

#### Session.Write

```go
func () Write(data []byte) (int, error)
```

Write writes data to the PTY (input to the agent).

#### Session.waitForExit

```go
func () waitForExit()
```

waitForExit waits for the PTY process to exit and updates state.


## type State

```go
type State string
```

State represents the session state.

### Constants

#### StateCreated, StateRunning, StateStopped

```go
const (
	// StateCreated indicates the session has been created but not started.
	StateCreated State = "created"

	// StateRunning indicates the session is actively running.
	StateRunning State = "running"

	// StateStopped indicates the session has been stopped.
	StateStopped State = "stopped"
)
```


