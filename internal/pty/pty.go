// Package pty provides PTY (pseudo-terminal) management for amux.
//
// This package wraps creack/pty to provide PTY creation, I/O, and lifecycle
// management. All PTY operations are agent-agnostic.
//
// See spec §4.2.4 and §7 for PTY requirements.
package pty

import (
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"
)

// PTY represents a pseudo-terminal.
type PTY struct {
	mu      sync.Mutex
	file    *os.File
	cmd     *exec.Cmd
	size    *pty.Winsize
	closed  bool
}

// Open creates a new PTY for the given command.
func Open(cmd *exec.Cmd) (*PTY, error) {
	f, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	return &PTY{
		file: f,
		cmd:  cmd,
	}, nil
}

// Read reads from the PTY.
func (p *PTY) Read(buf []byte) (int, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return 0, io.EOF
	}
	f := p.file
	p.mu.Unlock()

	return f.Read(buf)
}

// Write writes to the PTY.
func (p *PTY) Write(data []byte) (int, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return 0, io.ErrClosedPipe
	}
	f := p.file
	p.mu.Unlock()

	return f.Write(data)
}

// Resize changes the PTY window size.
func (p *PTY) Resize(rows, cols uint16) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return io.ErrClosedPipe
	}

	size := &pty.Winsize{
		Rows: rows,
		Cols: cols,
	}

	if err := pty.Setsize(p.file, size); err != nil {
		return err
	}

	p.size = size
	return nil
}

// Close closes the PTY.
func (p *PTY) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true
	return p.file.Close()
}

// Wait waits for the PTY command to exit.
func (p *PTY) Wait() error {
	return p.cmd.Wait()
}

// File returns the underlying PTY file descriptor.
func (p *PTY) File() *os.File {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.file
}
