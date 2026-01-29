package agent

import (
	"os/exec"
	"testing"
)

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	// git init
	execCmd(t, dir, "git", "init")
	// config user
	execCmd(t, dir, "git", "config", "user.email", "test@example.com")
	execCmd(t, dir, "git", "config", "user.name", "Test User")
	// commit something so HEAD exists
	execCmd(t, dir, "touch", "README.md")
	execCmd(t, dir, "git", "add", "README.md")
	execCmd(t, dir, "git", "commit", "-m", "Initial commit")
	execCmd(t, dir, "git", "branch", "-M", "main")
}

func execCmd(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Command %s %v failed: %v\nOutput: %s", name, args, err, out)
	}
}

