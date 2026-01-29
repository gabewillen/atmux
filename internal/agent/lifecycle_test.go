package agent

import (
	"context"
	"testing"

	"github.com/stateforward/amux/pkg/api"
)

func TestLifecycleTransitions(t *testing.T) {
	ctx := context.Background()

	agent := &api.Agent{
		ID:      api.GenerateID(),
		Name:    "test-agent",
		About:   "A test agent",
		Adapter: "test-adapter",
	}

	actor := NewAgentActor(ctx, agent)

	// Verify initial state is pending
	if actor.GetSimpleState() != StatePending {
		t.Errorf("Initial state = %q, want %q", actor.GetSimpleState(), StatePending)
	}

	// Test Pending → Starting
	actor.StartAgent(ctx)

	if actor.GetSimpleState() != StateStarting {
		t.Errorf("After Start, state = %q, want %q", actor.GetSimpleState(), StateStarting)
	}

	// Test Starting → Running
	actor.Ready(ctx)

	if actor.GetSimpleState() != StateRunning {
		t.Errorf("After Ready, state = %q, want %q", actor.GetSimpleState(), StateRunning)
	}

	// Test Running → Terminated
	actor.StopAgent(ctx)

	if actor.GetSimpleState() != StateTerminated {
		t.Errorf("After Stop, state = %q, want %q", actor.GetSimpleState(), StateTerminated)
	}
}

func TestLifecycleErrorTransition(t *testing.T) {
	ctx := context.Background()

	agent := &api.Agent{
		ID:      api.GenerateID(),
		Name:    "test-agent",
		About:   "A test agent",
		Adapter: "test-adapter",
	}

	actor := NewAgentActor(ctx, agent)

	// Error from pending state
	actor.ErrorAgent(ctx, nil)

	if actor.GetSimpleState() != StateErrored {
		t.Errorf("After Error, state = %q, want %q", actor.GetSimpleState(), StateErrored)
	}
}

func TestLifecycleErrorFromRunning(t *testing.T) {
	ctx := context.Background()

	agent := &api.Agent{
		ID:      api.GenerateID(),
		Name:    "test-agent",
		About:   "A test agent",
		Adapter: "test-adapter",
	}

	actor := NewAgentActor(ctx, agent)

	// Transition to Running
	actor.StartAgent(ctx)
	actor.Ready(ctx)

	if actor.GetSimpleState() != StateRunning {
		t.Fatalf("Expected running state, got %q", actor.GetSimpleState())
	}

	// Error from running state (per spec §5.4, error can be triggered from any state)
	actor.ErrorAgent(ctx, nil)

	if actor.GetSimpleState() != StateErrored {
		t.Errorf("After Error from running, state = %q, want %q", actor.GetSimpleState(), StateErrored)
	}
}

func TestLifecycleModel(t *testing.T) {
	// Verify the lifecycle model is defined correctly
	if LifecycleModel.Name() != "agent.lifecycle" {
		t.Errorf("LifecycleModel.Name() = %q, want %q", LifecycleModel.Name(), "agent.lifecycle")
	}

	// Verify all required states exist in the model
	members := LifecycleModel.Members()
	// States are qualified, so we need to check for the full name
	requiredStates := []string{
		"/agent.lifecycle/pending",
		"/agent.lifecycle/starting",
		"/agent.lifecycle/running",
		"/agent.lifecycle/terminated",
		"/agent.lifecycle/errored",
	}
	for _, state := range requiredStates {
		if _, ok := members[state]; !ok {
			t.Errorf("LifecycleModel missing required state: %q", state)
		}
	}
}
