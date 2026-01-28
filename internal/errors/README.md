# package errors

`import "github.com/stateforward/amux/internal/errors"`

Package errors implements error handling conventions for the amux project

- `ErrNotImplemented, ErrInvalidInput, ErrInternal` — Sentinel errors for the common package
- `func Wrapf(err error, format string, args ...interface{}) error` — Wrapf wraps an error with additional context using fmt.Errorf Usage: errors.Wrapf(err, "context: %w", err) - following spec convention

### Variables

#### ErrNotImplemented, ErrInvalidInput, ErrInternal

```go
var (
	// ErrNotImplemented is returned when a feature is not yet implemented
	ErrNotImplemented = errors.New("not implemented")

	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")

	// ErrInternal is returned for internal errors
	ErrInternal = errors.New("internal error")
)
```

Sentinel errors for the common package


### Functions

#### Wrapf

```go
func Wrapf(err error, format string, args ...interface{}) error
```

Wrapf wraps an error with additional context using fmt.Errorf
Usage: errors.Wrapf(err, "context: %w", err) - following spec convention


