# package pty

`import "github.com/agentflare-ai/amux/internal/pty"`

Package pty provides PTY creation and I/O helpers.

- `type Pair` — Pair represents a PTY master/slave pair.

## type Pair

```go
type Pair struct {
	Master *os.File
	Slave  *os.File
}
```

Pair represents a PTY master/slave pair.

### Functions returning Pair

#### Open

```go
func Open() (*Pair, error)
```

Open allocates a new PTY pair.


### Methods

#### Pair.Close

```go
func () Close() error
```

Close closes the PTY pair.


