# package pty

`import "github.com/agentflare-ai/amux/internal/pty"`

Package pty provides PTY creation and lifecycle for amux (spec §4.2.4, §7).
amux owns the PTY for each agent; the monitor observes raw output from the PTY.

- `type Session` — Session represents an owned PTY session for an agent (spec §7, B.5).

## type Session

```go
type Session struct {
	cmd    *exec.Cmd
	pty    *os.File
	mu     sync.Mutex
	closed bool
}
```

Session represents an owned PTY session for an agent (spec §7, B.5).
The PTY is created and owned by amux; the slave side is used as the agent's terminal.

### Functions returning Session

#### NewSession

```go
func NewSession(workDir string, name string, args []string, env []string) (*Session, error)
```

NewSession starts a command in a new PTY with the given working directory and environment.
The command runs with the PTY as its stdin/stdout/stderr. Caller must call Close when done.


### Methods

#### Session.Close

```go
func () Close() error
```

Close closes the PTY and waits for the command to exit.
Idempotent; safe to call multiple times.

#### Session.OutputStream

```go
func () OutputStream() io.Reader
```

StdoutPipe is not used; PTY owns stdin/stdout/stderr. Exposed for tests that need raw stream.
OutputStream returns a reader for the PTY master (agent output). Do not close the returned reader.

#### Session.PTY

```go
func () PTY() *os.File
```

PTY returns the master end of the PTY for reading output and writing input.
Do not close it; use Session.Close to close the session.

#### Session.Process

```go
func () Process() *os.Process
```

Process returns the underlying process, or nil if not started.

#### Session.Read

```go
func () Read(p []byte) (n int, err error)
```

Read reads from the PTY master (agent output).

#### Session.Resize

```go
func () Resize(rows, cols uint16) error
```

Resize sets the PTY window size (spec §4.2.4).

#### Session.Wait

```go
func () Wait() error
```

Wait blocks until the command exits and returns its error.

#### Session.Write

```go
func () Write(p []byte) (n int, err error)
```

Write writes to the PTY master (agent input).


