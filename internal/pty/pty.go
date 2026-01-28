package pty

import (
	"fmt"
	"os"

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

// Close closes the PTY pair.
func (p *Pair) Close() error {
	if p == nil {
		return nil
	}
	var firstErr error
	if p.Master != nil {
		if err := p.Master.Close(); err != nil {
			firstErr = err
		}
	}
	if p.Slave != nil {
		if err := p.Slave.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
