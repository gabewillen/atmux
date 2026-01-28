# package pty

`import "github.com/agentflare-ai/amux/internal/pty"`

Package pty provides PTY (pseudo-terminal) management for amux.

This package wraps creack/pty to provide PTY creation, I/O, and lifecycle
management. All PTY operations are agent-agnostic.

See spec §4.2.4 and §7 for PTY requirements.

- `type PTY` — PTY represents a pseudo-terminal.

## type PTY

```go
type PTY struct {
	mu     sync.Mutex
	file   *os.File
	cmd    *exec.Cmd
	size   *pty.Winsize
	closed bool
}
```

PTY represents a pseudo-terminal.

### Functions returning PTY

#### Open

```go
func Open(cmd *exec.Cmd) (*PTY, error)
```

Open creates a new PTY for the given command.


### Methods

#### PTY.Close

```go
func () Close() error
```

Close closes the PTY.

#### PTY.File

```go
func () File() *os.File
```

File returns the underlying PTY file descriptor.

#### PTY.Read

```go
func () Read(buf []byte) (int, error)
```

Read reads from the PTY.

#### PTY.Resize

```go
func () Resize(rows, cols uint16) error
```

Resize changes the PTY window size.

#### PTY.Wait

```go
func () Wait() error
```

Wait waits for the PTY command to exit.

#### PTY.Write

```go
func () Write(data []byte) (int, error)
```

Write writes to the PTY.


