package errors

import (
	"errors"
	"fmt"
)

// New returns a new error with the given message.
// It is a wrapper around generic errors.New.
func New(message string) error {
	return errors.New(message)
}

// Wrap returns an error annotating err with a stack trace
// at the point Wrap is called, and the supplied message.
// If err is nil, Wrap returns nil.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Is reports whether any error in err's chain matches target.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target, and if so, sets
// target to that error value and returns true. Otherwise, it returns false.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Common sentinel errors
var (
	ErrNotFound      = New("not found")
	ErrInvalidConfig = New("invalid configuration")
	ErrNotImplemented = New("not implemented")
)
