package agent

import (
	"testing"
	"time"
)

func TestReflectHSM(t *testing.T) {
	// Dummy test removed
}

// Helper to wait for state transitions (since HSM is async)
func waitForState(t *testing.T, a *AgentActor, targetState string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if a.State() == targetState {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for state %q, current: %q", targetState, a.State())
}

func TestAgentNew(t *testing.T) {
	a := NewAgent("Test Agent", "dummy-adapter", "/tmp/repo", nil)
	if a.ID() == 0 {
		t.Error("agent ID should not be 0")
	}
	if a.Data().Name != "Test Agent" {
		t.Errorf("got name %q, want 'Test Agent'", a.Data().Name)
	}
	// Verify initial state
	if s := a.State(); s != "/agent/pending" {
		t.Errorf("initial state = %q, want '/agent/pending'", s)
	}
}

func TestAgentLifecycle(t *testing.T) {
	a := NewAgent("LifecycleAgent", "dummy", "/tmp", nil)
	waitForState(t, a, "/agent/pending", time.Second)

	// Start
	a.Start()
	// When worktree is nil, StartAction immediately dispatches EventStarted,
	// so we transition directly to running/online.
	waitForState(t, a, "/agent/running/online", time.Second)

	// Activity -> Busy
	a.SendActivity()
	waitForState(t, a, "/agent/running/busy", time.Second)

	// Stop -> Terminated
	a.Stop()
	waitForState(t, a, "/agent/terminated", time.Second)
}
