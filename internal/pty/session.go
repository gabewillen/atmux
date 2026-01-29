// Package pty provides pseudo-terminal creation and I/O operations.
// This package handles raw PTY operations without any agent-specific logic.
//
// Uses creack/pty for cross-platform PTY management on Linux and macOS
// with non-blocking I/O via standard Go interfaces.
package pty

import (
	"errors"
	"fmt"
	"os"

	"github.com/creack/pty"
)

// Common sentinel errors for PTY operations.
var (
	// ErrPTYCreateFailed indicates PTY creation failed.
	ErrPTYCreateFailed = errors.New("PTY creation failed")

	// ErrInvalidSize indicates an invalid PTY window size.
	ErrInvalidSize = errors.New("invalid PTY size")

	// ErrPTYClosed indicates the PTY has been closed.
	ErrPTYClosed = errors.New("PTY closed")
)

// Session represents a PTY session with master and slave file descriptors.
type Session struct {
	master *os.File
	slave  *os.File
	size   *pty.Winsize
}

// NewSession creates a new PTY session.
func NewSession() (*Session, error) {
	master, slave, err := pty.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open PTY: %w", ErrPTYCreateFailed)
	}

	return &Session{
		master: master,
		slave:  slave,
		size:   &pty.Winsize{Rows: 24, Cols: 80},
	}, nil
}

// SetSize updates the PTY window size.
func (s *Session) SetSize(rows, cols uint16) error {
	if rows == 0 || cols == 0 {
		return fmt.Errorf("invalid dimensions %dx%d: %w", rows, cols, ErrInvalidSize)
	}

	s.size = &pty.Winsize{Rows: rows, Cols: cols}
	if err := pty.Setsize(s.master, s.size); err != nil {
		return fmt.Errorf("failed to set PTY size: %w", err)
	}
	return nil
}

// Master returns the master file descriptor for writing to the PTY.
func (s *Session) Master() *os.File {
	return s.master
}

// Slave returns the slave file descriptor for the child process.
func (s *Session) Slave() *os.File {
	return s.slave
}

// Close closes both master and slave file descriptors.
func (s *Session) Close() error {
	var lastErr error
	if s.slave != nil {
		if err := s.slave.Close(); err != nil {
			lastErr = fmt.Errorf("failed to close PTY slave: %w", err)
		}
		s.slave = nil
	}
	if s.master != nil {
		if err := s.master.Close(); err != nil {
			lastErr = fmt.Errorf("failed to close PTY master: %w", lastErr)
		}
		s.master = nil
	}
	if lastErr != nil {
		return fmt.Errorf("PTY close error: %w", ErrPTYClosed)
	}
	return nil
}

// SlaveName returns the name/path of the slave PTY device.
func (s *Session) SlaveName() string {
	if s.slave == nil {
		return ""
	}
	return s.slave.Name()
}