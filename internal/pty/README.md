# package pty

`import "github.com/agentflare-ai/amux/internal/pty"`

- `func Start(cmd *os.File) (*os.File, error)` — Start assigns a pseudo-terminal to the command.

### Functions

#### Start

```go
func Start(cmd *os.File) (*os.File, error)
```

Start assigns a pseudo-terminal to the command.
Phase 0: Wrapper for creack/pty.Start.


