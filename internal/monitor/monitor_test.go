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

	// 1. Initial state (should be nothing until input)
	
	// 2. Write data -> Expect ActivityDetected (only if transitioning back from idle/stuck, OR generally?)
	// The implementation only emits ActivityDetected if (isStuck || isIdle) is true.
	// So initially it might NOT emit ActivityDetected unless we force it to Idle first.
	// Actually, let's wait for Idle first.
	
	time.Sleep(200 * time.Millisecond)
	
	// Should have received EventIdle
	foundIdle := false
LoopIdle:
	for {
		select {
		case event := <-sub.C:
			if event.Type == agent.EventIdle {
				foundIdle = true
				break LoopIdle
			}
		default:
			break LoopIdle
		}
	}
	if !foundIdle {
		// Try waiting a bit more
		select {
		case event := <-sub.C:
			if event.Type == agent.EventIdle {
				foundIdle = true
			}
		case <-time.After(100 * time.Millisecond):
		}
	}
	if !foundIdle {
		t.Error("Did not receive EventIdle")
	}

	// 3. Write data -> Expect ActivityDetected (transition from Idle)
	go func() {
		pw.Write([]byte("data"))
	}()
	
	select {
	case event := <-sub.C:
		if event.Type != agent.EventActivityDetected {
			t.Errorf("Expected ActivityDetected event, got %s", event.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timed out waiting for activity detected")
	}
}