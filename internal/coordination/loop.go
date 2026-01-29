package coordination

import (
	"context"
	"fmt"
	"time"

	"github.com/agentflare-ai/amux/internal/agent"
)

// ObservationLoop runs the coordination cycle.
type ObservationLoop struct {
	Interval time.Duration
	Agent    *agent.Agent // Focus on single agent for now
	cancel   func()
}

// NewObservationLoop creates a new loop.
func NewObservationLoop(agent *agent.Agent, interval time.Duration) *ObservationLoop {
	return &ObservationLoop{
		Interval: interval,
		Agent:    agent,
	}
}

// Start begins the loop.
func (l *ObservationLoop) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	l.cancel = cancel

	ticker := time.NewTicker(l.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := l.tick(ctx); err != nil {
				// Log error
				fmt.Printf("Coordination tick error: %v\n", err)
			}
		}
	}
}

// Stop stops the loop.
func (l *ObservationLoop) Stop() {
	if l.cancel != nil {
		l.cancel()
	}
}

func (l *ObservationLoop) tick(ctx context.Context) error {
	// 1. Capture Snapshot
	snap, err := l.captureSnapshot()
	if err != nil {
		return err
	}

	// 2. Inference / Adapter Matching (Placeholder)
	// actions := adapter.Match(snap)
	
	// 3. Execute Actions (Placeholder)
	// for _, action := range actions { execute(action) }
	
	_ = snap
	return nil
}

func (l *ObservationLoop) captureSnapshot() (*Snapshot, error) {
	// Gather state from Agent (TUI, Process)
	// For Phase 9, we assume single active session.
	var tuiXML string
	
	// Iterate sessions to find active one with monitor
	// Note: Session struct in Phase 5 doesn't have Monitor field visible yet?
	// It was added in Manager (RemoteSession). Local session needs it too?
	// Monitor is created in SpawnAgent? No, SpawnAgent just starts PTY.
	// We need to attach Monitor.
	
	// For now, use stub or "No Session"
	tuiXML = "<screen>waiting...</screen>"
	
	return &Snapshot{
		Timestamp: time.Now(),
		AgentID:   l.Agent.ID,
		TUI:       tuiXML,
	}, nil
}
