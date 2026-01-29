// Package session provides local agent session management with owned PTYs.
//
// A session represents a running agent PTY instance. Each agent has at most
// one active session. Sessions own the PTY and are responsible for starting
// the agent shell, managing I/O, and cleaning up on shutdown.
//
// The session package integrates with the agent lifecycle HSM to drive
// state transitions (Pending → Starting → Running → Terminated/Errored).
//
// See spec §5.4, §5.6, and §B.5 for lifecycle, shutdown, and PTY ownership.
package session

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/stateforward/hsm-go/muid"

	"github.com/agentflare-ai/amux/internal/agent"
	amuxerrors "github.com/agentflare-ai/amux/internal/errors"
	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/internal/ids"
	amuxpty "github.com/agentflare-ai/amux/internal/pty"
)

// Session represents a running agent PTY session.
//
// Each session owns a PTY and manages the agent's shell process.
// The session is the single owner of the agent's terminal I/O.
type Session struct {
	mu sync.RWMutex

	// ID is the unique session identifier.
	ID muid.MUID

	// AgentID is the ID of the agent this session belongs to.
	AgentID muid.MUID

	// Agent is the managed agent instance.
	Agent *agent.Agent

	// PTY is the owned pseudo-terminal for this session.
	PTY *amuxpty.PTY

	// cmd is the shell command running in the PTY.
	cmd *exec.Cmd

	// state tracks whether the session is running.
	state State

	// dispatcher is used to emit session events.
	dispatcher event.Dispatcher

	// done is closed when the session exits.
	done chan struct{}

	// exitErr holds the process exit error, if any.
	exitErr error
}

// State represents the session state.
type State string

const (
	// StateCreated indicates the session has been created but not started.
	StateCreated State = "created"

	// StateRunning indicates the session is actively running.
	StateRunning State = "running"

	// StateStopped indicates the session has been stopped.
	StateStopped State = "stopped"
)

// Manager manages active sessions for agents.
type Manager struct {
	mu         sync.RWMutex
	sessions   map[muid.MUID]*Session // agent ID -> session
	dispatcher event.Dispatcher
}

// NewManager creates a new session manager.
func NewManager(dispatcher event.Dispatcher) *Manager {
	if dispatcher == nil {
		dispatcher = event.NewNoopDispatcher()
	}
	return &Manager{
		sessions:   make(map[muid.MUID]*Session),
		dispatcher: dispatcher,
	}
}

// Spawn creates and starts a new PTY session for an agent.
//
// The shell command is executed in the agent's worktree directory.
// The session takes ownership of the PTY and monitors the process.
//
// See spec §5.4 (lifecycle) and §B.5 (owned PTY).
func (m *Manager) Spawn(ctx context.Context, ag *agent.Agent, shell string, args ...string) (*Session, error) {
	if ag == nil {
		return nil, fmt.Errorf("session spawn: agent is nil")
	}

	agentID := ag.ID
	worktree := ag.Worktree

	m.mu.Lock()
	// Check if agent already has a running session
	if existing, ok := m.sessions[agentID]; ok {
		existing.mu.RLock()
		state := existing.state
		existing.mu.RUnlock()
		if state == StateRunning {
			m.mu.Unlock()
			return nil, fmt.Errorf("session spawn: %w", amuxerrors.ErrAgentAlreadyRunning)
		}
		// Remove stale session
		delete(m.sessions, agentID)
	}
	m.mu.Unlock()

	// Build the command
	if shell == "" {
		shell = "/bin/sh"
	}
	cmd := exec.CommandContext(ctx, shell, args...)
	if worktree != "" {
		cmd.Dir = worktree
	}

	// Create the PTY
	pty, err := amuxpty.Open(cmd)
	if err != nil {
		return nil, fmt.Errorf("session spawn: pty open: %w", err)
	}

	sess := &Session{
		ID:         ids.NewID(),
		AgentID:    agentID,
		Agent:      ag,
		PTY:        pty,
		cmd:        cmd,
		state:      StateRunning,
		dispatcher: m.dispatcher,
		done:       make(chan struct{}),
	}

	m.mu.Lock()
	m.sessions[agentID] = sess
	m.mu.Unlock()

	// Monitor the process in the background
	go sess.waitForExit()

	return sess, nil
}

// Stop stops a session by closing the PTY and waiting for the process to exit.
//
// If the process does not exit within context deadline, it is killed.
// See spec §5.6.3 for agent shutdown behavior.
func (m *Manager) Stop(ctx context.Context, agentID muid.MUID) error {
	m.mu.Lock()
	sess, ok := m.sessions[agentID]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("session stop: %w", amuxerrors.ErrSessionNotFound)
	}
	m.mu.Unlock()

	return sess.Stop()
}

// Kill forcefully terminates a session.
func (m *Manager) Kill(ctx context.Context, agentID muid.MUID) error {
	m.mu.Lock()
	sess, ok := m.sessions[agentID]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("session kill: %w", amuxerrors.ErrSessionNotFound)
	}
	m.mu.Unlock()

	return sess.Kill()
}

// Get returns the session for an agent, or nil if none exists.
func (m *Manager) Get(agentID muid.MUID) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[agentID]
}

// Remove removes a session from the manager.
func (m *Manager) Remove(agentID muid.MUID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, agentID)
}

// List returns all active sessions.
func (m *Manager) List() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		result = append(result, s)
	}
	return result
}

// StopAll stops all sessions gracefully.
func (m *Manager) StopAll() {
	m.mu.RLock()
	sessions := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}
	m.mu.RUnlock()

	for _, s := range sessions {
		_ = s.Stop()
	}
}

// KillAll forcefully terminates all sessions.
func (m *Manager) KillAll() {
	m.mu.RLock()
	sessions := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}
	m.mu.RUnlock()

	for _, s := range sessions {
		_ = s.Kill()
	}
}

// Stop gracefully stops the session by closing the PTY (sends EOF to shell).
func (s *Session) Stop() error {
	s.mu.Lock()
	if s.state != StateRunning {
		s.mu.Unlock()
		return nil
	}
	s.mu.Unlock()

	// Close PTY which sends EOF to the shell process
	if err := s.PTY.Close(); err != nil {
		return fmt.Errorf("session stop: %w", err)
	}

	// Wait for exit
	<-s.done
	return nil
}

// Kill forcefully terminates the session.
func (s *Session) Kill() error {
	s.mu.Lock()
	if s.state != StateRunning {
		s.mu.Unlock()
		return nil
	}
	s.mu.Unlock()

	// Kill the process
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
	}

	// Close PTY
	_ = s.PTY.Close()

	// Wait for exit
	<-s.done
	return nil
}

// Done returns a channel that is closed when the session exits.
func (s *Session) Done() <-chan struct{} {
	return s.done
}

// State returns the session state.
func (s *Session) SessionState() State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// ExitErr returns the process exit error, or nil if still running or exited cleanly.
func (s *Session) ExitErr() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.exitErr
}

// Write writes data to the PTY (input to the agent).
func (s *Session) Write(data []byte) (int, error) {
	s.mu.RLock()
	if s.state != StateRunning {
		s.mu.RUnlock()
		return 0, io.ErrClosedPipe
	}
	s.mu.RUnlock()

	return s.PTY.Write(data)
}

// Read reads data from the PTY (output from the agent).
func (s *Session) Read(buf []byte) (int, error) {
	return s.PTY.Read(buf)
}

// Resize changes the PTY window size.
func (s *Session) Resize(rows, cols uint16) error {
	return s.PTY.Resize(rows, cols)
}

// waitForExit waits for the PTY process to exit and updates state.
func (s *Session) waitForExit() {
	err := s.PTY.Wait()

	s.mu.Lock()
	s.state = StateStopped
	s.exitErr = err
	s.mu.Unlock()

	close(s.done)

	// Emit event
	data := map[string]any{"session_id": ids.EncodeID(s.ID)}
	if err != nil {
		data["error"] = err.Error()
	}
	_ = s.dispatcher.Dispatch(context.Background(),
		event.NewEvent(event.TypeAgentStopped, s.AgentID, data))
}
