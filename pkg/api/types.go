// Package api provides public API types for amux.
// types.go defines Agent, Session, and Location per spec §5.1, §5.5.9.
package api

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
