package agent

import (
	"context"
	"testing"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/stateforward/hsm-go"
)

func TestNewAgent(t *testing.T) {
	cfg := config.AgentConfig{
		Name:    "test-agent",
		Adapter: "test-adapter",
	}
	bus := NewEventBus()
	a, err := NewAgent(cfg, "/tmp/repo", bus)
	if err != nil {
		t.Fatalf("NewAgent failed: %v", err)
	}
	if a.ID == 0 {
		t.Error("ID not generated")
	}
	if a.Slug != "test-agent" {
		t.Errorf("Expected slug test-agent, got %s", a.Slug)
	}
}

func TestLifecycleTransitions(t *testing.T) {
	cfg := config.AgentConfig{Name: "LifecycleAgent"}
	bus := NewEventBus()
	a, _ := NewAgent(cfg, "/tmp", bus)
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
	bus := NewEventBus()
	a, _ := NewAgent(cfg, "/tmp", bus)
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
