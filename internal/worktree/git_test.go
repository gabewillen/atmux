package worktree

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func createTestRepo(t *testing.T) string {
	dir := t.TempDir()

	// Init git repo
	mp := []string{
		"init",
		"config user.email 'test@example.com'",
		"config user.name 'Test User'",
		"commit --allow-empty -m 'Initial commit'",
	}

	for _, cmdStr := range mp {
		args := strings.Split(cmdStr, " ")
		// handle quoted args poorly but sufficient for these simple ones
		// Actually "config user.email '...'" is hard to split by space.
		// Manual execution:
		var cmd *exec.Cmd
		if strings.HasPrefix(cmdStr, "config") {
			if strings.Contains(cmdStr, "user.email") {
				cmd = exec.Command("git", "config", "user.email", "test@example.com")
			} else {
				cmd = exec.Command("git", "config", "user.name", "Test User")
			}
		} else if strings.HasPrefix(cmdStr, "commit") {
			cmd = exec.Command("git", "commit", "--allow-empty", "-m", "Initial commit")
		} else {
			cmd = exec.Command("git", args...)
		}
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("failed to run git %s: %s", cmdStr, string(out))
		}
	}
	return dir
}

func TestIsDirty(t *testing.T) {
	repo := createTestRepo(t)

	// Initially clean
	dirty, err := IsDirty(repo)
	if err != nil {
		t.Fatalf("IsDirty failed: %v", err)
	}
	if dirty {
		t.Errorf("expected clean repo, got dirty")
	}

	// Create file (unstaged)
	os.WriteFile(filepath.Join(repo, "foo.txt"), []byte("foo"), 0644)
	dirty, err = IsDirty(repo)
	if err != nil {
		t.Fatalf("IsDirty failed: %v", err)
	}
	if !dirty {
		t.Errorf("expected dirty repo (unstaged), got clean")
	}

	// Stage file
	exec.Command("git", "-C", repo, "add", "foo.txt").Run()
	dirty, err = IsDirty(repo)
	if err != nil {
		t.Fatalf("IsDirty failed: %v", err)
	}
	if !dirty {
		t.Errorf("expected dirty repo (staged), got clean")
	}

	// Commit
	exec.Command("git", "-C", repo, "commit", "-m", "add foo").Run()
	dirty, err = IsDirty(repo)
	if err != nil {
		t.Fatalf("IsDirty failed: %v", err)
	}
	if dirty {
		t.Errorf("expected clean repo after commit, got dirty")
	}
}

func TestMerge(t *testing.T) {
	repo := createTestRepo(t)

	// Create feature branch
	if err := EnsureBranch(repo, "feature", "HEAD"); err != nil {
		t.Fatalf("EnsureBranch failed: %v", err)
	}

	// Switch to feature and adding commit
	// Since we are simulating, we can just use git commands directly
	exec.Command("git", "-C", repo, "checkout", "feature").Run()
	os.WriteFile(filepath.Join(repo, "bar.txt"), []byte("bar"), 0644)
	exec.Command("git", "-C", repo, "add", "bar.txt").Run()
	exec.Command("git", "-C", repo, "commit", "-m", "add bar").Run()

	// Switch back to master/main
	// Check what head branch is
	head, _ := GetHeadBranch(repo) // likely master or main
	// Actually createTestRepo returns a repo on 'master' or 'main'.
	// init might default to master.

	// But we need to switch back to base to merge.
	// In the real world agent worktree is separate.
	// Merge() assumes we are in the target repo/worktree.
	// Wait, Merge(repoPath, branch, strategy) runs `git merge <branch>` in `repoPath`.
	// So `repoPath` must be checked out to the TARGET branch (e.g. main).
	// In `MergeAgent`, `agent.RepoRoot` is the main checkout?
	// The agent's worktree is separate.
	// So `agent.RepoRoot` is typically the "base" repo.

	// Ensure we are on main
	// We were on feature.
	cmd := exec.Command("git", "checkout", "master")
	if head != "" && head != "feature" {
		cmd = exec.Command("git", "checkout", head)
	}
	cmd.Dir = repo
	if out, err := cmd.CombinedOutput(); err != nil {
		// try main
		cmd = exec.Command("git", "checkout", "main")
		cmd.Dir = repo
		if out2, err2 := cmd.CombinedOutput(); err2 != nil {
			t.Fatalf("failed to checkout master/main: %s / %s", string(out), string(out2))
		}
	}

	// Test Squash Merge
	if err := Merge(repo, "feature", "squash"); err != nil {
		t.Errorf("Merge squash failed: %v", err)
	}

	// Verify staged changes
	dirty, _ := IsDirty(repo)
	if !dirty {
		t.Errorf("expected dirty (staged) after squash merge")
	}

	// Reset
	exec.Command("git", "-C", repo, "reset", "--hard", "HEAD").Run()

	// Test Merge Commit (no-ff)
	// We need config user again? (Env might bleed but set locally in createTestRepo helps)
	if err := Merge(repo, "feature", "merge-commit"); err != nil {
		t.Errorf("Merge merge-commit failed: %v", err)
	}

	// Verify committed (not dirty)
	dirty, _ = IsDirty(repo)
	if dirty {
		t.Errorf("expected clean after merge-commit")
	}
}
