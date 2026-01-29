// Package agent provides agent orchestration: lifecycle, presence, and messaging.
// actor.go composes lifecycle and presence HSMs and wires dispatch.
package agent

import (
	"context"
	"sync"

	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/stateforward/hsm-go"
)

// Actor holds an agent's data and its lifecycle and presence state machines.
// Lifecycle and presence transitions emit events via the configured Dispatcher.
type Actor struct {
	Agent      *api.Agent
	Dispatcher protocol.Dispatcher

	lifecycle hsm.Instance
	presence  hsm.Instance
	mu        sync.RWMutex
}

// NewActor creates an actor for the given agent and dispatcher.
// The agent ID MUST be a valid runtime ID (non-zero). Call Start to run the HSMs.
func NewActor(agent *api.Agent, d protocol.Dispatcher) (*Actor, error) {
	if agent == nil {
		return nil, api.ErrInvalidConfig
	}
	if !api.ValidRuntimeID(agent.ID) {
		return nil, api.ErrInvalidConfig
	}
	return &Actor{
		Agent:      agent,
		Dispatcher: d,
	}, nil
}

// Start starts the lifecycle and presence HSMs. Call once after NewActor.
func (a *Actor) Start(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()

	lifecycleSM := &lifecycleActor{
		AgentID:   a.Agent.ID,
		Dispatcher: a.Dispatcher,
	}
	a.lifecycle = hsm.Started(ctx, lifecycleSM, &LifecycleModel)

	presenceSM := &presenceActor{
		AgentID:   a.Agent.ID,
		Dispatcher: a.Dispatcher,
	}
	a.presence = hsm.Started(ctx, presenceSM, &PresenceModel)
}

// LifecycleState returns the current lifecycle state name (e.g. agent.lifecycle.pending).
func (a *Actor) LifecycleState() string {
	a.mu.RLock()
	inst := a.lifecycle
	a.mu.RUnlock()
	if inst == nil {
		return LifecyclePending
	}
	return inst.State()
}

// PresenceState returns the current presence state name (e.g. agent.presence.online).
func (a *Actor) PresenceState() string {
	a.mu.RLock()
	inst := a.presence
	a.mu.RUnlock()
	if inst == nil {
		return PresenceOnline
	}
	return inst.State()
}

// DispatchLifecycle sends an event to the lifecycle HSM (start, ready, stop, error).
// It blocks until the transition is processed.
func (a *Actor) DispatchLifecycle(ctx context.Context, eventName string, data interface{}) {
	a.mu.RLock()
	inst := a.lifecycle
	a.mu.RUnlock()
	if inst == nil {
		return
	}
	<-hsm.Dispatch(ctx, inst, hsm.Event{Name: eventName, Data: data})
}

// DispatchPresence sends an event to the presence HSM (task.assigned, activity.detected, etc.).
// It blocks until the transition is processed.
func (a *Actor) DispatchPresence(ctx context.Context, eventName string, data interface{}) {
	a.mu.RLock()
	inst := a.presence
	a.mu.RUnlock()
	if inst == nil {
		return
	}
	<-hsm.Dispatch(ctx, inst, hsm.Event{Name: eventName, Data: data})
}
