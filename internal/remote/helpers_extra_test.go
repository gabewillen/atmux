package remote

import (
	"context"
	"testing"

	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestOutbox(t *testing.T) {
	outbox := NewOutbox(5)
	outbox.Enqueue("s1", []byte("abc"))
	outbox.Enqueue("s2", []byte("def"))
	items := outbox.Drain()
	if len(items) == 0 {
		t.Fatalf("expected items")
	}
	if outbox.Drain() != nil {
		t.Fatalf("expected empty after drain")
	}
}

func TestOutboxDisabled(t *testing.T) {
	var outbox *Outbox
	outbox.Enqueue("s", []byte("a"))
	if outbox.Drain() != nil {
		t.Fatalf("expected nil drain")
	}
	outbox = NewOutbox(0)
	outbox.Enqueue("s", []byte("a"))
	if outbox.Drain() != nil {
		t.Fatalf("expected empty")
	}
}

func TestHostPermissions(t *testing.T) {
	perms := HostPermissions("", api.MustParseHostID("host"), "")
	if len(perms.Pub.Allow) == 0 || len(perms.Sub.Allow) == 0 {
		t.Fatalf("expected permissions")
	}
}

func TestKVStoreErrors(t *testing.T) {
	if _, err := NewKVStore(nil, "bucket"); err == nil {
		t.Fatalf("expected js error")
	}
	var store *KVStore
	if err := store.Put(context.Background(), "key", []byte("value")); err == nil {
		t.Fatalf("expected put error")
	}
	if _, err := store.Get(context.Background(), "key"); err == nil {
		t.Fatalf("expected get error")
	}
	if _, err := store.ListKeys(context.Background(), ""); err == nil {
		t.Fatalf("expected list error")
	}
}

func TestDecodeEventPayload(t *testing.T) {
	if err := decodeEventPayload(nil, &struct{}{}); err == nil {
		t.Fatalf("expected decode error")
	}
	var roster []api.RosterEntry
	payload := []api.RosterEntry{{Kind: api.RosterAgent, Name: "alpha", RuntimeID: api.NewRuntimeID()}}
	if err := decodeEventPayload(payload, &roster); err != nil {
		t.Fatalf("decode: %v", err)
	}
	manager := &HostManager{agentIndex: map[api.AgentID]*remoteSession{}}
	manager.updateSessionPresence(api.NewAgentID(), "BUSY")
	event := protocol.Event{Name: agent.EventPresenceChanged, Payload: payload}
	manager.handlePresenceEvent(event)
}
