# package errors

`import "github.com/stateforward/amux/internal/errors"`

Package errors provides error handling conventions and sentinel errors for amux.

This package implements the error handling strategy per spec §4.2.5:
- All errors are wrapped with context using fmt.Errorf("context: %w", err)
- Sentinel errors are defined as package-level variables
- No errors are silently ignored

- `ErrNotFound, ErrInvalidInput, ErrAlreadyExists, ErrNotImplemented, ErrTimeout, ErrCancelled, ErrUnavailable, ErrPermissionDenied, ErrConflict, ErrInternal` — Sentinel errors for common error conditions across the application.
- `func Wrap(err error, context string) error` — Wrap wraps an error with context.
- `func Wrapf(err error, format string, args ...any) error` — Wrapf wraps an error with formatted context.

### Variables

#### ErrNotFound, ErrInvalidInput, ErrAlreadyExists, ErrNotImplemented, ErrTimeout, ErrCancelled, ErrUnavailable, ErrPermissionDenied, ErrConflict, ErrInternal

```go
var (
	// ErrNotFound indicates a requested resource was not found.
	ErrNotFound = errors.New("not found")

	// ErrInvalidInput indicates user-provided input failed validation.
	ErrInvalidInput = errors.New("invalid input")

	// ErrAlreadyExists indicates an attempt to create a resource that already exists.
	ErrAlreadyExists = errors.New("already exists")

	// ErrNotImplemented indicates functionality that is not yet implemented.
	ErrNotImplemented = errors.New("not implemented")

	// ErrTimeout indicates an operation exceeded its time limit.
	ErrTimeout = errors.New("timeout")

	// ErrCancelled indicates an operation was cancelled.
	ErrCancelled = errors.New("cancelled")

	// ErrUnavailable indicates a required resource or service is unavailable.
	ErrUnavailable = errors.New("unavailable")

	// ErrPermissionDenied indicates insufficient permissions for the operation.
	ErrPermissionDenied = errors.New("permission denied")

	// ErrConflict indicates a conflict with the current state.
	ErrConflict = errors.New("conflict")

	// ErrInternal indicates an internal error that should not occur.
	ErrInternal = errors.New("internal error")
)
```

Sentinel errors for common error conditions across the application.


### Functions

#### Wrap

```go
func Wrap(err error, context string) error
```

Wrap wraps an error with context. If err is nil, returns nil.
This is a convenience wrapper for fmt.Errorf with %w.

#### Wrapf

```go
func Wrapf(err error, format string, args ...any) error
```

Wrapf wraps an error with formatted context. If err is nil, returns nil.


