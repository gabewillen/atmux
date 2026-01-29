# package errors

`import "github.com/agentflare-ai/amux/internal/errors"`

- `ErrNotFound, ErrInvalidConfig, ErrNotImplemented` — Common sentinel errors
- `func As(err error, target any) bool` — As finds the first error in err's chain that matches target, and if so, sets target to that error value and returns true.
- `func Errorf(format string, args ...any) error` — Errorf returns a new error with the given format and args.
- `func Is(err, target error) bool` — Is reports whether any error in err's chain matches target.
- `func New(message string) error` — New returns a new error with the given message.
- `func Wrap(err error, message string) error` — Wrap returns an error annotating err with a stack trace at the point Wrap is called, and the supplied message.
- `func Wrapf(err error, format string, args ...any) error` — Wrapf returns an error annotating err with a stack trace at the point Wrapf is called, and the format specifier.

### Variables

#### ErrNotFound, ErrInvalidConfig, ErrNotImplemented

```go
var (
	ErrNotFound       = New("not found")
	ErrInvalidConfig  = New("invalid configuration")
	ErrNotImplemented = New("not implemented")
)
```

Common sentinel errors


### Functions

#### As

```go
func As(err error, target any) bool
```

As finds the first error in err's chain that matches target, and if so, sets
target to that error value and returns true. Otherwise, it returns false.

#### Errorf

```go
func Errorf(format string, args ...any) error
```

Errorf returns a new error with the given format and args.

#### Is

```go
func Is(err, target error) bool
```

Is reports whether any error in err's chain matches target.

#### New

```go
func New(message string) error
```

New returns a new error with the given message.
It is a wrapper around generic errors.New.

#### Wrap

```go
func Wrap(err error, message string) error
```

Wrap returns an error annotating err with a stack trace
at the point Wrap is called, and the supplied message.
If err is nil, Wrap returns nil.

#### Wrapf

```go
func Wrapf(err error, format string, args ...any) error
```

Wrapf returns an error annotating err with a stack trace
at the point Wrapf is called, and the format specifier.
If err is nil, Wrapf returns nil.


