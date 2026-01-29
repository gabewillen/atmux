package agent

import (
	"context"
	"testing"

	"github.com/stateforward/hsm-go/muid"

	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestManagerAdd(t *testing.T) {
	mgr := NewManager(event.NewNoopDispatcher())
	ctx := context.Background()

	agent, err := mgr.Add(ctx, api.Agent{
		Name:     "test-agent",
		Adapter:  "claude-code",
		RepoRoot: "/home/user/project",
	})

	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	if agent.ID == 0 {
		t.Error("Agent ID should be non-zero")
	}
	if agent.Slug == "" {
		t.Error("Agent Slug should be set")
	}
	if agent.Slug != "test-agent" {
		t.Errorf("Agent Slug = %q, want %q", agent.Slug, "test-agent")
	}
	if agent.Lifecycle() != api.LifecyclePending {
		t.Errorf("Agent Lifecycle = %q, want %q", agent.Lifecycle(), api.LifecyclePending)
	}
	if agent.Presence() != api.PresenceOffline {
		t.Errorf("Agent Presence = %q, want %q", agent.Presence(), api.PresenceOffline)
	}
}

func TestManagerAddValidation(t *testing.T) {
	mgr := NewManager(event.NewNoopDispatcher())
	ctx := context.Background()

	// Missing required fields
	_, err := mgr.Add(ctx, api.Agent{
		Name: "test-agent",
		// Missing Adapter and RepoRoot
	})
	if err == nil {
		t.Error("Add() should fail with missing required fields")
	}
}

func TestManagerAddSlugNormalization(t *testing.T) {
	mgr := NewManager(event.NewNoopDispatcher())
	ctx := context.Background()

	agent, err := mgr.Add(ctx, api.Agent{
		Name:     "My Agent Name!",
		Adapter:  "claude-code",
		RepoRoot: "/home/user/project",
	})

	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	if agent.Slug != "my-agent-name" {
		t.Errorf("Agent Slug = %q, want %q (normalized)", agent.Slug, "my-agent-name")
	}
}

func TestManagerAddSlugCollision(t *testing.T) {
	mgr := NewManager(event.NewNoopDispatcher())
	ctx := context.Background()

	// Add first agent
	agent1, err := mgr.Add(ctx, api.Agent{
		Name:     "frontend",
		Adapter:  "claude-code",
		RepoRoot: "/home/user/project",
	})
	if err != nil {
		t.Fatalf("Add() first agent failed: %v", err)
	}
	if agent1.Slug != "frontend" {
		t.Errorf("First agent Slug = %q, want %q", agent1.Slug, "frontend")
	}

	// Add second agent with same name - should get unique slug
	agent2, err := mgr.Add(ctx, api.Agent{
		Name:     "frontend",
		Adapter:  "cursor",
		RepoRoot: "/home/user/project",
	})
	if err != nil {
		t.Fatalf("Add() second agent failed: %v", err)
	}
	if agent2.Slug != "frontend-2" {
		t.Errorf("Second agent Slug = %q, want %q", agent2.Slug, "frontend-2")
	}
}

func TestManagerRemove(t *testing.T) {
	mgr := NewManager(event.NewNoopDispatcher())
	ctx := context.Background()

	agent, _ := mgr.Add(ctx, api.Agent{
		Name:     "test-agent",
		Adapter:  "claude-code",
		RepoRoot: "/home/user/project",
	})
	id := agent.ID
	slug := agent.Slug

	// Verify agent exists
	if got := mgr.Get(id); got == nil {
		t.Fatal("Agent should exist after Add()")
	}
	if !mgr.SlugExists(slug) {
		t.Fatal("Slug should exist after Add()")
	}

	// Remove the agent
	if err := mgr.Remove(ctx, id); err != nil {
		t.Fatalf("Remove() failed: %v", err)
	}

	// Verify agent is removed
	if got := mgr.Get(id); got != nil {
		t.Error("Agent should not exist after Remove()")
	}
	if mgr.SlugExists(slug) {
		t.Error("Slug should not exist after Remove()")
	}
}

func TestManagerGet(t *testing.T) {
	mgr := NewManager(event.NewNoopDispatcher())
	ctx := context.Background()

	agent, _ := mgr.Add(ctx, api.Agent{
		Name:     "test-agent",
		Adapter:  "claude-code",
		RepoRoot: "/home/user/project",
	})

	// Get existing agent
	got := mgr.Get(agent.ID)
	if got == nil {
		t.Fatal("Get() returned nil for existing agent")
	}
	if got.ID != agent.ID {
		t.Errorf("Get() returned wrong agent: got ID %d, want %d", got.ID, agent.ID)
	}

	// Get non-existing agent
	if mgr.Get(muid.MUID(999999)) != nil {
		t.Error("Get() should return nil for non-existing agent")
	}
}

func TestManagerGetBySlug(t *testing.T) {
	mgr := NewManager(event.NewNoopDispatcher())
	ctx := context.Background()

	agent, _ := mgr.Add(ctx, api.Agent{
		Name:     "test-agent",
		Adapter:  "claude-code",
		RepoRoot: "/home/user/project",
	})

	// Get existing agent by slug
	got := mgr.GetBySlug(agent.Slug)
	if got == nil {
		t.Fatal("GetBySlug() returned nil for existing agent")
	}
	if got.ID != agent.ID {
		t.Errorf("GetBySlug() returned wrong agent: got ID %d, want %d", got.ID, agent.ID)
	}

	// Get non-existing slug
	if mgr.GetBySlug("non-existing") != nil {
		t.Error("GetBySlug() should return nil for non-existing slug")
	}
}

func TestManagerList(t *testing.T) {
	mgr := NewManager(event.NewNoopDispatcher())
	ctx := context.Background()

	// Empty list
	if len(mgr.List()) != 0 {
		t.Error("List() should return empty slice for new manager")
	}

	// Add agents
	_, _ = mgr.Add(ctx, api.Agent{Name: "a1", Adapter: "claude-code", RepoRoot: "/r"})
	_, _ = mgr.Add(ctx, api.Agent{Name: "a2", Adapter: "cursor", RepoRoot: "/r"})
	_, _ = mgr.Add(ctx, api.Agent{Name: "a3", Adapter: "windsurf", RepoRoot: "/r"})

	if len(mgr.List()) != 3 {
		t.Errorf("List() returned %d agents, want 3", len(mgr.List()))
	}
}

func TestManagerRoster(t *testing.T) {
	mgr := NewManager(event.NewNoopDispatcher())
	ctx := context.Background()

	_, _ = mgr.Add(ctx, api.Agent{Name: "a1", Adapter: "claude-code", RepoRoot: "/r"})
	_, _ = mgr.Add(ctx, api.Agent{Name: "a2", Adapter: "cursor", RepoRoot: "/r"})

	roster := mgr.Roster()
	if len(roster) != 2 {
		t.Errorf("Roster() returned %d entries, want 2", len(roster))
	}

	for _, entry := range roster {
		if entry.Lifecycle != api.LifecyclePending {
			t.Errorf("Roster entry Lifecycle = %q, want %q", entry.Lifecycle, api.LifecyclePending)
		}
		if entry.Presence != api.PresenceOffline {
			t.Errorf("Roster entry Presence = %q, want %q", entry.Presence, api.PresenceOffline)
		}
	}
}

func TestAgentLifecycle(t *testing.T) {
	agent := &Agent{
		lifecycle: api.LifecyclePending,
	}

	if agent.Lifecycle() != api.LifecyclePending {
		t.Errorf("Lifecycle() = %q, want %q", agent.Lifecycle(), api.LifecyclePending)
	}

	agent.SetLifecycle(api.LifecycleRunning)
	if agent.Lifecycle() != api.LifecycleRunning {
		t.Errorf("After SetLifecycle(Running), Lifecycle() = %q, want %q", agent.Lifecycle(), api.LifecycleRunning)
	}
}

func TestAgentPresence(t *testing.T) {
	agent := &Agent{
		presence: api.PresenceOffline,
	}

	if agent.Presence() != api.PresenceOffline {
		t.Errorf("Presence() = %q, want %q", agent.Presence(), api.PresenceOffline)
	}

	agent.SetPresence(api.PresenceOnline)
	if agent.Presence() != api.PresenceOnline {
		t.Errorf("After SetPresence(Online), Presence() = %q, want %q", agent.Presence(), api.PresenceOnline)
	}
}
