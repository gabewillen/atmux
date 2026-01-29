package process

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"testing"
	"time"

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

func TestStartHookServer(t *testing.T) {
	tracker := NewTracker()
	tmpDir := t.TempDir()
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	if err := tracker.Start(ctx, tmpDir); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	
	// Wait for socket
	socketPath := tracker.SocketPath
	if socketPath == "" {
		t.Fatal("SocketPath not set")
	}
	
	// Verify file exists
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		t.Fatalf("Socket file %s does not exist", socketPath)
	}
	
	// Simulate hook connection
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()
	
	msg := HookMessage{
		PID:  999,
		PPID: 888,
		Cmd:  "test-cmd",
	}
	data, _ := json.Marshal(msg)
	conn.Write(data)
	conn.Write([]byte("\n"))
	
	// Wait for event
	select {
	case evt := <-tracker.Events:
		if evt.Type != EventSpawned {
			t.Errorf("Expected Spawned event, got %s", evt.Type)
		}
		if p, ok := evt.Payload.(*Process); ok {
			if p.PID != 999 {
				t.Errorf("Expected PID 999, got %d", p.PID)
			}
			if p.Command != "test-cmd" {
				t.Errorf("Expected cmd test-cmd, got %s", p.Command)
			}
		} else {
			t.Error("Payload is not *Process")
		}
	case <-time.After(1 * time.Second):
		t.Error("Timed out waiting for hook event")
	}
}