// Package api provides public API types for amux.
// types.go defines Agent, Session, Location, AgentMessage, and Roster types per spec §5.1, §5.5.9, §6.2–§6.4.
package api

import "time"

// Agent is the core data structure for an agent instance (spec §5.1).
// Lifecycle and presence are managed by HSMs in internal/agent; query via agent actor.
type Agent struct {
	ID       ID
	Name     string
	About    string
	Adapter  string // String reference to adapter name (agent-agnostic)
	RepoRoot string // Canonical repository root path (§3.23, §5.3.4)
	Worktree string // Absolute path to the agent's working directory within RepoRoot
	Location Location
}

// Location describes where an agent runs: local or SSH (spec §5.1).
type Location struct {
	Type     LocationType
	Host     string // SSH host or alias from ~/.ssh/config
	User     string // SSH user (optional if in ssh config)
	Port     int    // SSH port (optional if in ssh config)
	RepoPath string // Path to git repository root on target host
}

// LocationType is the agent location kind (spec §5.1).
type LocationType int

const (
	LocationLocal LocationType = iota
	LocationSSH
)

// Session holds session identity and reference to an agent (spec §5.5.9).
// PTY and replay buffer are owned by internal/agent; this type is the public/wire shape.
type Session struct {
	ID      ID
	AgentID ID
}

// AgentMessage is the inter-agent message payload (spec §6.4).
// From/To are set by the publishing component; ToSlug is the captured recipient token (e.g. agent_slug).
type AgentMessage struct {
	ID        ID
	From      ID
	To        ID
	ToSlug    string
	Content   string
	Timestamp time.Time // RFC 3339 UTC per spec §9.1.3.1
}

// RosterEntry is one participant in the roster (spec §6.2, §6.3).
// At minimum: agent_id, name, adapter, presence, repo_root (§12); About and CurrentTask for presence awareness.
type RosterEntry struct {
	AgentID     ID
	Name        string
	About       string
	Adapter     string
	Presence    string // HSM state name e.g. /agent.presence/online
	RepoRoot    string
	Kind        RosterKind
	CurrentTask string // Optional; for busy agents
}

// RosterKind is the participant type in the roster (spec §6.2).
type RosterKind int

const (
	RosterKindAgent   RosterKind = iota
	RosterKindManager
	RosterKindDirector
)

// Roster is the full list of participants (spec §6.2).
type Roster []RosterEntry
