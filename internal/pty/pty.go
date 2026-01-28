// Package pty provides PTY management for amux.
package pty

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	amuxerrors "github.com/agentflare-ai/amux/internal/errors"
	"github.com/creack/pty"
)

// PTY represents a pseudo-terminal session.
type PTY struct {
	// File descriptor for the PTY master
	master *os.File

	// File descriptor for the PTY slave
	slave *os.File

	// Command running in the PTY
	cmd *exec.Cmd

	// PTY size
	size *pty.Winsize

	// Context for cancellation
	ctx context.Context

	// Cancel function
	cancel context.CancelFunc
}

// Config contains PTY configuration.
type Config struct {
	// Initial window size
	WindowSize *pty.Winsize `json:"window_size"`

	// Command to run
	Command string `json:"command"`

	// Arguments for command
	Args []string `json:"args"`

	// Environment variables
	Env []string `json:"env"`

	// Working directory
	WorkingDir string `json:"working_dir"`
}

// New creates a new PTY session.
func New(ctx context.Context, config *Config) (*PTY, error) {
	if config == nil {
		return nil, amuxerrors.Wrap("creating PTY", amuxerrors.ErrInvalidConfig)
	}

	if config.Command == "" {
		return nil, amuxerrors.Wrap("creating PTY", amuxerrors.ErrInvalidConfig)
	}

	// Create context with cancellation
	ptyCtx, cancel := context.WithCancel(ctx)

	// Create command
	cmd := exec.CommandContext(ptyCtx, config.Command, config.Args...)

	// Set environment
	if len(config.Env) > 0 {
		cmd.Env = config.Env
	}

	// Set working directory
	if config.WorkingDir != "" {
		cmd.Dir = config.WorkingDir
	}

	// Start PTY
	var err error
	var ptyFile *os.File

	if config.WindowSize != nil {
		ptyFile, err = pty.StartWithSize(cmd, config.WindowSize)
	} else {
		ptyFile, err = pty.Start(cmd)
	}

	if err != nil {
		cancel()
		return nil, amuxerrors.Wrap("starting PTY", err)
	}

	return &PTY{
		master: ptyFile,
		cmd:    cmd,
		ctx:    ptyCtx,
		cancel: cancel,
		size:   config.WindowSize,
	}, nil
}

// Read reads data from the PTY.
func (p *PTY) Read(b []byte) (int, error) {
	if p.master == nil {
		return 0, amuxerrors.Wrap("reading from PTY", amuxerrors.ErrNotReady)
	}

	return p.master.Read(b)
}

// Write writes data to the PTY.
func (p *PTY) Write(b []byte) (int, error) {
	if p.master == nil {
		return 0, amuxerrors.Wrap("writing to PTY", amuxerrors.ErrNotReady)
	}

	return p.master.Write(b)
}

// Close closes the PTY session.
func (p *PTY) Close() error {
	var lastErr error

	// Cancel context
	if p.cancel != nil {
		p.cancel()
	}

	// Close PTY
	if p.master != nil {
		if err := p.master.Close(); err != nil {
			lastErr = amuxerrors.Wrap("closing PTY master", err)
		}
	}

	// Wait for command to finish
	if p.cmd != nil && p.cmd.Process != nil {
		if err := p.cmd.Wait(); err != nil {
			lastErr = amuxerrors.Wrap("waiting for command", err)
		}
	}

	return lastErr
}

// SetSize sets the PTY window size.
func (p *PTY) SetSize(cols, rows uint16) error {
	if p.master == nil {
		return amuxerrors.Wrap("setting PTY size", amuxerrors.ErrNotReady)
	}

	size := &pty.Winsize{
		Cols: cols,
		Rows: rows,
	}

	err := pty.Setsize(p.master, size)
	if err != nil {
		return amuxerrors.Wrap("setting PTY size", err)
	}

	p.size = size
	return nil
}

// Size returns the current PTY size.
func (p *PTY) Size() *pty.Winsize {
	if p.size == nil {
		return &pty.Winsize{Cols: 80, Rows: 24} // Default
	}

	return p.size
}

// Process returns the underlying process.
func (p *PTY) Process() *os.Process {
	if p.cmd == nil {
		return nil
	}

	return p.cmd.Process
}

// Demo demonstrates PTY functionality for Phase 0.
func Demo(ctx context.Context) error {
	// Create a simple PTY demo that runs 'echo' command
	config := &Config{
		Command: "/bin/echo",
		Args:    []string{"Hello", "from", "PTY"},
		WindowSize: &pty.Winsize{
			Cols: 80,
			Rows: 24,
		},
	}

	pty, err := New(ctx, config)
	if err != nil {
		return amuxerrors.Wrap("creating demo PTY", err)
	}
	defer pty.Close()

	// Read from PTY with timeout
	buf := make([]byte, 1024)
	done := make(chan error, 1)

	go func() {
		_, err := pty.Read(buf)
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			return amuxerrors.Wrap("reading from demo PTY", err)
		}
		fmt.Printf("PTY Demo Output: %s\n", string(buf))
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return amuxerrors.New("PTY demo timeout")
	}

	return nil
}
