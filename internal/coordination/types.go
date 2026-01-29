package coordination

import (
	"time"

	"github.com/agentflare-ai/amux/internal/process"
	"github.com/agentflare-ai/amux/pkg/api"
)

// Snapshot represents the state of the system at a point in time.
type Snapshot struct {
	Timestamp time.Time         `json:"timestamp"`
	AgentID   api.AgentID       `json:"agent_id"`
	TUI       string            `json:"tui_xml"` // XML representation
	Processes []process.Process `json:"processes"`
	// Additional context like recent logs or events
}

// Action represents a coordination action to be executed.
type Action struct {
	Type    string            `json:"type"`
	Target  api.AgentID       `json:"target"`
	Payload map[string]string `json:"payload"`
}

// Coordinator manages the loop.
type Coordinator interface {
	Start() error
	Stop() error
}

// Mode defines the coordination mode.
type Mode string

const (
	ModeAuto   Mode = "auto"
	ModeManual Mode = "manual"
)

// Config holds coordination settings.
type Config struct {
	Mode     Mode
	Interval time.Duration
}