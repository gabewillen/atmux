package agent

import (
	"context"
	"strings"

	"github.com/stateforward/hsm-go"
)

// Presence state constants matching spec §6.1.
const (
	StateOnline  = "online"
	StateBusy    = "busy"
	StateOffline = "offline"
	StateAway    = "away"
)

// Presence event constants per spec §6.5.
const (
	EventTaskAssigned   = "task.assigned"
	EventTaskCompleted  = "task.completed"
	EventPromptDetected = "prompt.detected"
	EventRateLimit      = "rate.limit"
	EventRateCleared    = "rate.cleared"
	EventStuckDetected  = "stuck.detected"
	EventActivityDetected = "activity.detected"
)

// PresenceModel defines the agent presence state machine per spec §6.5.
//
// State diagram:
//
//	                    ┌──────────────────┐
//	                    ▼                  │
//	┌────────┐    ┌─────────┐    ┌────────┐
//	│ Online │◀──▶│  Busy   │───▶│ Offline│
//	└────────┘    └─────────┘    └────────┘
//	     ▲              │              │
//	     │              ▼              │
//	     │         ┌────────┐          │
//	     └─────────│  Away  │◀─────────┘
//	               └────────┘
var PresenceModel = hsm.Define("agent.presence",
	hsm.State(StateOnline),
	hsm.State(StateBusy),
	hsm.State(StateOffline),
	hsm.State(StateAway),

	// Online ↔ Busy
	hsm.Transition(hsm.On(hsm.Event{Name: EventTaskAssigned}), hsm.Source(StateOnline), hsm.Target(StateBusy)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventTaskCompleted}), hsm.Source(StateBusy), hsm.Target(StateOnline)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventPromptDetected}), hsm.Source(StateBusy), hsm.Target(StateOnline)),

	// → Offline (rate limited)
	hsm.Transition(hsm.On(hsm.Event{Name: EventRateLimit}), hsm.Source(StateBusy), hsm.Target(StateOffline)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventRateLimit}), hsm.Source(StateOnline), hsm.Target(StateOffline)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventRateCleared}), hsm.Source(StateOffline), hsm.Target(StateOnline)),

	// → Away (unresponsive) from any state
	hsm.Transition(hsm.On(hsm.Event{Name: EventStuckDetected}), hsm.Source(StateOnline), hsm.Target(StateAway)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventStuckDetected}), hsm.Source(StateBusy), hsm.Target(StateAway)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventStuckDetected}), hsm.Source(StateOffline), hsm.Target(StateAway)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventActivityDetected}), hsm.Source(StateAway), hsm.Target(StateOnline)),

	hsm.Initial(hsm.Target(StateOnline)),
)

// PresenceActor wraps an Agent with a presence state machine.
// Per spec §6.1, presence indicates whether an agent can accept tasks.
type PresenceActor struct {
	hsm.HSM
	AgentID string // For logging/debugging
}

// NewPresenceActor creates a new PresenceActor with initialized presence HSM.
// Per spec §6.1, presence starts in the "online" state.
func NewPresenceActor(ctx context.Context, agentID string) *PresenceActor {
	actor := &PresenceActor{
		AgentID: agentID,
	}

	// Initialize and start presence HSM
	actor = hsm.Started(ctx, actor, &PresenceModel)

	return actor
}

// GetPresenceState returns the current presence state.
func (p *PresenceActor) GetPresenceState() string {
	return p.State()
}

// GetSimplePresenceState returns the current presence state without the qualified name prefix.
// For example, returns "online" instead of "/agent.presence/online".
func (p *PresenceActor) GetSimplePresenceState() string {
	state := p.State()
	// Extract the last component after the final slash
	if idx := strings.LastIndex(state, "/"); idx >= 0 {
		return state[idx+1:]
	}
	return state
}

// TaskAssigned transitions from Online to Busy.
func (p *PresenceActor) TaskAssigned(ctx context.Context) {
	done := hsm.Dispatch(ctx, p, hsm.Event{
		Name: EventTaskAssigned,
	})
	<-done
}

// TaskCompleted transitions from Busy to Online.
func (p *PresenceActor) TaskCompleted(ctx context.Context) {
	done := hsm.Dispatch(ctx, p, hsm.Event{
		Name: EventTaskCompleted,
	})
	<-done
}

// PromptDetected transitions from Busy to Online.
func (p *PresenceActor) PromptDetected(ctx context.Context) {
	done := hsm.Dispatch(ctx, p, hsm.Event{
		Name: EventPromptDetected,
	})
	<-done
}

// RateLimit transitions to Offline.
func (p *PresenceActor) RateLimit(ctx context.Context) {
	done := hsm.Dispatch(ctx, p, hsm.Event{
		Name: EventRateLimit,
	})
	<-done
}

// RateCleared transitions from Offline to Online.
func (p *PresenceActor) RateCleared(ctx context.Context) {
	done := hsm.Dispatch(ctx, p, hsm.Event{
		Name: EventRateCleared,
	})
	<-done
}

// StuckDetected transitions to Away from any state.
func (p *PresenceActor) StuckDetected(ctx context.Context) {
	done := hsm.Dispatch(ctx, p, hsm.Event{
		Name: EventStuckDetected,
	})
	<-done
}

// ActivityDetected transitions from Away to Online.
func (p *PresenceActor) ActivityDetected(ctx context.Context) {
	done := hsm.Dispatch(ctx, p, hsm.Event{
		Name: EventActivityDetected,
	})
	<-done
}
