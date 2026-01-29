# package pty

`import "github.com/agentflare-ai/amux/internal/pty"`

- `func Close(f *os.File) error` — Close closes the PTY file descriptor.
- `func Resize(f *os.File, rows, cols int) error` — Resize resizes the PTY.
- `func Start(command string, args []string, dir string) (*os.File, error)` — Start starts a process in a PTY.

### Functions

#### Close

```go
func Close(f *os.File) error
```

Close closes the PTY file descriptor.
Note: This often triggers SIGHUP to the process.

#### Resize

```go
func Resize(f *os.File, rows, cols int) error
```

Resize resizes the PTY.

#### Start

```go
func Start(command string, args []string, dir string) (*os.File, error)
```

Start starts a process in a PTY.
If dir is empty, uses current directory.


