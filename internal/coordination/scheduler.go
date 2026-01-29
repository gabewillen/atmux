package coordination

import (
	"context"

	"github.com/agentflare-ai/amux/internal/agent"
)

// Scheduler manages the periodic execution of the coordination loop.
type Scheduler struct {
	Config Config
	Agent  *agent.Agent
	Loop   *ObservationLoop
}

// NewScheduler creates a new scheduler.
func NewScheduler(cfg Config, agent *agent.Agent) *Scheduler {
	return &Scheduler{
		Config: cfg,
		Agent:  agent,
		Loop:   NewObservationLoop(agent, cfg.Interval),
	}
}

// Start starts the scheduler.
func (s *Scheduler) Start(ctx context.Context) {
	if s.Config.Mode == ModeManual {
		return
	}
	s.Loop.Start(ctx)
}

// Stop stops the scheduler.
func (s *Scheduler) Stop() {
	s.Loop.Stop()
}

// SetMode updates the coordination mode.
func (s *Scheduler) SetMode(mode Mode) {
	s.Config.Mode = mode
	if mode == ModeManual {
		s.Stop()
	} else {
		// Restart if not running?
		// Requires context management.
		// For simplicity, we assume Start is called at top level.
		// Dynamic restart is complex without a supervisor.
	}
}
