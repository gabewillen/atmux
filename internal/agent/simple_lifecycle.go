package agent

import (
	"context"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go"
	"log/slog"
)

// SimpleAgentLifecycle is a simplified test to understand hsm-go API
type SimpleAgentLifecycle struct {
	hsm.HSM
	ctx   context.Context
	agent *api.Agent
}

// NewSimpleAgentLifecycle creates a simple lifecycle state machine
func NewSimpleAgentLifecycle(ctx context.Context, agent *api.Agent) *SimpleAgentLifecycle {
	lifecycle := &SimpleAgentLifecycle{
		agent: agent,
	}

	// Define a simple model with 2 states and 1 transition
	model := hsm.Define(
		"simple_test",
		hsm.State("idle"),
		hsm.State("running"),
		hsm.Transition(
			hsm.On(hsm.Event{Name: "start"}),
			hsm.Source("idle"),
			hsm.Target("running"),
			hsm.Effect(func(ctx context.Context, sal *SimpleAgentLifecycle, event hsm.Event) {
				slog.Info("Transition: idle → running")
			}),
		),
		hsm.Initial(hsm.Target("idle")),
	)

	// Start the state machine
	hsm.Started(ctx, lifecycle, &model)

	return lifecycle
}

// SimpleAgentLifecycleEvent represents test events
var (
	SimpleStart = hsm.Event{Name: "start"}
)

// Start starts the agent
func (sal *SimpleAgentLifecycle) Start() error {
	<-hsm.Dispatch(sal.ctx, sal, SimpleStart)
	return nil
}

// Agent returns the associated agent
func (sal *SimpleAgentLifecycle) Agent() *api.Agent {
	return sal.agent
}
