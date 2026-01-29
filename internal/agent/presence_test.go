package agent

import (
	"context"
	"testing"
	"time"

	hsm "github.com/stateforward/hsm-go"

	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestPresenceHSMInitialState(t *testing.T) {
	agent := &Agent{
		Agent:    api.Agent{ID: 123, Name: "test", Slug: "test"},
		presence: api.PresenceOnline,
	}
	phsm := NewPresenceHSM(agent, nil)

	if phsm.PresenceState() != api.PresenceOnline {
		t.Errorf("Initial state = %q, want %q", phsm.PresenceState(), api.PresenceOnline)
	}
}

func TestPresenceHSMOnlineToBusy(t *testing.T) {
	agent := &Agent{
		Agent:    api.Agent{ID: 123, Name: "test", Slug: "test"},
		presence: api.PresenceOnline,
	}
	phsm := NewPresenceHSM(agent, event.NewNoopDispatcher())

	ctx := context.Background()
	instance := phsm.Start(ctx)

	// Dispatch task.assigned event
	<-DispatchTaskAssigned(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	if phsm.PresenceState() != api.PresenceBusy {
		t.Errorf("After task.assigned, state = %q, want %q", phsm.PresenceState(), api.PresenceBusy)
	}

	// Verify agent state is synchronized
	if agent.Presence() != api.PresenceBusy {
		t.Errorf("Agent presence = %q, want %q", agent.Presence(), api.PresenceBusy)
	}

	<-hsm.Stop(ctx, instance)
}

func TestPresenceHSMBusyToOnlineViaTaskCompleted(t *testing.T) {
	agent := &Agent{
		Agent:    api.Agent{ID: 123, Name: "test", Slug: "test"},
		presence: api.PresenceOnline,
	}
	phsm := NewPresenceHSM(agent, event.NewNoopDispatcher())

	ctx := context.Background()
	instance := phsm.Start(ctx)

	// Go to Busy first
	<-DispatchTaskAssigned(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	// Complete task
	<-DispatchTaskCompleted(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	if phsm.PresenceState() != api.PresenceOnline {
		t.Errorf("After task.completed, state = %q, want %q", phsm.PresenceState(), api.PresenceOnline)
	}

	<-hsm.Stop(ctx, instance)
}

func TestPresenceHSMBusyToOnlineViaPromptDetected(t *testing.T) {
	agent := &Agent{
		Agent:    api.Agent{ID: 123, Name: "test", Slug: "test"},
		presence: api.PresenceOnline,
	}
	phsm := NewPresenceHSM(agent, event.NewNoopDispatcher())

	ctx := context.Background()
	instance := phsm.Start(ctx)

	// Go to Busy first
	<-DispatchTaskAssigned(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	// Prompt detected
	<-DispatchPromptDetected(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	if phsm.PresenceState() != api.PresenceOnline {
		t.Errorf("After prompt.detected, state = %q, want %q", phsm.PresenceState(), api.PresenceOnline)
	}

	<-hsm.Stop(ctx, instance)
}

func TestPresenceHSMOnlineToOffline(t *testing.T) {
	agent := &Agent{
		Agent:    api.Agent{ID: 123, Name: "test", Slug: "test"},
		presence: api.PresenceOnline,
	}
	phsm := NewPresenceHSM(agent, event.NewNoopDispatcher())

	ctx := context.Background()
	instance := phsm.Start(ctx)

	// Rate limit from Online
	<-DispatchRateLimit(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	if phsm.PresenceState() != api.PresenceOffline {
		t.Errorf("After rate.limit from online, state = %q, want %q", phsm.PresenceState(), api.PresenceOffline)
	}

	<-hsm.Stop(ctx, instance)
}

func TestPresenceHSMBusyToOffline(t *testing.T) {
	agent := &Agent{
		Agent:    api.Agent{ID: 123, Name: "test", Slug: "test"},
		presence: api.PresenceOnline,
	}
	phsm := NewPresenceHSM(agent, event.NewNoopDispatcher())

	ctx := context.Background()
	instance := phsm.Start(ctx)

	// Go to Busy first
	<-DispatchTaskAssigned(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	// Rate limit from Busy
	<-DispatchRateLimit(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	if phsm.PresenceState() != api.PresenceOffline {
		t.Errorf("After rate.limit from busy, state = %q, want %q", phsm.PresenceState(), api.PresenceOffline)
	}

	<-hsm.Stop(ctx, instance)
}

func TestPresenceHSMOfflineToOnline(t *testing.T) {
	agent := &Agent{
		Agent:    api.Agent{ID: 123, Name: "test", Slug: "test"},
		presence: api.PresenceOnline,
	}
	phsm := NewPresenceHSM(agent, event.NewNoopDispatcher())

	ctx := context.Background()
	instance := phsm.Start(ctx)

	// Go to Offline first
	<-DispatchRateLimit(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	// Rate cleared
	<-DispatchRateCleared(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	if phsm.PresenceState() != api.PresenceOnline {
		t.Errorf("After rate.cleared, state = %q, want %q", phsm.PresenceState(), api.PresenceOnline)
	}

	<-hsm.Stop(ctx, instance)
}

func TestPresenceHSMOnlineToAway(t *testing.T) {
	agent := &Agent{
		Agent:    api.Agent{ID: 123, Name: "test", Slug: "test"},
		presence: api.PresenceOnline,
	}
	phsm := NewPresenceHSM(agent, event.NewNoopDispatcher())

	ctx := context.Background()
	instance := phsm.Start(ctx)

	// Stuck detected from Online
	<-DispatchStuckDetected(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	if phsm.PresenceState() != api.PresenceAway {
		t.Errorf("After stuck.detected from online, state = %q, want %q", phsm.PresenceState(), api.PresenceAway)
	}

	<-hsm.Stop(ctx, instance)
}

func TestPresenceHSMBusyToAway(t *testing.T) {
	agent := &Agent{
		Agent:    api.Agent{ID: 123, Name: "test", Slug: "test"},
		presence: api.PresenceOnline,
	}
	phsm := NewPresenceHSM(agent, event.NewNoopDispatcher())

	ctx := context.Background()
	instance := phsm.Start(ctx)

	// Go to Busy first
	<-DispatchTaskAssigned(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	// Stuck detected from Busy
	<-DispatchStuckDetected(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	if phsm.PresenceState() != api.PresenceAway {
		t.Errorf("After stuck.detected from busy, state = %q, want %q", phsm.PresenceState(), api.PresenceAway)
	}

	<-hsm.Stop(ctx, instance)
}

func TestPresenceHSMAwayToOnline(t *testing.T) {
	agent := &Agent{
		Agent:    api.Agent{ID: 123, Name: "test", Slug: "test"},
		presence: api.PresenceOnline,
	}
	phsm := NewPresenceHSM(agent, event.NewNoopDispatcher())

	ctx := context.Background()
	instance := phsm.Start(ctx)

	// Go to Away first
	<-DispatchStuckDetected(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	// Activity detected
	<-DispatchActivityDetected(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	if phsm.PresenceState() != api.PresenceOnline {
		t.Errorf("After activity.detected, state = %q, want %q", phsm.PresenceState(), api.PresenceOnline)
	}

	<-hsm.Stop(ctx, instance)
}

func TestPresenceHSMInvalidTransitions(t *testing.T) {
	agent := &Agent{
		Agent:    api.Agent{ID: 123, Name: "test", Slug: "test"},
		presence: api.PresenceOnline,
	}
	phsm := NewPresenceHSM(agent, event.NewNoopDispatcher())

	ctx := context.Background()
	instance := phsm.Start(ctx)

	// Try task.completed from Online (invalid - should be ignored)
	<-DispatchTaskCompleted(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	if phsm.PresenceState() != api.PresenceOnline {
		t.Errorf("After invalid task.completed, state = %q, want %q (unchanged)", phsm.PresenceState(), api.PresenceOnline)
	}

	// Try rate.cleared from Online (invalid - should be ignored)
	<-DispatchRateCleared(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	if phsm.PresenceState() != api.PresenceOnline {
		t.Errorf("After invalid rate.cleared, state = %q, want %q (unchanged)", phsm.PresenceState(), api.PresenceOnline)
	}

	<-hsm.Stop(ctx, instance)
}

func TestMonitorEventsToPresenceHSM(t *testing.T) {
	dispatcher := event.NewLocalDispatcher()
	mgr := NewManagerWithResolver(dispatcher, nil)
	ctx := context.Background()

	repoRoot := initTestRepo(t)

	ag, err := mgr.Add(ctx, api.Agent{
		Name:     "monitor-test",
		Adapter:  "claude-code",
		RepoRoot: repoRoot,
	})
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	agentID := ag.ID

	// Verify initial state is Online
	phsm := mgr.PresenceHSMFor(agentID)
	if phsm == nil {
		t.Fatal("PresenceHSMFor returned nil")
	}
	if phsm.PresenceState() != api.PresenceOnline {
		t.Fatalf("Initial state = %q, want %q", phsm.PresenceState(), api.PresenceOnline)
	}

	// Dispatch pty.activity event -> should transition to Busy
	_ = dispatcher.Dispatch(ctx, event.NewEvent(event.TypePTYActivity, agentID, nil))
	time.Sleep(50 * time.Millisecond)

	if phsm.PresenceState() != api.PresenceBusy {
		t.Errorf("After pty.activity, state = %q, want %q", phsm.PresenceState(), api.PresenceBusy)
	}

	// Dispatch pty.idle event -> should transition from Busy to Online
	_ = dispatcher.Dispatch(ctx, event.NewEvent(event.TypePTYIdle, agentID, nil))
	time.Sleep(50 * time.Millisecond)

	if phsm.PresenceState() != api.PresenceOnline {
		t.Errorf("After pty.idle, state = %q, want %q", phsm.PresenceState(), api.PresenceOnline)
	}

	// Dispatch pty.stuck event -> should transition to Away
	_ = dispatcher.Dispatch(ctx, event.NewEvent(event.TypePTYStuck, agentID, nil))
	time.Sleep(50 * time.Millisecond)

	if phsm.PresenceState() != api.PresenceAway {
		t.Errorf("After pty.stuck, state = %q, want %q", phsm.PresenceState(), api.PresenceAway)
	}
}

func TestPresenceHSMEventDispatch(t *testing.T) {
	agent := &Agent{
		Agent:    api.Agent{ID: 123, Name: "test", Slug: "test"},
		presence: api.PresenceOnline,
	}

	dispatcher := event.NewLocalDispatcher()
	phsm := NewPresenceHSM(agent, dispatcher)

	var receivedEvents []event.Event
	var mu = make(chan struct{}, 1)
	mu <- struct{}{}

	dispatcher.Subscribe(event.Subscription{
		Types: []event.Type{event.TypePresenceChanged},
		Handler: func(ctx context.Context, e event.Event) error {
			<-mu
			receivedEvents = append(receivedEvents, e)
			mu <- struct{}{}
			return nil
		},
	})

	ctx := context.Background()
	instance := phsm.Start(ctx)

	// Trigger some transitions
	<-DispatchTaskAssigned(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	<-DispatchTaskCompleted(ctx, instance)
	time.Sleep(10 * time.Millisecond)

	<-hsm.Stop(ctx, instance)

	// Check events were dispatched
	<-mu
	if len(receivedEvents) < 2 {
		t.Errorf("Expected at least 2 presence changed events, got %d", len(receivedEvents))
	}
	mu <- struct{}{}
}
