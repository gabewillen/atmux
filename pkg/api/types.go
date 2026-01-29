package api

import (
	"os"

	"github.com/stateforward/hsm-go/muid"
)

// Presence represents the availability state of an agent.
type Presence string

const (
	PresenceOnline  Presence = "online"
	PresenceBusy    Presence = "busy"
	PresenceOffline Presence = "offline"
	PresenceAway    Presence = "away"
)

// AgentState represents the lifecycle state of an agent.
type AgentState string

const (
	StatePending    AgentState = "pending"
	StateStarting   AgentState = "starting"
	StateRunning    AgentState = "running"
	StateTerminated AgentState = "terminated"
	StateErrored    AgentState = "errored"
)

// Agent represents a managed agent instance.
// This struct exposes public state; internal logic resides in the actor.
type Agent struct {
	ID       muid.MUID  `json:"id"`
	Slug     AgentSlug  `json:"slug"`
	Name     string     `json:"name"`
	Adapter  string     `json:"adapter"` // Name of the adapter (e.g., "claude-code")
	State    AgentState `json:"state"`
	Presence Presence   `json:"presence"`
	RepoRoot string     `json:"repo_root"` // Canonical absolute path
	Worktree string     `json:"worktree"`  // Absolute path to worktree
}

// Session represents an active PTY session for an agent.
type Session struct {
	ID      muid.MUID `json:"id"`
	AgentID muid.MUID `json:"agent_id"`
	PTY     *os.File  `json:"-"` // PTY file descriptor (not serialized)
}

// RosterEntry represents a simplified view of an agent for listing.
type RosterEntry struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"` // Combined state/presence summary
	Location string `json:"location"`
}
