package agent

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stateforward/hsm-go/muid"

	amuxerrors "github.com/agentflare-ai/amux/internal/errors"
	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/pkg/api"
)

// mockSessionHandle implements SessionHandle for testing.
type mockSessionHandle struct {
	done    chan struct{}
	exitErr error
	mu      sync.Mutex
}

func newMockHandle() *mockSessionHandle {
	return &mockSessionHandle{done: make(chan struct{})}
}

func (h *mockSessionHandle) Done() <-chan struct{} {
	return h.done
}

func (h *mockSessionHandle) ExitErr() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.exitErr
}

// exit simulates the session exiting.
func (h *mockSessionHandle) exit(err error) {
	h.mu.Lock()
	h.exitErr = err
	h.mu.Unlock()
	close(h.done)
}

// mockSessionSpawner implements SessionSpawner for testing.
type mockSessionSpawner struct {
	mu       sync.Mutex
	handles  map[muid.MUID]*mockSessionHandle
	spawnErr error
	stopErr  error
	killErr  error
	removed  map[muid.MUID]bool
}

func newMockSpawner() *mockSessionSpawner {
	return &mockSessionSpawner{
		handles: make(map[muid.MUID]*mockSessionHandle),
		removed: make(map[muid.MUID]bool),
	}
}

func (s *mockSessionSpawner) SpawnAgent(_ context.Context, ag *Agent, _ string, _ ...string) (SessionHandle, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.spawnErr != nil {
		return nil, s.spawnErr
	}
	handle := newMockHandle()
	s.handles[ag.ID] = handle
	return handle, nil
}

func (s *mockSessionSpawner) StopAgent(_ context.Context, agentID muid.MUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopErr != nil {
		return s.stopErr
	}
	if h, ok := s.handles[agentID]; ok {
		h.exit(nil)
	}
	return nil
}

func (s *mockSessionSpawner) KillAgent(_ context.Context, agentID muid.MUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.killErr != nil {
		return s.killErr
	}
	if h, ok := s.handles[agentID]; ok {
		h.exit(nil)
	}
	return nil
}

func (s *mockSessionSpawner) RemoveSession(agentID muid.MUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.removed[agentID] = true
	delete(s.handles, agentID)
}

func (s *mockSessionSpawner) wasRemoved(agentID muid.MUID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.removed[agentID]
}

// handleFor returns the mock handle for an agent, or nil.
func (s *mockSessionSpawner) handleFor(agentID muid.MUID) *mockSessionHandle {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.handles[agentID]
}

// newTestManagerWithAgent creates a Manager with a single agent and mock spawner.
// Returns the manager, agent, and spawner.
func newTestManagerWithAgent(t *testing.T) (*Manager, *Agent, *mockSessionSpawner) {
	t.Helper()

	mgr := NewManager(event.NewNoopDispatcher())
	spawner := newMockSpawner()
	mgr.SetSessionSpawner(spawner)

	ctx := context.Background()
	repoRoot := initTestRepo(t)

	ag, err := mgr.Add(ctx, api.Agent{
		Name:     "test-agent",
		Adapter:  "claude-code",
		RepoRoot: repoRoot,
	})
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	return mgr, ag, spawner
}

// waitForState waits for the lifecycle HSM to reach the expected state.
func waitForState(t *testing.T, mgr *Manager, agentID muid.MUID, expected api.LifecycleState) {
	t.Helper()
	deadline := time.After(2 * time.Second)
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
			t.Fatalf("timeout waiting for state %q, got %q", expected, actual)
		case <-time.After(5 * time.Millisecond):
			// retry
		}
	}
}

func TestStartAgentLifecycleTransitions(t *testing.T) {
	mgr, ag, _ := newTestManagerWithAgent(t)
	ctx := context.Background()

	// Before start: should be Pending
	waitForState(t, mgr, ag.ID, api.LifecyclePending)

	// Start the agent
	if err := mgr.Start(ctx, ag.ID, "/bin/sh"); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// After start: should be Running
	waitForState(t, mgr, ag.ID, api.LifecycleRunning)

	// Agent struct should also reflect Running
	if ag.Lifecycle() != api.LifecycleRunning {
		t.Errorf("agent.Lifecycle() = %q, want %q", ag.Lifecycle(), api.LifecycleRunning)
	}
}

func TestStartWithSpawnFailure(t *testing.T) {
	mgr, ag, spawner := newTestManagerWithAgent(t)
	ctx := context.Background()

	spawner.mu.Lock()
	spawner.spawnErr = errors.New("spawn failed")
	spawner.mu.Unlock()

	err := mgr.Start(ctx, ag.ID, "/bin/sh")
	if err == nil {
		t.Fatal("Start() should fail when spawn fails")
	}

	// Lifecycle should be Errored
	waitForState(t, mgr, ag.ID, api.LifecycleErrored)
}

func TestStopAgent(t *testing.T) {
	mgr, ag, spawner := newTestManagerWithAgent(t)
	ctx := context.Background()

	if err := mgr.Start(ctx, ag.ID, "/bin/sh"); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	waitForState(t, mgr, ag.ID, api.LifecycleRunning)

	// Stop the agent
	// The mock spawner's StopAgent closes the handle's done channel,
	// which triggers watchSession to dispatch DispatchStop.
	if err := mgr.Stop(ctx, ag.ID); err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}

	// After stop: should transition to Terminated
	waitForState(t, mgr, ag.ID, api.LifecycleTerminated)

	// Session should have been removed
	if !spawner.wasRemoved(ag.ID) {
		t.Error("session should be removed after stop")
	}
}

func TestKillAgent(t *testing.T) {
	mgr, ag, spawner := newTestManagerWithAgent(t)
	ctx := context.Background()

	if err := mgr.Start(ctx, ag.ID, "/bin/sh"); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	waitForState(t, mgr, ag.ID, api.LifecycleRunning)

	if err := mgr.Kill(ctx, ag.ID); err != nil {
		t.Fatalf("Kill() failed: %v", err)
	}

	// After kill: should transition to Terminated
	waitForState(t, mgr, ag.ID, api.LifecycleTerminated)

	if !spawner.wasRemoved(ag.ID) {
		t.Error("session should be removed after kill")
	}
}

func TestSessionCrash(t *testing.T) {
	mgr, ag, spawner := newTestManagerWithAgent(t)
	ctx := context.Background()

	if err := mgr.Start(ctx, ag.ID, "/bin/sh"); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	waitForState(t, mgr, ag.ID, api.LifecycleRunning)

	// Simulate unexpected crash (not via Stop/Kill)
	handle := spawner.handleFor(ag.ID)
	if handle == nil {
		t.Fatal("no handle for agent")
	}
	handle.exit(fmt.Errorf("signal: killed"))

	// Should transition to Errored (not Terminated)
	waitForState(t, mgr, ag.ID, api.LifecycleErrored)

	lhsm := mgr.LifecycleHSMFor(ag.ID)
	if lhsm.LastError() == nil {
		t.Error("LastError() should be set after crash")
	}
}

func TestStartNonExistentAgent(t *testing.T) {
	mgr := NewManager(event.NewNoopDispatcher())
	spawner := newMockSpawner()
	mgr.SetSessionSpawner(spawner)
	ctx := context.Background()

	err := mgr.Start(ctx, muid.MUID(999999), "/bin/sh")
	if err == nil {
		t.Error("Start() should fail for non-existent agent")
	}
	if !errors.Is(err, amuxerrors.ErrAgentNotFound) {
		t.Errorf("error should wrap ErrAgentNotFound, got: %v", err)
	}
}

func TestStartAgentInWrongState(t *testing.T) {
	mgr, ag, _ := newTestManagerWithAgent(t)
	ctx := context.Background()

	// Start the agent first
	if err := mgr.Start(ctx, ag.ID, "/bin/sh"); err != nil {
		t.Fatalf("First Start() failed: %v", err)
	}
	waitForState(t, mgr, ag.ID, api.LifecycleRunning)

	// Try to start again — should fail because agent is Running, not Pending
	err := mgr.Start(ctx, ag.ID, "/bin/sh")
	if err == nil {
		t.Error("Start() should fail when agent is already running")
	}
}

func TestStopAgentNotRunning(t *testing.T) {
	mgr, ag, _ := newTestManagerWithAgent(t)
	ctx := context.Background()

	// Agent is in Pending state, not Running
	err := mgr.Stop(ctx, ag.ID)
	if err == nil {
		t.Error("Stop() should fail when agent is not running")
	}
	if !errors.Is(err, amuxerrors.ErrAgentNotRunning) {
		t.Errorf("error should wrap ErrAgentNotRunning, got: %v", err)
	}
}

func TestKillAgentNotRunning(t *testing.T) {
	mgr, ag, _ := newTestManagerWithAgent(t)
	ctx := context.Background()

	err := mgr.Kill(ctx, ag.ID)
	if err == nil {
		t.Error("Kill() should fail when agent is not running")
	}
	if !errors.Is(err, amuxerrors.ErrAgentNotRunning) {
		t.Errorf("error should wrap ErrAgentNotRunning, got: %v", err)
	}
}

func TestHSMsCreatedOnAdd(t *testing.T) {
	mgr := NewManager(event.NewNoopDispatcher())
	ctx := context.Background()
	repoRoot := initTestRepo(t)

	ag, err := mgr.Add(ctx, api.Agent{
		Name:     "test-agent",
		Adapter:  "claude-code",
		RepoRoot: repoRoot,
	})
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	lhsm := mgr.LifecycleHSMFor(ag.ID)
	if lhsm == nil {
		t.Fatal("LifecycleHSMFor() returned nil after Add()")
	}
	if lhsm.LifecycleState() != api.LifecyclePending {
		t.Errorf("lifecycle state = %q, want %q", lhsm.LifecycleState(), api.LifecyclePending)
	}

	phsm := mgr.PresenceHSMFor(ag.ID)
	if phsm == nil {
		t.Fatal("PresenceHSMFor() returned nil after Add()")
	}
	if phsm.PresenceState() != api.PresenceOnline {
		t.Errorf("presence state = %q, want %q", phsm.PresenceState(), api.PresenceOnline)
	}
}

func TestHSMsCleanedOnRemove(t *testing.T) {
	mgr := NewManager(event.NewNoopDispatcher())
	ctx := context.Background()
	repoRoot := initTestRepo(t)

	ag, err := mgr.Add(ctx, api.Agent{
		Name:     "test-agent",
		Adapter:  "claude-code",
		RepoRoot: repoRoot,
	})
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}
	id := ag.ID

	if err := mgr.Remove(ctx, id, false); err != nil {
		t.Fatalf("Remove() failed: %v", err)
	}

	if mgr.LifecycleHSMFor(id) != nil {
		t.Error("LifecycleHSMFor() should return nil after Remove()")
	}
	if mgr.PresenceHSMFor(id) != nil {
		t.Error("PresenceHSMFor() should return nil after Remove()")
	}
}

func TestRemoveRunningAgent(t *testing.T) {
	mgr, ag, spawner := newTestManagerWithAgent(t)
	ctx := context.Background()

	if err := mgr.Start(ctx, ag.ID, "/bin/sh"); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	waitForState(t, mgr, ag.ID, api.LifecycleRunning)

	id := ag.ID

	// Remove should stop the session and clean up HSMs
	if err := mgr.Remove(ctx, id, false); err != nil {
		t.Fatalf("Remove() failed: %v", err)
	}

	// Session should have been cleaned up
	if !spawner.wasRemoved(id) {
		t.Error("session should be removed when removing a running agent")
	}

	// HSMs should be gone
	if mgr.LifecycleHSMFor(id) != nil {
		t.Error("LifecycleHSMFor() should return nil after Remove()")
	}
	if mgr.PresenceHSMFor(id) != nil {
		t.Error("PresenceHSMFor() should return nil after Remove()")
	}
}

func TestStartWithoutSpawner(t *testing.T) {
	mgr := NewManager(event.NewNoopDispatcher())
	ctx := context.Background()
	repoRoot := initTestRepo(t)

	ag, err := mgr.Add(ctx, api.Agent{
		Name:     "test-agent",
		Adapter:  "claude-code",
		RepoRoot: repoRoot,
	})
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	// No spawner configured
	err = mgr.Start(ctx, ag.ID, "/bin/sh")
	if err == nil {
		t.Error("Start() should fail without a session spawner")
	}
}

func TestLifecycleHSMForNonExistent(t *testing.T) {
	mgr := NewManager(event.NewNoopDispatcher())

	if mgr.LifecycleHSMFor(muid.MUID(999)) != nil {
		t.Error("LifecycleHSMFor() should return nil for non-existent agent")
	}
}

func TestPresenceHSMForNonExistent(t *testing.T) {
	mgr := NewManager(event.NewNoopDispatcher())

	if mgr.PresenceHSMFor(muid.MUID(999)) != nil {
		t.Error("PresenceHSMFor() should return nil for non-existent agent")
	}
}

// TestConcurrentRemoveDuringWatchSession verifies that Remove() of a running
// agent while watchSession is still blocked on <-handle.Done() does not race
// or panic. Remove deletes the HSMs from the map before stopping the session;
// when watchSession wakes up, it should see the agent is gone and return.
func TestConcurrentRemoveDuringWatchSession(t *testing.T) {
	mgr := NewManager(event.NewNoopDispatcher())
	ctx := context.Background()
	repoRoot := initTestRepo(t)

	// Use a spawner where StopAgent does NOT auto-close the done channel.
	// This lets us control the timing: Remove calls StopAgent, then we
	// close the handle manually to simulate the session finally exiting.
	spawner := newMockSpawner()
	spawner.mu.Lock()
	spawner.stopErr = nil
	spawner.mu.Unlock()
	mgr.SetSessionSpawner(spawner)

	ag, err := mgr.Add(ctx, api.Agent{
		Name:     "race-agent",
		Adapter:  "claude-code",
		RepoRoot: repoRoot,
	})
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	// Override StopAgent to NOT close the handle immediately.
	// We'll close it after Remove has already deleted the HSMs.
	delayedSpawner := &delayedStopSpawner{
		mockSessionSpawner: newMockSpawner(),
		stopCh:             make(chan struct{}),
	}
	mgr.SetSessionSpawner(delayedSpawner)

	if err := mgr.Start(ctx, ag.ID, "/bin/sh"); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	waitForState(t, mgr, ag.ID, api.LifecycleRunning)

	handle := delayedSpawner.handleFor(ag.ID)
	if handle == nil {
		t.Fatal("no handle for agent")
	}

	id := ag.ID

	// Remove in a goroutine. This will:
	// 1. Delete HSMs from the map
	// 2. Call StopAgent (which signals stopCh but does NOT close the handle)
	// 3. Try to stop the HSMs
	removeDone := make(chan error, 1)
	go func() {
		removeDone <- mgr.Remove(ctx, id, false)
	}()

	// Wait for StopAgent to be called
	select {
	case <-delayedSpawner.stopCh:
		// StopAgent was called — Remove has already deleted HSMs from map
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for StopAgent to be called")
	}

	// Now close the handle. watchSession will wake up but the HSMs are
	// already deleted from the map — it should return without panic.
	handle.exit(nil)

	// Wait for Remove to finish
	select {
	case err := <-removeDone:
		if err != nil {
			t.Fatalf("Remove() failed: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for Remove() to complete")
	}

	// Verify cleanup
	if mgr.LifecycleHSMFor(id) != nil {
		t.Error("LifecycleHSMFor() should return nil after Remove()")
	}
}

// delayedStopSpawner is a mock spawner where StopAgent signals a channel
// instead of immediately closing the session handle.
type delayedStopSpawner struct {
	*mockSessionSpawner
	stopCh chan struct{}
}

func (s *delayedStopSpawner) StopAgent(_ context.Context, agentID muid.MUID) error {
	// Signal that StopAgent was called, but do NOT close the handle.
	// The handle must be closed externally to control timing.
	select {
	case s.stopCh <- struct{}{}:
	default:
	}
	return nil
}
