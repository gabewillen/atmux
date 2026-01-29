package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	hsm "github.com/stateforward/hsm-go"

	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestLifecycleHSMInitialState(t *testing.T) {
	agent := &Agent{
		Agent:     api.Agent{ID: 123, Name: "test", Slug: "test"},
		lifecycle: api.LifecyclePending,
	}
	lhsm := NewLifecycleHSM(agent, nil)

	if lhsm.LifecycleState() != api.LifecyclePending {
		t.Errorf("Initial state = %q, want %q", lhsm.LifecycleState(), api.LifecyclePending)
	}
}

func TestLifecycleHSMPendingToStarting(t *testing.T) {
	agent := &Agent{
		Agent:     api.Agent{ID: 123, Name: "test", Slug: "test"},
		lifecycle: api.LifecyclePending,
	}
	lhsm := NewLifecycleHSM(agent, event.NewNoopDispatcher())

	ctx := context.Background()
	instance := lhsm.Start(ctx)

	// Dispatch "start" event
	<-DispatchStart(ctx, instance)

	// Wait for state machine to process
	time.Sleep(10 * time.Millisecond)

	if lhsm.LifecycleState() != api.LifecycleStarting {
		t.Errorf("After start, state = %q, want %q", lhsm.LifecycleState(), api.LifecycleStarting)
	}

	// Verify agent state is synchronized
	if agent.Lifecycle() != api.LifecycleStarting {
		t.Errorf("Agent lifecycle = %q, want %q", agent.Lifecycle(), api.LifecycleStarting)
	}

	<-hsm.Stop(ctx, instance)
}

func TestLifecycleHSMStartingToRunning(t *testing.T) {
	agent := &Agent{
		Agent:     api.Agent{ID: 123, Name: "test", Slug: "test"},
		lifecycle: api.LifecyclePending,
	}
	lhsm := NewLifecycleHSM(agent, event.NewNoopDispatcher())

	ctx := context.Background()
	instance := lhsm.Start(ctx)

	// Transition to Starting first
	<-DispatchStart(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	// Dispatch "ready" event
	<-DispatchReady(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	if lhsm.LifecycleState() != api.LifecycleRunning {
		t.Errorf("After ready, state = %q, want %q", lhsm.LifecycleState(), api.LifecycleRunning)
	}

	<-hsm.Stop(ctx, instance)
}

func TestLifecycleHSMRunningToTerminated(t *testing.T) {
	agent := &Agent{
		Agent:     api.Agent{ID: 123, Name: "test", Slug: "test"},
		lifecycle: api.LifecyclePending,
	}
	lhsm := NewLifecycleHSM(agent, event.NewNoopDispatcher())

	ctx := context.Background()
	instance := lhsm.Start(ctx)

	// Full lifecycle: pending → starting → running → terminated
	<-DispatchStart(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	<-DispatchReady(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	<-DispatchStop(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	if lhsm.LifecycleState() != api.LifecycleTerminated {
		t.Errorf("After stop, state = %q, want %q", lhsm.LifecycleState(), api.LifecycleTerminated)
	}

	<-hsm.Stop(ctx, instance)
}

func TestLifecycleHSMErrorFromPending(t *testing.T) {
	agent := &Agent{
		Agent:     api.Agent{ID: 123, Name: "test", Slug: "test"},
		lifecycle: api.LifecyclePending,
	}
	lhsm := NewLifecycleHSM(agent, event.NewNoopDispatcher())

	ctx := context.Background()
	instance := lhsm.Start(ctx)

	testErr := errors.New("test error")
	<-DispatchError(ctx, instance, testErr)
	time.Sleep(10 * time.Millisecond)

	if lhsm.LifecycleState() != api.LifecycleErrored {
		t.Errorf("After error, state = %q, want %q", lhsm.LifecycleState(), api.LifecycleErrored)
	}

	if lhsm.LastError() == nil {
		t.Error("LastError() should not be nil")
	} else if lhsm.LastError().Error() != testErr.Error() {
		t.Errorf("LastError() = %q, want %q", lhsm.LastError().Error(), testErr.Error())
	}

	<-hsm.Stop(ctx, instance)
}

func TestLifecycleHSMErrorFromStarting(t *testing.T) {
	agent := &Agent{
		Agent:     api.Agent{ID: 123, Name: "test", Slug: "test"},
		lifecycle: api.LifecyclePending,
	}
	lhsm := NewLifecycleHSM(agent, event.NewNoopDispatcher())

	ctx := context.Background()
	instance := lhsm.Start(ctx)

	<-DispatchStart(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	testErr := errors.New("startup failed")
	<-DispatchError(ctx, instance, testErr)
	time.Sleep(10 * time.Millisecond)

	if lhsm.LifecycleState() != api.LifecycleErrored {
		t.Errorf("After error from starting, state = %q, want %q", lhsm.LifecycleState(), api.LifecycleErrored)
	}

	<-hsm.Stop(ctx, instance)
}

func TestLifecycleHSMErrorFromRunning(t *testing.T) {
	agent := &Agent{
		Agent:     api.Agent{ID: 123, Name: "test", Slug: "test"},
		lifecycle: api.LifecyclePending,
	}
	lhsm := NewLifecycleHSM(agent, event.NewNoopDispatcher())

	ctx := context.Background()
	instance := lhsm.Start(ctx)

	<-DispatchStart(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	<-DispatchReady(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	testErr := errors.New("runtime crash")
	<-DispatchError(ctx, instance, testErr)
	time.Sleep(10 * time.Millisecond)

	if lhsm.LifecycleState() != api.LifecycleErrored {
		t.Errorf("After error from running, state = %q, want %q", lhsm.LifecycleState(), api.LifecycleErrored)
	}

	<-hsm.Stop(ctx, instance)
}

func TestLifecycleHSMInvalidTransition(t *testing.T) {
	agent := &Agent{
		Agent:     api.Agent{ID: 123, Name: "test", Slug: "test"},
		lifecycle: api.LifecyclePending,
	}
	lhsm := NewLifecycleHSM(agent, event.NewNoopDispatcher())

	ctx := context.Background()
	instance := lhsm.Start(ctx)

	// Try to go directly from pending to running (should be ignored)
	<-DispatchReady(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	// State should still be pending
	if lhsm.LifecycleState() != api.LifecyclePending {
		t.Errorf("After invalid transition, state = %q, want %q", lhsm.LifecycleState(), api.LifecyclePending)
	}

	// Try to stop from pending (should be ignored)
	<-DispatchStop(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	if lhsm.LifecycleState() != api.LifecyclePending {
		t.Errorf("After invalid stop, state = %q, want %q", lhsm.LifecycleState(), api.LifecyclePending)
	}

	<-hsm.Stop(ctx, instance)
}

func TestLifecycleHSMWithEventCollection(t *testing.T) {
	agent := &Agent{
		Agent:     api.Agent{ID: 123, Name: "test", Slug: "test"},
		lifecycle: api.LifecyclePending,
	}

	// Use a local dispatcher to collect events
	dispatcher := event.NewLocalDispatcher()
	lhsm := NewLifecycleHSM(agent, dispatcher)

	var receivedEvents []event.Event
	var mu = make(chan struct{}, 1)
	mu <- struct{}{}

	dispatcher.Subscribe(event.Subscription{
		Types: []event.Type{
			event.TypeAgentStarting,
			event.TypeAgentStarted,
			event.TypeAgentStopping,
			event.TypeAgentTerminated,
		},
		Handler: func(ctx context.Context, e event.Event) error {
			<-mu
			receivedEvents = append(receivedEvents, e)
			mu <- struct{}{}
			return nil
		},
	})

	ctx := context.Background()
	instance := lhsm.Start(ctx)

	// Full lifecycle
	<-DispatchStart(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	<-DispatchReady(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	<-DispatchStop(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	<-hsm.Stop(ctx, instance)

	// Check collected events
	<-mu
	expectedTypes := []event.Type{
		event.TypeAgentStarting,
		event.TypeAgentStarted,
		event.TypeAgentStopping,
		event.TypeAgentTerminated,
	}

	if len(receivedEvents) != len(expectedTypes) {
		t.Errorf("Received %d events, want %d", len(receivedEvents), len(expectedTypes))
	}

	for i, expected := range expectedTypes {
		if i < len(receivedEvents) && receivedEvents[i].Type != expected {
			t.Errorf("Event[%d].Type = %q, want %q", i, receivedEvents[i].Type, expected)
		}
	}
	mu <- struct{}{}
}
