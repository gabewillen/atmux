package session

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/pty"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go"
)

var (
	// ErrSessionRunning is returned when a session is already running.
	ErrSessionRunning = errors.New("session already running")
	// ErrSessionNotRunning is returned when a session is not running.
	ErrSessionNotRunning = errors.New("session not running")
	// ErrSessionInvalid is returned when session configuration is invalid.
	ErrSessionInvalid = errors.New("session invalid")
)

// Command describes the command used to start an agent.
type Command struct {
	// Argv is the command argv.
	Argv []string
	// Env holds additional environment variables.
	Env []string
}

// Config configures session behavior.
type Config struct {
	// DrainTimeout controls graceful shutdown duration.
	DrainTimeout time.Duration
}

// LocalSession owns a PTY and process for a local agent.
type LocalSession struct {
	mu            sync.Mutex
	agent         *agent.Agent
	meta          api.Session
	command       Command
	worktree      string
	ptyPair       *pty.Pair
	cmd           *exec.Cmd
	done          chan error
	stopRequested bool
	forcedKill    bool
	config        Config
}

// NewLocalSession constructs a LocalSession for an agent.
func NewLocalSession(meta api.Session, runtime *agent.Agent, command Command, worktree string, cfg Config) (*LocalSession, error) {
	if runtime == nil {
		return nil, fmt.Errorf("new session: %w", ErrSessionInvalid)
	}
	if len(command.Argv) == 0 {
		return nil, fmt.Errorf("new session: %w", ErrSessionInvalid)
	}
	if worktree == "" {
		return nil, fmt.Errorf("new session: %w", ErrSessionInvalid)
	}
	if cfg.DrainTimeout <= 0 {
		cfg.DrainTimeout = 30 * time.Second
	}
	return &LocalSession{
		agent:    runtime,
		meta:     meta,
		command:  command,
		worktree: worktree,
		config:   cfg,
	}, nil
}

// Start launches the PTY session.
func (s *LocalSession) Start(ctx context.Context) error {
	if ctx.Err() != nil {
		return fmt.Errorf("session start: %w", ctx.Err())
	}
	s.mu.Lock()
	if s.cmd != nil {
		s.mu.Unlock()
		return fmt.Errorf("session start: %w", ErrSessionRunning)
	}
	s.agent.Start(ctx)
	hsm.Dispatch(ctx, s.agent.Lifecycle, hsm.Event{Name: agent.EventStart})
	cmd := exec.CommandContext(ctx, s.command.Argv[0], s.command.Argv[1:]...)
	cmd.Dir = s.worktree
	cmd.Env = append(os.Environ(), s.command.Env...)
	master, err := pty.Start(cmd)
	if err != nil {
		s.mu.Unlock()
		hsm.Dispatch(ctx, s.agent.Lifecycle, hsm.Event{Name: agent.EventError, Data: err})
		return fmt.Errorf("session start: %w", err)
	}
	s.ptyPair = &pty.Pair{Master: master}
	s.cmd = cmd
	s.done = make(chan error, 1)
	s.stopRequested = false
	s.forcedKill = false
	go s.wait(ctx)
	s.mu.Unlock()
	hsm.Dispatch(ctx, s.agent.Lifecycle, hsm.Event{Name: agent.EventReady})
	return nil
}

// Attach returns a duplicate of the PTY master for interactive use.
func (s *LocalSession) Attach() (*os.File, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ptyPair == nil || s.ptyPair.Master == nil {
		return nil, fmt.Errorf("session attach: %w", ErrSessionNotRunning)
	}
	dup, err := dupFile(s.ptyPair.Master)
	if err != nil {
		return nil, fmt.Errorf("session attach: %w", err)
	}
	return dup, nil
}

// Stop requests graceful termination of the session.
func (s *LocalSession) Stop(ctx context.Context) error {
	if ctx.Err() != nil {
		return fmt.Errorf("session stop: %w", ctx.Err())
	}
	s.mu.Lock()
	cmd := s.cmd
	if cmd == nil {
		s.mu.Unlock()
		return fmt.Errorf("session stop: %w", ErrSessionNotRunning)
	}
	s.stopRequested = true
	s.mu.Unlock()
	if err := sendTerminate(cmd.Process); err != nil {
		return fmt.Errorf("session stop: %w", err)
	}
	return s.waitForExit(ctx, true)
}

// Kill forces session termination.
func (s *LocalSession) Kill(ctx context.Context) error {
	if ctx.Err() != nil {
		return fmt.Errorf("session kill: %w", ctx.Err())
	}
	s.mu.Lock()
	cmd := s.cmd
	if cmd == nil {
		s.mu.Unlock()
		return fmt.Errorf("session kill: %w", ErrSessionNotRunning)
	}
	s.stopRequested = true
	s.forcedKill = true
	s.mu.Unlock()
	if err := cmd.Process.Kill(); err != nil {
		return fmt.Errorf("session kill: %w", err)
	}
	return s.waitForExit(ctx, true)
}

// Restart stops and starts the session.
func (s *LocalSession) Restart(ctx context.Context) error {
	if err := s.Stop(ctx); err != nil {
		return fmt.Errorf("session restart: %w", err)
	}
	if err := s.Start(ctx); err != nil {
		return fmt.Errorf("session restart: %w", err)
	}
	return nil
}

// Meta returns the session metadata.
func (s *LocalSession) Meta() api.Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.meta
}

func (s *LocalSession) wait(ctx context.Context) {
	err := s.cmd.Wait()
	s.mu.Lock()
	s.cmd = nil
	s.done <- err
	close(s.done)
	pair := s.ptyPair
	s.ptyPair = nil
	stopRequested := s.stopRequested
	forcedKill := s.forcedKill
	s.stopRequested = false
	s.forcedKill = false
	s.mu.Unlock()
	if pair != nil {
		if closeErr := pair.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	if err != nil && !stopRequested || forcedKill {
		hsm.Dispatch(ctx, s.agent.Lifecycle, hsm.Event{Name: agent.EventError, Data: err})
		return
	}
	hsm.Dispatch(ctx, s.agent.Lifecycle, hsm.Event{Name: agent.EventStop})
}

func (s *LocalSession) waitForExit(ctx context.Context, allowExitError bool) error {
	s.mu.Lock()
	done := s.done
	timeout := s.config.DrainTimeout
	s.mu.Unlock()
	if done == nil {
		return fmt.Errorf("session wait: %w", ErrSessionNotRunning)
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	select {
	case err := <-done:
		if err != nil && !allowExitError {
			return fmt.Errorf("session wait: %w", err)
		}
		return nil
	case <-time.After(timeout):
		s.mu.Lock()
		if s.cmd != nil && s.cmd.Process != nil {
			s.forcedKill = true
			_ = s.cmd.Process.Kill()
		}
		s.mu.Unlock()
		select {
		case err := <-done:
			if err != nil && !allowExitError {
				return fmt.Errorf("session wait: %w", err)
			}
			return nil
		case <-ctx.Done():
			return fmt.Errorf("session wait: %w", ctx.Err())
		}
	case <-ctx.Done():
		return fmt.Errorf("session wait: %w", ctx.Err())
	}
}
