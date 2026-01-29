package session

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/monitor"
	"github.com/agentflare-ai/amux/internal/process"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/pty"
	"github.com/agentflare-ai/amux/pkg/api"
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
	dispatcher    protocol.Dispatcher
	monitor       *monitor.Monitor
	tracker       *process.Tracker
	ptyPair       *pty.Pair
	cmd           *exec.Cmd
	done          chan error
	stopRequested bool
	forcedKill    bool
	config        Config
	outputMu      sync.Mutex
	outputs       map[uint64]net.Conn
	nextOutputID  uint64
	writeMu       sync.Mutex
	observerMu    sync.Mutex
	observers     []func([]byte)
}

// NewLocalSession constructs a LocalSession for an agent.
func NewLocalSession(meta api.Session, runtime *agent.Agent, command Command, worktree string, matcher adapter.PatternMatcher, dispatcher protocol.Dispatcher, cfg Config) (*LocalSession, error) {
	if len(command.Argv) == 0 {
		return nil, fmt.Errorf("new session: %w", ErrSessionInvalid)
	}
	if worktree == "" {
		return nil, fmt.Errorf("new session: %w", ErrSessionInvalid)
	}
	if dispatcher == nil {
		return nil, fmt.Errorf("new session: %w", ErrSessionInvalid)
	}
	if cfg.DrainTimeout <= 0 {
		cfg.DrainTimeout = 30 * time.Second
	}
	if matcher == nil {
		matcher = &adapter.NoopMatcher{}
	}
	return &LocalSession{
		agent:      runtime,
		meta:       meta,
		command:    command,
		worktree:   worktree,
		dispatcher: dispatcher,
		monitor:    monitor.NewMonitor(matcher),
		tracker:    &process.Tracker{},
		config:     cfg,
		outputs:    make(map[uint64]net.Conn),
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
	if s.agent != nil {
		s.agent.Start(ctx)
		if s.tracker != nil {
			if err := s.tracker.Start(ctx, s.agent.ID.Value()); err != nil {
				_ = s.agent.EmitLifecycle(ctx, agent.EventError, err)
				s.mu.Unlock()
				return fmt.Errorf("session start: %w", err)
			}
		}
		_ = s.agent.EmitLifecycle(ctx, agent.EventStart, nil)
	}
	cmd := exec.CommandContext(ctx, s.command.Argv[0], s.command.Argv[1:]...)
	cmd.Dir = s.worktree
	cmd.Env = append(os.Environ(), s.command.Env...)
	master, err := pty.Start(cmd)
	if err != nil {
		s.mu.Unlock()
		if s.agent != nil {
			_ = s.agent.EmitLifecycle(ctx, agent.EventError, err)
		}
		return fmt.Errorf("session start: %w", err)
	}
	s.ptyPair = &pty.Pair{Master: master}
	s.cmd = cmd
	s.done = make(chan error, 1)
	s.stopRequested = false
	s.forcedKill = false
	go s.wait(ctx)
	go s.readOutput(ctx, master)
	s.mu.Unlock()
	if s.agent != nil {
		_ = s.agent.EmitLifecycle(ctx, agent.EventReady, nil)
	}
	return nil
}

// Attach returns a stream for interactive use.
func (s *LocalSession) Attach() (net.Conn, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ptyPair == nil || s.ptyPair.Master == nil {
		return nil, fmt.Errorf("session attach: %w", ErrSessionNotRunning)
	}
	local, remote := net.Pipe()
	id := atomic.AddUint64(&s.nextOutputID, 1)
	s.outputMu.Lock()
	s.outputs[id] = local
	s.outputMu.Unlock()
	go s.forwardInput(local, id)
	return remote, nil
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

// Done returns a channel that closes when the session exits.
func (s *LocalSession) Done() <-chan error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.done
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
		if s.agent != nil {
			_ = s.agent.EmitLifecycle(ctx, agent.EventError, err)
		}
		return
	}
	if s.agent != nil {
		_ = s.agent.EmitLifecycle(ctx, agent.EventStop, nil)
	}
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

func (s *LocalSession) readOutput(ctx context.Context, master *os.File) {
	if master == nil {
		return
	}
	buf := make([]byte, 4096)
	for {
		n, err := master.Read(buf)
		if n > 0 {
			chunk := append([]byte(nil), buf[:n]...)
			s.fanout(chunk)
			s.notifyObservers(chunk)
			s.handleOutput(ctx, chunk)
		}
		if err != nil {
			if err == io.EOF {
				return
			}
			return
		}
	}
}

func (s *LocalSession) fanout(chunk []byte) {
	s.outputMu.Lock()
	conns := make([]net.Conn, 0, len(s.outputs))
	for _, conn := range s.outputs {
		conns = append(conns, conn)
	}
	s.outputMu.Unlock()
	for _, conn := range conns {
		if _, err := conn.Write(chunk); err != nil {
			_ = conn.Close()
			s.removeOutput(conn)
		}
	}
}

// AddOutputObserver registers a callback for PTY output.
func (s *LocalSession) AddOutputObserver(observer func([]byte)) {
	if s == nil || observer == nil {
		return
	}
	s.observerMu.Lock()
	s.observers = append(s.observers, observer)
	s.observerMu.Unlock()
}

func (s *LocalSession) notifyObservers(chunk []byte) {
	s.observerMu.Lock()
	observers := append([]func([]byte){}, s.observers...)
	s.observerMu.Unlock()
	for _, observer := range observers {
		observer(chunk)
	}
}

func (s *LocalSession) removeOutput(target net.Conn) {
	s.outputMu.Lock()
	defer s.outputMu.Unlock()
	for id, conn := range s.outputs {
		if conn == target {
			delete(s.outputs, id)
			return
		}
	}
}

func (s *LocalSession) forwardInput(conn net.Conn, id uint64) {
	if conn == nil {
		return
	}
	defer func() {
		_ = conn.Close()
		s.outputMu.Lock()
		delete(s.outputs, id)
		s.outputMu.Unlock()
	}()
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if n > 0 {
			_ = s.Send(buf[:n])
		}
		if err != nil {
			return
		}
	}
}

// Send writes input bytes to the PTY.
func (s *LocalSession) Send(input []byte) error {
	if len(input) == 0 {
		return nil
	}
	s.mu.Lock()
	master := (*os.File)(nil)
	if s.ptyPair != nil {
		master = s.ptyPair.Master
	}
	s.mu.Unlock()
	if master == nil {
		return fmt.Errorf("session send: %w", ErrSessionNotRunning)
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if _, err := master.Write(input); err != nil {
		return fmt.Errorf("session send: %w", err)
	}
	return nil
}

func (s *LocalSession) handleOutput(ctx context.Context, chunk []byte) {
	if s.monitor == nil || len(chunk) == 0 {
		return
	}
	if s.agent != nil {
		_ = s.agent.EmitPresence(ctx, agent.EventActivity, nil)
	}
	matches, err := s.monitor.Scan(ctx, bytes.NewReader(chunk))
	if err != nil {
		return
	}
	for _, match := range matches {
		pattern := strings.ToLower(strings.TrimSpace(match.Pattern))
		switch pattern {
		case "prompt":
			if s.agent != nil {
				_ = s.agent.EmitPresence(ctx, agent.EventPromptDetected, match)
			}
		case "rate_limit":
			if s.agent != nil {
				_ = s.agent.EmitPresence(ctx, agent.EventRateLimit, match)
			}
		case "completion":
			if s.agent != nil {
				_ = s.agent.EmitPresence(ctx, agent.EventTaskCompleted, match)
			}
		case "message":
			if s.agent == nil {
				continue
			}
			var payload api.OutboundMessage
			if err := json.Unmarshal([]byte(match.Text), &payload); err != nil {
				continue
			}
			if payload.ToSlug == "" || payload.Content == "" {
				continue
			}
			agentID := s.agent.ID
			payload.AgentID = &agentID
			_ = s.dispatcher.Publish(ctx, protocol.Subject("events", "message"), protocol.Event{
				Name:       "message.outbound",
				Payload:    payload,
				OccurredAt: time.Now().UTC(),
			})
		}
	}
}
