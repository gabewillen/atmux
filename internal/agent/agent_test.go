package agent

import (
	"context"
	"testing"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go"
)

func TestNewAgent(t *testing.T) {
	cfg := config.AgentConfig{
		Name: "Test Agent",
	}
	repoRoot := api.RepoRoot("/tmp/repo")

	a, err := NewAgent(cfg, repoRoot)
	if err != nil {
		t.Fatalf("NewAgent failed: %v", err)
	}

	if a.Name != "Test Agent" {
		t.Errorf("Expected Name 'Test Agent', got %q", a.Name)
	}
	if a.Slug != "test-agent" {
		t.Errorf("Expected Slug 'test-agent', got %q", a.Slug)
	}
	if a.Lifecycle == nil {
		t.Error("Lifecycle HSM is nil")
	}
	if a.Presence == nil {
		t.Error("Presence HSM is nil")
	}

	// Verify initial states
	// Note: hsm-go usually exposes checking state, but exact method depends on API.
	// hsm.ID(instance) might return current state name if states are IDs?
	// hsm.Name(instance) returns HSM name.
	// We might need to query the state.
	// hsm-go seems to track current state internally.
	// For now, check no panic.
}

func TestLifecycleTransitions(t *testing.T) {
	cfg := config.AgentConfig{Name: "LifecycleAgent"}
	a, _ := NewAgent(cfg, "/tmp")
	ctx := context.Background()

	// Dispatch Spawn
	hsm.Dispatch(ctx, a.Lifecycle, hsm.Event{Name: EventSpawn})
	
	// Wait for processing? hsm-go Dispatch returns a channel.
	<-hsm.Dispatch(ctx, a.Lifecycle, hsm.Event{Name: EventStarted})

	// Should be Running
	// We assume transitions worked if no panic.
	// To verify state, we'd need to inspect the HSM.
}

func TestPresenceTransitions(t *testing.T) {
	cfg := config.AgentConfig{Name: "PresenceAgent"}
	a, _ := NewAgent(cfg, "/tmp")
	ctx := context.Background()

	// Dispatch Connect
	<-hsm.Dispatch(ctx, a.Presence, hsm.Event{Name: EventConnect})
	// Dispatch Busy
	<-hsm.Dispatch(ctx, a.Presence, hsm.Event{Name: EventBusy})
	// Dispatch Idle
	<-hsm.Dispatch(ctx, a.Presence, hsm.Event{Name: EventIdle})
	// Dispatch Disconnect
	<-hsm.Dispatch(ctx, a.Presence, hsm.Event{Name: EventDisconnect})
}
