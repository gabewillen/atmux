package api

// RosterKind describes the roster entry type.
type RosterKind string

const (
	// RosterAgent represents a managed agent.
	RosterAgent RosterKind = "agent"
	// RosterManager represents a host manager participant.
	RosterManager RosterKind = "manager"
	// RosterDirector represents the director participant.
	RosterDirector RosterKind = "director"
)

// RosterEntry describes a roster entry for an agent or system participant.
type RosterEntry struct {
	Kind      RosterKind `json:"kind"`
	RuntimeID RuntimeID  `json:"runtime_id"`
	AgentID   *AgentID   `json:"agent_id,omitempty"`
	Name      string     `json:"name"`
	About     string     `json:"about,omitempty"`
	Adapter   AdapterRef `json:"adapter,omitempty"`
	RepoRoot  string     `json:"repo_root,omitempty"`
	Worktree  string     `json:"worktree,omitempty"`
	Slug      string     `json:"slug,omitempty"`
	Presence  string     `json:"presence"`
	Task      string     `json:"task,omitempty"`
	Location  *Location  `json:"location,omitempty"`
}
