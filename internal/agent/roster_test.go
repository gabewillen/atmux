package agent

import (
	"context"
	"testing"

	"github.com/stateforward/hsm-go/muid"

	"github.com/stateforward/amux/internal/event"
	"github.com/stateforward/amux/pkg/api"
)

func TestRosterStoreUpsertAndList(t *testing.T) {
	dispatcher := event.NewLocalDispatcher()
	store := NewRosterStore(dispatcher)

	agentID := muid.Make()
	ag := &api.Agent{
		ID:       agentID,
		Name:     "alpha",
		About:    "test agent",
		Adapter:  "test-adapter",
		RepoRoot: "/tmp/repo",
	}

	ctx := context.Background()
	if err := store.UpsertAgent(ctx, ag, StateOnline); err != nil {
		t.Fatalf("UpsertAgent returned error: %v", err)
	}

	roster := store.List()
	if len(roster) != 1 {
		t.Fatalf("List() length = %d, want %d", len(roster), 1)
	}

	entry := roster[0]
	if entry.AgentID != agentID {
		t.Errorf("Roster entry AgentID = %v, want %v", entry.AgentID, agentID)
	}
	if entry.Name != "alpha" {
		t.Errorf("Roster entry Name = %q, want %q", entry.Name, "alpha")
	}
	if entry.Adapter != "test-adapter" {
		t.Errorf("Roster entry Adapter = %q, want %q", entry.Adapter, "test-adapter")
	}
	if entry.Presence != StateOnline {
		t.Errorf("Roster entry Presence = %q, want %q", entry.Presence, StateOnline)
	}
	if entry.RepoRoot != "/tmp/repo" {
		t.Errorf("Roster entry RepoRoot = %q, want %q", entry.RepoRoot, "/tmp/repo")
	}
}

func TestRosterStoreEventsEmitted(t *testing.T) {
	dispatcher := event.NewLocalDispatcher()
	store := NewRosterStore(dispatcher)

	ctx := context.Background()

	// Subscribe to presence.changed and roster.updated events.
	presenceCh, err := dispatcher.Subscribe(ctx, event.TypeFilter{Prefix: EventTypePresenceChanged})
	if err != nil {
		t.Fatalf("Subscribe presence.changed returned error: %v", err)
	}
	rosterCh, err := dispatcher.Subscribe(ctx, event.TypeFilter{Prefix: EventTypeRosterUpdated})
	if err != nil {
		t.Fatalf("Subscribe roster.updated returned error: %v", err)
	}

	agentID := muid.Make()
	ag := &api.Agent{
		ID:       agentID,
		Name:     "beta",
		Adapter:  "test-adapter",
		RepoRoot: "/tmp/repo2",
	}

	if err := store.UpsertAgent(ctx, ag, StateBusy); err != nil {
		t.Fatalf("UpsertAgent returned error: %v", err)
	}

	// Verify presence.changed event.
	select {
	case ev := <-presenceCh:
		basic, ok := ev.(event.BasicEvent)
		if !ok {
			t.Fatalf("presence event type = %T, want event.BasicEvent", ev)
		}
		if basic.EventType != EventTypePresenceChanged {
			t.Errorf("presence event type = %q, want %q", basic.EventType, EventTypePresenceChanged)
		}
		payload, ok := basic.Payload.(PresenceChangedPayload)
		if !ok {
			t.Fatalf("presence payload type = %T, want PresenceChangedPayload", basic.Payload)
		}
		if payload.AgentID != agentID {
			t.Errorf("presence payload AgentID = %v, want %v", payload.AgentID, agentID)
		}
		if payload.Presence != StateBusy {
			t.Errorf("presence payload Presence = %q, want %q", payload.Presence, StateBusy)
		}
	default:
		t.Fatal("expected presence.changed event, got none")
	}

	// Verify roster.updated event.
	select {
	case ev := <-rosterCh:
		basic, ok := ev.(event.BasicEvent)
		if !ok {
			t.Fatalf("roster event type = %T, want event.BasicEvent", ev)
		}
		if basic.EventType != EventTypeRosterUpdated {
			t.Errorf("roster event type = %q, want %q", basic.EventType, EventTypeRosterUpdated)
		}
		entries, ok := basic.Payload.([]api.RosterEntry)
		if !ok {
			t.Fatalf("roster payload type = %T, want []api.RosterEntry", basic.Payload)
		}
		if len(entries) != 1 {
			t.Fatalf("roster entries length = %d, want %d", len(entries), 1)
		}
		if entries[0].AgentID != agentID {
			t.Errorf("roster entry AgentID = %v, want %v", entries[0].AgentID, agentID)
		}
	default:
		t.Fatal("expected roster.updated event, got none")
	}
}
