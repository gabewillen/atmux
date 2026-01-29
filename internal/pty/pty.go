package pty

import (
	"os"
	"os/exec"

	"github.com/creack/pty"

	"github.com/agentflare-ai/amux/internal/errors"
)

// Start starts a process in a PTY.
// If dir is empty, uses current directory.
func Start(command string, args []string, dir string) (*os.File, error) {
	c := exec.Command(command, args...)
	if dir != "" {
		c.Dir = dir
	}

	// Start the command with a PTY.
	ptmx, err := pty.Start(c)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to start pty for command %q", command)
	}

	return ptmx, nil
}

// Resize resizes the PTY.
func Resize(f *os.File, rows, cols int) error {
	w, h := cols, rows
	if err := pty.Setsize(f, &pty.Winsize{Rows: uint16(h), Cols: uint16(w)}); err != nil {
		return errors.Wrap(err, "failed to resize pty")
	}
	return nil
}

// Close closes the PTY file descriptor.
// Note: This often triggers SIGHUP to the process.
func Close(f *os.File) error {
	return f.Close()
}
