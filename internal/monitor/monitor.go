package monitor

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/tui"
)

// ErrMatcherRequired is returned when a matcher is missing.
var ErrMatcherRequired = errors.New("matcher required")

// EventType describes a PTY monitor event type.
type EventType string

const (
	// EventActivity indicates PTY output activity.
	EventActivity EventType = "activity.detected"
	// EventInactivity indicates PTY output inactivity.
	EventInactivity EventType = "inactivity.detected"
	// EventStuck indicates a stuck PTY (no output for stuck timeout).
	EventStuck EventType = "stuck.detected"
)

// Event captures a PTY monitor observation.
type Event struct {
	Type  EventType
	At    time.Time
	Since time.Duration
}

// Options configures monitor behavior.
type Options struct {
	// IdleTimeout triggers an inactivity event after no output.
	IdleTimeout time.Duration
	// StuckTimeout triggers a stuck event after prolonged inactivity.
	StuckTimeout time.Duration
	// TUIEnabled enables terminal decoding.
	TUIEnabled bool
	// TUIRows sets the initial decoder row count.
	TUIRows int
	// TUICols sets the initial decoder column count.
	TUICols int
}

// Monitor scans PTY output, detects timeouts, and optionally decodes TUI output.
type Monitor struct {
	matcher     adapter.PatternMatcher
	events      chan Event
	idleTimeout time.Duration
	stuckTimeout time.Duration
	decoder     *tui.Decoder
	mu          sync.Mutex
	lastActivity time.Time
	idleSeq     uint64
	stuckSeq    uint64
	idleTimer   *time.Timer
	stuckTimer  *time.Timer
	closed      bool
}

// NewMonitor constructs a monitor with the provided matcher and options.
func NewMonitor(matcher adapter.PatternMatcher, opts Options) (*Monitor, error) {
	if matcher == nil {
		return nil, fmt.Errorf("new monitor: %w", ErrMatcherRequired)
	}
	monitor := &Monitor{
		matcher:      matcher,
		events:       make(chan Event, 32),
		idleTimeout:  opts.IdleTimeout,
		stuckTimeout: opts.StuckTimeout,
		lastActivity: time.Now().UTC(),
	}
	if opts.TUIEnabled {
		monitor.decoder = tui.NewDecoder(tui.Config{Rows: opts.TUIRows, Cols: opts.TUICols})
	}
	monitor.resetTimers(monitor.lastActivity)
	return monitor, nil
}

// Events returns the monitor event channel.
func (m *Monitor) Events() <-chan Event {
	if m == nil {
		ch := make(chan Event)
		close(ch)
		return ch
	}
	return m.events
}

// Close stops the monitor and releases resources.
func (m *Monitor) Close() error {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil
	}
	m.closed = true
	if m.idleTimer != nil {
		m.idleTimer.Stop()
	}
	if m.stuckTimer != nil {
		m.stuckTimer.Stop()
	}
	m.idleSeq++
	m.stuckSeq++
	close(m.events)
	m.mu.Unlock()
	return nil
}

// Observe processes PTY output bytes and returns any pattern matches.
func (m *Monitor) Observe(ctx context.Context, output []byte) ([]adapter.PatternMatch, error) {
	if m == nil {
		return nil, fmt.Errorf("monitor observe: %w", ErrMatcherRequired)
	}
	if len(output) == 0 {
		return nil, nil
	}
	now := time.Now().UTC()
	m.resetTimers(now)
	m.emit(Event{Type: EventActivity, At: now})
	if m.decoder != nil {
		_ = m.decoder.Write(output)
	}
	return m.matcher.Match(ctx, output)
}

// Resize updates the TUI decoder geometry.
func (m *Monitor) Resize(rows, cols int) {
	if m == nil || m.decoder == nil {
		return
	}
	m.decoder.Resize(rows, cols)
}

// TUIXML returns the latest TUI XML snapshot.
func (m *Monitor) TUIXML() string {
	if m == nil || m.decoder == nil {
		return ""
	}
	return m.decoder.EncodeXML()
}

func (m *Monitor) resetTimers(now time.Time) {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return
	}
	m.lastActivity = now
	if m.idleTimeout > 0 {
		m.idleSeq++
		idleSeq := m.idleSeq
		if m.idleTimer != nil {
			m.idleTimer.Stop()
		}
		m.idleTimer = time.AfterFunc(m.idleTimeout, func() {
			m.fireIdle(idleSeq)
		})
	}
	if m.stuckTimeout > 0 {
		m.stuckSeq++
		stuckSeq := m.stuckSeq
		if m.stuckTimer != nil {
			m.stuckTimer.Stop()
		}
		m.stuckTimer = time.AfterFunc(m.stuckTimeout, func() {
			m.fireStuck(stuckSeq)
		})
	}
	m.mu.Unlock()
}

func (m *Monitor) fireIdle(seq uint64) {
	m.mu.Lock()
	if m.closed || seq != m.idleSeq {
		m.mu.Unlock()
		return
	}
	since := time.Since(m.lastActivity)
	at := time.Now().UTC()
	m.mu.Unlock()
	m.emit(Event{Type: EventInactivity, At: at, Since: since})
}

func (m *Monitor) fireStuck(seq uint64) {
	m.mu.Lock()
	if m.closed || seq != m.stuckSeq {
		m.mu.Unlock()
		return
	}
	since := time.Since(m.lastActivity)
	at := time.Now().UTC()
	m.mu.Unlock()
	m.emit(Event{Type: EventStuck, At: at, Since: since})
}

func (m *Monitor) emit(event Event) {
	if m == nil {
		return
	}
	m.mu.Lock()
	closed := m.closed
	m.mu.Unlock()
	if closed {
		return
	}
	defer func() {
		_ = recover()
	}()
	select {
	case m.events <- event:
	default:
	}
}
