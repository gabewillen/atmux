// Package agent provides agent orchestration: lifecycle, presence, and messaging.
// presence.go implements the Presence HSM per spec §4.2.3, §6.1, §6.5.
package agent

import (
	"context"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go"
)

// Presence state names (spec §6.1); HSM returns qualified names like /agent.presence/online.
const (
	PresenceOnline  = "/agent.presence/online"
	PresenceBusy     = "/agent.presence/busy"
	PresenceOffline  = "/agent.presence/offline"
	PresenceAway     = "/agent.presence/away"
)

// Presence event names for dispatch (spec §6.5, §5.5.8).
const (
	EventPresenceTaskAssigned        = "task.assigned"
	EventPresenceTaskCompleted      = "task.completed"
	EventPresencePromptDetected      = "prompt.detected"
	EventPresenceRateLimit          = "rate.limit"
	EventPresenceRateCleared        = "rate.cleared"
	EventPresenceStuckDetected      = "stuck.detected"
	EventPresenceActivityDetected   = "activity.detected"
	EventPresenceConnectionLost    = "connection.lost"    // Remote disconnect → Away (spec §5.5.8)
	EventPresenceConnectionRecovered = "connection.recovered" // Reconnect + replay → Online (spec §5.5.8)
)

// presenceActor holds HSM state and dispatch hook for agent presence.
type presenceActor struct {
	hsm.HSM
	AgentID   api.ID
	Dispatcher protocol.Dispatcher
}

// PresenceModel defines the presence HSM (spec §6.5).
// Online ↔ Busy ↔ Offline ↔ Away.
var PresenceModel = hsm.Define("agent.presence",
	hsm.State("online"),
	hsm.State("busy"),
	hsm.State("offline"),
	hsm.State("away"),

	// Online ↔ Busy
	hsm.Transition(hsm.On(hsm.Event{Name: EventPresenceTaskAssigned}), hsm.Source("online"), hsm.Target("busy"),
		hsm.Effect(func(ctx context.Context, a *presenceActor, _ hsm.Event) {
			emitPresenceChanged(ctx, a.Dispatcher, a.AgentID, PresenceBusy)
		})),
	hsm.Transition(hsm.On(hsm.Event{Name: EventPresenceTaskCompleted}), hsm.Source("busy"), hsm.Target("online"),
		hsm.Effect(func(ctx context.Context, a *presenceActor, _ hsm.Event) {
			emitPresenceChanged(ctx, a.Dispatcher, a.AgentID, PresenceOnline)
		})),
	hsm.Transition(hsm.On(hsm.Event{Name: EventPresencePromptDetected}), hsm.Source("busy"), hsm.Target("online"),
		hsm.Effect(func(ctx context.Context, a *presenceActor, _ hsm.Event) {
			emitPresenceChanged(ctx, a.Dispatcher, a.AgentID, PresenceOnline)
		})),

	// → Offline (rate limited)
	hsm.Transition(hsm.On(hsm.Event{Name: EventPresenceRateLimit}), hsm.Source("busy"), hsm.Target("offline"),
		hsm.Effect(func(ctx context.Context, a *presenceActor, _ hsm.Event) {
			emitPresenceChanged(ctx, a.Dispatcher, a.AgentID, PresenceOffline)
		})),
	hsm.Transition(hsm.On(hsm.Event{Name: EventPresenceRateLimit}), hsm.Source("online"), hsm.Target("offline"),
		hsm.Effect(func(ctx context.Context, a *presenceActor, _ hsm.Event) {
			emitPresenceChanged(ctx, a.Dispatcher, a.AgentID, PresenceOffline)
		})),
	hsm.Transition(hsm.On(hsm.Event{Name: EventPresenceRateCleared}), hsm.Source("offline"), hsm.Target("online"),
		hsm.Effect(func(ctx context.Context, a *presenceActor, _ hsm.Event) {
			emitPresenceChanged(ctx, a.Dispatcher, a.AgentID, PresenceOnline)
		})),

	// → Away (unresponsive) from any state (spec §6.5)
	hsm.Transition(hsm.On(hsm.Event{Name: EventPresenceStuckDetected}), hsm.Source("online"), hsm.Target("away"),
		hsm.Effect(func(ctx context.Context, a *presenceActor, _ hsm.Event) {
			emitPresenceChanged(ctx, a.Dispatcher, a.AgentID, PresenceAway)
		})),
	hsm.Transition(hsm.On(hsm.Event{Name: EventPresenceStuckDetected}), hsm.Source("busy"), hsm.Target("away"),
		hsm.Effect(func(ctx context.Context, a *presenceActor, _ hsm.Event) {
			emitPresenceChanged(ctx, a.Dispatcher, a.AgentID, PresenceAway)
		})),
	hsm.Transition(hsm.On(hsm.Event{Name: EventPresenceStuckDetected}), hsm.Source("offline"), hsm.Target("away"),
		hsm.Effect(func(ctx context.Context, a *presenceActor, _ hsm.Event) {
			emitPresenceChanged(ctx, a.Dispatcher, a.AgentID, PresenceAway)
		})),
	hsm.Transition(hsm.On(hsm.Event{Name: EventPresenceActivityDetected}), hsm.Source("away"), hsm.Target("online"),
		hsm.Effect(func(ctx context.Context, a *presenceActor, _ hsm.Event) {
			emitPresenceChanged(ctx, a.Dispatcher, a.AgentID, PresenceOnline)
		})),

	// Remote disconnect → Away from any state (spec §5.5.8)
	hsm.Transition(hsm.On(hsm.Event{Name: EventPresenceConnectionLost}), hsm.Source("online"), hsm.Target("away"),
		hsm.Effect(func(ctx context.Context, a *presenceActor, _ hsm.Event) {
			emitPresenceChanged(ctx, a.Dispatcher, a.AgentID, PresenceAway)
		})),
	hsm.Transition(hsm.On(hsm.Event{Name: EventPresenceConnectionLost}), hsm.Source("busy"), hsm.Target("away"),
		hsm.Effect(func(ctx context.Context, a *presenceActor, _ hsm.Event) {
			emitPresenceChanged(ctx, a.Dispatcher, a.AgentID, PresenceAway)
		})),
	hsm.Transition(hsm.On(hsm.Event{Name: EventPresenceConnectionLost}), hsm.Source("offline"), hsm.Target("away"),
		hsm.Effect(func(ctx context.Context, a *presenceActor, _ hsm.Event) {
			emitPresenceChanged(ctx, a.Dispatcher, a.AgentID, PresenceAway)
		})),
	// Reconnect + replay → Online from Away (spec §5.5.8)
	hsm.Transition(hsm.On(hsm.Event{Name: EventPresenceConnectionRecovered}), hsm.Source("away"), hsm.Target("online"),
		hsm.Effect(func(ctx context.Context, a *presenceActor, _ hsm.Event) {
			emitPresenceChanged(ctx, a.Dispatcher, a.AgentID, PresenceOnline)
		})),

	hsm.Initial(hsm.Target("online")),
)

func emitPresenceChanged(ctx context.Context, d protocol.Dispatcher, agentID api.ID, state string) {
	if d == nil {
		return
	}
	_ = d.Dispatch(ctx, protocol.Event{
		Type: "presence.changed",
		Data: map[string]interface{}{"agent_id": api.EncodeID(agentID), "state": state},
	})
}
