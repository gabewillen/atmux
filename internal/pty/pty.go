package pty

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/creack/pty"
)

// Pair represents a PTY master/slave pair.
type Pair struct {
	Master *os.File
	Slave  *os.File
}

var (
	// ErrInvalidSize is returned when a PTY size is invalid.
	ErrInvalidSize = errors.New("invalid size")
	// ErrInvalidPTY is returned when a PTY file is missing.
	ErrInvalidPTY = errors.New("invalid pty")
)

// Open allocates a new PTY pair.
func Open() (*Pair, error) {
	master, slave, err := pty.Open()
	if err != nil {
		return nil, fmt.Errorf("pty open: %w", err)
	}
	return &Pair{Master: master, Slave: slave}, nil
}

// Start starts a command with a new PTY and returns the master.
func Start(cmd *exec.Cmd) (*os.File, error) {
	if cmd == nil {
		return nil, fmt.Errorf("pty start: command is nil")
	}
	master, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("pty start: %w", err)
	}
	return master, nil
}

// Resize sets the PTY window size for the provided master.
func Resize(master *os.File, rows, cols uint16) error {
	if master == nil {
		return fmt.Errorf("pty resize: %w", ErrInvalidPTY)
	}
	if rows == 0 || cols == 0 {
		return fmt.Errorf("pty resize: %w", ErrInvalidSize)
	}
	if err := pty.Setsize(master, &pty.Winsize{Rows: rows, Cols: cols}); err != nil {
		return fmt.Errorf("pty resize: %w", err)
	}
	return nil
}

// Close closes the PTY pair.
func (p *Pair) Close() error {
	if p == nil {
		return nil
	}
	var firstErr error
	if p.Master != nil {
		if err := p.Master.Close(); err != nil {
			firstErr = fmt.Errorf("pty close master: %w", err)
		}
	}
	if p.Slave != nil {
		if err := p.Slave.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("pty close slave: %w", err)
		}
	}
	return firstErr
}
