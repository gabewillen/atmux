package main

import (
	"os"
	"testing"
)

func TestRunGitInvalid(t *testing.T) {
	if err := runGit(nil); err == nil {
		t.Fatalf("expected usage error")
	}
	if err := runGit([]string{"unknown"}); err == nil {
		t.Fatalf("expected unknown error")
	}
}

func TestRunGitMergeMissingRef(t *testing.T) {
	if err := runGitMerge(nil); err == nil {
		t.Fatalf("expected missing ref error")
	}
}

func TestRunGitMergeInvalidFlag(t *testing.T) {
	if err := runGitMerge([]string{"--bad"}); err == nil {
		t.Fatalf("expected flag parse error")
	}
}

func TestRunGitMergeCommand(t *testing.T) {
	repoRoot, _, cleanup := setupDaemonSocket(t)
	defer cleanup()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Setenv("HOME", t.TempDir())
	if err := runGit([]string{"merge", "--id", "1"}); err != nil {
		t.Fatalf("runGit merge: %v", err)
	}
}
