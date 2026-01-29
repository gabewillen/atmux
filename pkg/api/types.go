// Package api provides public types for the Agent Multiplexer (amux).
//
// This package contains the stable API types that may be imported by external
// packages. All types in this package are agent-agnostic; the Agent.Adapter
// field is a string reference to an adapter name, not a typed dependency.
package api

import (
	"strings"

	"github.com/stateforward/hsm-go/muid"
)

// Agent represents an active agent instance with a name, description,
// assigned adapter, and dedicated worktree.
//
// The Adapter field is a string reference, not a typed dependency.
// The agent structure has no knowledge of specific adapter implementations.
// The adapter is loaded dynamically by name through the WASM registry.
type Agent struct {
	// ID is the globally unique identifier assigned to this agent at runtime.
	// Used for HSM identity, event routing, and the remote protocol field agent_id.
	// Must be non-zero (0 is reserved per spec §3.22).
	ID muid.MUID

	// Name is the configured agent name.
	Name string

	// Slug is the filesystem-safe identifier derived from Name.
	// Used for worktree directory names and git branch names.
	// See spec §5.3.1 for normalization rules.
	Slug string

	// About is a description of the agent's purpose.
	About string

	// Adapter is a string reference to the adapter name (agent-agnostic).
	// Example values: "claude-code", "cursor", "windsurf"
	Adapter string

	// RepoRoot is the canonical repository root path for this agent.
	// See spec §3.23 for canonicalization rules.
	RepoRoot string

	// Worktree is the absolute path to the agent's working directory within RepoRoot.
	// Located at .amux/worktrees/{agent_slug}/.
	Worktree string

	// Location specifies where the agent runs (local or SSH).
	Location Location
}

// AgentValidationError represents an error validating an Agent.
type AgentValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *AgentValidationError) Error() string {
	return "invalid agent " + e.Field + ": " + e.Message
}

// Validate checks that the Agent meets all invariants:
//   - ID must be non-zero (spec §3.22)
//   - Name must be non-empty
//   - Slug must be non-empty
//   - Adapter must be non-empty
//   - RepoRoot must be non-empty
//
// Returns nil if valid, or an AgentValidationError describing the first violation.
func (a *Agent) Validate() error {
	if a.ID == 0 {
		return &AgentValidationError{Field: "ID", Message: "must be non-zero (0 is reserved)"}
	}
	if a.Name == "" {
		return &AgentValidationError{Field: "Name", Message: "must not be empty"}
	}
	if a.Slug == "" {
		return &AgentValidationError{Field: "Slug", Message: "must not be empty"}
	}
	if a.Adapter == "" {
		return &AgentValidationError{Field: "Adapter", Message: "must not be empty"}
	}
	if a.RepoRoot == "" {
		return &AgentValidationError{Field: "RepoRoot", Message: "must not be empty"}
	}
	if a.Location.Type == LocationSSH && a.Location.Host == "" {
		return &AgentValidationError{Field: "Location.Host", Message: "must not be empty for SSH agents"}
	}
	return nil
}

// Location specifies where an agent runs.
type Location struct {
	// Type is the location type (Local or SSH).
	Type LocationType

	// Host is the SSH host or alias from ~/.ssh/config.
	// Only used when Type is LocationSSH.
	Host string

	// User is the SSH user (optional if defined in ssh config).
	User string

	// Port is the SSH port (optional if defined in ssh config).
	Port int

	// RepoPath is the path to the git repository root on the target host.
	// Required for SSH agents; optional for local agents to select a non-default repo.
	RepoPath string
}

// LocationType indicates whether an agent runs locally or via SSH.
type LocationType int

const (
	// LocationLocal indicates the agent runs on the local machine.
	LocationLocal LocationType = iota

	// LocationSSH indicates the agent runs on a remote host via SSH.
	LocationSSH
)

// String returns the string representation of the location type.
func (lt LocationType) String() string {
	switch lt {
	case LocationLocal:
		return "local"
	case LocationSSH:
		return "ssh"
	default:
		return "unknown"
	}
}

// ParseLocationType parses a case-insensitive string into a LocationType.
// Returns LocationLocal for "local" and LocationSSH for "ssh".
// Returns an error for any other value.
func ParseLocationType(s string) (LocationType, error) {
	switch strings.ToLower(s) {
	case "local":
		return LocationLocal, nil
	case "ssh":
		return LocationSSH, nil
	default:
		return LocationLocal, &InvalidLocationTypeError{Value: s}
	}
}

// InvalidLocationTypeError is returned when parsing an invalid location type string.
type InvalidLocationTypeError struct {
	Value string
}

// Error implements the error interface.
func (e *InvalidLocationTypeError) Error() string {
	return "invalid location type: " + e.Value + " (expected \"local\" or \"ssh\")"
}

// LifecycleState represents the state of an agent's lifecycle.
// See spec §5.4 for the lifecycle state machine.
type LifecycleState string

const (
	// LifecyclePending indicates the agent is pending initialization.
	LifecyclePending LifecycleState = "pending"

	// LifecycleStarting indicates the agent is starting up.
	LifecycleStarting LifecycleState = "starting"

	// LifecycleRunning indicates the agent is running.
	LifecycleRunning LifecycleState = "running"

	// LifecycleTerminated indicates the agent has terminated normally.
	// This is a final state.
	LifecycleTerminated LifecycleState = "terminated"

	// LifecycleErrored indicates the agent has terminated with an error.
	// This is a final state.
	LifecycleErrored LifecycleState = "errored"
)

// IsFinal returns true if this is a terminal lifecycle state.
// Final states are Terminated and Errored.
func (s LifecycleState) IsFinal() bool {
	return s == LifecycleTerminated || s == LifecycleErrored
}

// IsValid returns true if this is a recognized lifecycle state.
func (s LifecycleState) IsValid() bool {
	switch s {
	case LifecyclePending, LifecycleStarting, LifecycleRunning, LifecycleTerminated, LifecycleErrored:
		return true
	default:
		return false
	}
}

// PresenceState represents the availability state of an agent.
// See spec §6.1 and §6.5 for presence states and transitions.
type PresenceState string

const (
	// PresenceOnline indicates the agent is available to accept tasks.
	PresenceOnline PresenceState = "online"

	// PresenceBusy indicates the agent is working on a task.
	PresenceBusy PresenceState = "busy"

	// PresenceOffline indicates the agent is rate-limited or temporarily unavailable.
	PresenceOffline PresenceState = "offline"

	// PresenceAway indicates the agent is connected but not responsive
	// (e.g., stuck, or remote connection lost).
	PresenceAway PresenceState = "away"
)

// CanAcceptTasks returns true if the agent can accept new tasks.
// Only Online agents can accept tasks.
func (s PresenceState) CanAcceptTasks() bool {
	return s == PresenceOnline
}

// IsValid returns true if this is a recognized presence state.
func (s PresenceState) IsValid() bool {
	switch s {
	case PresenceOnline, PresenceBusy, PresenceOffline, PresenceAway:
		return true
	default:
		return false
	}
}

// Session represents an amux session containing one or more agent PTYs.
type Session struct {
	// ID is the unique session identifier.
	// Must be non-zero (0 is reserved per spec §3.22).
	ID muid.MUID

	// Agents is the list of agent IDs in this session.
	// All agent IDs must be non-zero.
	Agents []muid.MUID
}

// SessionValidationError represents an error validating a Session.
type SessionValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *SessionValidationError) Error() string {
	return "invalid session " + e.Field + ": " + e.Message
}

// Validate checks that the Session meets all invariants:
//   - ID must be non-zero (spec §3.22)
//   - All agent IDs must be non-zero
//
// Returns nil if valid, or a SessionValidationError describing the first violation.
func (s *Session) Validate() error {
	if s.ID == 0 {
		return &SessionValidationError{Field: "ID", Message: "must be non-zero (0 is reserved)"}
	}
	for i, id := range s.Agents {
		if id == 0 {
			return &SessionValidationError{
				Field:   "Agents",
				Message: "agent ID at index " + itoa(i) + " must be non-zero",
			}
		}
	}
	return nil
}

// HasAgent returns true if the session contains the given agent ID.
func (s *Session) HasAgent(id muid.MUID) bool {
	for _, agentID := range s.Agents {
		if agentID == id {
			return true
		}
	}
	return false
}

// AddAgent adds an agent ID to the session if not already present.
// Returns true if the agent was added, false if already present.
func (s *Session) AddAgent(id muid.MUID) bool {
	if s.HasAgent(id) {
		return false
	}
	s.Agents = append(s.Agents, id)
	return true
}

// RemoveAgent removes an agent ID from the session.
// Returns true if the agent was removed, false if not present.
func (s *Session) RemoveAgent(id muid.MUID) bool {
	for i, agentID := range s.Agents {
		if agentID == id {
			s.Agents = append(s.Agents[:i], s.Agents[i+1:]...)
			return true
		}
	}
	return false
}

// itoa converts an integer to a string without importing strconv.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	if i < 0 {
		return "-" + itoa(-i)
	}
	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	return string(digits)
}

// RosterEntry represents an agent in the roster with presence information.
type RosterEntry struct {
	// Agent is the agent information.
	Agent Agent

	// Lifecycle is the current lifecycle state.
	Lifecycle LifecycleState

	// Presence is the current presence state.
	Presence PresenceState
}

// SpecVersion is the version of the specification this implementation targets.
const SpecVersion = "v1.22"
