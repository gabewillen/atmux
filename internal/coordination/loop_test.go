package coordination

import (
	"context"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/config"
)

func TestLoop_Scheduler(t *testing.T) {
	cfg := config.AgentConfig{Name: "test-agent"}
	bus := agent.NewEventBus()
	a, _ := agent.NewAgent(cfg, "/tmp", bus) // Mock agent
	
	loop := NewLoop(a, 10*time.Millisecond)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go loop.Start(ctx)

	// Let it run for a few ticks
	time.Sleep(50 * time.Millisecond)
	
	loop.Stop()
	
	// Verify no panic and basic execution (logs would show errors)
}
