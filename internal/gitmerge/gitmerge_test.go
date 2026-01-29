package gitmerge

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	amuxerrors "github.com/agentflare-ai/amux/internal/errors"
	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/internal/ids"
)

// initTestRepo creates a temporary git repository with an initial commit.
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
	gitRun(t, dir, "add", ".")
	gitRun(t, dir, "commit", "-m", "initial commit")

	return dir
}

// createWorktreeWithCommit creates a worktree branch with a new commit.
func createWorktreeWithCommit(t *testing.T, repoRoot, slug, filename, content string) {
	t.Helper()
	branch := "amux/" + slug

	// Create a branch from HEAD
	gitRun(t, repoRoot, "branch", branch)

	// Create worktree directory
	wtDir := filepath.Join(repoRoot, ".amux", "worktrees", slug)
	if err := os.MkdirAll(filepath.Dir(wtDir), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	gitRun(t, repoRoot, "worktree", "add", wtDir, branch)

	// Make a change in the worktree
	if err := os.WriteFile(filepath.Join(wtDir, filename), []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	gitRun(t, wtDir, "add", ".")
	gitRun(t, wtDir, "commit", "-m", "add "+filename)
}

func gitRun(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}
	return string(output)
}

func TestParseStrategy(t *testing.T) {
	tests := []struct {
		input string
		want  Strategy
		ok    bool
	}{
		{"merge-commit", StrategyMergeCommit, true},
		{"squash", StrategySquash, true},
		{"rebase", StrategyRebase, true},
		{"ff-only", StrategyFFOnly, true},
		{"invalid", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		got, err := ParseStrategy(tt.input)
		if tt.ok {
			if err != nil {
				t.Errorf("ParseStrategy(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseStrategy(%q) = %q, want %q", tt.input, got, tt.want)
			}
		} else {
			if err == nil {
				t.Errorf("ParseStrategy(%q) should fail", tt.input)
			}
		}
	}
}

func TestValidStrategies(t *testing.T) {
	strategies := ValidStrategies()
	if len(strategies) != 4 {
		t.Errorf("ValidStrategies() returned %d, want 4", len(strategies))
	}
}

func TestMergeCommit(t *testing.T) {
	repoRoot := initTestRepo(t)
	createWorktreeWithCommit(t, repoRoot, "test-agent", "feature.txt", "feature content\n")

	executor := NewExecutor(event.NewNoopDispatcher())
	ctx := context.Background()

	// Get the current branch as base
	baseBranch := gitBranch(t, repoRoot)

	result, err := executor.Execute(ctx, Request{
		RepoRoot:     repoRoot,
		AgentSlug:    "test-agent",
		Strategy:     StrategyMergeCommit,
		BaseBranch:   baseBranch,
		TargetBranch: baseBranch,
		AgentID:      ids.NewID(),
	})
	if err != nil {
		t.Fatalf("Execute(merge-commit) failed: %v", err)
	}

	if result.Strategy != StrategyMergeCommit {
		t.Errorf("result.Strategy = %q, want %q", result.Strategy, StrategyMergeCommit)
	}
	if result.CommitSHA == "" {
		t.Error("result.CommitSHA should not be empty")
	}
	if result.Conflict {
		t.Error("result.Conflict should be false")
	}

	// Verify the file exists on the target branch
	if _, err := os.Stat(filepath.Join(repoRoot, "feature.txt")); os.IsNotExist(err) {
		t.Error("feature.txt should exist after merge-commit")
	}
}

func TestSquash(t *testing.T) {
	repoRoot := initTestRepo(t)
	createWorktreeWithCommit(t, repoRoot, "test-agent", "feature.txt", "feature content\n")

	executor := NewExecutor(event.NewNoopDispatcher())
	ctx := context.Background()

	baseBranch := gitBranch(t, repoRoot)

	result, err := executor.Execute(ctx, Request{
		RepoRoot:     repoRoot,
		AgentSlug:    "test-agent",
		Strategy:     StrategySquash,
		BaseBranch:   baseBranch,
		TargetBranch: baseBranch,
		AgentID:      ids.NewID(),
	})
	if err != nil {
		t.Fatalf("Execute(squash) failed: %v", err)
	}

	if result.Strategy != StrategySquash {
		t.Errorf("result.Strategy = %q, want %q", result.Strategy, StrategySquash)
	}
	if result.Conflict {
		t.Error("result.Conflict should be false")
	}
}

func TestFFOnly(t *testing.T) {
	repoRoot := initTestRepo(t)
	createWorktreeWithCommit(t, repoRoot, "test-agent", "feature.txt", "feature content\n")

	executor := NewExecutor(event.NewNoopDispatcher())
	ctx := context.Background()

	baseBranch := gitBranch(t, repoRoot)

	result, err := executor.Execute(ctx, Request{
		RepoRoot:     repoRoot,
		AgentSlug:    "test-agent",
		Strategy:     StrategyFFOnly,
		BaseBranch:   baseBranch,
		TargetBranch: baseBranch,
		AgentID:      ids.NewID(),
	})
	if err != nil {
		t.Fatalf("Execute(ff-only) failed: %v", err)
	}

	if result.Strategy != StrategyFFOnly {
		t.Errorf("result.Strategy = %q, want %q", result.Strategy, StrategyFFOnly)
	}
}

func TestRebase(t *testing.T) {
	repoRoot := initTestRepo(t)
	createWorktreeWithCommit(t, repoRoot, "test-agent", "feature.txt", "feature content\n")

	executor := NewExecutor(event.NewNoopDispatcher())
	ctx := context.Background()

	baseBranch := gitBranch(t, repoRoot)

	result, err := executor.Execute(ctx, Request{
		RepoRoot:     repoRoot,
		AgentSlug:    "test-agent",
		Strategy:     StrategyRebase,
		BaseBranch:   baseBranch,
		TargetBranch: baseBranch,
		AgentID:      ids.NewID(),
	})
	if err != nil {
		t.Fatalf("Execute(rebase) failed: %v", err)
	}

	if result.Strategy != StrategyRebase {
		t.Errorf("result.Strategy = %q, want %q", result.Strategy, StrategyRebase)
	}
}

func TestMissingTargetBranch(t *testing.T) {
	repoRoot := initTestRepo(t)
	executor := NewExecutor(event.NewNoopDispatcher())
	ctx := context.Background()

	_, err := executor.Execute(ctx, Request{
		RepoRoot:  repoRoot,
		AgentSlug: "test-agent",
		Strategy:  StrategySquash,
		// No TargetBranch or BaseBranch
		AgentID: ids.NewID(),
	})
	if err == nil {
		t.Error("Execute() without target/base branch should fail")
	}
}

func TestPreconditionBranchNotFound(t *testing.T) {
	repoRoot := initTestRepo(t)
	executor := NewExecutor(event.NewNoopDispatcher())
	ctx := context.Background()

	baseBranch := gitBranch(t, repoRoot)

	_, err := executor.Execute(ctx, Request{
		RepoRoot:     repoRoot,
		AgentSlug:    "nonexistent",
		Strategy:     StrategySquash,
		BaseBranch:   baseBranch,
		TargetBranch: baseBranch,
		AgentID:      ids.NewID(),
	})
	if err == nil {
		t.Error("Execute() with non-existent source branch should fail")
	}
	if !amuxerrors.Is(err, amuxerrors.ErrBranchNotFound) {
		t.Errorf("error should wrap ErrBranchNotFound, got: %v", err)
	}
}

func TestPreconditionDirtyWorktree(t *testing.T) {
	repoRoot := initTestRepo(t)
	createWorktreeWithCommit(t, repoRoot, "test-agent", "feature.txt", "content\n")

	// Make the worktree dirty
	wtDir := filepath.Join(repoRoot, ".amux", "worktrees", "test-agent")
	if err := os.WriteFile(filepath.Join(wtDir, "dirty.txt"), []byte("dirty"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	executor := NewExecutor(event.NewNoopDispatcher())
	ctx := context.Background()

	baseBranch := gitBranch(t, repoRoot)

	_, err := executor.Execute(ctx, Request{
		RepoRoot:     repoRoot,
		AgentSlug:    "test-agent",
		Strategy:     StrategySquash,
		BaseBranch:   baseBranch,
		TargetBranch: baseBranch,
		AllowDirty:   false,
		AgentID:      ids.NewID(),
	})
	if err == nil {
		t.Error("Execute() with dirty worktree and AllowDirty=false should fail")
	}
}

func TestPreconditionDirtyWorktreeAllowed(t *testing.T) {
	repoRoot := initTestRepo(t)
	createWorktreeWithCommit(t, repoRoot, "test-agent", "feature.txt", "content\n")

	// Make the worktree dirty
	wtDir := filepath.Join(repoRoot, ".amux", "worktrees", "test-agent")
	if err := os.WriteFile(filepath.Join(wtDir, "dirty.txt"), []byte("dirty"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	executor := NewExecutor(event.NewNoopDispatcher())
	ctx := context.Background()

	baseBranch := gitBranch(t, repoRoot)

	// Should succeed with AllowDirty=true
	_, err := executor.Execute(ctx, Request{
		RepoRoot:     repoRoot,
		AgentSlug:    "test-agent",
		Strategy:     StrategySquash,
		BaseBranch:   baseBranch,
		TargetBranch: baseBranch,
		AllowDirty:   true,
		AgentID:      ids.NewID(),
	})
	if err != nil {
		t.Fatalf("Execute() with AllowDirty=true should succeed: %v", err)
	}
}

func TestMergeEvents(t *testing.T) {
	repoRoot := initTestRepo(t)
	createWorktreeWithCommit(t, repoRoot, "test-agent", "feature.txt", "content\n")

	dispatcher := event.NewLocalDispatcher()
	executor := NewExecutor(dispatcher)
	ctx := context.Background()

	var receivedTypes []event.Type
	dispatcher.Subscribe(event.Subscription{
		Handler: func(ctx context.Context, evt event.Event) error {
			receivedTypes = append(receivedTypes, evt.Type)
			return nil
		},
	})

	baseBranch := gitBranch(t, repoRoot)

	_, err := executor.Execute(ctx, Request{
		RepoRoot:     repoRoot,
		AgentSlug:    "test-agent",
		Strategy:     StrategySquash,
		BaseBranch:   baseBranch,
		TargetBranch: baseBranch,
		AgentID:      ids.NewID(),
	})
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	// Should have received requested and completed events
	hasRequested := false
	hasCompleted := false
	for _, evtType := range receivedTypes {
		switch evtType {
		case event.TypeGitMergeRequested:
			hasRequested = true
		case event.TypeGitMergeCompleted:
			hasCompleted = true
		}
	}
	if !hasRequested {
		t.Error("git.merge.requested event should be emitted")
	}
	if !hasCompleted {
		t.Error("git.merge.completed event should be emitted")
	}
}

func TestInvalidStrategy(t *testing.T) {
	repoRoot := initTestRepo(t)
	createWorktreeWithCommit(t, repoRoot, "test-agent", "feature.txt", "content\n")

	executor := NewExecutor(event.NewNoopDispatcher())
	ctx := context.Background()

	baseBranch := gitBranch(t, repoRoot)

	_, err := executor.Execute(ctx, Request{
		RepoRoot:     repoRoot,
		AgentSlug:    "test-agent",
		Strategy:     Strategy("invalid"),
		BaseBranch:   baseBranch,
		TargetBranch: baseBranch,
		AgentID:      ids.NewID(),
	})
	if err == nil {
		t.Error("Execute() with invalid strategy should fail")
	}
}

// gitBranch returns the current branch name.
func gitBranch(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "symbolic-ref", "--quiet", "--short", "HEAD")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("git symbolic-ref failed: %v", err)
	}
	return string(output[:len(output)-1]) // trim newline
}
