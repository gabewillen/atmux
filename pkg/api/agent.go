package api

import (
	"strings"
	"time"

	"github.com/stateforward/hsm-go/muid"
)

// Agent represents a coding agent instance managed by amux.
// Per spec §5.1, an agent consists of the required properties.
// The Adapter field is a string reference, not a typed dependency - the agent structure
// has no knowledge of specific adapter implementations.
type Agent struct {
	// ID is the runtime identifier for this agent instance.
	// Per spec §3.21, this is a globally unique identifier used for HSM identity,
	// event routing, and the remote protocol field `agent_id`.
	ID muid.MUID

	// Name is the human-readable name for this agent.
	Name string

	// About is a description of the agent's role or purpose.
	About string

	// Adapter is a string reference to the adapter name (e.g., "claude-code", "cursor").
	// Per spec §5.1, this is agent-agnostic - the adapter is loaded dynamically by name
	// through the WASM registry.
	Adapter string

	// RepoRoot is the canonical repository root path for this agent.
	// Per spec §5.1 and §3.23, this is the canonicalized absolute path.
	RepoRoot string

	// Worktree is the absolute path to the agent's working directory within RepoRoot.
	// Per spec §5.3.1, this is typically `.amux/worktrees/{agent_slug}/`.
	Worktree string

	// Location specifies where this agent runs (local or SSH).
	Location Location

	// Lifecycle and Presence are managed by HSMs (see §5.4, §6.1).
	// Query via: agent.Lifecycle.State(), agent.Presence.State()
	// These are not stored directly in the struct but managed by the hsm-go state machines.
}

// Location specifies where an agent runs.
// Per spec §5.1, agents can run locally or on a remote host via SSH.
type Location struct {
	// Type indicates whether this is a local or SSH location.
	Type LocationType

	// Host is the SSH host or alias from ~/.ssh/config.
	// Only used when Type is LocationSSH.
	Host string

	// User is the SSH user (optional if configured in ssh config).
	// Only used when Type is LocationSSH.
	User string

	// Port is the SSH port (optional if configured in ssh config).
	// Only used when Type is LocationSSH.
	Port int

	// RepoPath is the path to the git repository root on the target host.
	// Per spec §5.1:
	// - Required for SSH agents
	// - Optional for local agents to select a non-default repo
	RepoPath string
}

// LocationType indicates whether an agent runs locally or remotely.
type LocationType int

const (
	// LocationLocal indicates the agent runs on the same host as the director.
	LocationLocal LocationType = iota

	// LocationSSH indicates the agent runs on a remote host accessed via SSH.
	LocationSSH
)

// String returns the string representation of a LocationType.
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

// ParseLocationType parses a location type string (case-insensitive).
// Per spec §5.1, valid values are "local" and "ssh".
func ParseLocationType(s string) (LocationType, error) {
	switch s {
	case "local", "Local", "LOCAL":
		return LocationLocal, nil
	case "ssh", "SSH", "Ssh":
		return LocationSSH, nil
	default:
		return LocationLocal, ErrInvalidLocationType
	}
}

// Session represents a collection of agents managed together.
// Per spec §3.5, a session contains one or more agent PTYs.
type Session struct {
	// ID is the runtime identifier for this session.
	ID muid.MUID

	// Agents is the list of agents in this session.
	Agents []*Agent

	// CreatedAt is the timestamp when the session was created.
	// We'll add this in a later phase when we implement full session management.
}

// RosterEntry represents a single entry in the current roster for agent.list.
// Per spec §6.2 and §12.4.5, the roster MUST expose at least agent_id, name,
// adapter, presence, and repo_root.
type RosterEntry struct {
	AgentID  muid.MUID
	Name     string
	Adapter  string
	Presence string
	RepoRoot string
}

// AgentMessage represents a message between agents, host managers, or the director.
// Per spec §6.4, agents can communicate with each other using these messages.
type AgentMessage struct {
	// ID is the unique identifier for this message.
	ID muid.MUID

	// From is the sender runtime ID (set by publishing component).
	From muid.MUID

	// To is the recipient runtime ID (set by publishing component, or BroadcastID).
	To muid.MUID

	// ToSlug is the recipient token captured from text (typically agent_slug); case-insensitive.
	ToSlug string

	// Content is the message content.
	Content string

	// Timestamp is when the message was sent.
	// Per spec §6.4 and §9.1.3.1, timestamps are encoded as RFC3339 UTC strings in JSON.
	Timestamp time.Time
}

// ValidateAgent validates the core invariants for an Agent.
// It is intended for use by constructors and callers that assemble Agent
// values before handing them to the rest of the system.
func ValidateAgent(ag *Agent) error {
	if ag == nil {
		return ErrInvalidAgent
	}

	if ag.ID == BroadcastID {
		return ErrReservedID
	}

	if strings.TrimSpace(ag.Name) == "" {
		return ErrInvalidAgent
	}
	if strings.TrimSpace(ag.Adapter) == "" {
		return ErrInvalidAgent
	}
	if strings.TrimSpace(ag.RepoRoot) == "" {
		return ErrInvalidAgent
	}

	switch ag.Location.Type {
	case LocationLocal:
		// RepoPath is optional for local agents; RepoRoot is already required.
	case LocationSSH:
		if strings.TrimSpace(ag.Location.RepoPath) == "" {
			return ErrInvalidAgent
		}
	default:
		return ErrInvalidLocationType
	}

	return nil
}
