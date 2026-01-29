package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
)

// Agent represents a runtime agent instance with lifecycle and presence state machines.
type Agent struct {
	api.Agent
	Lifecycle  *Lifecycle
	Presence   *Presence
	dispatcher protocol.Dispatcher
	mu         sync.RWMutex
	lastErr    error
}

// NewAgent constructs a new agent with lifecycle and presence state machines.
func NewAgent(meta api.Agent, dispatcher protocol.Dispatcher) (*Agent, error) {
	if dispatcher == nil {
		dispatcher = &protocol.NoopDispatcher{}
	}
	if meta.ID.IsZero() {
		meta.ID = api.NewAgentID()
	}
	if err := meta.Validate(); err != nil {
		return nil, fmt.Errorf("new agent: %w", err)
	}
	agent := &Agent{
		Agent:      meta,
		dispatcher: dispatcher,
	}
	agent.Lifecycle = NewLifecycle(agent, dispatcher)
	agent.Presence = NewPresence(agent, dispatcher)
	return agent, nil
}

// Start starts the lifecycle and presence state machines.
func (a *Agent) Start(ctx context.Context) {
	if a.Lifecycle != nil {
		a.Lifecycle.Start(ctx)
	}
	if a.Presence != nil {
		a.Presence.Start(ctx)
	}
}

// LastError returns the last error observed by the agent state machines.
func (a *Agent) LastError() error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.lastErr
}

func (a *Agent) recordError(err error) {
	if err == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.lastErr = err
}
