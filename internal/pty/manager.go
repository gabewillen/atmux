// Package pty provides PTY session management for agents.
// This file implements the owned PTY session model per spec requirements.
package pty

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/stateforward/hsm-go/muid"
	
	"github.com/copilot-claude-sonnet-4/amux/internal/ids"
)

// Additional errors for the session manager
var (
	ErrManagedSessionNotFound = errors.New("managed session not found")
	ErrManagedSessionClosed   = errors.New("managed session closed")
)

// SessionState represents the state of a PTY session.
type SessionState string

const (
	// SessionStateStarting indicates the session is starting up.
	SessionStateStarting SessionState = "starting"

	// SessionStateRunning indicates the session is active.
	SessionStateRunning SessionState = "running"

	// SessionStateTerminated indicates the session has ended normally.
	SessionStateTerminated SessionState = "terminated"

	// SessionStateErrored indicates the session ended with an error.
	SessionStateErrored SessionState = "errored"
)

// ManagedSession represents an owned PTY session for an agent.
type ManagedSession struct {
	ID          muid.MUID
	AgentID     muid.MUID
	Command     []string
	WorkingDir  string
	State       SessionState
	CreatedAt   time.Time
	StartedAt   *time.Time
	EndedAt     *time.Time
	ExitCode    *int

	// Internal fields
	pty       *os.File
	process   *os.Process
	cmd       *exec.Cmd
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex
	output    chan []byte
	closed    bool
	waitGroup sync.WaitGroup
}

// SessionManager manages multiple PTY sessions.
type SessionManager struct {
	sessions map[muid.MUID]*ManagedSession
	mu       sync.RWMutex
}

// NewSessionManager creates a new PTY session manager.
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[muid.MUID]*ManagedSession),
	}
}

// CreateSession creates a new PTY session for the given agent.
func (m *SessionManager) CreateSession(agentID muid.MUID, command []string, workingDir string) (*ManagedSession, error) {
	if len(command) == 0 {
		return nil, fmt.Errorf("command required")
	}

	if workingDir == "" {
		var err error
		workingDir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	// Verify working directory exists
	if _, err := os.Stat(workingDir); err != nil {
		return nil, fmt.Errorf("working directory %s does not exist: %w", workingDir, err)
	}

	sessionID := muid.MUID(ids.NewAgentID())
	ctx, cancel := context.WithCancel(context.Background())

	session := &ManagedSession{
		ID:         sessionID,
		AgentID:    agentID,
		Command:    command,
		WorkingDir: workingDir,
		State:      SessionStateStarting,
		CreatedAt:  time.Now().UTC(),
		ctx:        ctx,
		cancel:     cancel,
		output:     make(chan []byte, 1000), // Buffer output
	}

	m.mu.Lock()
	m.sessions[sessionID] = session
	m.mu.Unlock()

	return session, nil
}

// StartSession starts the PTY session and begins command execution.
func (m *SessionManager) StartSession(sessionID muid.MUID) error {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return err
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	if session.State != SessionStateStarting {
		return fmt.Errorf("session not in starting state: %s", session.State)
	}

	// Create the PTY
	ptyFile, ttyFile, err := pty.Open()
	if err != nil {
		session.State = SessionStateErrored
		return fmt.Errorf("failed to create PTY: %w", err)
	}

	// Set up the command
	cmd := exec.CommandContext(session.ctx, session.Command[0], session.Command[1:]...)
	cmd.Dir = session.WorkingDir
	cmd.Stdout = ttyFile
	cmd.Stderr = ttyFile
	cmd.Stdin = ttyFile

	// Set process group for proper signal handling
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		ptyFile.Close()
		ttyFile.Close()
		session.State = SessionStateErrored
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Update session state
	now := time.Now().UTC()
	session.pty = ptyFile
	session.process = cmd.Process
	session.cmd = cmd
	session.State = SessionStateRunning
	session.StartedAt = &now

	// Close TTY file in parent process
	ttyFile.Close()

	// Start output monitoring
	session.waitGroup.Add(2)
	go session.monitorOutput()
	go session.waitForCompletion()

	return nil
}

// GetSession returns a PTY session by ID.
func (m *SessionManager) GetSession(sessionID muid.MUID) (*ManagedSession, error) {
	m.mu.RLock()
	session, exists := m.sessions[sessionID]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrManagedSessionNotFound
	}

	return session, nil
}

// GetSessionByAgent returns the active PTY session for an agent.
func (m *SessionManager) GetSessionByAgent(agentID muid.MUID) (*ManagedSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, session := range m.sessions {
		if session.AgentID == agentID && session.State == SessionStateRunning {
			return session, nil
		}
	}

	return nil, ErrManagedSessionNotFound
}

// ListSessions returns all PTY sessions.
func (m *SessionManager) ListSessions() []*ManagedSession {
	m.mu.RLock()
	sessions := make([]*ManagedSession, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}
	m.mu.RUnlock()

	return sessions
}

// TerminateSession terminates a PTY session.
func (m *SessionManager) TerminateSession(sessionID muid.MUID) error {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return err
	}

	return session.Terminate()
}

// ManagedSession methods

// Write sends input to the PTY session.
func (s *ManagedSession) Write(data []byte) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.State != SessionStateRunning {
		return fmt.Errorf("session not running: %s", s.State)
	}

	if s.pty == nil {
		return ErrManagedSessionClosed
	}

	_, err := s.pty.Write(data)
	return err
}

// ReadOutput reads buffered output from the PTY session.
func (s *ManagedSession) ReadOutput() <-chan []byte {
	return s.output
}

// Terminate gracefully terminates the PTY session.
func (s *ManagedSession) Terminate() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.State == SessionStateTerminated || s.State == SessionStateErrored {
		return nil // Already terminated
	}

	// Cancel context to signal shutdown
	if s.cancel != nil {
		s.cancel()
	}

	// Send SIGTERM to process group
	if s.process != nil {
		if err := syscall.Kill(-s.process.Pid, syscall.SIGTERM); err != nil {
			// If SIGTERM fails, try SIGKILL
			syscall.Kill(-s.process.Pid, syscall.SIGKILL)
		}
	}

	// Wait for cleanup to complete with timeout
	done := make(chan bool, 1)
	go func() {
		s.waitGroup.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Normal cleanup completed
	case <-time.After(5 * time.Second):
		// Force cleanup if timeout
		if s.process != nil {
			syscall.Kill(-s.process.Pid, syscall.SIGKILL)
		}
	}

	return nil
}

// Kill forcefully kills the PTY session.
func (s *ManagedSession) Kill() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.State == SessionStateTerminated || s.State == SessionStateErrored {
		return nil // Already terminated
	}

	// Cancel context
	if s.cancel != nil {
		s.cancel()
	}

	// Send SIGKILL to process group
	if s.process != nil {
		syscall.Kill(-s.process.Pid, syscall.SIGKILL)
	}

	// Wait for cleanup to complete with timeout
	done := make(chan bool, 1)
	go func() {
		s.waitGroup.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Normal cleanup completed
	case <-time.After(2 * time.Second):
		// Force cleanup if timeout
	}

	return nil
}

// GetState returns the current session state safely.
func (s *ManagedSession) GetState() SessionState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.State
}

// GetInfo returns session information safely.
func (s *ManagedSession) GetInfo() SessionInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return SessionInfo{
		ID:         s.ID,
		AgentID:    s.AgentID,
		Command:    s.Command,
		WorkingDir: s.WorkingDir,
		State:      s.State,
		CreatedAt:  s.CreatedAt,
		StartedAt:  s.StartedAt,
		EndedAt:    s.EndedAt,
		ExitCode:   s.ExitCode,
	}
}

// SessionInfo provides read-only session information.
type SessionInfo struct {
	ID         muid.MUID
	AgentID    muid.MUID
	Command    []string
	WorkingDir string
	State      SessionState
	CreatedAt  time.Time
	StartedAt  *time.Time
	EndedAt    *time.Time
	ExitCode   *int
}

// Internal methods

// monitorOutput continuously reads from PTY and buffers output
func (s *ManagedSession) monitorOutput() {
	defer s.waitGroup.Done()

	buffer := make([]byte, 1024)
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			if s.pty == nil {
				return
			}

			// Set read timeout
			s.pty.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			n, err := s.pty.Read(buffer)

			if err != nil {
				if err == io.EOF {
					return // PTY closed
				}
				if os.IsTimeout(err) {
					continue // Timeout is expected
				}
				return // Other error
			}

			if n > 0 {
				// Copy data to avoid buffer reuse issues
				data := make([]byte, n)
				copy(data, buffer[:n])

				// Send to output channel (non-blocking)
				select {
				case s.output <- data:
				default:
					// Channel full, drop oldest data
					select {
					case <-s.output:
						s.output <- data
					default:
					}
				}
			}
		}
	}
}

// waitForCompletion waits for the process to complete and updates state
func (s *ManagedSession) waitForCompletion() {
	defer s.waitGroup.Done()

	if s.cmd == nil {
		return
	}

	// Wait for process to complete
	err := s.cmd.Wait()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Update session state
	now := time.Now().UTC()
	s.EndedAt = &now

	if err != nil {
		s.State = SessionStateErrored
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode := exitError.ExitCode()
			s.ExitCode = &exitCode
		} else {
			exitCode := -1
			s.ExitCode = &exitCode
		}
	} else {
		s.State = SessionStateTerminated
		exitCode := 0
		s.ExitCode = &exitCode
	}

	// Close PTY
	if s.pty != nil {
		s.pty.Close()
		s.pty = nil
	}

	// Close output channel
	if !s.closed {
		close(s.output)
		s.closed = true
	}
}