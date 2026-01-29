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
	router     *EventRouter
	dispatcher protocol.Dispatcher
	mu         sync.RWMutex
	lastErr    error
}

// NewAgent constructs a new agent with lifecycle and presence state machines.
func NewAgent(meta api.Agent, dispatcher protocol.Dispatcher) (*Agent, error) {
	if dispatcher == nil {
		return nil, fmt.Errorf("new agent: %w", ErrDispatcherRequired)
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
	agent.router = NewEventRouter(agent, dispatcher)
	lifecycle, err := NewLifecycle(agent, dispatcher)
	if err != nil {
		return nil, fmt.Errorf("new agent: %w", err)
	}
	presence, err := NewPresence(agent, dispatcher)
	if err != nil {
		return nil, fmt.Errorf("new agent: %w", err)
	}
	agent.Lifecycle = lifecycle
	agent.Presence = presence
	return agent, nil
}

// Start starts the lifecycle and presence state machines.
func (a *Agent) Start(ctx context.Context) {
	if a.router != nil {
		if err := a.router.Start(ctx); err != nil {
			a.recordError(err)
		}
	}
	if a.Lifecycle != nil {
		a.Lifecycle.Start(ctx)
	}
	if a.Presence != nil {
		a.Presence.Start(ctx)
	}
}

// EmitLifecycle publishes a lifecycle event through the dispatcher.
func (a *Agent) EmitLifecycle(ctx context.Context, name string, payload any) error {
	if a.router == nil {
		return fmt.Errorf("emit lifecycle: %w", ErrDispatcherRequired)
	}
	if err := a.router.EmitLifecycle(ctx, name, payload); err != nil {
		a.recordError(err)
		return err
	}
	return nil
}

// EmitPresence publishes a presence event through the dispatcher.
func (a *Agent) EmitPresence(ctx context.Context, name string, payload any) error {
	if a.router == nil {
		return fmt.Errorf("emit presence: %w", ErrDispatcherRequired)
	}
	if err := a.router.EmitPresence(ctx, name, payload); err != nil {
		a.recordError(err)
		return err
	}
	return nil
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
