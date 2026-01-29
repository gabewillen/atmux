// Package monitor provides PTY output monitoring for amux.
//
// The monitor observes PTY output to detect activity, inactivity, and
// state changes. Pattern matching is delegated to adapters via the
// PatternMatcher interface.
//
// See spec §7 for PTY monitoring requirements.
package monitor

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/event"
	"github.com/stateforward/hsm-go/muid"
)

// Monitor observes PTY output and emits events.
type Monitor struct {
	mu           sync.Mutex
	agentID      muid.MUID
	reader       io.Reader
	matcher      adapter.PatternMatcher
	dispatcher   event.Dispatcher
	idleTimeout  time.Duration
	stuckTimeout time.Duration
	running      bool
	cancel       context.CancelFunc
}

// Config holds monitor configuration.
type Config struct {
	// AgentID is the agent being monitored.
	AgentID muid.MUID

	// Reader is the PTY output reader.
	Reader io.Reader

	// Matcher is the pattern matcher (may be noop).
	Matcher adapter.PatternMatcher

	// Dispatcher is the event dispatcher.
	Dispatcher event.Dispatcher

	// IdleTimeout is the idle detection timeout.
	IdleTimeout time.Duration

	// StuckTimeout is the stuck detection timeout.
	StuckTimeout time.Duration
}

// New creates a new monitor.
func New(cfg Config) *Monitor {
	if cfg.Matcher == nil {
		cfg.Matcher = adapter.NewNoopPatternMatcher()
	}
	if cfg.Dispatcher == nil {
		cfg.Dispatcher = event.NewNoopDispatcher()
	}

	return &Monitor{
		agentID:      cfg.AgentID,
		reader:       cfg.Reader,
		matcher:      cfg.Matcher,
		dispatcher:   cfg.Dispatcher,
		idleTimeout:  cfg.IdleTimeout,
		stuckTimeout: cfg.StuckTimeout,
	}
}

// Start begins monitoring.
func (m *Monitor) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil
	}
	m.running = true

	ctx, m.cancel = context.WithCancel(ctx)
	m.mu.Unlock()

	go m.run(ctx)
	return nil
}

// Stop stops monitoring.
func (m *Monitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	m.running = false
	if m.cancel != nil {
		m.cancel()
	}
}

func (m *Monitor) run(ctx context.Context) {
	buf := make([]byte, 4096)
	lastActivity := time.Now()
	lastPatternChange := time.Now()
	var lastPatternResult bool

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Read with timeout
		n, err := m.reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				return
			}
			// Handle read error
			continue
		}

		if n > 0 {
			now := time.Now()
			lastActivity = now

			// Emit activity event
			_ = m.dispatcher.Dispatch(ctx, event.NewEvent(event.TypePTYActivity, m.agentID, nil))

			// Check for pattern matches
			matches := m.matcher.Match(buf[:n])
			matched := len(matches) > 0
			if matched != lastPatternResult {
				lastPatternChange = now
				lastPatternResult = matched
			}
			for _, match := range matches {
				_ = m.dispatcher.Dispatch(ctx, event.NewEvent(event.TypePTYOutput, m.agentID, match))
			}
		}

		// Check for idle timeout
		if m.idleTimeout > 0 && time.Since(lastActivity) > m.idleTimeout {
			_ = m.dispatcher.Dispatch(ctx, event.NewEvent(event.TypePTYIdle, m.agentID, nil))
		}

		// Check for stuck timeout (pattern result unchanged for too long)
		if m.stuckTimeout > 0 && time.Since(lastPatternChange) > m.stuckTimeout {
			_ = m.dispatcher.Dispatch(ctx, event.NewEvent(event.TypePTYStuck, m.agentID, nil))
			// Reset to avoid repeated stuck events
			lastPatternChange = time.Now()
		}
	}
}
