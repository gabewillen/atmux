package agent

import (
	"github.com/agentflare-ai/amux/pkg/api"
)

// RosterEntry represents an agent in the roster.
type RosterEntry struct {
	AgentID  api.AgentID       `json:"agent_id"`
	Name     string            `json:"name"`
	Adapter  string            `json:"adapter"`
	Presence api.PresenceState `json:"presence"`
	RepoRoot api.RepoRoot      `json:"repo_root"`
}

// GetRoster generates the roster from the current configuration and runtime state.
// Since agents are currently managed in memory via specific instantiated Agent structs,
// we need a registry or manager to look them up.
// For Phase 2/4, we haven't implemented a central "AgentManager" struct yet, 
// but we have config.Config.
// However, the runtime state (HSM) is in the *Agent struct, not in config.
// So we need to traverse the active *Agent instances.
// We'll define a simple registry map here or assume the caller passes the active agents.

// AgentRegistry tracks active agents.
type AgentRegistry struct {
	Agents map[api.AgentID]*Agent
}

// NewRegistry creates a new registry.
func NewRegistry() *AgentRegistry {
	return &AgentRegistry{
		Agents: make(map[api.AgentID]*Agent),
	}
}

// Register adds an agent to the registry.
func (r *AgentRegistry) Register(a *Agent) {
	r.Agents[a.ID] = a
}

// GetRoster returns the list of agents.
func (r *AgentRegistry) GetRoster() []RosterEntry {
	roster := make([]RosterEntry, 0, len(r.Agents))
	for _, a := range r.Agents {
		entry := RosterEntry{
			AgentID:  a.ID,
			Name:     a.Name,
			Adapter:  a.Config.Adapter,
			Presence: a.GetPresence(),
			RepoRoot: a.RepoRoot,
		}
		roster = append(roster, entry)
	}
	return roster
}

// GetPresence returns the current presence state of the agent.
func (a *Agent) GetPresence() api.PresenceState {
	return a.CurrentPresence
}
