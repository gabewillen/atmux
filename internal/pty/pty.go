package pty

import (
	"os"
	"os/exec"

	"github.com/creack/pty"
)

// PTY wraps a pseudo-terminal file and its associated command.
type PTY struct {
	File *os.File
	Cmd  *exec.Cmd
}

// Start starts a command in a new PTY.
func Start(cmd *exec.Cmd) (*PTY, error) {
	f, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}
	return &PTY{
		File: f,
		Cmd:  cmd,
	}, nil
}

// Resize resizes the PTY window.
func (p *PTY) Resize(rows, cols uint16) error {
	return pty.Setsize(p.File, &pty.Winsize{
		Rows: rows,
		Cols: cols,
	})
}

// Close closes the PTY file.
func (p *PTY) Close() error {
	return p.File.Close()
}