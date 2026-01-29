// Package manager - pty.go provides PTY creation for managed sessions.
package manager

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/creack/pty"
)

// startPTY starts a command in a new PTY and returns the master file descriptor.
// The master FD is used for reading output and writing input.
func startPTY(cmd *exec.Cmd) (io.ReadWriteCloser, error) {
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("pty start: %w", err)
	}
	return ptmx, nil
}
