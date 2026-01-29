# package pty

`import "github.com/agentflare-ai/amux/internal/pty"`

- `type PTY` — PTY wraps a pseudo-terminal file and its associated command.

## type PTY

```go
type PTY struct {
	File *os.File
	Cmd  *exec.Cmd
}
```

PTY wraps a pseudo-terminal file and its associated command.

### Functions returning PTY

#### Start

```go
func Start(cmd *exec.Cmd) (*PTY, error)
```

Start starts a command in a new PTY.


### Methods

#### PTY.Close

```go
func () Close() error
```

Close closes the PTY file.

#### PTY.Resize

```go
func () Resize(rows, cols uint16) error
```

Resize resizes the PTY window.


