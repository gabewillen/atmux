package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/stateforward/hsm-go"
)

// EventRouter routes lifecycle and presence events through NATS subjects.
type EventRouter struct {
	agent      *Agent
	dispatcher protocol.Dispatcher
	mu         sync.Mutex
	started    bool
	subs       []protocol.Subscription
}

// NewEventRouter constructs a router for an agent.
func NewEventRouter(agent *Agent, dispatcher protocol.Dispatcher) *EventRouter {
	return &EventRouter{agent: agent, dispatcher: dispatcher}
}

// Start subscribes to agent event subjects.
func (r *EventRouter) Start(ctx context.Context) error {
	if r == nil || r.dispatcher == nil {
		return fmt.Errorf("router start: %w", ErrDispatcherRequired)
	}
	r.mu.Lock()
	if r.started {
		r.mu.Unlock()
		return nil
	}
	r.started = true
	r.mu.Unlock()
	lifecycleSubject := protocol.Subject("events", "agent", r.agent.ID.String(), "lifecycle")
	presenceSubject := protocol.Subject("events", "agent", r.agent.ID.String(), "presence")
	lifecycleSub, err := r.dispatcher.Subscribe(ctx, lifecycleSubject, func(event protocol.Event) {
		hsm.Dispatch(context.Background(), r.agent.Lifecycle, hsm.Event{Name: event.Name, Data: event.Payload})
	})
	if err != nil {
		return fmt.Errorf("router subscribe: %w", err)
	}
	presenceSub, err := r.dispatcher.Subscribe(ctx, presenceSubject, func(event protocol.Event) {
		hsm.Dispatch(context.Background(), r.agent.Presence, hsm.Event{Name: event.Name, Data: event.Payload})
	})
	if err != nil {
		_ = lifecycleSub.Unsubscribe()
		return fmt.Errorf("router subscribe: %w", err)
	}
	r.mu.Lock()
	r.subs = append(r.subs, lifecycleSub, presenceSub)
	r.mu.Unlock()
	return nil
}

// EmitLifecycle publishes a lifecycle event.
func (r *EventRouter) EmitLifecycle(ctx context.Context, name string, payload any) error {
	return r.emit(ctx, protocol.Subject("events", "agent", r.agent.ID.String(), "lifecycle"), name, payload)
}

// EmitPresence publishes a presence event.
func (r *EventRouter) EmitPresence(ctx context.Context, name string, payload any) error {
	return r.emit(ctx, protocol.Subject("events", "agent", r.agent.ID.String(), "presence"), name, payload)
}

func (r *EventRouter) emit(ctx context.Context, subject string, name string, payload any) error {
	if r == nil || r.dispatcher == nil {
		return fmt.Errorf("router emit: %w", ErrDispatcherRequired)
	}
	if name == "" {
		return fmt.Errorf("router emit: %w", ErrDispatcherRequired)
	}
	event := protocol.Event{Name: name, Payload: payload, OccurredAt: time.Now().UTC()}
	if err := r.dispatcher.Publish(ctx, subject, event); err != nil {
		return fmt.Errorf("router emit: %w", err)
	}
	return nil
}
