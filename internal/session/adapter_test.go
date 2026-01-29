package session

import (
	"context"
	"testing"
	"time"

	"github.com/stateforward/hsm-go/muid"

	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/pkg/api"
)

// Compile-time check: Adapter must satisfy agent.SessionSpawner.
var _ agent.SessionSpawner = (*Adapter)(nil)

func TestNewAdapter(t *testing.T) {
	mgr := NewManager(event.NewNoopDispatcher())
	adapter := NewAdapter(mgr)

	if adapter == nil {
		t.Fatal("NewAdapter() returned nil")
	}
	if adapter.mgr != mgr {
		t.Error("adapter.mgr should reference the provided manager")
	}
}

// waitForLifecycleState polls the lifecycle HSM until it reaches the expected
// state or times out.
func waitForLifecycleState(t *testing.T, mgr *agent.Manager, agentID muid.MUID, expected api.LifecycleState) {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		lhsm := mgr.LifecycleHSMFor(agentID)
		if lhsm != nil && lhsm.LifecycleState() == expected {
			return
		}
		select {
		case <-deadline:
			actual := api.LifecycleState("nil")
			if lhsm != nil {
				actual = lhsm.LifecycleState()
			}
			t.Fatalf("timeout waiting for lifecycle state %q, got %q", expected, actual)
		case <-time.After(5 * time.Millisecond):
			// retry
		}
	}
}

// TestIntegrationStartStop wires agent.Manager and session.Manager through the
// Adapter and runs a real PTY session through the full lifecycle:
// Pending → Starting → Running → (Stop) → Terminated.
func TestIntegrationStartStop(t *testing.T) {
	dispatcher := event.NewNoopDispatcher()
	agentMgr := agent.NewManager(dispatcher)
	sessMgr := NewManager(dispatcher)
	adapter := NewAdapter(sessMgr)
	agentMgr.SetSessionSpawner(adapter)

	ctx := context.Background()
	repoRoot := initTestRepo(t)

	ag, err := agentMgr.Add(ctx, api.Agent{
		Name:     "integration-agent",
		Adapter:  "claude-code",
		RepoRoot: repoRoot,
	})
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	// Verify HSMs created
	lhsm := agentMgr.LifecycleHSMFor(ag.ID)
	if lhsm == nil {
		t.Fatal("LifecycleHSMFor() returned nil")
	}
	if lhsm.LifecycleState() != api.LifecyclePending {
		t.Fatalf("initial lifecycle = %q, want %q", lhsm.LifecycleState(), api.LifecyclePending)
	}

	// Start: spawns a real shell process via PTY
	if err := agentMgr.Start(ctx, ag.ID, "/bin/sh", "-c", "sleep 60"); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Wait for Running state
	waitForLifecycleState(t, agentMgr, ag.ID, api.LifecycleRunning)

	// Verify session exists in session manager
	sess := sessMgr.Get(ag.ID)
	if sess == nil {
		t.Fatal("session should exist after Start()")
	}
	if sess.SessionState() != StateRunning {
		t.Errorf("session state = %q, want %q", sess.SessionState(), StateRunning)
	}

	// Stop the agent
	if err := agentMgr.Stop(ctx, ag.ID); err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}

	// Wait for Terminated state
	waitForLifecycleState(t, agentMgr, ag.ID, api.LifecycleTerminated)

	// Verify agent struct reflects Terminated
	if ag.Lifecycle() != api.LifecycleTerminated {
		t.Errorf("agent.Lifecycle() = %q, want %q", ag.Lifecycle(), api.LifecycleTerminated)
	}
}

// TestIntegrationStartKill wires through the full lifecycle with Kill.
func TestIntegrationStartKill(t *testing.T) {
	dispatcher := event.NewNoopDispatcher()
	agentMgr := agent.NewManager(dispatcher)
	sessMgr := NewManager(dispatcher)
	adapter := NewAdapter(sessMgr)
	agentMgr.SetSessionSpawner(adapter)

	ctx := context.Background()
	repoRoot := initTestRepo(t)

	ag, err := agentMgr.Add(ctx, api.Agent{
		Name:     "kill-agent",
		Adapter:  "claude-code",
		RepoRoot: repoRoot,
	})
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	if err := agentMgr.Start(ctx, ag.ID, "/bin/sh", "-c", "sleep 60"); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	waitForLifecycleState(t, agentMgr, ag.ID, api.LifecycleRunning)

	if err := agentMgr.Kill(ctx, ag.ID); err != nil {
		t.Fatalf("Kill() failed: %v", err)
	}

	waitForLifecycleState(t, agentMgr, ag.ID, api.LifecycleTerminated)
}

// TestIntegrationSessionCrash verifies that an unexpected process exit
// transitions the lifecycle to Errored (not Terminated).
func TestIntegrationSessionCrash(t *testing.T) {
	dispatcher := event.NewNoopDispatcher()
	agentMgr := agent.NewManager(dispatcher)
	sessMgr := NewManager(dispatcher)
	adapter := NewAdapter(sessMgr)
	agentMgr.SetSessionSpawner(adapter)

	ctx := context.Background()
	repoRoot := initTestRepo(t)

	ag, err := agentMgr.Add(ctx, api.Agent{
		Name:     "crash-agent",
		Adapter:  "claude-code",
		RepoRoot: repoRoot,
	})
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	// Start with a command that exits on its own (simulating a crash)
	if err := agentMgr.Start(ctx, ag.ID, "/bin/sh", "-c", "exit 1"); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// The process will exit quickly with code 1. watchSession should
	// see that stopping was not set and transition to Errored.
	waitForLifecycleState(t, agentMgr, ag.ID, api.LifecycleErrored)

	lhsm := agentMgr.LifecycleHSMFor(ag.ID)
	if lhsm.LastError() == nil {
		t.Error("LastError() should be set after crash")
	}
}

// TestIntegrationRemoveRunningAgent verifies Remove stops the session and
// cleans up HSMs when the agent is running.
func TestIntegrationRemoveRunningAgent(t *testing.T) {
	dispatcher := event.NewNoopDispatcher()
	agentMgr := agent.NewManager(dispatcher)
	sessMgr := NewManager(dispatcher)
	adapter := NewAdapter(sessMgr)
	agentMgr.SetSessionSpawner(adapter)

	ctx := context.Background()
	repoRoot := initTestRepo(t)

	ag, err := agentMgr.Add(ctx, api.Agent{
		Name:     "remove-agent",
		Adapter:  "claude-code",
		RepoRoot: repoRoot,
	})
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	if err := agentMgr.Start(ctx, ag.ID, "/bin/sh", "-c", "sleep 60"); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	waitForLifecycleState(t, agentMgr, ag.ID, api.LifecycleRunning)

	id := ag.ID

	if err := agentMgr.Remove(ctx, id, false); err != nil {
		t.Fatalf("Remove() failed: %v", err)
	}

	// HSMs should be cleaned up
	if agentMgr.LifecycleHSMFor(id) != nil {
		t.Error("LifecycleHSMFor() should return nil after Remove()")
	}

	// Session should be cleaned up
	if sessMgr.Get(id) != nil {
		t.Error("session should be removed after Remove()")
	}
}
