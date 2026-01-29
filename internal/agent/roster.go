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

// GetPresence returns the current presence state of the agent.
func (a *Agent) GetPresence() api.PresenceState {
	return a.CurrentPresence
}
