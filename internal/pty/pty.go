// Package pty implements PTY management (generic PTY operations)
package pty

import (
	"errors"

	// Import creack/pty to ensure it's included in the dependencies
	_ "github.com/creack/pty"
)

// ErrPTYFailure is returned when a PTY operation fails
var ErrPTYFailure = errors.New("pty operation failed")