package process

import (
	"testing"
	"time"

	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go/muid"
)

func TestTracker(t *testing.T) {
	tracker := NewTracker()

	pid := 1234
	procID := api.ProcessID(muid.Make())
	agentID := api.AgentID(1)

	p := &Process{
		PID:       pid,
		AgentID:   agentID,
		ProcessID: procID,
		Command:   "ls",
		StartedAt: time.Now(),
	}

	// Test Spawn
	tracker.TrackSpawn(p)

	// Verify internal state
	got, ok := tracker.GetProcess(pid)
	if !ok {
		t.Fatal("Process not tracked")
	}
	if !got.Running {
		t.Error("Process should be running")
	}

	// Verify Event
	select {
	case e := <-tracker.Events:
		if e.Type != EventSpawned {
			t.Errorf("Expected spawned event, got %s", e.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for spawn event")
	}

	// Test Exit
	if err := tracker.TrackExit(pid, 0); err != nil {
		t.Fatalf("TrackExit failed: %v", err)
	}

	got, _ = tracker.GetProcess(pid)
	if got.Running {
		t.Error("Process should not be running")
	}
	if got.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", got.ExitCode)
	}

	// Verify Exit Event
	select {
	case e := <-tracker.Events:
		if e.Type != EventExited {
			t.Errorf("Expected exited event, got %s", e.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for exit event")
	}
}
