// Package process provides generic child process tracking functionality.
// This package observes processes generically without any agent-specific logic,
// supporting both hook-based interception and polling fallback.
package process

import (
	"errors"
	"fmt"
	"os"
)

// Common sentinel errors for process operations.
var (
	// ErrProcessNotFound indicates a process was not found.
	ErrProcessNotFound = errors.New("process not found")

	// ErrTrackingFailed indicates process tracking initialization failed.
	ErrTrackingFailed = errors.New("tracking failed")

	// ErrInterceptionUnavailable indicates process interception is not available.
	ErrInterceptionUnavailable = errors.New("interception unavailable")
)

// Tracker manages child process monitoring.
// Uses LD_PRELOAD/DYLD_INSERT_LIBRARIES with polling fallback.
type Tracker struct {
	hookAvailable bool
}

// NewTracker creates a new process tracker.
func NewTracker() (*Tracker, error) {
	tracker := &Tracker{
		hookAvailable: false, // Will be implemented in Phase 6
	}

	return tracker, nil
}

// Track begins monitoring a process by PID.
func (t *Tracker) Track(pid int) error {
	if pid <= 0 {
		return fmt.Errorf("invalid PID %d: %w", pid, ErrProcessNotFound)
	}

	// Verify process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", err, ErrProcessNotFound)
	}
	_ = process // Use process in actual implementation

	// Implementation deferred to Phase 6
	return fmt.Errorf("process tracking not implemented: %w", ErrTrackingFailed)
}

// IsHookAvailable returns whether process interception hooks are available.
func (t *Tracker) IsHookAvailable() bool {
	return t.hookAvailable
}