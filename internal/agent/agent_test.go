// Package agent implements tests for agent orchestration (lifecycle, presence, messaging)
package agent

import (
	"context"
	"testing"

	"github.com/stateforward/hsm-go/muid"
	"github.com/stateforward/amux/pkg/api"
)

// TestNewAgentActor tests creating a new agent actor
func TestNewAgentActor(t *testing.T) {
	agentData := &api.Agent{
		ID:      muid.Make(),
		Name:    "test-agent",
		Adapter: "test-adapter",
	}

	actor, err := NewAgentActor(agentData, nil)
	if err != nil {
		t.Fatalf("Unexpected error creating agent actor: %v", err)
	}

	if actor == nil {
		t.Fatal("Expected agent actor to be created")
	}

	if actor.ID != agentData.ID {
		t.Errorf("Expected ID %v, got %v", agentData.ID, actor.ID)
	}

	if actor.Agent.Name != agentData.Name {
		t.Errorf("Expected name %s, got %s", agentData.Name, actor.Agent.Name)
	}

	// Check initial states
	if actor.CurrentLifecycleState() != LifecyclePending {
		t.Errorf("Expected initial lifecycle state %s, got %s", LifecyclePending, actor.CurrentLifecycleState())
	}

	if actor.CurrentPresenceState() != PresenceOffline {
		t.Errorf("Expected initial presence state %s, got %s", PresenceOffline, actor.CurrentPresenceState())
	}
}

// TestAgentActorLifecycle tests the agent lifecycle state machine
func TestAgentActorLifecycle(t *testing.T) {
	agentData := &api.Agent{
		ID:      muid.Make(),
		Name:    "test-agent",
		Adapter: "test-adapter",
	}

	actor, err := NewAgentActor(agentData, nil)
	if err != nil {
		t.Fatalf("Unexpected error creating agent actor: %v", err)
	}

	ctx := context.Background()

	// Test Pending -> Starting
	err = actor.Start(ctx)
	if err != nil {
		t.Fatalf("Unexpected error starting agent: %v", err)
	}

	if actor.CurrentLifecycleState() != LifecycleStarting {
		t.Errorf("Expected lifecycle state %s, got %s", LifecycleStarting, actor.CurrentLifecycleState())
	}

	// Test Starting -> Running
	err = actor.Ready(ctx)
	if err != nil {
		t.Fatalf("Unexpected error setting agent ready: %v", err)
	}

	if actor.CurrentLifecycleState() != LifecycleRunning {
		t.Errorf("Expected lifecycle state %s, got %s", LifecycleRunning, actor.CurrentLifecycleState())
	}

	// Test Running -> Terminated
	err = actor.Terminate(ctx)
	if err != nil {
		t.Fatalf("Unexpected error terminating agent: %v", err)
	}

	if actor.CurrentLifecycleState() != LifecycleTerminated {
		t.Errorf("Expected lifecycle state %s, got %s", LifecycleTerminated, actor.CurrentLifecycleState())
	}
}

// TestAgentActorPresence tests the agent presence state machine
func TestAgentActorPresence(t *testing.T) {
	agentData := &api.Agent{
		ID:      muid.Make(),
		Name:    "test-agent",
		Adapter: "test-adapter",
	}

	actor, err := NewAgentActor(agentData, nil)
	if err != nil {
		t.Fatalf("Unexpected error creating agent actor: %v", err)
	}

	ctx := context.Background()

	// Test Offline -> Online
	err = actor.Connect(ctx)
	if err != nil {
		t.Fatalf("Unexpected error connecting agent: %v", err)
	}

	if actor.CurrentPresenceState() != PresenceOnline {
		t.Errorf("Expected presence state %s, got %s", PresenceOnline, actor.CurrentPresenceState())
	}

	// Test Online -> Busy
	err = actor.SetBusy(ctx)
	if err != nil {
		t.Fatalf("Unexpected error setting agent busy: %v", err)
	}

	if actor.CurrentPresenceState() != PresenceBusy {
		t.Errorf("Expected presence state %s, got %s", PresenceBusy, actor.CurrentPresenceState())
	}

	// Test Busy -> Online
	err = actor.SetAvailable(ctx)
	if err != nil {
		t.Fatalf("Unexpected error setting agent available: %v", err)
	}

	if actor.CurrentPresenceState() != PresenceOnline {
		t.Errorf("Expected presence state %s, got %s", PresenceOnline, actor.CurrentPresenceState())
	}

	// Test Online -> Away
	err = actor.SetAway(ctx)
	if err != nil {
		t.Fatalf("Unexpected error setting agent away: %v", err)
	}

	if actor.CurrentPresenceState() != PresenceAway {
		t.Errorf("Expected presence state %s, got %s", PresenceAway, actor.CurrentPresenceState())
	}

	// Test Away -> Online
	err = actor.SetBack(ctx)
	if err != nil {
		t.Fatalf("Unexpected error setting agent back: %v", err)
	}

	if actor.CurrentPresenceState() != PresenceOnline {
		t.Errorf("Expected presence state %s, got %s", PresenceOnline, actor.CurrentPresenceState())
	}

	// Test Online -> Offline
	err = actor.Disconnect(ctx)
	if err != nil {
		t.Fatalf("Unexpected error disconnecting agent: %v", err)
	}

	if actor.CurrentPresenceState() != PresenceOffline {
		t.Errorf("Expected presence state %s, got %s", PresenceOffline, actor.CurrentPresenceState())
	}
}

// TestAgentActorErrorStates tests error state transitions
func TestAgentActorErrorStates(t *testing.T) {
	agentData := &api.Agent{
		ID:      muid.Make(),
		Name:    "test-agent",
		Adapter: "test-adapter",
	}

	actor, err := NewAgentActor(agentData, nil)
	if err != nil {
		t.Fatalf("Unexpected error creating agent actor: %v", err)
	}

	ctx := context.Background()

	// Test error transition from Running state
	err = actor.Start(ctx)
	if err != nil {
		t.Fatalf("Unexpected error starting agent: %v", err)
	}
	err = actor.Ready(ctx)
	if err != nil {
		t.Fatalf("Unexpected error setting agent ready: %v", err)
	}

	err = actor.Error(ctx, nil)
	if err != nil {
		t.Fatalf("Unexpected error setting agent error: %v", err)
	}

	if actor.CurrentLifecycleState() != LifecycleErrored {
		t.Errorf("Expected lifecycle state %s, got %s", LifecycleErrored, actor.CurrentLifecycleState())
	}

	// Reset for next test
	actor, err = NewAgentActor(agentData, nil)
	if err != nil {
		t.Fatalf("Unexpected error creating agent actor: %v", err)
	}

	// Test fatal error transition from Pending state
	err = actor.FatalError(ctx, nil)
	if err != nil {
		t.Fatalf("Unexpected error setting agent fatal error: %v", err)
	}

	if actor.CurrentLifecycleState() != LifecycleErrored {
		t.Errorf("Expected lifecycle state %s, got %s", LifecycleErrored, actor.CurrentLifecycleState())
	}
}