package manager

import (
	"encoding/json"
	"testing"

	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestDecodeEventPayload(t *testing.T) {
	t.Parallel()
	var out struct {
		Name string `json:"name"`
	}
	if err := decodeEventPayload(nil, &out); err == nil {
		t.Fatalf("expected error for nil payload")
	}
	if err := decodeEventPayload(map[string]any{"name": "ok"}, &out); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if out.Name != "ok" {
		t.Fatalf("unexpected payload: %#v", out)
	}
	if err := decodeEventPayload(make(chan int), &out); err == nil {
		t.Fatalf("expected marshal error")
	}
}

func TestSortRosterAndLocation(t *testing.T) {
	t.Parallel()
	entries := []api.RosterEntry{
		{Kind: api.RosterAgent, Name: "b", RuntimeID: api.NewRuntimeID()},
		{Kind: api.RosterDirector, Name: "director", RuntimeID: api.NewRuntimeID()},
		{Kind: api.RosterManager, Name: "manager", RuntimeID: api.NewRuntimeID()},
		{Kind: api.RosterAgent, Name: "a", RuntimeID: api.NewRuntimeID()},
	}
	sortRoster(entries)
	if entries[0].Kind != api.RosterDirector || entries[1].Kind != api.RosterManager {
		t.Fatalf("unexpected roster order: %#v", entries)
	}
	state := &agentState{
		config: config.AgentConfig{
			Location: config.AgentLocationConfig{Type: "local", Host: "host", RepoPath: "/tmp/repo"},
		},
	}
	location := locationForState(state)
	if location == nil || location.Type != api.LocationLocal {
		t.Fatalf("unexpected location: %#v", location)
	}
	state.runtime = &agent.Agent{Agent: api.Agent{Location: api.Location{Type: api.LocationLocal}}}
	location = locationForState(state)
	if location == nil || location.Type != api.LocationLocal {
		t.Fatalf("unexpected runtime location: %#v", location)
	}
}

func TestRosterEntryDefaults(t *testing.T) {
	t.Parallel()
	mgr := &Manager{}
	entry := mgr.rosterEntry(api.NewAgentID(), nil)
	if entry.Presence != agent.PresenceOffline {
		t.Fatalf("expected offline presence")
	}
	payload, err := json.Marshal(entry)
	if err != nil || len(payload) == 0 {
		t.Fatalf("expected roster entry to marshal")
	}
}

