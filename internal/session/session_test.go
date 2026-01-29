package session

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/internal/ids"
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

func newTestAgent(t *testing.T) *agent.Agent {
	t.Helper()
	repoRoot := initTestRepo(t)
	mgr := agent.NewManager(event.NewNoopDispatcher())
	ctx := context.Background()

	ag, err := mgr.Add(ctx, api.Agent{
		Name:     "test-agent",
		Adapter:  "claude-code",
		RepoRoot: repoRoot,
	})
	if err != nil {
		t.Fatalf("create test agent: %v", err)
	}
	return ag
}

func newTestAgentWithRepo(t *testing.T, repoRoot string) *agent.Agent {
	t.Helper()
	mgr := agent.NewManager(event.NewNoopDispatcher())
	ctx := context.Background()

	ag, err := mgr.Add(ctx, api.Agent{
		Name:     "agent-" + ids.EncodeID(ids.NewID()),
		Adapter:  "claude-code",
		RepoRoot: repoRoot,
	})
	if err != nil {
		t.Fatalf("create test agent: %v", err)
	}
	return ag
}

func TestManagerSpawn(t *testing.T) {
	smgr := NewManager(event.NewNoopDispatcher())
	ag := newTestAgent(t)
	ctx := context.Background()

	// Spawn a simple command that exits quickly
	sess, err := smgr.Spawn(ctx, ag, "/bin/sh", "-c", "echo hello")
	if err != nil {
		t.Fatalf("Spawn() failed: %v", err)
	}

	if sess.ID == 0 {
		t.Error("session ID should be non-zero")
	}
	if sess.AgentID != ag.ID {
		t.Errorf("session AgentID = %d, want %d", sess.AgentID, ag.ID)
	}
	if sess.SessionState() != StateRunning {
		t.Errorf("session state = %q, want %q", sess.SessionState(), StateRunning)
	}

	// Read output
	buf := make([]byte, 1024)
	n, _ := sess.Read(buf)
	if n == 0 {
		t.Error("expected to read some output from spawned session")
	}

	// Wait for exit
	select {
	case <-sess.Done():
		// OK
	case <-time.After(5 * time.Second):
		t.Fatal("session did not exit within timeout")
	}

	if sess.SessionState() != StateStopped {
		t.Errorf("session state after exit = %q, want %q", sess.SessionState(), StateStopped)
	}
}

func TestManagerSpawnDuplicate(t *testing.T) {
	smgr := NewManager(event.NewNoopDispatcher())
	ag := newTestAgent(t)
	ctx := context.Background()

	// Spawn a long-running command
	_, err := smgr.Spawn(ctx, ag, "/bin/sh", "-c", "sleep 60")
	if err != nil {
		t.Fatalf("First Spawn() failed: %v", err)
	}

	// Second spawn should fail
	_, err = smgr.Spawn(ctx, ag, "/bin/sh", "-c", "echo hello")
	if err == nil {
		t.Error("Second Spawn() for same agent should fail")
	}

	// Clean up
	_ = smgr.Kill(ctx, ag.ID)
}

func TestManagerStop(t *testing.T) {
	smgr := NewManager(event.NewNoopDispatcher())
	ag := newTestAgent(t)
	ctx := context.Background()

	_, err := smgr.Spawn(ctx, ag, "/bin/sh", "-c", "sleep 60")
	if err != nil {
		t.Fatalf("Spawn() failed: %v", err)
	}

	// Stop should close the PTY and cause the shell to exit
	if err := smgr.Stop(ctx, ag.ID); err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}

	sess := smgr.Get(ag.ID)
	if sess == nil {
		t.Fatal("session should still be in manager after Stop")
	}
	if sess.SessionState() != StateStopped {
		t.Errorf("session state = %q, want %q", sess.SessionState(), StateStopped)
	}
}

func TestManagerKill(t *testing.T) {
	smgr := NewManager(event.NewNoopDispatcher())
	ag := newTestAgent(t)
	ctx := context.Background()

	sess, err := smgr.Spawn(ctx, ag, "/bin/sh", "-c", "sleep 60")
	if err != nil {
		t.Fatalf("Spawn() failed: %v", err)
	}

	if err := smgr.Kill(ctx, ag.ID); err != nil {
		t.Fatalf("Kill() failed: %v", err)
	}

	select {
	case <-sess.Done():
		// OK
	case <-time.After(5 * time.Second):
		t.Fatal("session did not exit after Kill")
	}
}

func TestManagerStopNotFound(t *testing.T) {
	smgr := NewManager(event.NewNoopDispatcher())
	ctx := context.Background()

	err := smgr.Stop(ctx, ids.NewID())
	if err == nil {
		t.Error("Stop() for non-existing session should fail")
	}
}

func TestManagerGet(t *testing.T) {
	smgr := NewManager(event.NewNoopDispatcher())
	ag := newTestAgent(t)
	ctx := context.Background()

	// Before spawn
	if smgr.Get(ag.ID) != nil {
		t.Error("Get() should return nil before Spawn")
	}

	_, err := smgr.Spawn(ctx, ag, "/bin/sh", "-c", "echo test")
	if err != nil {
		t.Fatalf("Spawn() failed: %v", err)
	}

	// After spawn
	sess := smgr.Get(ag.ID)
	if sess == nil {
		t.Error("Get() should return session after Spawn")
	}
}

func TestManagerList(t *testing.T) {
	smgr := NewManager(event.NewNoopDispatcher())
	ctx := context.Background()

	if len(smgr.List()) != 0 {
		t.Error("List() should be empty initially")
	}

	repoRoot := initTestRepo(t)

	// Create multiple agents and sessions
	for i := range 3 {
		ag := newTestAgentWithRepo(t, repoRoot)
		_, err := smgr.Spawn(ctx, ag, "/bin/sh", "-c", "sleep 60")
		if err != nil {
			t.Fatalf("Spawn() %d: %v", i, err)
		}
	}

	list := smgr.List()
	if len(list) != 3 {
		t.Errorf("List() = %d sessions, want 3", len(list))
	}

	// Clean up
	smgr.KillAll()
}

func TestManagerRemoveSession(t *testing.T) {
	smgr := NewManager(event.NewNoopDispatcher())
	ag := newTestAgent(t)
	ctx := context.Background()

	_, err := smgr.Spawn(ctx, ag, "/bin/sh", "-c", "echo done")
	if err != nil {
		t.Fatalf("Spawn() failed: %v", err)
	}

	smgr.Remove(ag.ID)

	if smgr.Get(ag.ID) != nil {
		t.Error("Get() should return nil after Remove")
	}
}

func TestSessionWrite(t *testing.T) {
	smgr := NewManager(event.NewNoopDispatcher())
	ag := newTestAgent(t)
	ctx := context.Background()

	sess, err := smgr.Spawn(ctx, ag, "/bin/sh")
	if err != nil {
		t.Fatalf("Spawn() failed: %v", err)
	}

	// Write to the session
	n, err := sess.Write([]byte("echo hello\n"))
	if err != nil {
		t.Fatalf("Write() failed: %v", err)
	}
	if n == 0 {
		t.Error("Write() should write bytes")
	}

	// Clean up
	_ = smgr.Kill(ctx, ag.ID)
}

func TestStopAll(t *testing.T) {
	smgr := NewManager(event.NewNoopDispatcher())
	ctx := context.Background()

	repoRoot := initTestRepo(t)

	var sessions []*Session
	for i := range 3 {
		ag := newTestAgentWithRepo(t, repoRoot)
		sess, err := smgr.Spawn(ctx, ag, "/bin/sh", "-c", "sleep 60")
		if err != nil {
			t.Fatalf("Spawn() %d: %v", i, err)
		}
		sessions = append(sessions, sess)
	}

	smgr.StopAll()

	for i, sess := range sessions {
		select {
		case <-sess.Done():
			// OK
		case <-time.After(5 * time.Second):
			t.Fatalf("session %d did not stop", i)
		}
	}
}
