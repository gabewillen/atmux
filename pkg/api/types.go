package api

import (
	"time"

	"github.com/stateforward/hsm-go/muid"
)

// Agent represents a managed agent instance.
type Agent struct {
	ID       muid.MUID `json:"id"`
	Name     string    `json:"name"`
	About    string    `json:"about,omitempty"`
	Adapter  string    `json:"adapter"` // String reference to adapter name
	RepoPath string    `json:"repo_path,omitempty"`
	Location Location  `json:"location"`
	Presence string    `json:"presence"` // Online, Busy, Offline, Away
	Status   string    `json:"status"`   // Pending, Starting, Running, Terminated, Errored
}

// Location defines where the agent runs.
type Location struct {
	Type string `json:"type"` // "local", "ssh"
	Host string `json:"host,omitempty"`
}

// Session represents an active PTY session.
type Session struct {
	ID        muid.MUID `json:"id"`
	AgentID   muid.MUID `json:"agent_id"`
	HostID    string    `json:"host_id"` // Node ID where session runs
	StartedAt time.Time `json:"started_at"`
}
