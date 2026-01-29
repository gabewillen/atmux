// Package errors implements error handling conventions for the amux project
package errors

import (
	"errors"
	"fmt"
)

// Sentinel errors for the common package
var (
	// ErrNotImplemented is returned when a feature is not yet implemented
	ErrNotImplemented = errors.New("not implemented")

	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")

	// ErrInternal is returned for internal errors
	ErrInternal = errors.New("internal error")
)

// Wrapf wraps an error with additional context using fmt.Errorf
// Usage: errors.Wrapf(err, "processing file %s", filename) - following spec convention
func Wrapf(err error, format string, args ...interface{}) error {
	// Append the error as the last argument and add the wrapping directive
	allArgs := append(args, err)
	fullFormat := format + ": %w"
	return fmt.Errorf(fullFormat, allArgs...)
}

// Wrap wraps an error with additional context
// Usage: errors.Wrap(err, "context") - following spec convention
func Wrap(err error, context string) error {
	return fmt.Errorf("%s: %w", context, err)
}