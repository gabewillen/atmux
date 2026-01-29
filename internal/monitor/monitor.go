package monitor

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/stateforward/hsm-go"
)

// Config holds monitor configuration.
type Config struct {
	IdleTimeout time.Duration
	BufferSize  int
}

// Monitor watches a PTY stream for activity and inactivity.
type Monitor struct {
	cfg        Config
	reader     io.Reader
	actor      hsm.Instance
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.Mutex
	lastActive time.Time
}

// NewMonitor creates a new PTY monitor.
func NewMonitor(cfg Config, reader io.Reader, actor hsm.Instance) *Monitor {
	if cfg.BufferSize == 0 {
		cfg.BufferSize = 4096
	}
	return &Monitor{
		cfg:    cfg,
		reader: reader,
		actor:  actor,
	}
}

// Start begins monitoring the PTY.
func (m *Monitor) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	m.cancelFunc = cancel
	m.lastActive = time.Now()

	m.wg.Add(2)
	go m.readLoop(ctx)
	go m.timeoutLoop(ctx)
}

// Stop stops the monitor.
func (m *Monitor) Stop() {
	if m.cancelFunc != nil {
		m.cancelFunc()
	}
	m.wg.Wait()
}

func (m *Monitor) readLoop(ctx context.Context) {
	defer m.wg.Done()
	buf := make([]byte, m.cfg.BufferSize)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Read is blocking, but we rely on Close() of PTY or context cancel to break out
			// In practice, pty.Read returns error on close.
			n, err := m.reader.Read(buf)
			if n > 0 {
				m.mu.Lock()
				m.lastActive = time.Now()
				m.mu.Unlock()

				// Dispatch activity event
				hsm.Dispatch(ctx, m.actor, hsm.Event{
					Name: "agent.presence.activity",
				})

				// TODO: Send data to TUI decoder and Adapter runtime
			}
			if err != nil {
				if err != io.EOF {
					// Logic for read error (process exit usually)
				}
				return
			}
		}
	}
}

func (m *Monitor) timeoutLoop(ctx context.Context) {
	defer m.wg.Done()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.mu.Lock()
			elapsed := time.Since(m.lastActive)
			m.mu.Unlock()

			if elapsed > m.cfg.IdleTimeout {
				hsm.Dispatch(ctx, m.actor, hsm.Event{
					Name: "agent.presence.inactivity",
				})
			}
		}
	}
}
