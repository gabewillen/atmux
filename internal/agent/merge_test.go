package agent

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestSelectMergeStrategy(t *testing.T) {
	// Setup a dummy git repo
	tmpRepo := t.TempDir()
	
	// Init git
	cmd := exec.Command("git", "init", "-b", "main")
	cmd.Dir = tmpRepo
	if err := cmd.Run(); err != nil {
		// Fallback for older git
		exec.Command("git", "init", tmpRepo).Run()
		exec.Command("git", "-C", tmpRepo, "checkout", "-b", "main").Run()
	}
	
	// Need a commit for HEAD to be valid?
	// git symbolic-ref HEAD works even if unborn?
	// Actually symbolic-ref works on unborn branches.
	
	repoRoot := api.RepoRoot(tmpRepo)

	// Test 1: Configured target branch
	cfg1 := config.GitConfig{
		Merge: config.MergeConfig{
			Strategy:     "rebase",
			TargetBranch: "develop",
		},
	}
	strat, target, err := SelectMergeStrategy(cfg1, repoRoot)
	if err != nil {
		t.Fatalf("SelectMergeStrategy failed: %v", err)
	}
	if strat != MergeRebase {
		t.Errorf("Expected rebase, got %s", strat)
	}
	if target != "develop" {
		t.Errorf("Expected develop, got %s", target)
	}

	// Test 2: Auto-detect base branch
	cfg2 := config.GitConfig{
		Merge: config.MergeConfig{
			Strategy: "squash",
		},
	}
	strat, target, err = SelectMergeStrategy(cfg2, repoRoot)
	if err != nil {
		t.Fatalf("SelectMergeStrategy failed: %v", err)
	}
	if strat != MergeSquash {
		t.Errorf("Expected squash, got %s", strat)
	}
	if target != "main" {
		t.Errorf("Expected main, got %s", target)
	}
}

func TestSelectMergeStrategy_DetachedHead(t *testing.T) {
	tmpRepo := t.TempDir()
	
	// Init and commit
	exec.Command("git", "init", tmpRepo).Run()
	configCmd(tmpRepo, "user.email", "test@example.com")
	configCmd(tmpRepo, "user.name", "Test")
	
	// Create a commit
	f := filepath.Join(tmpRepo, "file")
	os.WriteFile(f, []byte("content"), 0644)
	exec.Command("git", "-C", tmpRepo, "add", ".").Run()
	exec.Command("git", "-C", tmpRepo, "commit", "-m", "init").Run()
	
	// Detach head
	exec.Command("git", "-C", tmpRepo, "checkout", "--detach", "HEAD").Run()
	
	repoRoot := api.RepoRoot(tmpRepo)
	cfg := config.GitConfig{
		Merge: config.MergeConfig{Strategy: "squash"},
	}
	
	_, _, err := SelectMergeStrategy(cfg, repoRoot)
	if err == nil {
		t.Error("Expected error for detached HEAD without configured target")
	}
}

func configCmd(dir, key, val string) {
	exec.Command("git", "-C", dir, "config", key, val).Run()
}
