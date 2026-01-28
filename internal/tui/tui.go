// Package tui implements terminal screen decoding and TUI XML encoding (agent-agnostic)
package tui

import "errors"

// ErrTUIDecode is returned when TUI decoding fails
var ErrTUIDecode = errors.New("tui decode failed")