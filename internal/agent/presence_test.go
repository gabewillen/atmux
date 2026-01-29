package agent

import (
	"context"
	"testing"
)

func TestPresenceTransitions(t *testing.T) {
	ctx := context.Background()
	presence := NewPresenceActor(ctx, "test-agent")

	// Verify initial state is online
	if presence.GetSimplePresenceState() != StateOnline {
		t.Errorf("Initial presence = %q, want %q", presence.GetSimplePresenceState(), StateOnline)
	}

	// Test Online → Busy
	presence.TaskAssigned(ctx)
	if presence.GetSimplePresenceState() != StateBusy {
		t.Errorf("After TaskAssigned, presence = %q, want %q", presence.GetSimplePresenceState(), StateBusy)
	}

	// Test Busy → Online (via task completed)
	presence.TaskCompleted(ctx)
	if presence.GetSimplePresenceState() != StateOnline {
		t.Errorf("After TaskCompleted, presence = %q, want %q", presence.GetSimplePresenceState(), StateOnline)
	}

	// Test Online → Busy → Online (via prompt detected)
	presence.TaskAssigned(ctx)
	presence.PromptDetected(ctx)
	if presence.GetSimplePresenceState() != StateOnline {
		t.Errorf("After PromptDetected, presence = %q, want %q", presence.GetSimplePresenceState(), StateOnline)
	}
}

func TestPresenceRateLimiting(t *testing.T) {
	ctx := context.Background()
	presence := NewPresenceActor(ctx, "test-agent")

	// Test Online → Offline (rate limit)
	presence.RateLimit(ctx)
	if presence.GetSimplePresenceState() != StateOffline {
		t.Errorf("After RateLimit from online, presence = %q, want %q", presence.GetSimplePresenceState(), StateOffline)
	}

	// Test Offline → Online (rate cleared)
	presence.RateCleared(ctx)
	if presence.GetSimplePresenceState() != StateOnline {
		t.Errorf("After RateCleared, presence = %q, want %q", presence.GetSimplePresenceState(), StateOnline)
	}

	// Test Busy → Offline (rate limit)
	presence.TaskAssigned(ctx)
	presence.RateLimit(ctx)
	if presence.GetSimplePresenceState() != StateOffline {
		t.Errorf("After RateLimit from busy, presence = %q, want %q", presence.GetSimplePresenceState(), StateOffline)
	}
}

func TestPresenceAwayTransitions(t *testing.T) {
	ctx := context.Background()

	// Test stuck from Online
	presence := NewPresenceActor(ctx, "test-agent-1")
	presence.StuckDetected(ctx)
	if presence.GetSimplePresenceState() != StateAway {
		t.Errorf("After StuckDetected from online, presence = %q, want %q", presence.GetSimplePresenceState(), StateAway)
	}

	// Test stuck from Busy
	presence2 := NewPresenceActor(ctx, "test-agent-2")
	presence2.TaskAssigned(ctx)
	presence2.StuckDetected(ctx)
	if presence2.GetSimplePresenceState() != StateAway {
		t.Errorf("After StuckDetected from busy, presence = %q, want %q", presence2.GetSimplePresenceState(), StateAway)
	}

	// Test stuck from Offline
	presence3 := NewPresenceActor(ctx, "test-agent-3")
	presence3.RateLimit(ctx)
	presence3.StuckDetected(ctx)
	if presence3.GetSimplePresenceState() != StateAway {
		t.Errorf("After StuckDetected from offline, presence = %q, want %q", presence3.GetSimplePresenceState(), StateAway)
	}

	// Test Away → Online (activity detected)
	presence.ActivityDetected(ctx)
	if presence.GetSimplePresenceState() != StateOnline {
		t.Errorf("After ActivityDetected, presence = %q, want %q", presence.GetSimplePresenceState(), StateOnline)
	}
}

func TestPresenceModel(t *testing.T) {
	// Verify the presence model is defined correctly
	if PresenceModel.Name() != "agent.presence" {
		t.Errorf("PresenceModel.Name() = %q, want %q", PresenceModel.Name(), "agent.presence")
	}

	// Verify all required states exist in the model
	members := PresenceModel.Members()
	// States are qualified
	requiredStates := []string{
		"/agent.presence/online",
		"/agent.presence/busy",
		"/agent.presence/offline",
		"/agent.presence/away",
	}
	for _, state := range requiredStates {
		if _, ok := members[state]; !ok {
			t.Errorf("PresenceModel missing required state: %q", state)
		}
	}
}
