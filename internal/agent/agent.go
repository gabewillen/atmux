package agent

import (
	"context"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go"
)

// Agent represents a runtime agent instance.
type Agent struct {
	hsm.HSM
	ID      api.ID
	Adapter adapter.Adapter
}

// NewAgent constructs a new agent with a fresh ID.
func NewAgent(adapter adapter.Adapter) *Agent {
	return &Agent{ID: api.NewID(), Adapter: adapter}
}

// Start starts the agent state machine.
func (a *Agent) Start(ctx context.Context, model *hsm.Model) {
	hsm.Started(ctx, a, model)
}
