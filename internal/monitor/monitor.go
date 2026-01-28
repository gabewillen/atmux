package monitor

import (
	"context"
	"fmt"
	"io"

	"github.com/agentflare-ai/amux/internal/adapter"
)

// Monitor scans PTY output with an adapter matcher.
type Monitor struct {
	matcher adapter.PatternMatcher
}

// NewMonitor constructs a monitor with the provided matcher.
func NewMonitor(matcher adapter.PatternMatcher) *Monitor {
	if matcher == nil {
		matcher = &adapter.NoopMatcher{}
	}
	return &Monitor{matcher: matcher}
}

// Scan reads from r and emits pattern matches.
func (m *Monitor) Scan(ctx context.Context, r io.Reader) ([]adapter.PatternMatch, error) {
	if r == nil {
		return nil, fmt.Errorf("monitor scan: reader is nil")
	}
	buf := make([]byte, 4096)
	n, err := r.Read(buf)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("monitor scan: %w", err)
	}
	if n == 0 {
		return nil, nil
	}
	return m.matcher.Match(ctx, buf[:n])
}
