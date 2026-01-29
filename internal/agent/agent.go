package agent

import (
	"context"

	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go"
	"github.com/stateforward/hsm-go/muid"
)

// AgentActor wraps the public Agent struct and manages its state via HSM.
type AgentActor struct {
	hsm.HSM
	data api.Agent
}

// NewAgent creates a new AgentActor.
func NewAgent(name, adapter, repoRoot string) *AgentActor {
	id := muid.Make() // 64-bit snowflake ID
	slug := api.NewAgentSlug(name)

	a := &AgentActor{
		data: api.Agent{
			ID:       id,
			Slug:     slug,
			Name:     name,
			Adapter:  adapter,
			RepoRoot: repoRoot,
			State:    api.StatePending,
			Presence: api.PresenceOffline,
		},
	}

	// Initialize HSM
	// We must use hsm.New to initialize the embedded HSM before starting.
	a = hsm.New(a, &AgentModel)

	// Start the HSM
	hsm.Start(context.Background(), a, &AgentModel)
	return a
}

// ID returns the agent's ID.
func (a *AgentActor) ID() api.AgentID {
	return a.data.ID
}

// Data returns a copy of the public agent data.
func (a *AgentActor) Data() api.Agent {
	return a.data
}

// Start initiates the agent start sequence.
func (a *AgentActor) Start() {
	hsm.Dispatch(context.Background(), a, hsm.Event{Name: EventStart})
}

// Stop initiates the agent stop sequence.
func (a *AgentActor) Stop() {
	hsm.Dispatch(context.Background(), a, hsm.Event{Name: EventStop})
}

// SendActivity signals activity to the agent presence HSM.
func (a *AgentActor) SendActivity() {
	hsm.Dispatch(context.Background(), a, hsm.Event{Name: EventActivityDetected})
}
