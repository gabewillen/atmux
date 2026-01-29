package agent

import (
	"context"
	"testing"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go"
)

func TestGetRoster(t *testing.T) {
	// Setup
	reg := NewRegistry()
	
	// Create Agent
	cfg := config.AgentConfig{Name: "RosterAgent", Adapter: "test"}
	a, err := NewAgent(cfg, "/tmp")
	if err != nil {
		t.Fatalf("NewAgent failed: %v", err)
	}
	
	reg.Register(a)
	
	// Verify Initial Roster
	roster := reg.GetRoster()
	if len(roster) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(roster))
	}
	if roster[0].Presence != api.PresenceOffline {
		t.Errorf("Expected initial presence Offline, got %s", roster[0].Presence)
	}
	
	// Trigger Presence Change
	ctx := context.Background()
	<-hsm.Dispatch(ctx, a.Presence, hsm.Event{Name: EventConnect})
	
	// Verify Updated Roster
	roster = reg.GetRoster()
	if roster[0].Presence != api.PresenceOnline {
		t.Errorf("Expected updated presence Online, got %s", roster[0].Presence)
	}
}
