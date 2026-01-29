package pty

import (
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
