package monitor

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestMonitor_Activity(t *testing.T) {
	bus := agent.NewEventBus()
	sub := bus.Subscribe()
	defer sub.Close()

	// Mock input
	input := new(bytes.Buffer)
	mon := NewMonitor(api.AgentID(1), bus, input)
	mon.ActivityTimeout = 100 * time.Millisecond
	mon.CheckInterval = 50 * time.Millisecond // Fast checks

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Write data
	input.WriteString("data")

	// Start monitor
	go mon.Start(ctx)

	// Wait for read
	time.Sleep(50 * time.Millisecond)
	
	// Check bus for "activity"
	select {
	case e := <-sub.C:
		if e.Type != agent.EventActivity {
			t.Errorf("Expected activity event, got %v", e.Type)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for activity")
	}
}
