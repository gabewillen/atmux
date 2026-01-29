package agent

import (
	"context"
	"testing"

	"github.com/agentflare-ai/amux/pkg/api"
)

func TestPresenceTransitions(t *testing.T) {
	ctx := context.Background()
	agent := &api.Agent{ID: api.NextRuntimeID(), Name: "test", Adapter: "test"}
	disp := &capturingDispatcher{}
	act, err := NewActor(agent, disp)
	if err != nil {
		t.Fatalf("NewActor: %v", err)
	}
	act.Start(ctx)

	// online → busy
	act.DispatchPresence(ctx, EventPresenceTaskAssigned, nil)
	if st := act.PresenceState(); st != PresenceBusy {
		t.Errorf("after task.assigned: state = %q, want %q", st, PresenceBusy)
	}

	// busy → online (task.completed)
	act.DispatchPresence(ctx, EventPresenceTaskCompleted, nil)
	if st := act.PresenceState(); st != PresenceOnline {
		t.Errorf("after task.completed: state = %q, want %q", st, PresenceOnline)
	}

	// online → busy → online (prompt.detected)
	act.DispatchPresence(ctx, EventPresenceTaskAssigned, nil)
	act.DispatchPresence(ctx, EventPresencePromptDetected, nil)
	if st := act.PresenceState(); st != PresenceOnline {
		t.Errorf("after prompt.detected: state = %q, want %q", st, PresenceOnline)
	}
}

func TestPresenceOfflineAway(t *testing.T) {
	ctx := context.Background()
	agent := &api.Agent{ID: api.NextRuntimeID(), Name: "test", Adapter: "test"}
	disp := &capturingDispatcher{}
	act, err := NewActor(agent, disp)
	if err != nil {
		t.Fatalf("NewActor: %v", err)
	}
	act.Start(ctx)

	// * → offline (rate.limit)
	act.DispatchPresence(ctx, EventPresenceRateLimit, nil)
	if st := act.PresenceState(); st != PresenceOffline {
		t.Errorf("after rate.limit: state = %q, want %q", st, PresenceOffline)
	}

	// offline → online (rate.cleared)
	act.DispatchPresence(ctx, EventPresenceRateCleared, nil)
	if st := act.PresenceState(); st != PresenceOnline {
		t.Errorf("after rate.cleared: state = %q, want %q", st, PresenceOnline)
	}

	// * → away (stuck.detected)
	act.DispatchPresence(ctx, EventPresenceStuckDetected, nil)
	if st := act.PresenceState(); st != PresenceAway {
		t.Errorf("after stuck.detected: state = %q, want %q", st, PresenceAway)
	}

	// away → online (activity.detected)
	act.DispatchPresence(ctx, EventPresenceActivityDetected, nil)
	if st := act.PresenceState(); st != PresenceOnline {
		t.Errorf("after activity.detected: state = %q, want %q", st, PresenceOnline)
	}

	events := disp.eventsWithType("presence.changed")
	if len(events) < 4 {
		t.Errorf("want at least 4 presence.changed events, got %d", len(events))
	}
}
