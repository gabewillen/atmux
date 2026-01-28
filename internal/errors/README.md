# package errors

`import "github.com/agentflare-ai/amux/internal/errors"`

Package errors provides sentinel errors and error handling conventions for amux.

This package follows the error handling conventions specified in §4.2.5:
- Errors are wrapped with context using fmt.Errorf("context: %w", err)
- Sentinel errors are defined as package-level variables using errors.New()
- Error checking is not deferred; errors are handled at the point of occurrence

- `ErrAdapterNotFound, ErrAdapterLoadFailed, ErrAdapterCallFailed, ErrAdapterManifestInvalid, ErrAdapterVersionIncompatible` — Adapter errors
- `ErrAgentNotFound, ErrAgentAlreadyExists, ErrAgentNotRunning, ErrAgentSlugCollision` — Agent errors
- `ErrConfigNotFound, ErrConfigInvalid, ErrConfigParseError` — Configuration errors
- `ErrConnectionFailed, ErrHandshakeFailed, ErrHostNotConnected, ErrSessionConflict` — Remote errors
- `ErrModelNotFound, ErrModelLoadFailed, ErrInferenceUnavailable` — Inference errors
- `ErrNotFound, ErrAlreadyExists, ErrInvalidInput, ErrNotReady, ErrTimeout, ErrClosed, ErrPermissionDenied, ErrNotImplemented` — Sentinel errors for common error conditions.
- `ErrNotInRepository, ErrWorktreeCreateFailed, ErrWorktreeRemoveFailed` — Repository errors
- `func As(err error, target any) bool` — As finds the first error in err's chain that matches target.
- `func Is(err, target error) bool` — Is reports whether any error in err's chain matches target.
- `func Join(errs ...error) error` — Join returns an error that wraps the given errors.
- `func New(text string) error` — New returns a new error with the given message.
- `func Wrap(context string, err error) error` — Wrap wraps an error with additional context.
- `func Wrapf(err error, format string, args ...any) error` — Wrapf wraps an error with formatted context.

### Variables

#### ErrNotFound, ErrAlreadyExists, ErrInvalidInput, ErrNotReady, ErrTimeout, ErrClosed, ErrPermissionDenied, ErrNotImplemented

```go
var (
	// ErrNotFound indicates a requested resource was not found.
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists indicates a resource already exists.
	ErrAlreadyExists = errors.New("already exists")

	// ErrInvalidInput indicates invalid input was provided.
	ErrInvalidInput = errors.New("invalid input")

	// ErrNotReady indicates the system is not ready for the operation.
	ErrNotReady = errors.New("not ready")

	// ErrTimeout indicates an operation timed out.
	ErrTimeout = errors.New("timeout")

	// ErrClosed indicates an operation was attempted on a closed resource.
	ErrClosed = errors.New("closed")

	// ErrPermissionDenied indicates the operation is not permitted.
	ErrPermissionDenied = errors.New("permission denied")

	// ErrNotImplemented indicates a feature is not yet implemented.
	ErrNotImplemented = errors.New("not implemented")
)
```

Sentinel errors for common error conditions.
These errors can be checked using errors.Is().

#### ErrConfigNotFound, ErrConfigInvalid, ErrConfigParseError

```go
var (
	// ErrConfigNotFound indicates a configuration file was not found.
	ErrConfigNotFound = errors.New("configuration not found")

	// ErrConfigInvalid indicates a configuration is invalid.
	ErrConfigInvalid = errors.New("configuration invalid")

	// ErrConfigParseError indicates a configuration parsing error.
	ErrConfigParseError = errors.New("configuration parse error")
)
```

Configuration errors

#### ErrAgentNotFound, ErrAgentAlreadyExists, ErrAgentNotRunning, ErrAgentSlugCollision

```go
var (
	// ErrAgentNotFound indicates an agent was not found.
	ErrAgentNotFound = fmt.Errorf("agent %w", ErrNotFound)

	// ErrAgentAlreadyExists indicates an agent already exists.
	ErrAgentAlreadyExists = fmt.Errorf("agent %w", ErrAlreadyExists)

	// ErrAgentNotRunning indicates an operation requires a running agent.
	ErrAgentNotRunning = errors.New("agent not running")

	// ErrAgentSlugCollision indicates an agent slug collision.
	ErrAgentSlugCollision = errors.New("agent slug collision")
)
```

Agent errors

#### ErrAdapterNotFound, ErrAdapterLoadFailed, ErrAdapterCallFailed, ErrAdapterManifestInvalid, ErrAdapterVersionIncompatible

```go
var (
	// ErrAdapterNotFound indicates an adapter was not found.
	ErrAdapterNotFound = fmt.Errorf("adapter %w", ErrNotFound)

	// ErrAdapterLoadFailed indicates an adapter failed to load.
	ErrAdapterLoadFailed = errors.New("adapter load failed")

	// ErrAdapterCallFailed indicates an adapter call failed.
	ErrAdapterCallFailed = errors.New("adapter call failed")

	// ErrAdapterManifestInvalid indicates an adapter manifest is invalid.
	ErrAdapterManifestInvalid = errors.New("adapter manifest invalid")

	// ErrAdapterVersionIncompatible indicates an adapter version is incompatible.
	ErrAdapterVersionIncompatible = errors.New("adapter version incompatible")
)
```

Adapter errors

#### ErrNotInRepository, ErrWorktreeCreateFailed, ErrWorktreeRemoveFailed

```go
var (
	// ErrNotInRepository indicates the operation requires a git repository.
	ErrNotInRepository = errors.New("not in a git repository")

	// ErrWorktreeCreateFailed indicates worktree creation failed.
	ErrWorktreeCreateFailed = errors.New("worktree creation failed")

	// ErrWorktreeRemoveFailed indicates worktree removal failed.
	ErrWorktreeRemoveFailed = errors.New("worktree removal failed")
)
```

Repository errors

#### ErrConnectionFailed, ErrHandshakeFailed, ErrHostNotConnected, ErrSessionConflict

```go
var (
	// ErrConnectionFailed indicates a connection failed.
	ErrConnectionFailed = errors.New("connection failed")

	// ErrHandshakeFailed indicates a handshake failed.
	ErrHandshakeFailed = errors.New("handshake failed")

	// ErrHostNotConnected indicates the host is not connected.
	ErrHostNotConnected = errors.New("host not connected")

	// ErrSessionConflict indicates a session conflict.
	ErrSessionConflict = errors.New("session conflict")
)
```

Remote errors

#### ErrModelNotFound, ErrModelLoadFailed, ErrInferenceUnavailable

```go
var (
	// ErrModelNotFound indicates a model was not found.
	ErrModelNotFound = errors.New("model not found")

	// ErrModelLoadFailed indicates a model failed to load.
	ErrModelLoadFailed = errors.New("model load failed")

	// ErrInferenceUnavailable indicates inference is unavailable.
	ErrInferenceUnavailable = errors.New("inference unavailable")
)
```

Inference errors


### Functions

#### As

```go
func As(err error, target any) bool
```

As finds the first error in err's chain that matches target.
This is a re-export of errors.As for convenience.

#### Is

```go
func Is(err, target error) bool
```

Is reports whether any error in err's chain matches target.
This is a re-export of errors.Is for convenience.

#### Join

```go
func Join(errs ...error) error
```

Join returns an error that wraps the given errors.
This is a re-export of errors.Join for convenience.

#### New

```go
func New(text string) error
```

New returns a new error with the given message.
This is a re-export of errors.New for convenience.

#### Wrap

```go
func Wrap(context string, err error) error
```

Wrap wraps an error with additional context.
This is a convenience function that follows the spec convention:
fmt.Errorf("context: %w", err)

#### Wrapf

```go
func Wrapf(err error, format string, args ...any) error
```

Wrapf wraps an error with formatted context.


