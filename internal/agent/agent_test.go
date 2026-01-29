package agent

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stateforward/hsm-go/muid"

	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/pkg/api"
)

// initTestRepo creates a temporary git repository for testing.
func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	commands := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}
	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git command %v failed: %v\n%s", args, err, output)
		}
	}

	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Test\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = dir
	if output, err := addCmd.CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %v\n%s", err, output)
	}
	commitCmd := exec.Command("git", "commit", "-m", "initial commit")
	commitCmd.Dir = dir
	if output, err := commitCmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit failed: %v\n%s", err, output)
	}

	return dir
}

func TestManagerAdd(t *testing.T) {
	repoRoot := initTestRepo(t)
	mgr := NewManager(event.NewNoopDispatcher())
	ctx := context.Background()

	agent, err := mgr.Add(ctx, api.Agent{
		Name:     "test-agent",
		Adapter:  "claude-code",
		RepoRoot: repoRoot,
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
	if agent.Worktree == "" {
		t.Error("Agent Worktree should be set")
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

	// Missing adapter - should fail validation
	_, err := mgr.Add(ctx, api.Agent{
		Name: "test-agent",
		// Missing Adapter
	})
	if err == nil {
		t.Error("Add() should fail with missing required fields")
	}
}

func TestManagerAddSlugNormalization(t *testing.T) {
	repoRoot := initTestRepo(t)
	mgr := NewManager(event.NewNoopDispatcher())
	ctx := context.Background()

	agent, err := mgr.Add(ctx, api.Agent{
		Name:     "My Agent Name!",
		Adapter:  "claude-code",
		RepoRoot: repoRoot,
	})

	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	if agent.Slug != "my-agent-name" {
		t.Errorf("Agent Slug = %q, want %q (normalized)", agent.Slug, "my-agent-name")
	}
}

func TestManagerAddSlugCollision(t *testing.T) {
	repoRoot := initTestRepo(t)
	mgr := NewManager(event.NewNoopDispatcher())
	ctx := context.Background()

	// Add first agent
	agent1, err := mgr.Add(ctx, api.Agent{
		Name:     "frontend",
		Adapter:  "claude-code",
		RepoRoot: repoRoot,
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
		RepoRoot: repoRoot,
	})
	if err != nil {
		t.Fatalf("Add() second agent failed: %v", err)
	}
	if agent2.Slug != "frontend-2" {
		t.Errorf("Second agent Slug = %q, want %q", agent2.Slug, "frontend-2")
	}
}

func TestManagerRemove(t *testing.T) {
	repoRoot := initTestRepo(t)
	mgr := NewManager(event.NewNoopDispatcher())
	ctx := context.Background()

	agent, err := mgr.Add(ctx, api.Agent{
		Name:     "test-agent",
		Adapter:  "claude-code",
		RepoRoot: repoRoot,
	})
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}
	id := agent.ID
	slug := agent.Slug

	// Verify agent exists
	if got := mgr.Get(id); got == nil {
		t.Fatal("Agent should exist after Add()")
	}
	if !mgr.SlugExists(slug) {
		t.Fatal("Slug should exist after Add()")
	}

	// Remove the agent (preserve branch)
	if err := mgr.Remove(ctx, id, false); err != nil {
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

func TestManagerRemoveNotFound(t *testing.T) {
	mgr := NewManager(event.NewNoopDispatcher())
	ctx := context.Background()

	err := mgr.Remove(ctx, muid.MUID(999999), false)
	if err == nil {
		t.Error("Remove() should fail for non-existing agent")
	}
}

func TestManagerGet(t *testing.T) {
	repoRoot := initTestRepo(t)
	mgr := NewManager(event.NewNoopDispatcher())
	ctx := context.Background()

	agent, err := mgr.Add(ctx, api.Agent{
		Name:     "test-agent",
		Adapter:  "claude-code",
		RepoRoot: repoRoot,
	})
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

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
	repoRoot := initTestRepo(t)
	mgr := NewManager(event.NewNoopDispatcher())
	ctx := context.Background()

	agent, err := mgr.Add(ctx, api.Agent{
		Name:     "test-agent",
		Adapter:  "claude-code",
		RepoRoot: repoRoot,
	})
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

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
	repoRoot := initTestRepo(t)
	mgr := NewManager(event.NewNoopDispatcher())
	ctx := context.Background()

	// Empty list
	if len(mgr.List()) != 0 {
		t.Error("List() should return empty slice for new manager")
	}

	// Add agents
	_, _ = mgr.Add(ctx, api.Agent{Name: "a1", Adapter: "claude-code", RepoRoot: repoRoot})
	_, _ = mgr.Add(ctx, api.Agent{Name: "a2", Adapter: "cursor", RepoRoot: repoRoot})
	_, _ = mgr.Add(ctx, api.Agent{Name: "a3", Adapter: "windsurf", RepoRoot: repoRoot})

	if len(mgr.List()) != 3 {
		t.Errorf("List() returned %d agents, want 3", len(mgr.List()))
	}
}

func TestManagerRoster(t *testing.T) {
	repoRoot := initTestRepo(t)
	mgr := NewManager(event.NewNoopDispatcher())
	ctx := context.Background()

	_, _ = mgr.Add(ctx, api.Agent{Name: "a1", Adapter: "claude-code", RepoRoot: repoRoot})
	_, _ = mgr.Add(ctx, api.Agent{Name: "a2", Adapter: "cursor", RepoRoot: repoRoot})

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

func TestManagerAddWithWorktree(t *testing.T) {
	repoRoot := initTestRepo(t)

	resolver := &paths.Resolver{}
	_ = resolver.SetRepoRoot(repoRoot)

	mgr := NewManagerWithResolver(event.NewNoopDispatcher(), resolver)
	ctx := context.Background()

	agent, err := mgr.Add(ctx, api.Agent{
		Name:     "test-agent",
		Adapter:  "claude-code",
		RepoRoot: repoRoot,
	})
	if err != nil {
		t.Fatalf("Add() with worktree failed: %v", err)
	}

	// Verify worktree was created
	if agent.Worktree == "" {
		t.Error("Agent Worktree should be set after Add()")
	}

	expectedWT := filepath.Join(repoRoot, ".amux", "worktrees", "test-agent")
	if agent.Worktree != expectedWT {
		t.Errorf("Agent Worktree = %q, want %q", agent.Worktree, expectedWT)
	}

	// Verify worktree directory exists
	if _, err := os.Stat(agent.Worktree); os.IsNotExist(err) {
		t.Error("worktree directory should exist after Add()")
	}
}

func TestManagerAddSSHRequiresRepoPath(t *testing.T) {
	mgr := NewManager(event.NewNoopDispatcher())
	ctx := context.Background()

	_, err := mgr.Add(ctx, api.Agent{
		Name:    "remote-agent",
		Adapter: "claude-code",
		Location: api.Location{
			Type: api.LocationSSH,
			Host: "remote-host",
			// Missing RepoPath
		},
	})
	if err == nil {
		t.Error("Add() SSH agent without repo_path should fail")
	}
}

func TestManagerBaseBranch(t *testing.T) {
	repoRoot := initTestRepo(t)

	resolver := &paths.Resolver{}
	_ = resolver.SetRepoRoot(repoRoot)

	mgr := NewManagerWithResolver(event.NewNoopDispatcher(), resolver)
	ctx := context.Background()

	// Before adding any agent, no base branch should be recorded
	_, ok := mgr.BaseBranch(repoRoot)
	if ok {
		t.Error("BaseBranch should not be recorded before adding agents")
	}

	// Add an agent to trigger base branch recording
	_, err := mgr.Add(ctx, api.Agent{
		Name:     "test-agent",
		Adapter:  "claude-code",
		RepoRoot: repoRoot,
	})
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	// Base branch should now be recorded
	branch, ok := mgr.BaseBranch(repoRoot)
	if !ok {
		t.Error("BaseBranch should be recorded after adding an agent")
	}
	if branch != "main" && branch != "master" {
		t.Errorf("BaseBranch = %q, want 'main' or 'master'", branch)
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
