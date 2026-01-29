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
	ActivityTimeout time.Duration // idle timeout
	StuckTimeout    time.Duration
	CheckInterval   time.Duration
	
	// State
	lastActivity time.Time
	isStuck      bool
	isIdle       bool
}

// Hook allows injecting logic on data read (e.g. pattern matching).
type Hook func(data []byte)

// NewMonitor creates a new monitor.
func NewMonitor(agentID api.AgentID, bus *agent.EventBus, input io.Reader) *Monitor {
	return &Monitor{
		AgentID:         agentID,
		Bus:             bus,
		Input:           input,
		ActivityTimeout: 30 * time.Second, // Default per Spec §7.5
		StuckTimeout:    5 * time.Minute,  // Default per Spec §7.5
		CheckInterval:   5 * time.Second,
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
				
				// Reset states
				if m.isStuck || m.isIdle {
					m.isStuck = false
					m.isIdle = false
					// Emit activity detected (transitions Away/Idle -> Online/Busy)
					if m.Bus != nil {
						m.Bus.Publish(agent.BusEvent{
							Type:    agent.EventActivityDetected,
							Source:  m.AgentID,
							Payload: "bytes",
						})
					}
				}

				data := make([]byte, n)
				copy(data, buf[:n])
				
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

	// Idle/Stuck check loop
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if m.Bus == nil {
					continue
				}
				
				since := time.Since(m.lastActivity)

				// Check Stuck
				if since > m.StuckTimeout {
					if !m.isStuck {
						m.isStuck = true
						m.Bus.Publish(agent.BusEvent{
							Type:   agent.EventStuck,
							Source: m.AgentID,
						})
					}
					// If stuck, we are also idle, but Stuck takes precedence for presence (Away)
					continue 
				}

				// Check Idle
				if since > m.ActivityTimeout {
					if !m.isIdle {
						m.isIdle = true
						m.Bus.Publish(agent.BusEvent{
							Type:   agent.EventIdle,
							Source: m.AgentID,
						})
					}
				}
			}
		}
	}()
}