package agent

import (
	"context"
	"sync"
	"testing"

	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/agentflare-ai/amux/internal/protocol"
)

// capturingDispatcher records dispatched events for tests.
type capturingDispatcher struct {
	mu     sync.Mutex
	events []protocol.Event
}

func (c *capturingDispatcher) Dispatch(ctx context.Context, event protocol.Event) error {
	c.mu.Lock()
	c.events = append(c.events, event)
	c.mu.Unlock()
	return nil
}

func (c *capturingDispatcher) Subscribe(filter protocol.EventFilter) (<-chan protocol.Event, func()) {
	ch := make(chan protocol.Event, 10)
	return ch, func() { close(ch) }
}

func (c *capturingDispatcher) eventsWithType(typ string) []protocol.Event {
	c.mu.Lock()
	defer c.mu.Unlock()
	var out []protocol.Event
	for _, e := range c.events {
		if e.Type == typ {
			out = append(out, e)
		}
	}
	return out
}

func TestLifecycleTransitions(t *testing.T) {
	ctx := context.Background()
	agent := &api.Agent{ID: api.NextRuntimeID(), Name: "test", Adapter: "test"}
	disp := &capturingDispatcher{}
	act, err := NewActor(agent, disp)
	if err != nil {
		t.Fatalf("NewActor: %v", err)
	}
	act.Start(ctx)

	// pending → starting
	act.DispatchLifecycle(ctx, EventLifecycleStart, nil)
	if st := act.LifecycleState(); st != LifecycleStarting {
		t.Errorf("after start: state = %q, want %q", st, LifecycleStarting)
	}

	// starting → running
	act.DispatchLifecycle(ctx, EventLifecycleReady, nil)
	if st := act.LifecycleState(); st != LifecycleRunning {
		t.Errorf("after ready: state = %q, want %q", st, LifecycleRunning)
	}

	// running → terminated
	act.DispatchLifecycle(ctx, EventLifecycleStop, nil)
	if st := act.LifecycleState(); st != LifecycleTerminated {
		t.Errorf("after stop: state = %q, want %q", st, LifecycleTerminated)
	}

	// events emitted
	lifecycleEvents := disp.eventsWithType("lifecycle.changed")
	if len(lifecycleEvents) < 3 {
		t.Errorf("want at least 3 lifecycle.changed events, got %d", len(lifecycleEvents))
	}
}

func TestLifecycleErrorPath(t *testing.T) {
	ctx := context.Background()
	agent := &api.Agent{ID: api.NextRuntimeID(), Name: "test", Adapter: "test"}
	disp := &capturingDispatcher{}
	act, err := NewActor(agent, disp)
	if err != nil {
		t.Fatalf("NewActor: %v", err)
	}
	act.Start(ctx)

	// * → errored
	act.DispatchLifecycle(ctx, EventLifecycleError, nil)
	if st := act.LifecycleState(); st != LifecycleErrored {
		t.Errorf("after error: state = %q, want %q", st, LifecycleErrored)
	}
	lifecycleEvents := disp.eventsWithType("lifecycle.changed")
	if len(lifecycleEvents) < 1 {
		t.Errorf("want at least 1 lifecycle.changed on error, got %d", len(lifecycleEvents))
	}
}

func TestNewActorRejectsZeroID(t *testing.T) {
	agent := &api.Agent{ID: api.BroadcastID, Name: "test", Adapter: "test"}
	_, err := NewActor(agent, nil)
	if err == nil {
		t.Error("NewActor with zero ID should error")
	}
}
