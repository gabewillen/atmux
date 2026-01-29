package agent

import (
	"testing"

	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go/muid"
)

func TestRoster(t *testing.T) {
	r := NewRoster()

	// 1. Add Agent
	id1 := muid.Make()
	slug1 := api.NewAgentSlug("agent-1")
	a1 := &AgentActor{
		data: api.Agent{
			ID:       id1,
			Slug:     slug1,
			Name:     "agent-1",
			State:    api.StateRunning,
			Presence: api.PresenceOnline,
		},
	}
	r.Add(a1)

	// 2. Get Agent
	got := r.Get(id1)
	if got != a1 {
		t.Errorf("Get(id1) = %v, want %v", got, a1)
	}

	// 3. List Agents
	entries := r.List()
	if len(entries) != 1 {
		t.Errorf("List() count = %d, want 1", len(entries))
	}
	if entries[0].Name != "agent-1" {
		t.Errorf("List()[0].Name = %s, want agent-1", entries[0].Name)
	}
	if entries[0].Status != "online" {
		t.Errorf("List()[0].Status = %s, want online", entries[0].Status)
	}

	// 4. Add second agent
	id2 := muid.Make()
	slug2 := api.NewAgentSlug("agent-2")
	a2 := &AgentActor{
		data: api.Agent{
			ID:    id2,
			Slug:  slug2,
			Name:  "agent-2",
			State: api.StatePending,
		},
	}
	r.Add(a2)

	entries = r.List()
	if len(entries) != 2 {
		t.Errorf("List() count = %d, want 2", len(entries))
	}

	// 5. Remove Agent
	r.Remove(id1)
	if r.Get(id1) != nil {
		t.Errorf("Get(id1) after remove should be nil")
	}
	entries = r.List()
	if len(entries) != 1 {
		t.Errorf("List() count after remove = %d, want 1", len(entries))
	}
	if entries[0].Name != "agent-2" {
		t.Errorf("List()[0].Name = %s, want agent-2", entries[0].Name)
	}
}
