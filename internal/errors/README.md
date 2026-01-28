# package errors

`import "github.com/agentflare-ai/amux/internal/errors"`

Package errors provides common error handling conventions and sentinel errors for amux.

- `ErrInvalidConfig, ErrAgentNotFound, ErrSessionNotFound, ErrHostNotFound, ErrInvalidState, ErrNotReady, ErrTimeout, ErrDisconnected` — Sentinel errors defined at package level.
- `func Wrap(context string, err error) error` — Wrap wraps an error with context, following the convention fmt.Errorf("context: %w", err).
- `type Error` — Error represents a sentinel error.

### Variables

#### ErrInvalidConfig, ErrAgentNotFound, ErrSessionNotFound, ErrHostNotFound, ErrInvalidState, ErrNotReady, ErrTimeout, ErrDisconnected

```go
var (
	// ErrInvalidConfig is returned when configuration is invalid.
	ErrInvalidConfig = New("invalid configuration")

	// ErrAgentNotFound is returned when an agent cannot be found.
	ErrAgentNotFound = New("agent not found")

	// ErrSessionNotFound is returned when a session cannot be found.
	ErrSessionNotFound = New("session not found")

	// ErrHostNotFound is returned when a host cannot be found.
	ErrHostNotFound = New("host not found")

	// ErrInvalidState is returned when an operation is invalid for the current state.
	ErrInvalidState = New("invalid state")

	// ErrNotReady is returned when a component is not ready for the operation.
	ErrNotReady = New("not ready")

	// ErrTimeout is returned when an operation times out.
	ErrTimeout = New("operation timeout")

	// ErrDisconnected is returned when a connection is lost.
	ErrDisconnected = New("disconnected")
)
```

Sentinel errors defined at package level.


### Functions

#### Wrap

```go
func Wrap(context string, err error) error
```

Wrap wraps an error with context, following the convention fmt.Errorf("context: %w", err).


## type Error

```go
type Error struct {
	message string
}
```

Error represents a sentinel error.

### Functions returning Error

#### New

```go
func New(message string) *Error
```

New creates a new sentinel error.


### Methods

#### Error.Error

```go
func () Error() string
```

Error implements the error interface.

#### Error.Is

```go
func () Is(target error) bool
```

Is returns true if the target error is the same sentinel error.


