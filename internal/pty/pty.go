package pty

import (
	"os"

	"github.com/creack/pty"
)

// Start assigns a pseudo-terminal to the command.
// Phase 0: Wrapper for creack/pty.Start.
func Start(cmd *os.File) (*os.File, error) {
	// This is just a placeholder signature matching usage pattern.
	// In reality creack/pty.Start takes *exec.Cmd.
	// We just want to pin the dependency.
	_ = pty.Winsize{}
	return nil, nil
}
