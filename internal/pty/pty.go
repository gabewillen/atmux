// Package pty provides PTY creation and lifecycle for amux (spec §4.2.4, §7).
// amux owns the PTY for each agent; the monitor observes raw output from the PTY.
package pty

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/creack/pty/v2"
)

// Session represents an owned PTY session for an agent (spec §7, B.5).
// The PTY is created and owned by amux; the slave side is used as the agent's terminal.
type Session struct {
	cmd    *exec.Cmd
	pty    *os.File
	mu     sync.Mutex
	closed bool
}

// NewSession starts a command in a new PTY with the given working directory and environment.
// The command runs with the PTY as its stdin/stdout/stderr. Caller must call Close when done.
func NewSession(workDir string, name string, args []string, env []string) (*Session, error) {
	if workDir == "" {
		return nil, fmt.Errorf("work dir is required")
	}
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("absolute work dir: %w", err)
	}
	cmd := exec.Command(name, args...)
	cmd.Dir = absWorkDir
	cmd.Env = env
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("pty start: %w", err)
	}
	return &Session{cmd: cmd, pty: ptmx}, nil
}

// PTY returns the master end of the PTY for reading output and writing input.
// Do not close it; use Session.Close to close the session.
func (s *Session) PTY() *os.File {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pty
}

// Read reads from the PTY master (agent output).
func (s *Session) Read(p []byte) (n int, err error) {
	f := s.PTY()
	if f == nil {
		return 0, os.ErrClosed
	}
	return f.Read(p)
}

// Write writes to the PTY master (agent input).
func (s *Session) Write(p []byte) (n int, err error) {
	f := s.PTY()
	if f == nil {
		return 0, os.ErrClosed
	}
	return f.Write(p)
}

// Resize sets the PTY window size (spec §4.2.4).
func (s *Session) Resize(rows, cols uint16) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pty == nil {
		return os.ErrClosed
	}
	return pty.Setsize(s.pty, &pty.Winsize{Rows: rows, Cols: cols})
}

// Close closes the PTY and waits for the command to exit.
// Idempotent; safe to call multiple times.
func (s *Session) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	ptmx := s.pty
	s.pty = nil
	s.mu.Unlock()

	if ptmx != nil {
		_ = ptmx.Close()
	}
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Signal(os.Interrupt)
		_ = s.cmd.Wait()
	}
	return nil
}

// Wait blocks until the command exits and returns its error.
func (s *Session) Wait() error {
	if s.cmd == nil {
		return nil
	}
	return s.cmd.Wait()
}

// Process returns the underlying process, or nil if not started.
func (s *Session) Process() *os.Process {
	if s.cmd == nil {
		return nil
	}
	return s.cmd.Process
}

// StdoutPipe is not used; PTY owns stdin/stdout/stderr. Exposed for tests that need raw stream.
// OutputStream returns a reader for the PTY master (agent output). Do not close the returned reader.
func (s *Session) OutputStream() io.Reader {
	return s.PTY()
}
