package monitor

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go/muid"
)

func TestMonitor_Activity(t *testing.T) {
	bus := agent.NewEventBus()
	agentID := api.AgentID(muid.Make())
	
	// Create pipe
	pr, pw := io.Pipe()
	
	mon := NewMonitor(agentID, bus, pr)
	mon.ActivityTimeout = 100 * time.Millisecond
	mon.CheckInterval = 10 * time.Millisecond
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	mon.Start(ctx)
	
	// Subscribe to bus
	sub := bus.Subscribe()
	defer sub.Close()

	// Write data
	go func() {
		pw.Write([]byte("data"))
	}()
	
	// Expect Activity event
	select {
	case event := <-sub.C:
		if event.Type != agent.EventActivity {
			t.Errorf("Expected Activity event, got %s", event.Type)
		}
	case <-time.After(50 * time.Millisecond):
		t.Error("Timed out waiting for activity")
	}

	// Wait for Idle (Online)
	// Last activity was just now. Timeout is 100ms.
	time.Sleep(200 * time.Millisecond)
	
	foundIdle := false
	// Drain channel
Loop:
	for {
		select {
		case event := <-sub.C:
			if event.Type == agent.EventPresenceUpdate {
				if state, ok := event.Payload.(api.PresenceState); ok && state == api.PresenceOnline {
					foundIdle = true
					break Loop
				}
			}
		default:
			break Loop
		}
	}
	
	if !foundIdle {
		t.Error("Did not receive Idle/Online event after timeout")
	}
}