// Package monitor provides PTY observation and timeout detection.
// This package observes PTY sessions generically and delegates pattern
// matching to adapters, maintaining zero agent-specific knowledge.
package monitor

import (
	"errors"
	"fmt"
	"time"
)

// Common sentinel errors for monitoring operations.
var (
	// ErrMonitorStopped indicates the monitor has been stopped.
	ErrMonitorStopped = errors.New("monitor stopped")

	// ErrInvalidTimeout indicates an invalid timeout configuration.
	ErrInvalidTimeout = errors.New("invalid timeout")

	// ErrObservationFailed indicates a failure in PTY observation.
	ErrObservationFailed = errors.New("observation failed")
)

// Observer monitors PTY sessions without agent-specific knowledge.
// Pattern matching and activity detection is delegated to adapters.
type Observer struct {
	timeout time.Duration
	stopped bool
}

// NewObserver creates a new PTY observer with the given timeout.
func NewObserver(timeout time.Duration) (*Observer, error) {
	if timeout <= 0 {
		return nil, fmt.Errorf("timeout must be positive: %w", ErrInvalidTimeout)
	}

	return &Observer{
		timeout: timeout,
		stopped: false,
	}, nil
}

// Start begins monitoring a PTY session.
// The actual monitoring implementation is deferred to Phase 3.
func (o *Observer) Start() error {
	if o.stopped {
		return fmt.Errorf("observer already stopped: %w", ErrMonitorStopped)
	}

	// Implementation deferred to Phase 3
	return fmt.Errorf("PTY monitoring not implemented: %w", ErrObservationFailed)
}

// Stop halts PTY monitoring.
func (o *Observer) Stop() error {
	o.stopped = true
	return nil
}