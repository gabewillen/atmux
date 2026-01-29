package agent

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go"
)

type recordDispatcher struct {
	mu       sync.Mutex
	subjects []string
	events   []protocol.Event
}

func (r *recordDispatcher) Publish(ctx context.Context, subject string, event protocol.Event) error {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	r.subjects = append(r.subjects, subject)
	r.events = append(r.events, event)
	return nil
}

func (r *recordDispatcher) Subscribe(ctx context.Context, subject string, handler func(protocol.Event)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}

func (r *recordDispatcher) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error {
	_ = ctx
	_ = subject
	_ = payload
	_ = reply
	return nil
}

func (r *recordDispatcher) SubscribeRaw(ctx context.Context, subject string, handler func(protocol.Message)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}

func (r *recordDispatcher) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (protocol.Message, error) {
	_ = ctx
	_ = subject
	_ = payload
	_ = timeout
	return protocol.Message{}, nil
}

func (r *recordDispatcher) MaxPayload() int {
	return 1024 * 1024
}

func (r *recordDispatcher) Closed() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func TestLifecycleTransitions(t *testing.T) {
	dispatcher := &recordDispatcher{}
	agent := &Agent{Agent: api.Agent{ID: api.NewAgentID()}, dispatcher: dispatcher}
	lifecycle, err := NewLifecycle(agent, dispatcher)
	if err != nil {
		t.Fatalf("new lifecycle: %v", err)
	}
	started := hsm.Started(context.Background(), lifecycle, &LifecycleModel)
	<-hsm.Dispatch(started.Context(), started, hsm.Event{Name: EventStart})
	if started.State() != "/agent.lifecycle/starting" {
		t.Fatalf("unexpected state: %s", started.State())
	}
	<-hsm.Dispatch(started.Context(), started, hsm.Event{Name: EventReady})
	if started.State() != "/agent.lifecycle/running" {
		t.Fatalf("unexpected state: %s", started.State())
	}
	<-hsm.Dispatch(started.Context(), started, hsm.Event{Name: EventStop})
	if started.State() != "/agent.lifecycle/terminated" {
		t.Fatalf("unexpected state: %s", started.State())
	}
	if len(dispatcher.events) == 0 {
		t.Fatalf("expected lifecycle events")
	}
}

func TestPresenceTransitions(t *testing.T) {
	dispatcher := &recordDispatcher{}
	agent := &Agent{Agent: api.Agent{ID: api.NewAgentID()}, dispatcher: dispatcher}
	presence, err := NewPresence(agent, dispatcher)
	if err != nil {
		t.Fatalf("new presence: %v", err)
	}
	started := hsm.Started(context.Background(), presence, &PresenceModel)
	if started.State() != "/agent.presence/online" {
		t.Fatalf("unexpected state: %s", started.State())
	}
	<-hsm.Dispatch(started.Context(), started, hsm.Event{Name: EventTaskAssigned})
	if started.State() != "/agent.presence/busy" {
		t.Fatalf("unexpected state: %s", started.State())
	}
	<-hsm.Dispatch(started.Context(), started, hsm.Event{Name: EventRateLimit})
	if started.State() != "/agent.presence/offline" {
		t.Fatalf("unexpected state: %s", started.State())
	}
	<-hsm.Dispatch(started.Context(), started, hsm.Event{Name: EventRateCleared})
	if started.State() != "/agent.presence/online" {
		t.Fatalf("unexpected state: %s", started.State())
	}
	<-hsm.Dispatch(started.Context(), started, hsm.Event{Name: EventStuckDetected})
	if started.State() != "/agent.presence/away" {
		t.Fatalf("unexpected state: %s", started.State())
	}
	<-hsm.Dispatch(started.Context(), started, hsm.Event{Name: EventActivity})
	if started.State() != "/agent.presence/online" {
		t.Fatalf("unexpected state: %s", started.State())
	}
	if len(dispatcher.events) == 0 {
		t.Fatalf("expected presence events")
	}
}

func TestLifecycleShutdownTransitions(t *testing.T) {
	dispatcher := &recordDispatcher{}
	agent := &Agent{Agent: api.Agent{ID: api.NewAgentID()}, dispatcher: dispatcher}
	lifecycle, err := NewLifecycle(agent, dispatcher)
	if err != nil {
		t.Fatalf("new lifecycle: %v", err)
	}
	started := hsm.Started(context.Background(), lifecycle, &LifecycleModel)
	<-hsm.Dispatch(started.Context(), started, hsm.Event{Name: EventStart})
	<-hsm.Dispatch(started.Context(), started, hsm.Event{Name: EventReady})
	<-hsm.Dispatch(started.Context(), started, hsm.Event{Name: EventShutdownInitiated})
	if started.State() != "/agent.lifecycle/terminated" {
		t.Fatalf("unexpected state: %s", started.State())
	}
}

func TestLifecycleShutdownForceTransitions(t *testing.T) {
	dispatcher := &recordDispatcher{}
	agent := &Agent{Agent: api.Agent{ID: api.NewAgentID()}, dispatcher: dispatcher}
	lifecycle, err := NewLifecycle(agent, dispatcher)
	if err != nil {
		t.Fatalf("new lifecycle: %v", err)
	}
	started := hsm.Started(context.Background(), lifecycle, &LifecycleModel)
	<-hsm.Dispatch(started.Context(), started, hsm.Event{Name: EventStart})
	<-hsm.Dispatch(started.Context(), started, hsm.Event{Name: EventReady})
	<-hsm.Dispatch(started.Context(), started, hsm.Event{Name: EventShutdownForce})
	if started.State() != "/agent.lifecycle/terminated" {
		t.Fatalf("unexpected state: %s", started.State())
	}
}
