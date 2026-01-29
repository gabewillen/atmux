package worktree

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// initTestRepo creates a temporary git repository for testing.
func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Initialize a git repo with an initial commit
	commands := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}
	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git init command %v failed: %v\n%s", args, err, output)
		}
	}

	// Create a file and commit
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

func TestCreate(t *testing.T) {
	repoRoot := initTestRepo(t)
	mgr := NewManager()

	wtDir, err := mgr.Create(repoRoot, "test-agent")
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Verify worktree directory exists
	if _, err := os.Stat(wtDir); os.IsNotExist(err) {
		t.Errorf("worktree directory does not exist: %s", wtDir)
	}

	// Verify path matches expected pattern
	expected := filepath.Join(repoRoot, ".amux", "worktrees", "test-agent")
	if wtDir != expected {
		t.Errorf("worktree path = %q, want %q", wtDir, expected)
	}

	// Verify .git file exists (worktree indicator)
	gitFile := filepath.Join(wtDir, ".git")
	info, err := os.Stat(gitFile)
	if err != nil {
		t.Fatalf(".git file not found in worktree: %v", err)
	}
	if info.IsDir() {
		t.Error(".git should be a file in a worktree, not a directory")
	}
}

func TestCreateBranchNaming(t *testing.T) {
	repoRoot := initTestRepo(t)
	mgr := NewManager()

	_, err := mgr.Create(repoRoot, "my-feature")
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Verify the branch amux/my-feature was created
	if !mgr.branchExists(repoRoot, "amux/my-feature") {
		t.Error("branch amux/my-feature should exist after worktree creation")
	}
}

func TestCreateIdempotent(t *testing.T) {
	repoRoot := initTestRepo(t)
	mgr := NewManager()

	// First creation
	wtDir1, err := mgr.Create(repoRoot, "test-agent")
	if err != nil {
		t.Fatalf("First Create() failed: %v", err)
	}

	// Second creation should return same path without error
	wtDir2, err := mgr.Create(repoRoot, "test-agent")
	if err != nil {
		t.Fatalf("Second Create() failed: %v", err)
	}

	if wtDir1 != wtDir2 {
		t.Errorf("idempotent Create() returned different paths: %q vs %q", wtDir1, wtDir2)
	}
}

func TestCreateNoRepo(t *testing.T) {
	mgr := NewManager()

	_, err := mgr.Create("", "test")
	if err == nil {
		t.Error("Create() with empty repoRoot should fail")
	}
}

func TestCreateNotGitRepo(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager()

	_, err := mgr.Create(dir, "test")
	if err == nil {
		t.Error("Create() in non-git directory should fail")
	}
}

func TestRemove(t *testing.T) {
	repoRoot := initTestRepo(t)
	mgr := NewManager()

	// Create a worktree
	_, err := mgr.Create(repoRoot, "test-agent")
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Remove the worktree (preserve branch)
	if err := mgr.Remove(repoRoot, "test-agent", false); err != nil {
		t.Fatalf("Remove() failed: %v", err)
	}

	// Verify worktree directory is gone
	wtDir := filepath.Join(repoRoot, ".amux", "worktrees", "test-agent")
	if _, err := os.Stat(wtDir); !os.IsNotExist(err) {
		t.Error("worktree directory should not exist after Remove()")
	}

	// Branch should still exist (deleteBranch=false)
	if !mgr.branchExists(repoRoot, "amux/test-agent") {
		t.Error("branch should be preserved when deleteBranch=false")
	}
}

func TestRemoveWithBranchDelete(t *testing.T) {
	repoRoot := initTestRepo(t)
	mgr := NewManager()

	_, err := mgr.Create(repoRoot, "test-agent")
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Remove with branch deletion
	if err := mgr.Remove(repoRoot, "test-agent", true); err != nil {
		t.Fatalf("Remove() with branch delete failed: %v", err)
	}

	// Branch should be deleted
	if mgr.branchExists(repoRoot, "amux/test-agent") {
		t.Error("branch should be deleted when deleteBranch=true")
	}
}

func TestExists(t *testing.T) {
	repoRoot := initTestRepo(t)
	mgr := NewManager()

	if mgr.Exists(repoRoot, "test-agent") {
		t.Error("Exists() should return false before creation")
	}

	_, err := mgr.Create(repoRoot, "test-agent")
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if !mgr.Exists(repoRoot, "test-agent") {
		t.Error("Exists() should return true after creation")
	}
}

func TestPath(t *testing.T) {
	mgr := NewManager()

	path := mgr.Path("/home/user/project", "my-agent")
	expected := "/home/user/project/.amux/worktrees/my-agent"
	if path != expected {
		t.Errorf("Path() = %q, want %q", path, expected)
	}
}

func TestBranchName(t *testing.T) {
	tests := []struct {
		slug string
		want string
	}{
		{"frontend-dev", "amux/frontend-dev"},
		{"backend", "amux/backend"},
		{"test-runner", "amux/test-runner"},
	}

	for _, tt := range tests {
		got := BranchName(tt.slug)
		if got != tt.want {
			t.Errorf("BranchName(%q) = %q, want %q", tt.slug, got, tt.want)
		}
	}
}

func TestIsDirty(t *testing.T) {
	repoRoot := initTestRepo(t)
	mgr := NewManager()

	_, err := mgr.Create(repoRoot, "test-agent")
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Fresh worktree should be clean
	dirty, err := mgr.IsDirty(repoRoot, "test-agent")
	if err != nil {
		t.Fatalf("IsDirty() failed: %v", err)
	}
	if dirty {
		t.Error("fresh worktree should not be dirty")
	}

	// Make the worktree dirty
	wtDir := filepath.Join(repoRoot, ".amux", "worktrees", "test-agent")
	if err := os.WriteFile(filepath.Join(wtDir, "newfile.txt"), []byte("dirty"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	dirty, err = mgr.IsDirty(repoRoot, "test-agent")
	if err != nil {
		t.Fatalf("IsDirty() failed: %v", err)
	}
	if !dirty {
		t.Error("worktree with uncommitted changes should be dirty")
	}
}

func TestBaseBranch(t *testing.T) {
	repoRoot := initTestRepo(t)
	mgr := NewManager()

	branch, err := mgr.BaseBranch(repoRoot)
	if err != nil {
		t.Fatalf("BaseBranch() failed: %v", err)
	}

	// Default branch should be "main" or "master"
	if branch != "main" && branch != "master" {
		t.Errorf("BaseBranch() = %q, want 'main' or 'master'", branch)
	}
}

func TestMultipleWorktrees(t *testing.T) {
	repoRoot := initTestRepo(t)
	mgr := NewManager()

	slugs := []string{"frontend", "backend", "testing"}
	for _, slug := range slugs {
		_, err := mgr.Create(repoRoot, slug)
		if err != nil {
			t.Fatalf("Create(%q) failed: %v", slug, err)
		}
	}

	// All should exist
	for _, slug := range slugs {
		if !mgr.Exists(repoRoot, slug) {
			t.Errorf("worktree %q should exist", slug)
		}
		if !mgr.branchExists(repoRoot, BranchPrefix+slug) {
			t.Errorf("branch amux/%s should exist", slug)
		}
	}

	// Remove one
	if err := mgr.Remove(repoRoot, "backend", false); err != nil {
		t.Fatalf("Remove(backend) failed: %v", err)
	}

	// Others should still exist
	if !mgr.Exists(repoRoot, "frontend") {
		t.Error("frontend worktree should still exist")
	}
	if !mgr.Exists(repoRoot, "testing") {
		t.Error("testing worktree should still exist")
	}
	if mgr.Exists(repoRoot, "backend") {
		t.Error("backend worktree should not exist after removal")
	}
}
