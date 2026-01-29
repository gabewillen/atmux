package monitor

import (
	"bytes"
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stateforward/hsm-go/muid"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/event"
)

// collectingDispatcher records dispatched events.
type collectingDispatcher struct {
	mu     sync.Mutex
	events []event.Event
}

func (d *collectingDispatcher) Dispatch(_ context.Context, evt event.Event) error {
	d.mu.Lock()
	d.events = append(d.events, evt)
	d.mu.Unlock()
	return nil
}

func (d *collectingDispatcher) Subscribe(_ event.Subscription) func() {
	return func() {}
}

func (d *collectingDispatcher) Close() error { return nil }

func (d *collectingDispatcher) eventTypes() []event.Type {
	d.mu.Lock()
	defer d.mu.Unlock()
	types := make([]event.Type, len(d.events))
	for i, e := range d.events {
		types[i] = e.Type
	}
	return types
}

func (d *collectingDispatcher) hasType(t event.Type) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, e := range d.events {
		if e.Type == t {
			return true
		}
	}
	return false
}

func TestMonitor_EmitsActivityEvent(t *testing.T) {
	disp := &collectingDispatcher{}
	r := bytes.NewReader([]byte("hello world"))

	m := New(Config{
		AgentID:    muid.Make(),
		Reader:     r,
		Matcher:    adapter.NewNoopPatternMatcher(),
		Dispatcher: disp,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	if err := m.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Wait for reader to be consumed and goroutine to exit
	time.Sleep(100 * time.Millisecond)
	m.Stop()

	if !disp.hasType(event.TypePTYActivity) {
		t.Error("expected TypePTYActivity event")
	}
}

func TestMonitor_IdleTimeout(t *testing.T) {
	disp := &collectingDispatcher{}

	// slowReader blocks indefinitely, causing idle timeout
	sr := &slowReader{delay: 200 * time.Millisecond}

	m := New(Config{
		AgentID:     muid.Make(),
		Reader:      sr,
		Matcher:     adapter.NewNoopPatternMatcher(),
		Dispatcher:  disp,
		IdleTimeout: 50 * time.Millisecond,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	if err := m.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Wait for idle timeout to trigger
	time.Sleep(300 * time.Millisecond)
	m.Stop()

	if !disp.hasType(event.TypePTYIdle) {
		t.Error("expected TypePTYIdle event from idle timeout")
	}
}

func TestMonitor_StuckTimeout(t *testing.T) {
	disp := &collectingDispatcher{}

	// continuousReader provides data but with no pattern changes
	cr := &continuousReader{data: []byte("same data"), interval: 10 * time.Millisecond}

	m := New(Config{
		AgentID:      muid.Make(),
		Reader:       cr,
		Matcher:      adapter.NewNoopPatternMatcher(), // never matches
		Dispatcher:   disp,
		StuckTimeout: 50 * time.Millisecond,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	if err := m.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Wait for stuck timeout to trigger
	time.Sleep(200 * time.Millisecond)
	m.Stop()

	if !disp.hasType(event.TypePTYStuck) {
		types := disp.eventTypes()
		t.Errorf("expected TypePTYStuck event, got types: %v", types)
	}
}

func TestMonitor_StopIdempotent(t *testing.T) {
	m := New(Config{
		AgentID:    muid.Make(),
		Reader:     bytes.NewReader(nil),
		Dispatcher: event.NewNoopDispatcher(),
	})

	// Stop without start should not panic
	m.Stop()
	m.Stop()
}

func TestMonitor_StartIdempotent(t *testing.T) {
	m := New(Config{
		AgentID:    muid.Make(),
		Reader:     bytes.NewReader([]byte("data")),
		Dispatcher: event.NewNoopDispatcher(),
	})

	ctx := context.Background()
	if err := m.Start(ctx); err != nil {
		t.Fatalf("First Start: %v", err)
	}
	// Second start should be a no-op
	if err := m.Start(ctx); err != nil {
		t.Fatalf("Second Start: %v", err)
	}
	m.Stop()
}

// slowReader blocks for a given delay on each Read, then returns 0 bytes.
type slowReader struct {
	delay time.Duration
}

func (r *slowReader) Read(p []byte) (int, error) {
	time.Sleep(r.delay)
	return 0, nil
}

// continuousReader produces data at regular intervals.
type continuousReader struct {
	data     []byte
	interval time.Duration
}

func (r *continuousReader) Read(p []byte) (int, error) {
	time.Sleep(r.interval)
	n := copy(p, r.data)
	return n, nil
}
