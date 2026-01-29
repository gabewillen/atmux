package agent

import (
	"context"
	"testing"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestRosterStore_AddListRemove(t *testing.T) {
	ctx := context.Background()
	disp := protocol.NewDispatcher()
	store := NewRosterStore(disp)

	id := api.NextRuntimeID()
	store.Add(ctx, api.RosterEntry{
		AgentID:  id,
		Name:     "alice",
		Adapter:  "test",
		Presence: PresenceOnline,
		Kind:     api.RosterKindAgent,
	})
	list := store.List()
	if len(list) != 1 {
		t.Fatalf("List() len = %d, want 1", len(list))
	}
	if list[0].AgentID != id || list[0].Name != "alice" {
		t.Errorf("entry = %+v", list[0])
	}

	store.Remove(ctx, id)
	list = store.List()
	if len(list) != 0 {
		t.Errorf("after Remove: List() len = %d, want 0", len(list))
	}
}

func TestRosterStore_UpdatePresence(t *testing.T) {
	ctx := context.Background()
	disp := protocol.NewDispatcher()
	store := NewRosterStore(disp)

	id := api.NextRuntimeID()
	store.Add(ctx, api.RosterEntry{AgentID: id, Name: "bob", Presence: PresenceOnline, Kind: api.RosterKindAgent})
	store.UpdatePresence(ctx, id, PresenceBusy)
	list := store.List()
	if len(list) != 1 || list[0].Presence != PresenceBusy {
		t.Errorf("UpdatePresence: list[0].Presence = %q, want %q", list[0].Presence, PresenceBusy)
	}
}

func TestRosterStore_RejectsZeroID(t *testing.T) {
	ctx := context.Background()
	store := NewRosterStore(protocol.NewDispatcher())
	store.Add(ctx, api.RosterEntry{AgentID: api.BroadcastID, Name: "x", Kind: api.RosterKindAgent})
	list := store.List()
	if len(list) != 0 {
		t.Errorf("Add with zero ID should not add entry, got len %d", len(list))
	}
}

func TestRosterStore_EmitsRosterUpdated(t *testing.T) {
	ctx := context.Background()
	disp := protocol.NewDispatcher()
	ch, unsub := disp.Subscribe(protocol.EventFilter{Types: []string{"roster.updated"}})
	defer unsub()

	store := NewRosterStore(disp)
	id := api.NextRuntimeID()
	store.Add(ctx, api.RosterEntry{AgentID: id, Name: "sub", Presence: PresenceOnline, Kind: api.RosterKindAgent})

	ev := <-ch
	if ev.Type != "roster.updated" {
		t.Errorf("event Type = %q, want roster.updated", ev.Type)
	}
	if list, ok := ev.Data.(api.Roster); !ok || len(list) != 1 {
		t.Errorf("event Data = %T %v", ev.Data, ev.Data)
	}
}
