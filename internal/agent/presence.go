// Package agent implements presence state machine
// using hsm-go library. This handles presence states: Online ↔ Busy ↔ Offline ↔ Away
package agent

import (
	"context"
	"log/slog"
	"time"

	"github.com/stateforward/hsm-go"

	"github.com/agentflare-ai/amux/pkg/api"
)

// PresenceStateMachine represents agent presence state machine.
// States: Online ↔ Busy ↔ Offline ↔ Away
type PresenceStateMachine struct {
	hsm.HSM
	agentID api.AgentID
	model   *hsm.Model
	ctx     context.Context
	info    *api.PresenceInfo
}

// PresenceEvent represents events that can trigger presence state transitions.
var (
	// EventGoOnline triggers transition to online presence.
	EventGoOnline = hsm.Event{Name: "go_online"}

	// EventGoBusy triggers transition to busy presence.
	EventGoBusy = hsm.Event{Name: "go_busy"}

	// EventGoOffline triggers transition to offline presence.
	EventGoOffline = hsm.Event{Name: "go_offline"}

	// EventGoAway triggers transition to away presence.
	EventGoAway = hsm.Event{Name: "go_away"}

	// EventActivityDetected updates last activity time and may trigger away→online.
	EventActivityDetected = hsm.Event{Name: "activity_detected"}

	// EventInactivityTimeout triggers transition from online/Busy to away.
	EventInactivityTimeout = hsm.Event{Name: "inactivity_timeout"}
)

// NewPresenceStateMachine creates a new presence state machine for an agent.
func NewPresenceStateMachine(ctx context.Context, agentID api.AgentID) *PresenceStateMachine {
	psm := &PresenceStateMachine{
		agentID: agentID,
		ctx:     ctx,
		info:    api.NewPresenceInfo(agentID, api.PresenceOnline),
	}

	// Build state machine model
	psm.model = psm.buildStateMachine()

	// Start the state machine
	hsm.Started(ctx, psm, psm.model)

	return psm
}

// buildStateMachine constructs the presence state machine model.
func (psm *PresenceStateMachine) buildStateMachine() *hsm.Model {
	agentID := api.IDToString(psm.agentID)

	return hsm.Define(
		"presence_"+agentID,

		// Define states
		hsm.State("online",
			hsm.Transition(
				hsm.On(EventGoBusy),
				hsm.Source("online"),
				hsm.Target("busy"),
				hsm.Effect(func(ctx context.Context, psm *PresenceStateMachine, event hsm.Event) {
					psm.logPresenceChange("online", "busy")
				}),
			),
			hsm.Transition(
				hsm.On(EventGoOffline),
				hsm.Source("online"),
				hsm.Target("offline"),
				hsm.Effect(func(ctx context.Context, psm *PresenceStateMachine, event hsm.Event) {
					psm.logPresenceChange("online", "offline")
				}),
			),
			hsm.Transition(
				hsm.On(EventGoAway),
				hsm.Source("online"),
				hsm.Target("away"),
				hsm.Effect(func(ctx context.Context, psm *PresenceStateMachine, event hsm.Event) {
					psm.logPresenceChange("online", "away")
				}),
			),
		),

		hsm.State("busy",
			hsm.Transition(
				hsm.On(EventGoOnline),
				hsm.Source("busy"),
				hsm.Target("online"),
				hsm.Effect(func(ctx context.Context, psm *PresenceStateMachine, event hsm.Event) {
					psm.logPresenceChange("busy", "online")
				}),
			),
			hsm.Transition(
				hsm.On(EventGoOffline),
				hsm.Source("busy"),
				hsm.Target("offline"),
				hsm.Effect(func(ctx context.Context, psm *PresenceStateMachine, event hsm.Event) {
					psm.logPresenceChange("busy", "offline")
				}),
			),
			hsm.Transition(
				hsm.On(EventGoAway),
				hsm.Source("busy"),
				hsm.Target("away"),
				hsm.Effect(func(ctx context.Context, psm *PresenceStateMachine, event hsm.Event) {
					psm.logPresenceChange("busy", "away")
				}),
			),
		),

		hsm.State("offline",
			hsm.Transition(
				hsm.On(EventGoOnline),
				hsm.Source("offline"),
				hsm.Target("online"),
				hsm.Effect(func(ctx context.Context, psm *PresenceStateMachine, event hsm.Event) {
					psm.logPresenceChange("offline", "online")
				}),
			),
		),

		hsm.State("away",
			hsm.Transition(
				hsm.On(EventGoOnline),
				hsm.Source("away"),
				hsm.Target("online"),
				hsm.Effect(func(ctx context.Context, psm *PresenceStateMachine, event hsm.Event) {
					psm.logPresenceChange("away", "online")
				}),
			),
			hsm.Transition(
				hsm.On(EventActivityDetected),
				hsm.Source("away"),
				hsm.Target("online"),
				hsm.Effect(func(ctx context.Context, psm *PresenceStateMachine, event hsm.Event) {
					psm.logPresenceChange("away", "online")
					psm.updateActivity()
				}),
			),
		),

		// Set initial state
		hsm.Initial(hsm.Target("online")),
	)
}

// logPresenceChange logs presence state transitions for debugging.
func (psm *PresenceStateMachine) logPresenceChange(from, to string) {
	slog.Info("Presence state transition",
		"agent_id", psm.agentID,
		"from", from,
		"to", to,
	)
}

// updateActivity updates the last activity timestamp and log.
func (psm *PresenceStateMachine) updateActivity() {
	now := time.Now()
	psm.info.UpdatePresence(api.PresenceOnline)
	psm.info.LastActivity = now

	slog.Info("Activity detected",
		"agent_id", psm.agentID,
		"timestamp", now,
	)
}

// CurrentState returns the current presence state.
func (psm *PresenceStateMachine) CurrentState() string {
	return hsm.State(psm)
}

// CurrentPresenceInfo returns the current presence information.
func (psm *PresenceStateMachine) CurrentPresenceInfo() *api.PresenceInfo {
	return psm.info
}

// GoOnline transitions the agent to online presence.
func (psm *PresenceStateMachine) GoOnline() error {
	<-hsm.Dispatch(psm.ctx, psm, EventGoOnline)
	return nil
}

// GoBusy transitions the agent to busy presence.
func (psm *PresenceStateMachine) GoBusy() error {
	<-hsm.Dispatch(psm.ctx, psm, EventGoBusy)
	return nil
}

// GoOffline transitions the agent to offline presence.
func (psm *PresenceStateMachine) GoOffline() error {
	<-hsm.Dispatch(psm.ctx, psm, EventGoOffline)
	return nil
}

// GoAway transitions the agent to away presence.
func (psm *PresenceStateMachine) GoAway() error {
	<-hsm.Dispatch(psm.ctx, psm, EventGoAway)
	return nil
}

// ActivityDetected updates activity and may trigger away→online transition.
func (psm *PresenceStateMachine) ActivityDetected() error {
	<-hsm.Dispatch(psm.ctx, psm, EventActivityDetected)
	return nil
}

// InactivityTimeout transitions from online/Busy to away after timeout.
func (psm *PresenceStateMachine) InactivityTimeout() error {
	<-hsm.Dispatch(psm.ctx, psm, EventInactivityTimeout)
	return nil
}

// AgentID returns the agent ID associated with this presence state machine.
func (psm *PresenceStateMachine) AgentID() api.AgentID {
	return psm.agentID
}
