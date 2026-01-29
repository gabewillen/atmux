// Package agent provides agent orchestration: lifecycle, presence, and messaging.
// lifecycle.go implements the Agent lifecycle HSM per spec §4.2.3, §5.4.
package agent

import (
	"context"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go"
)

// Lifecycle state names (spec §5.4); HSM returns qualified names like /agent.lifecycle/pending.
const (
	LifecyclePending    = "/agent.lifecycle/pending"
	LifecycleStarting   = "/agent.lifecycle/starting"
	LifecycleRunning    = "/agent.lifecycle/running"
	LifecycleTerminated = "/agent.lifecycle/terminated"
	LifecycleErrored    = "/agent.lifecycle/errored"
)

// Lifecycle event names for dispatch (spec §5.4).
const (
	EventLifecycleStart   = "start"
	EventLifecycleReady   = "ready"
	EventLifecycleStop    = "stop"
	EventLifecycleError  = "error"
)

// lifecycleActor holds HSM state and dispatch hook for agent lifecycle.
type lifecycleActor struct {
	hsm.HSM
	AgentID   api.ID
	Dispatcher protocol.Dispatcher
}

// LifecycleModel defines the agent lifecycle HSM (spec §5.4).
// Pending → Starting → Running → Terminated/Errored.
var LifecycleModel = hsm.Define("agent.lifecycle",
	hsm.State("pending"),
	hsm.State("starting"),
	hsm.State("running"),
	hsm.State("terminated", hsm.Final("terminated")),
	hsm.State("errored", hsm.Final("errored")),

	hsm.Transition(hsm.On(hsm.Event{Name: EventLifecycleStart}), hsm.Source("pending"), hsm.Target("starting"),
		hsm.Effect(func(ctx context.Context, a *lifecycleActor, _ hsm.Event) {
			emitLifecycleChanged(ctx, a.Dispatcher, a.AgentID, LifecycleStarting)
		})),
	hsm.Transition(hsm.On(hsm.Event{Name: EventLifecycleReady}), hsm.Source("starting"), hsm.Target("running"),
		hsm.Effect(func(ctx context.Context, a *lifecycleActor, _ hsm.Event) {
			emitLifecycleChanged(ctx, a.Dispatcher, a.AgentID, LifecycleRunning)
		})),
	hsm.Transition(hsm.On(hsm.Event{Name: EventLifecycleStop}), hsm.Source("running"), hsm.Target("terminated"),
		hsm.Effect(func(ctx context.Context, a *lifecycleActor, _ hsm.Event) {
			emitLifecycleChanged(ctx, a.Dispatcher, a.AgentID, LifecycleTerminated)
		})),
	// error from any non-final state (spec §5.4)
	hsm.Transition(hsm.On(hsm.Event{Name: EventLifecycleError}), hsm.Source("pending"), hsm.Target("errored"),
		hsm.Effect(func(ctx context.Context, a *lifecycleActor, _ hsm.Event) {
			emitLifecycleChanged(ctx, a.Dispatcher, a.AgentID, LifecycleErrored)
		})),
	hsm.Transition(hsm.On(hsm.Event{Name: EventLifecycleError}), hsm.Source("starting"), hsm.Target("errored"),
		hsm.Effect(func(ctx context.Context, a *lifecycleActor, _ hsm.Event) {
			emitLifecycleChanged(ctx, a.Dispatcher, a.AgentID, LifecycleErrored)
		})),
	hsm.Transition(hsm.On(hsm.Event{Name: EventLifecycleError}), hsm.Source("running"), hsm.Target("errored"),
		hsm.Effect(func(ctx context.Context, a *lifecycleActor, _ hsm.Event) {
			emitLifecycleChanged(ctx, a.Dispatcher, a.AgentID, LifecycleErrored)
		})),

	hsm.Initial(hsm.Target("pending")),
)

func emitLifecycleChanged(ctx context.Context, d protocol.Dispatcher, agentID api.ID, state string) {
	if d == nil {
		return
	}
	_ = d.Dispatch(ctx, protocol.Event{
		Type: "lifecycle.changed",
		Data: map[string]interface{}{"agent_id": api.EncodeID(agentID), "state": state},
	})
}
