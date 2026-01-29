package monitor

import (
	"context"
	"io"
	"time"

	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/pkg/api"
)

// Monitor observes PTY output.
type Monitor struct {
	AgentID api.AgentID
	Bus     *agent.EventBus
	Input   io.Reader
	
	// Configuration
	ActivityTimeout time.Duration
	CheckInterval   time.Duration
	
	// State
	lastActivity time.Time
}

// Hook allows injecting logic on data read (e.g. pattern matching).
type Hook func(data []byte)

// NewMonitor creates a new monitor.
func NewMonitor(agentID api.AgentID, bus *agent.EventBus, input io.Reader) *Monitor {
	return &Monitor{
		AgentID:         agentID,
		Bus:             bus,
		Input:           input,
		ActivityTimeout: 30 * time.Second, // Default
		CheckInterval:   5 * time.Second,  // Default
		lastActivity:    time.Now(),
	}
}

// Start runs the monitoring loop.
func (m *Monitor) Start(ctx context.Context, hooks ...Hook) {
	// Activity ticker
	interval := m.CheckInterval
	if interval == 0 {
		interval = 5 * time.Second
	}
	ticker := time.NewTicker(interval)

	// Read loop
	buf := make([]byte, 4096)
	go func() {
		defer ticker.Stop()
		for {
			n, err := m.Input.Read(buf)
			if n > 0 {
				m.lastActivity = time.Now()
				data := make([]byte, n)
				copy(data, buf[:n])
				
				// Publish activity event
				if m.Bus != nil {
					m.Bus.Publish(agent.BusEvent{
						Type:    agent.EventActivity,
						Source:  m.AgentID,
						Payload: "bytes",
					})
				}

				// Run hooks (e.g. pattern matching, TUI capture)
				for _, h := range hooks {
					h(data)
				}
			}
			if err != nil {
				if err != io.EOF {
					// Log error
				}
				return
			}
		}
	}()

	// Idle check loop
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if m.Bus == nil {
					continue
				}
				if time.Since(m.lastActivity) > m.ActivityTimeout {
					// Emit Idle
					m.Bus.Publish(agent.BusEvent{
						Type:   agent.EventPresenceUpdate,
						Source: m.AgentID,
						Payload: api.PresenceOnline, // Online == Idle/Ready
					})
				} else {
					// Recently active -> Busy
					m.Bus.Publish(agent.BusEvent{
						Type:   agent.EventPresenceUpdate,
						Source: m.AgentID,
						Payload: api.PresenceBusy,
					})
				}
			}
		}
	}()
}