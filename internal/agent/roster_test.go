package agent

import (
	"testing"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestRegistry_Roster(t *testing.T) {
	// Setup
	_ = NewRegistry() // Redundant as we use GlobalRegistry
	bus := NewEventBus()
	
	cfg := config.AgentConfig{
		Name:    "test-agent",
		Adapter: "test-adapter",
	}
	// NewAgent registers to GlobalRegistry
	// Reset GlobalRegistry for test
	GlobalRegistry = NewRegistry()
	
	if _, err := NewAgent(cfg, "/tmp", bus); err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Check List
	roster := GlobalRegistry.GetRoster()
	if len(roster) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(roster))
	}
	if roster[0].Name != "test-agent" {
		t.Errorf("Expected name 'test-agent', got %s", roster[0].Name)
	}
	if roster[0].Presence != api.PresenceOffline {
		t.Errorf("Expected Offline (default), got %s", roster[0].Presence)
	}
}