# package pty

`import "github.com/agentflare-ai/amux/internal/pty"`

Package pty provides PTY management for amux.

- `func Demo(ctx context.Context) error` — Demo demonstrates PTY functionality for Phase 0.
- `type Config` — Config contains PTY configuration.
- `type PTY` — PTY represents a pseudo-terminal session.

### Functions

#### Demo

```go
func Demo(ctx context.Context) error
```

Demo demonstrates PTY functionality for Phase 0.


## type Config

```go
type Config struct {
	// Initial window size
	WindowSize *pty.Winsize `json:"window_size"`

	// Command to run
	Command string `json:"command"`

	// Arguments for command
	Args []string `json:"args"`

	// Environment variables
	Env []string `json:"env"`

	// Working directory
	WorkingDir string `json:"working_dir"`
}
```

Config contains PTY configuration.

## type PTY

```go
type PTY struct {
	// File descriptor for the PTY master
	master *os.File

	// File descriptor for the PTY slave
	slave *os.File

	// Command running in the PTY
	cmd *exec.Cmd

	// PTY size
	size *pty.Winsize

	// Context for cancellation
	ctx context.Context

	// Cancel function
	cancel context.CancelFunc
}
```

PTY represents a pseudo-terminal session.

### Functions returning PTY

#### New

```go
func New(ctx context.Context, config *Config) (*PTY, error)
```

New creates a new PTY session.


### Methods

#### PTY.Close

```go
func () Close() error
```

Close closes the PTY session.

#### PTY.Process

```go
func () Process() *os.Process
```

Process returns the underlying process.

#### PTY.Read

```go
func () Read(b []byte) (int, error)
```

Read reads data from the PTY.

#### PTY.SetSize

```go
func () SetSize(cols, rows uint16) error
```

SetSize sets the PTY window size.

#### PTY.Size

```go
func () Size() *pty.Winsize
```

Size returns the current PTY size.

#### PTY.Write

```go
func () Write(b []byte) (int, error)
```

Write writes data to the PTY.


