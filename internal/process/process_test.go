package process

import (
	"testing"

	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go/muid"
)

func TestTracker(t *testing.T) {
	tracker := NewTracker()
	
	p := &Process{
		PID:       123,
		AgentID:   api.AgentID(muid.Make()),
		ProcessID: api.ProcessID(muid.Make()),
		Command:   "ls",
		Running:   true,
	}

	// Test Spawn
	tracker.TrackSpawn(p)
	
	retrieved, ok := tracker.GetProcess(123)
	if !ok {
		t.Fatal("Process not tracked")
	}
	if !retrieved.Running {
		t.Error("Process should be running")
	}

	select {
	case evt := <-tracker.Events:
		if evt.Type != EventSpawned {
			t.Errorf("Expected Spawned event, got %s", evt.Type)
		}
	default:
		t.Error("No spawn event emitted")
	}

	// Test Exit
	if err := tracker.TrackExit(123, 0); err != nil {
		t.Errorf("TrackExit failed: %v", err)
	}

	retrieved, _ = tracker.GetProcess(123)
	if retrieved.Running {
		t.Error("Process should not be running")
	}
	if retrieved.ExitCode != 0 {
		t.Error("Exit code mismatch")
	}

	select {
	case evt := <-tracker.Events:
		if evt.Type != EventExited {
			t.Errorf("Expected Exited event, got %s", evt.Type)
		}
	default:
		t.Error("No exit event emitted")
	}
}