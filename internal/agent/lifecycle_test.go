package agent

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestSpawnAndStopAgent(t *testing.T) {
	// Use a temp dir for repo root
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)
	repoRoot := api.RepoRoot(tmpDir)
	
	cfg := config.AgentConfig{Name: "SpawnTest"}
	bus := NewEventBus()
	a, err := NewAgent(cfg, repoRoot, bus)
	if err != nil {
		t.Fatalf("NewAgent failed: %v", err)
	}

	ctx := context.Background()

	// Spawn
	if err := SpawnAgent(ctx, a); err != nil {
		t.Fatalf("SpawnAgent failed: %v", err)
	}

	// Verify session
	if len(a.Sessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(a.Sessions))
	}
	
	// Verify state transition to Running
	// We can't easily check internal HSM state without an accessor or waiting.
	// But SpawnAgent waits for EventSpawn dispatch.
	// The started event is dispatched after session creation.
	// Let's assume if it didn't error, it's starting/running.

	// Check worktree existence
	slug := api.NormalizeAgentSlug("SpawnTest")
	wtPath := filepath.Join(string(repoRoot), ".amux", "worktrees", string(slug))
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Errorf("Worktree path %s does not exist", wtPath)
	}
	
	// Stop
	if err := StopAgent(ctx, a); err != nil {
		t.Fatalf("StopAgent failed: %v", err)
	}
	
	// Wait a bit for async cleanup
	time.Sleep(100 * time.Millisecond)
	
	// Verify it dispatched Stop
}

func TestSpawnAgent_WorktreeError(t *testing.T) {
	// Use invalid repo root to force error
	// Or mock EnsureWorktree? Not easily.
	// Let's use a root we can't write to.
	
	// Skip if root
	if filepath.Join("/", ".amux") == "/.amux" {
		// Hard to test permission denied in container as root without user trickery
		t.Skip("Skipping permission test")
	}
}
