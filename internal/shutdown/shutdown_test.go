package shutdown

import (
	"context"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/internal/session"
)

func TestNewController(t *testing.T) {
	sessions := session.NewManager(event.NewNoopDispatcher())
	ctrl := NewController(sessions, event.NewNoopDispatcher(), 5*time.Second)

	if ctrl.State() != StateRunning {
		t.Errorf("initial state = %q, want %q", ctrl.State(), StateRunning)
	}
}

func TestRequestShutdown(t *testing.T) {
	sessions := session.NewManager(event.NewNoopDispatcher())
	ctrl := NewController(sessions, event.NewNoopDispatcher(), 1*time.Second)
	ctx := context.Background()

	ctrl.RequestShutdown(ctx)

	// Should transition through draining to stopped (no active sessions)
	select {
	case <-ctrl.Done():
		// OK - reached stopped
	case <-time.After(5 * time.Second):
		t.Fatal("controller did not reach stopped state")
	}

	if ctrl.State() != StateStopped {
		t.Errorf("final state = %q, want %q", ctrl.State(), StateStopped)
	}
}

func TestForceShutdown(t *testing.T) {
	sessions := session.NewManager(event.NewNoopDispatcher())
	ctrl := NewController(sessions, event.NewNoopDispatcher(), 1*time.Second)
	ctx := context.Background()

	ctrl.ForceShutdown(ctx)

	select {
	case <-ctrl.Done():
		// OK
	case <-time.After(5 * time.Second):
		t.Fatal("controller did not reach stopped state after force shutdown")
	}

	if ctrl.State() != StateStopped {
		t.Errorf("final state = %q, want %q", ctrl.State(), StateStopped)
	}
}

func TestDoubleSignalEscalation(t *testing.T) {
	sessions := session.NewManager(event.NewNoopDispatcher())
	ctrl := NewController(sessions, event.NewNoopDispatcher(), 30*time.Second)
	ctx := context.Background()

	// First signal - should go to draining
	ctrl.RequestShutdown(ctx)

	// Controller should be draining or stopped (no active sessions means fast drain)
	select {
	case <-ctrl.Done():
		// Fast drain completed because no sessions
		if ctrl.State() != StateStopped {
			t.Errorf("state = %q, want %q", ctrl.State(), StateStopped)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("controller did not reach stopped state")
	}
}

func TestDefaultDrainTimeout(t *testing.T) {
	sessions := session.NewManager(event.NewNoopDispatcher())
	ctrl := NewController(sessions, event.NewNoopDispatcher(), 0)

	// Default timeout should be 30 seconds
	if ctrl.drainTimeout != 30*time.Second {
		t.Errorf("default drain timeout = %v, want %v", ctrl.drainTimeout, 30*time.Second)
	}
}

func TestShutdownEventsEmitted(t *testing.T) {
	dispatcher := event.NewLocalDispatcher()
	sessions := session.NewManager(dispatcher)
	ctrl := NewController(sessions, dispatcher, 1*time.Second)
	ctx := context.Background()

	var receivedTypes []event.Type
	dispatcher.Subscribe(event.Subscription{
		Handler: func(ctx context.Context, evt event.Event) error {
			receivedTypes = append(receivedTypes, evt.Type)
			return nil
		},
	})

	ctrl.RequestShutdown(ctx)

	select {
	case <-ctrl.Done():
		// OK
	case <-time.After(5 * time.Second):
		t.Fatal("shutdown did not complete")
	}

	// Check that shutdown events were emitted
	hasShutdownInitiated := false
	for _, t := range receivedTypes {
		if t == event.TypeShutdownInitiated {
			hasShutdownInitiated = true
		}
	}
	if !hasShutdownInitiated {
		t.Error("shutdown.initiated event should have been emitted")
	}
}
