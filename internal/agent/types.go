package agent

import (
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
	RepoRoot api.RepoRoot
	Config   config.AgentConfig

	// State machines
	Lifecycle hsm.Instance
	Presence  hsm.Instance

	// Sessions active for this agent
	Sessions map[api.SessionID]*Session
}

// Session represents a running session of an agent.
type Session struct {
	ID        api.SessionID
	AgentID   api.AgentID
	HostID    api.HostID
	StartedAt time.Time
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
