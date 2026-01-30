package monitor

import (
	"bytes"
	"context"
	"errors"
	"testing"

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

func TestMonitorNewRequiresMatcher(t *testing.T) {
	if _, err := NewMonitor(nil); err == nil {
		t.Fatalf("expected error for nil matcher")
	}
}

func TestMonitorScan(t *testing.T) {
	mon, err := NewMonitor(stubMatcher{match: adapter.PatternMatch{Pattern: "prompt", Text: "ready"}})
	if err != nil {
		t.Fatalf("new monitor: %v", err)
	}
	if _, err := mon.Scan(context.Background(), nil); err == nil {
		t.Fatalf("expected error for nil reader")
	}
	matches, err := mon.Scan(context.Background(), bytes.NewBufferString("hello"))
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(matches) != 1 || matches[0].Pattern != "prompt" {
		t.Fatalf("unexpected matches: %#v", matches)
	}
}

func TestMonitorScanMatcherError(t *testing.T) {
	wantErr := errors.New("match failed")
	mon, err := NewMonitor(stubMatcher{err: wantErr})
	if err != nil {
		t.Fatalf("new monitor: %v", err)
	}
	if _, err := mon.Scan(context.Background(), bytes.NewBufferString("oops")); err != wantErr {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}
