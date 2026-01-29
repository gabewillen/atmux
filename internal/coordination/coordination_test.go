package coordination

import (
	"context"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/config"
)

func TestObservationLoop(t *testing.T) {
	cfg := config.AgentConfig{Name: "test-agent"}
	bus := agent.NewEventBus()
	a, _ := agent.NewAgent(cfg, "/tmp", bus)
	
	loop := NewObservationLoop(a, 10*time.Millisecond)
	
	ctx, cancel := context.WithCancel(context.Background())
	go loop.Start(ctx)
	
	time.Sleep(50 * time.Millisecond)
	cancel()
}

func TestExecutor_Inject(t *testing.T) {
	// Stub
}
