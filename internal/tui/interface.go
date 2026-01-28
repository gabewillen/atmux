// Package tui provides terminal user interface functionality.
// This package handles TUI rendering and interaction without any
// agent-specific knowledge.
package tui

import (
	"errors"
	"fmt"
)

// Common sentinel errors for TUI operations.
var (
	// ErrRenderFailed indicates TUI rendering failed.
	ErrRenderFailed = errors.New("render failed")

	// ErrInputInvalid indicates invalid input was received.
	ErrInputInvalid = errors.New("invalid input")

	// ErrScreenNotInitialized indicates the TUI screen is not initialized.
	ErrScreenNotInitialized = errors.New("screen not initialized")
)

// Interface represents the terminal user interface.
// Implementation deferred to Phase 5.
type Interface struct {
	initialized bool
}

// NewInterface creates a new TUI interface.
func NewInterface() (*Interface, error) {
	return &Interface{
		initialized: false,
	}, nil
}

// Initialize sets up the TUI screen.
func (ui *Interface) Initialize() error {
	// Implementation deferred to Phase 5
	return fmt.Errorf("TUI initialization not implemented: %w", ErrScreenNotInitialized)
}

// Render updates the TUI display.
func (ui *Interface) Render() error {
	if !ui.initialized {
		return fmt.Errorf("TUI not initialized: %w", ErrScreenNotInitialized)
	}

	// Implementation deferred to Phase 5
	return fmt.Errorf("TUI rendering not implemented: %w", ErrRenderFailed)
}

// Close shuts down the TUI interface.
func (ui *Interface) Close() error {
	ui.initialized = false
	return nil
}