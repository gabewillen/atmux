// Package errors provides common error handling conventions and sentinel errors for amux.
package errors

import "fmt"

// Wrap wraps an error with context, following the convention fmt.Errorf("context: %w", err).
func Wrap(context string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// Sentinel errors defined at package level.
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

	// ErrNotFound is returned when a resource cannot be found.
	ErrNotFound = New("not found")
)

// Error represents a sentinel error.
type Error struct {
	message string
}

// New creates a new sentinel error.
func New(message string) *Error {
	return &Error{message: message}
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.message
}

// Is returns true if the target error is the same sentinel error.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.message == t.message
}
