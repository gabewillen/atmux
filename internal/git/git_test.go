// Package git provides git repository queries and merge-strategy helpers.
package git

import (
	"os/exec"
	"testing"
)

func TestIsRepo_NotDir(t *testing.T) {
	ok, err := IsRepo("/nonexistent")
	if err != nil {
		t.Fatalf("IsRepo: %v", err)
	}
	if ok {
		t.Error("expected false for nonexistent path")
	}
}

func TestIsRepo_NotGit(t *testing.T) {
	dir := t.TempDir()
	ok, err := IsRepo(dir)
	if err != nil {
		t.Fatalf("IsRepo: %v", err)
	}
	if ok {
		t.Error("expected false for non-git dir")
	}
}

func TestIsRepo_InsideRepo(t *testing.T) {
	dir := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %s: %v", out, err)
	}
	ok, err := IsRepo(dir)
	if err != nil {
		t.Fatalf("IsRepo: %v", err)
	}
	if !ok {
		t.Error("expected true for git repo")
	}
}

func TestValidStrategy(t *testing.T) {
	for _, s := range ValidMergeStrategies {
		if !ValidStrategy(s) {
			t.Errorf("ValidStrategy(%q) = false, want true", s)
		}
	}
	if ValidStrategy("invalid") {
		t.Error("ValidStrategy(\"invalid\") = true, want false")
	}
}

func TestResolveTargetBranch_Configured(t *testing.T) {
	dir := t.TempDir()
	got, err := ResolveTargetBranch(dir, "main")
	if err != nil {
		t.Fatalf("ResolveTargetBranch: %v", err)
	}
	if got != "main" {
		t.Errorf("ResolveTargetBranch(_, \"main\") = %q, want \"main\"", got)
	}
}
