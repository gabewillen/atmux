package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var (
	// ErrInvalidLocationType is returned for unknown location types.
	ErrInvalidLocationType = errors.New("invalid location type")
	// ErrInvalidLocation is returned when a location violates invariants.
	ErrInvalidLocation = errors.New("invalid location")
	// ErrInvalidAgent is returned when an agent violates invariants.
	ErrInvalidAgent = errors.New("invalid agent")
	// ErrInvalidSession is returned when a session violates invariants.
	ErrInvalidSession = errors.New("invalid session")
)

func validateRepoRoot(repoRoot string) error {
	if strings.TrimSpace(repoRoot) == "" {
		return fmt.Errorf("repo root: %w", ErrInvalidAgent)
	}
	if !filepath.IsAbs(repoRoot) {
		return fmt.Errorf("repo root: %w", ErrInvalidAgent)
	}
	return nil
}

func validateWorktree(repoRoot, worktree string) error {
	if strings.TrimSpace(worktree) == "" {
		return fmt.Errorf("worktree: %w", ErrInvalidAgent)
	}
	if !filepath.IsAbs(worktree) {
		return fmt.Errorf("worktree: %w", ErrInvalidAgent)
	}
	cleanRoot := filepath.Clean(repoRoot)
	cleanWorktree := filepath.Clean(worktree)
	rel, err := filepath.Rel(cleanRoot, cleanWorktree)
	if err != nil {
		return fmt.Errorf("worktree: %w", err)
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("worktree: %w", ErrInvalidAgent)
	}
	return nil
}

// LocationType describes where an agent runs.
type LocationType int

const (
	// LocationLocal represents a local agent.
	LocationLocal LocationType = iota
	// LocationSSH represents an agent running on a remote SSH host.
	LocationSSH
)

// String returns the string form of the location type.
func (t LocationType) String() string {
	switch t {
	case LocationLocal:
		return "local"
	case LocationSSH:
		return "ssh"
	default:
		return "unknown"
	}
}

// ParseLocationType parses a string into a location type.
func ParseLocationType(raw string) (LocationType, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "local":
		return LocationLocal, nil
	case "ssh":
		return LocationSSH, nil
	default:
		return LocationLocal, fmt.Errorf("parse location type: %w", ErrInvalidLocationType)
	}
}

// MarshalJSON encodes the location type as a string.
func (t LocationType) MarshalJSON() ([]byte, error) {
	if t != LocationLocal && t != LocationSSH {
		return nil, fmt.Errorf("marshal location type: %w", ErrInvalidLocationType)
	}
	return json.Marshal(t.String())
}

// UnmarshalJSON decodes a location type from a string.
func (t *LocationType) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("unmarshal location type: %w", err)
	}
	parsed, err := ParseLocationType(raw)
	if err != nil {
		return fmt.Errorf("unmarshal location type: %w", err)
	}
	*t = parsed
	return nil
}

// Location describes where an agent should run.
type Location struct {
	Type     LocationType `json:"type"`
	Host     string       `json:"host,omitempty"`
	User     string       `json:"user,omitempty"`
	Port     int          `json:"port,omitempty"`
	RepoPath string       `json:"repo_path,omitempty"`
}

// Validate checks location invariants.
func (l Location) Validate() error {
	switch l.Type {
	case LocationLocal:
		// repo_path is optional for local agents.
	case LocationSSH:
		if strings.TrimSpace(l.Host) == "" {
			return fmt.Errorf("location: %w", ErrInvalidLocation)
		}
		if strings.TrimSpace(l.RepoPath) == "" {
			return fmt.Errorf("location: %w", ErrInvalidLocation)
		}
	default:
		return fmt.Errorf("location: %w", ErrInvalidLocationType)
	}
	if l.Port < 0 || l.Port > 65535 {
		return fmt.Errorf("location: invalid port %d", l.Port)
	}
	return nil
}

// Agent describes the core metadata for a managed agent.
type Agent struct {
	ID       AgentID    `json:"agent_id"`
	Name     string     `json:"name"`
	About    string     `json:"about"`
	Adapter  AdapterRef `json:"adapter"`
	RepoRoot string     `json:"repo_root"`
	Worktree string     `json:"worktree"`
	Location Location   `json:"location"`
}

// NewAgent constructs a new agent with a fresh ID.
func NewAgent(name, about string, adapter AdapterRef, repoRoot, worktree string, location Location) (Agent, error) {
	return NewAgentWithID(NewAgentID(), name, about, adapter, repoRoot, worktree, location)
}

// NewAgentWithID constructs a new agent with the provided ID.
func NewAgentWithID(id AgentID, name, about string, adapter AdapterRef, repoRoot, worktree string, location Location) (Agent, error) {
	agent := Agent{
		ID:       id,
		Name:     name,
		About:    about,
		Adapter:  adapter,
		RepoRoot: repoRoot,
		Worktree: worktree,
		Location: location,
	}
	if err := agent.Validate(); err != nil {
		return Agent{}, err
	}
	return agent, nil
}

// Validate checks agent invariants.
func (a Agent) Validate() error {
	if a.ID.IsZero() {
		return fmt.Errorf("agent: %w", ErrZeroID)
	}
	if strings.TrimSpace(a.Name) == "" {
		return fmt.Errorf("agent: %w", ErrInvalidAgent)
	}
	if strings.TrimSpace(a.About) == "" {
		return fmt.Errorf("agent: %w", ErrInvalidAgent)
	}
	if strings.TrimSpace(string(a.Adapter)) == "" {
		return fmt.Errorf("agent: %w", ErrInvalidAgent)
	}
	if err := validateRepoRoot(a.RepoRoot); err != nil {
		return fmt.Errorf("agent: %w", err)
	}
	if err := validateWorktree(a.RepoRoot, a.Worktree); err != nil {
		return fmt.Errorf("agent: %w", err)
	}
	if err := a.Location.Validate(); err != nil {
		return fmt.Errorf("agent: %w", err)
	}
	return nil
}

// Session describes runtime session metadata for an agent.
type Session struct {
	ID       SessionID `json:"session_id"`
	AgentID  AgentID   `json:"agent_id"`
	RepoRoot string    `json:"repo_root"`
	Worktree string    `json:"worktree"`
	Location Location  `json:"location"`
}

// NewSession constructs a new session with a fresh ID.
func NewSession(agentID AgentID, repoRoot, worktree string, location Location) (Session, error) {
	return NewSessionWithID(NewSessionID(), agentID, repoRoot, worktree, location)
}

// NewSessionWithID constructs a new session with the provided ID.
func NewSessionWithID(id SessionID, agentID AgentID, repoRoot, worktree string, location Location) (Session, error) {
	session := Session{
		ID:       id,
		AgentID:  agentID,
		RepoRoot: repoRoot,
		Worktree: worktree,
		Location: location,
	}
	if err := session.Validate(); err != nil {
		return Session{}, err
	}
	return session, nil
}

// Validate checks session invariants.
func (s Session) Validate() error {
	if s.ID.IsZero() {
		return fmt.Errorf("session: %w", ErrZeroID)
	}
	if s.AgentID.IsZero() {
		return fmt.Errorf("session: %w", ErrZeroID)
	}
	if err := validateRepoRoot(s.RepoRoot); err != nil {
		return fmt.Errorf("session: %w", err)
	}
	if err := validateWorktree(s.RepoRoot, s.Worktree); err != nil {
		return fmt.Errorf("session: %w", err)
	}
	if err := s.Location.Validate(); err != nil {
		return fmt.Errorf("session: %w", err)
	}
	return nil
}
