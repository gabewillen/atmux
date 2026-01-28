# package tui

`import "github.com/copilot-claude-sonnet-4/amux/internal/tui"`

Package tui provides terminal user interface functionality.
This package handles TUI rendering and interaction without any
agent-specific knowledge.

- `ErrRenderFailed, ErrInputInvalid, ErrScreenNotInitialized` — Common sentinel errors for TUI operations.
- `type Interface` — Interface represents the terminal user interface.

### Variables

#### ErrRenderFailed, ErrInputInvalid, ErrScreenNotInitialized

```go
var (
	// ErrRenderFailed indicates TUI rendering failed.
	ErrRenderFailed = errors.New("render failed")

	// ErrInputInvalid indicates invalid input was received.
	ErrInputInvalid = errors.New("invalid input")

	// ErrScreenNotInitialized indicates the TUI screen is not initialized.
	ErrScreenNotInitialized = errors.New("screen not initialized")
)
```

Common sentinel errors for TUI operations.


## type Interface

```go
type Interface struct {
	initialized bool
}
```

Interface represents the terminal user interface.
Implementation deferred to Phase 5.

### Functions returning Interface

#### NewInterface

```go
func NewInterface() (*Interface, error)
```

NewInterface creates a new TUI interface.


### Methods

#### Interface.Close

```go
func () Close() error
```

Close shuts down the TUI interface.

#### Interface.Initialize

```go
func () Initialize() error
```

Initialize sets up the TUI screen.

#### Interface.Render

```go
func () Render() error
```

Render updates the TUI display.


