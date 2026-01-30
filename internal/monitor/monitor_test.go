package monitor

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
)

type stubMatcher struct {
	match adapter.PatternMatch
	err   error
}

func (s stubMatcher) Match(ctx context.Context, output []byte) ([]adapter.PatternMatch, error) {
	_ = ctx
	if s.err != nil {
		return nil, s.err
	}
	if len(output) == 0 {
		return nil, nil
	}
	return []adapter.PatternMatch{s.match}, nil
}

type countMatcher struct {
	calls int
}

func (c *countMatcher) Match(ctx context.Context, output []byte) ([]adapter.PatternMatch, error) {
	_ = ctx
	c.calls++
	return nil, nil
}

func TestMonitorNewRequiresMatcher(t *testing.T) {
	if _, err := NewMonitor(nil, Options{}); err == nil {
		t.Fatalf("expected error for nil matcher")
	}
}

func TestMonitorObserve(t *testing.T) {
	mon, err := NewMonitor(stubMatcher{match: adapter.PatternMatch{Pattern: "prompt", Text: "ready"}}, Options{})
	if err != nil {
		t.Fatalf("new monitor: %v", err)
	}
	matches, err := mon.Observe(context.Background(), []byte("hello"))
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(matches) != 1 || matches[0].Pattern != "prompt" {
		t.Fatalf("unexpected matches: %#v", matches)
	}
}

func TestMonitorObserveMatcherError(t *testing.T) {
	wantErr := errors.New("match failed")
	mon, err := NewMonitor(stubMatcher{err: wantErr}, Options{})
	if err != nil {
		t.Fatalf("new monitor: %v", err)
	}
	if _, err := mon.Observe(context.Background(), []byte("oops")); err != wantErr {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func TestMonitorObserveEmpty(t *testing.T) {
	matcher := &countMatcher{}
	mon, err := NewMonitor(matcher, Options{})
	if err != nil {
		t.Fatalf("new monitor: %v", err)
	}
	defer func() {
		_ = mon.Close()
	}()
	matches, err := mon.Observe(context.Background(), nil)
	if err != nil {
		t.Fatalf("observe: %v", err)
	}
	if matches != nil {
		t.Fatalf("expected no matches, got %#v", matches)
	}
	if matcher.calls != 0 {
		t.Fatalf("expected matcher not to run, got %d", matcher.calls)
	}
}

func TestMonitorTimeoutEvents(t *testing.T) {
	mon, err := NewMonitor(stubMatcher{}, Options{
		IdleTimeout:  20 * time.Millisecond,
		StuckTimeout: 40 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("new monitor: %v", err)
	}
	defer func() {
		_ = mon.Close()
	}()
	timeout := time.NewTimer(200 * time.Millisecond)
	defer timeout.Stop()
	seenIdle := false
	seenStuck := false
	for !(seenIdle && seenStuck) {
		select {
		case event := <-mon.Events():
			switch event.Type {
			case EventInactivity:
				seenIdle = true
			case EventStuck:
				seenStuck = true
			}
		case <-timeout.C:
			t.Fatalf("expected idle and stuck events")
		}
	}
}

func TestMonitorTUIXMLAndResize(t *testing.T) {
	mon, err := NewMonitor(stubMatcher{}, Options{
		TUIEnabled: true,
		TUIRows:    1,
		TUICols:    5,
	})
	if err != nil {
		t.Fatalf("new monitor: %v", err)
	}
	defer func() {
		_ = mon.Close()
	}()
	if _, err := mon.Observe(context.Background(), []byte("hi")); err != nil {
		t.Fatalf("observe: %v", err)
	}
	xml := mon.TUIXML()
	if !strings.Contains(xml, ">h</r>") {
		t.Fatalf("expected tui xml, got %s", xml)
	}
	mon.Resize(1, 2)
	xml = mon.TUIXML()
	if !strings.Contains(xml, `cols="2"`) {
		t.Fatalf("expected resized xml, got %s", xml)
	}
}

func TestMonitorNilAndClose(t *testing.T) {
	var mon *Monitor
	if _, err := mon.Observe(context.Background(), []byte("hi")); err == nil {
		t.Fatalf("expected error for nil monitor")
	}
	if mon.TUIXML() != "" {
		t.Fatalf("expected empty xml for nil monitor")
	}
	mon.Resize(1, 1)
	if err := mon.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	ch := mon.Events()
	if _, ok := <-ch; ok {
		t.Fatalf("expected closed channel")
	}
	mon2, err := NewMonitor(stubMatcher{}, Options{})
	if err != nil {
		t.Fatalf("new monitor: %v", err)
	}
	if err := mon2.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if err := mon2.Close(); err != nil {
		t.Fatalf("close second: %v", err)
	}
}
