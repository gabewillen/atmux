package agent

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/stateforward/amux/internal/config"
	"github.com/stateforward/amux/internal/paths"
)

func TestAddLocalAgentCreatesWorktreeAndConfig(t *testing.T) {
	ctx := context.Background()

	repoDir := t.TempDir()

	// Initialize a git repository
	cmd := exec.Command("git", "init", repoDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	// Canonicalize repo root via paths resolver
	resolver, err := paths.NewResolver(repoDir)
	if err != nil {
		t.Fatalf("NewResolver failed: %v", err)
	}

	cfg := config.DefaultConfig()

	agent, slug, err := AddLocalAgent(ctx, cfg, AddLocalAgentOptions{
		Name:    "Backend Dev",
		About:   "Backend developer agent",
		Adapter: "test-adapter",
		RepoRoot: repoDir,
	})
	if err != nil {
		t.Fatalf("AddLocalAgent returned error: %v", err)
	}

	if agent == nil {
		t.Fatal("expected non-nil agent")
	}

	if slug == "" {
		t.Error("expected non-empty slug")
	}

	// Verify worktree path matches resolver
	expectedWorktree := resolver.WorktreePath(slug)
	if agent.Worktree != expectedWorktree {
		t.Fatalf("agent.Worktree = %q, want %q", agent.Worktree, expectedWorktree)
	}

	// Verify the worktree directory exists on disk
	if _, err := os.Stat(expectedWorktree); err != nil {
		t.Fatalf("expected worktree directory to exist at %q: %v", expectedWorktree, err)
	}

	// Verify config contains the agent entry
	if len(cfg.Agents) != 1 {
		t.Fatalf("expected 1 agent in config, got %d", len(cfg.Agents))
	}

	if cfg.Agents[0].Name != "Backend Dev" {
		t.Errorf("config agent name = %q, want %q", cfg.Agents[0].Name, "Backend Dev")
	}

	// Ensure slug uniqueness logic will generate a different slug for a second agent
	_, slug2, err := AddLocalAgent(ctx, cfg, AddLocalAgentOptions{
		Name:    "Backend Dev", // same name
		About:   "Second backend agent",
		Adapter: "test-adapter",
		RepoRoot: repoDir,
	})
	if err != nil {
		t.Fatalf("AddLocalAgent (second) returned error: %v", err)
	}

	if slug2 == slug {
		t.Errorf("expected second slug to differ from first; got %q", slug2)
	}

	// Start a local PTY-backed session and verify it uses the agent worktree
	session, err := StartLocalSession(ctx, agent, []string{"sh", "-c", "echo ready"}, nil)
	if err != nil {
		t.Fatalf("StartLocalSession returned error: %v", err)
	}
	defer session.Stop()

	if session.Cmd == nil {
		t.Fatal("expected non-nil session.Cmd")
	}
	if session.PTY == nil {
		t.Fatal("expected non-nil session.PTY")
	}
	if session.Cmd.Dir != agent.Worktree {
		t.Errorf("session.Cmd.Dir = %q, want %q", session.Cmd.Dir, agent.Worktree)
	}
}

func TestSelectMergeStrategyDefaultsAndMapping(t *testing.T) {
	cfg := config.DefaultConfig()

	// Default config uses squash
	if got := SelectMergeStrategy(cfg); got != MergeStrategySquash {
		t.Fatalf("default strategy = %q, want %q", got, MergeStrategySquash)
	}

	cfg.Git.Merge.Strategy = "merge-commit"
	if got := SelectMergeStrategy(cfg); got != MergeStrategyMergeCommit {
		t.Errorf("strategy 'merge-commit' mapped to %q, want %q", got, MergeStrategyMergeCommit)
	}

	cfg.Git.Merge.Strategy = "rebase"
	if got := SelectMergeStrategy(cfg); got != MergeStrategyRebase {
		t.Errorf("strategy 'rebase' mapped to %q, want %q", got, MergeStrategyRebase)
	}

	cfg.Git.Merge.Strategy = "ff-only"
	if got := SelectMergeStrategy(cfg); got != MergeStrategyFFOnly {
		t.Errorf("strategy 'ff-only' mapped to %q, want %q", got, MergeStrategyFFOnly)
	}

	cfg.Git.Merge.Strategy = "unknown"
	if got := SelectMergeStrategy(cfg); got != MergeStrategySquash {
		t.Errorf("unknown strategy mapped to %q, want default %q", got, MergeStrategySquash)
	}
}
