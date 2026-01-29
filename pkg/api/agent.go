package api

import (
	"time"
)

// Agent represents an active agent instance with configuration and runtime state.
type Agent struct {
	// ID is the globally unique runtime identifier for this agent.
	ID AgentID `json:"id"`

	// Name is the human-readable name of the agent.
	Name string `json:"name"`

	// Description is an optional description of the agent's purpose.
	Description string `json:"description,omitempty"`

	// Adapter is the string identifier of the WASM adapter to use.
	Adapter string `json:"adapter"`

	// AgentSlug is the filesystem-safe derived identifier.
	AgentSlug AgentSlug `json:"agent_slug"`

	// RepoRoot is the canonical absolute path to the git repository root.
	RepoRoot RepoRoot `json:"repo_root"`

	// WorkDir is the working directory for the agent (worktree path).
	WorkDir string `json:"work_dir"`

	// Command is the command and arguments to start the agent process.
	Command []string `json:"command"`

	// Environment is the environment variables for the agent process.
	Environment map[string]string `json:"environment,omitempty"`

	// CreatedAt is when the agent was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the agent was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// Session represents a running agent PTY session.
type Session struct {
	// ID is the globally unique runtime identifier for this session.
	ID SessionID `json:"id"`

	// AgentID is the ID of the agent that owns this session.
	AgentID AgentID `json:"agent_id"`

	// State is the current state of the session (pending, running, terminated, etc.).
	State SessionState `json:"state"`

	// PTYFile is the path to the PTY device file.
	PTYFile string `json:"pty_file,omitempty"`

	// PID is the process ID of the agent process.
	PID int `json:"pid,omitempty"`

	// ExitCode is the exit code if the process has terminated.
	ExitCode *int `json:"exit_code,omitempty"`

	// StartedAt is when the session was started.
	StartedAt *time.Time `json:"started_at,omitempty"`

	// FinishedAt is when the session finished.
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

// SessionState represents the current state of a session.
type SessionState string

const (
	// SessionStatePending is the initial state when a session is requested but not yet started.
	SessionStatePending SessionState = "pending"

	// SessionStateStarting is when the session is being started.
	SessionStateStarting SessionState = "starting"

	// SessionStateRunning is when the agent process is actively running.
	SessionStateRunning SessionState = "running"

	// SessionStateTerminated is when the session has completed normally.
	SessionStateTerminated SessionState = "terminated"

	// SessionStateErrored is when the session has failed with an error.
	SessionStateErrored SessionState = "errored"
)

// Presence represents the availability state of an agent.
type Presence string

const (
	// PresenceOnline indicates the agent is available and ready to accept tasks.
	PresenceOnline Presence = "online"

	// PresenceBusy indicates the agent is currently working on a task.
	PresenceBusy Presence = "busy"

	// PresenceOffline indicates the agent is not connected.
	PresenceOffline Presence = "offline"

	// PresenceAway indicates the agent is connected but temporarily unavailable.
	PresenceAway Presence = "away"
)

// PresenceInfo contains current presence information for an agent.
type PresenceInfo struct {
	// AgentID is the ID of the agent.
	AgentID AgentID `json:"agent_id"`

	// Presence is the current presence state.
	Presence Presence `json:"presence"`

	// LastActivity is when the agent was last active.
	LastActivity time.Time `json:"last_activity"`

	// Status is an optional status message.
	Status string `json:"status,omitempty"`

	// UpdatedAt is when this presence info was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// NewAgent creates a new Agent with the given parameters.
func NewAgent(name, description, adapter string, repoRoot RepoRoot, workDir string, command []string) *Agent {
	now := time.Now()
	return &Agent{
		ID:          NewID(),
		Name:        name,
		Description: description,
		Adapter:     adapter,
		AgentSlug:   NormalizeAgentSlug(name),
		RepoRoot:    repoRoot,
		WorkDir:     workDir,
		Command:     command,
		Environment: make(map[string]string),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// NewSession creates a new Session for an agent.
func NewSession(agentID AgentID) *Session {
	return &Session{
		ID:      NewID(),
		AgentID: agentID,
		State:   SessionStatePending,
	}
}

// NewPresenceInfo creates new presence information for an agent.
func NewPresenceInfo(agentID AgentID, presence Presence) *PresenceInfo {
	now := time.Now()
	return &PresenceInfo{
		AgentID:      agentID,
		Presence:     presence,
		LastActivity: now,
		UpdatedAt:    now,
	}
}

// IsValid checks if an Agent has the required fields set.
func (a *Agent) IsValid() bool {
	return a != nil &&
		IDIsValid(a.ID) &&
		a.Name != "" &&
		a.Adapter != "" &&
		a.AgentSlug != "" &&
		a.RepoRoot != "" &&
		len(a.Command) > 0
}

// IsRunning checks if the session is in a running state.
func (s *Session) IsRunning() bool {
	return s.State == SessionStateRunning
}

// IsFinished checks if the session has completed (terminated or errored).
func (s *Session) IsFinished() bool {
	return s.State == SessionStateTerminated || s.State == SessionStateErrored
}

// UpdatePresence updates the presence information with the current time.
func (p *PresenceInfo) UpdatePresence(presence Presence) {
	now := time.Now()
	p.Presence = presence
	p.LastActivity = now
	p.UpdatedAt = now
}

// UpdateStatus updates the status message and last activity time.
func (p *PresenceInfo) UpdateStatus(status string) {
	now := time.Now()
	p.Status = status
	p.LastActivity = now
	p.UpdatedAt = now
}
