package agent

import (
	"os"
	"os/exec"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go"
)

// Agent represents the runtime state of an agent.
type Agent struct {
	ID       api.AgentID
	Slug     api.AgentSlug
	Name     string
	About    string
	Adapter  string
	RepoRoot api.RepoRoot
	Worktree string // Absolute path to the agent's working directory
	Config   config.AgentConfig

	// State machines
	Lifecycle hsm.Instance
	Presence  hsm.Instance
	
	// CurrentPresence tracks the current presence state (updated by HSM).
	CurrentPresence api.PresenceState

	// Sessions active for this agent
	Sessions map[api.SessionID]*Session
}

// Session represents a running session of an agent.
type Session struct {
	ID        api.SessionID
	AgentID   api.AgentID
	HostID    api.HostID
	StartedAt time.Time
	
	// Runtime
	Cmd *exec.Cmd
	PTY *os.File
}

// LifecycleState represents the lifecycle state of an agent.
type LifecycleState string

const (
	LifecyclePending    LifecycleState = "Pending"
	LifecycleStarting   LifecycleState = "Starting"
	LifecycleRunning    LifecycleState = "Running"
	LifecycleTerminated LifecycleState = "Terminated"
	LifecycleErrored    LifecycleState = "Errored"
)

// PresenceState represents the presence state of an agent.
type PresenceState string

const (
	PresenceOnline  PresenceState = "Online"
	PresenceBusy    PresenceState = "Busy"
	PresenceOffline PresenceState = "Offline"
	PresenceAway    PresenceState = "Away"
)
