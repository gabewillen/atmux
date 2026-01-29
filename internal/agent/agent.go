// Package agent implements agent orchestration (lifecycle, presence, messaging)
package agent

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/stateforward/hsm-go/muid"
	"github.com/stateforward/amux/internal/event"
	"github.com/stateforward/amux/pkg/api"
)

// ErrInvalidAgent is returned when an agent is invalid
var ErrInvalidAgent = errors.New("invalid agent")

// LifecycleState represents the agent lifecycle state
type LifecycleState string

const (
	LifecyclePending    LifecycleState = "pending"
	LifecycleStarting   LifecycleState = "starting"
	LifecycleRunning    LifecycleState = "running"
	LifecycleTerminated LifecycleState = "terminated"
	LifecycleErrored    LifecycleState = "errored"
)

// PresenceState represents the agent presence state
type PresenceState string

const (
	PresenceOnline PresenceState = "online"
	PresenceBusy   PresenceState = "busy"
	PresenceOffline PresenceState = "offline"
	PresenceAway   PresenceState = "away"
)

// AgentActor manages an agent's lifecycle and presence state machines
type AgentActor struct {
	ID             muid.MUID
	Agent          *api.Agent
	lifecycleState LifecycleState
	presenceState  PresenceState
	stateMutex     sync.RWMutex
	eventHandler   func(event interface{})
}

// NewAgentActor creates a new agent actor with initialized state machines
func NewAgentActor(agent *api.Agent, eventHandler func(event interface{})) (*AgentActor, error) {
	if agent == nil {
		return nil, fmt.Errorf("invalid agent: %w", ErrInvalidAgent)
	}

	actor := &AgentActor{
		ID:             agent.ID,
		Agent:          agent,
		lifecycleState: LifecyclePending,
		presenceState:  PresenceOffline,
		eventHandler:   eventHandler,
	}

	return actor, nil
}

// Start initiates the agent lifecycle by transitioning from Pending to Starting
func (a *AgentActor) Start(ctx context.Context) error {
	a.stateMutex.Lock()
	defer a.stateMutex.Unlock()

	if a.lifecycleState != LifecyclePending {
		return fmt.Errorf("cannot start agent: invalid lifecycle state: %w", ErrInvalidAgent)
	}

	oldState := a.lifecycleState
	a.lifecycleState = LifecycleStarting

	// Emit lifecycle change event
	if a.eventHandler != nil {
		a.eventHandler(map[string]interface{}{
			"type":     "lifecycle_change",
			"agent_id": a.ID,
			"from":     oldState,
			"to":       a.lifecycleState,
			"event":    "start",
		})
	}

	return nil
}

// Ready signals that the agent is ready, transitioning from Starting to Running
func (a *AgentActor) Ready(ctx context.Context) error {
	a.stateMutex.Lock()
	defer a.stateMutex.Unlock()

	if a.lifecycleState != LifecycleStarting {
		return fmt.Errorf("cannot mark agent as ready: invalid lifecycle state: %w", ErrInvalidAgent)
	}

	oldState := a.lifecycleState
	a.lifecycleState = LifecycleRunning

	// Emit lifecycle change event
	if a.eventHandler != nil {
		a.eventHandler(map[string]interface{}{
			"type":     "lifecycle_change",
			"agent_id": a.ID,
			"from":     oldState,
			"to":       a.lifecycleState,
			"event":    "ready",
		})
	}

	return nil
}

// Terminate signals graceful termination, transitioning to Terminated state
func (a *AgentActor) Terminate(ctx context.Context) error {
	a.stateMutex.Lock()
	defer a.stateMutex.Unlock()

	if a.lifecycleState != LifecycleRunning {
		return fmt.Errorf("cannot terminate agent: invalid lifecycle state: %w", ErrInvalidAgent)
	}

	oldState := a.lifecycleState
	a.lifecycleState = LifecycleTerminated

	// Emit lifecycle change event
	if a.eventHandler != nil {
		a.eventHandler(map[string]interface{}{
			"type":     "lifecycle_change",
			"agent_id": a.ID,
			"from":     oldState,
			"to":       a.lifecycleState,
			"event":    "terminate",
		})
	}

	return nil
}

// Error signals an error condition, transitioning to Errored state
func (a *AgentActor) Error(ctx context.Context, err error) error {
	a.stateMutex.Lock()
	defer a.stateMutex.Unlock()

	oldState := a.lifecycleState
	a.lifecycleState = LifecycleErrored

	// Emit lifecycle change event
	if a.eventHandler != nil {
		a.eventHandler(map[string]interface{}{
			"type":     "lifecycle_change",
			"agent_id": a.ID,
			"from":     oldState,
			"to":       a.lifecycleState,
			"event":    "error",
			"error":    err.Error(),
		})
	}

	return nil
}

// FatalError signals a fatal error, transitioning to Errored state
func (a *AgentActor) FatalError(ctx context.Context, err error) error {
	a.stateMutex.Lock()
	defer a.stateMutex.Unlock()

	oldState := a.lifecycleState
	a.lifecycleState = LifecycleErrored

	// Emit lifecycle change event
	if a.eventHandler != nil {
		a.eventHandler(map[string]interface{}{
			"type":     "lifecycle_change",
			"agent_id": a.ID,
			"from":     oldState,
			"to":       a.lifecycleState,
			"event":    "fatal_error",
			"error":    err.Error(),
		})
	}

	return nil
}

// CurrentLifecycleState returns the current lifecycle state
func (a *AgentActor) CurrentLifecycleState() LifecycleState {
	a.stateMutex.RLock()
	defer a.stateMutex.RUnlock()
	return a.lifecycleState
}

// Connect brings the agent online, transitioning from Offline to Online
func (a *AgentActor) Connect(ctx context.Context) error {
	a.stateMutex.Lock()
	defer a.stateMutex.Unlock()

	if a.presenceState != PresenceOffline {
		return fmt.Errorf("cannot connect agent: invalid presence state: %w", ErrInvalidAgent)
	}

	oldState := a.presenceState
	a.presenceState = PresenceOnline

	// Emit presence change event
	if a.eventHandler != nil {
		a.eventHandler(map[string]interface{}{
			"type":     "presence_change",
			"agent_id": a.ID,
			"from":     oldState,
			"to":       a.presenceState,
			"event":    "connect",
		})
	}

	return nil
}

// Disconnect takes the agent offline, transitioning to Offline state
func (a *AgentActor) Disconnect(ctx context.Context) error {
	a.stateMutex.Lock()
	defer a.stateMutex.Unlock()

	oldState := a.presenceState
	a.presenceState = PresenceOffline

	// Emit presence change event
	if a.eventHandler != nil {
		a.eventHandler(map[string]interface{}{
			"type":     "presence_change",
			"agent_id": a.ID,
			"from":     oldState,
			"to":       a.presenceState,
			"event":    "disconnect",
		})
	}

	return nil
}

// SetBusy marks the agent as busy, transitioning from Online to Busy
func (a *AgentActor) SetBusy(ctx context.Context) error {
	a.stateMutex.Lock()
	defer a.stateMutex.Unlock()

	if a.presenceState != PresenceOnline {
		return fmt.Errorf("cannot set agent busy: invalid presence state: %w", ErrInvalidAgent)
	}

	oldState := a.presenceState
	a.presenceState = PresenceBusy

	// Emit presence change event
	if a.eventHandler != nil {
		a.eventHandler(map[string]interface{}{
			"type":     "presence_change",
			"agent_id": a.ID,
			"from":     oldState,
			"to":       a.presenceState,
			"event":    "busy",
		})
	}

	return nil
}

// SetAvailable marks the agent as available, transitioning from Busy to Online
func (a *AgentActor) SetAvailable(ctx context.Context) error {
	a.stateMutex.Lock()
	defer a.stateMutex.Unlock()

	if a.presenceState != PresenceBusy {
		return fmt.Errorf("cannot set agent available: invalid presence state: %w", ErrInvalidAgent)
	}

	oldState := a.presenceState
	a.presenceState = PresenceOnline

	// Emit presence change event
	if a.eventHandler != nil {
		a.eventHandler(map[string]interface{}{
			"type":     "presence_change",
			"agent_id": a.ID,
			"from":     oldState,
			"to":       a.presenceState,
			"event":    "available",
		})
	}

	return nil
}

// SetAway marks the agent as away, transitioning from Online to Away
func (a *AgentActor) SetAway(ctx context.Context) error {
	a.stateMutex.Lock()
	defer a.stateMutex.Unlock()

	if a.presenceState != PresenceOnline {
		return fmt.Errorf("cannot set agent away: invalid presence state: %w", ErrInvalidAgent)
	}

	oldState := a.presenceState
	a.presenceState = PresenceAway

	// Emit presence change event
	if a.eventHandler != nil {
		a.eventHandler(map[string]interface{}{
			"type":     "presence_change",
			"agent_id": a.ID,
			"from":     oldState,
			"to":       a.presenceState,
			"event":    "away",
		})
	}

	return nil
}

// SetBack marks the agent as back, transitioning from Away to Online
func (a *AgentActor) SetBack(ctx context.Context) error {
	a.stateMutex.Lock()
	defer a.stateMutex.Unlock()

	if a.presenceState != PresenceAway {
		return fmt.Errorf("cannot set agent back: invalid presence state: %w", ErrInvalidAgent)
	}

	oldState := a.presenceState
	a.presenceState = PresenceOnline

	// Emit presence change event
	if a.eventHandler != nil {
		a.eventHandler(map[string]interface{}{
			"type":     "presence_change",
			"agent_id": a.ID,
			"from":     oldState,
			"to":       a.presenceState,
			"event":    "back",
		})
	}

	return nil
}

// CurrentPresenceState returns the current presence state
func (a *AgentActor) CurrentPresenceState() PresenceState {
	a.stateMutex.RLock()
	defer a.stateMutex.RUnlock()
	return a.presenceState
}

// HandleEvent processes an event that may trigger state transitions
// This satisfies the spec requirement that events from the PTY monitor
// can trigger state transitions via hsm.Dispatch() or equivalent
func (a *AgentActor) HandleEvent(ctx context.Context, eventType string, eventData map[string]interface{}) error {
	switch eventType {
	case "activity_detected":
		// If agent is away, bring back online
		if a.CurrentPresenceState() == PresenceAway {
			return a.SetBack(ctx)
		}
		// If agent is busy, it might become available
		if a.CurrentPresenceState() == PresenceBusy {
			// Optionally set available based on event data
			if reason, ok := eventData["reason"]; ok && reason == "prompt_detected" {
				return a.SetAvailable(ctx)
			}
		}
	case "rate_limit_detected":
		// Set agent offline if rate limited
		if a.CurrentPresenceState() == PresenceOnline || a.CurrentPresenceState() == PresenceBusy {
			return a.Disconnect(ctx)
		}
	case "stuck_detected":
		// Set agent away if stuck
		return a.SetAway(ctx)
	case "task_assigned":
		// Set agent busy when a task is assigned
		if a.CurrentPresenceState() == PresenceOnline {
			return a.SetBusy(ctx)
		}
	case "task_completed":
		// Set agent available when task is completed
		if a.CurrentPresenceState() == PresenceBusy {
			return a.SetAvailable(ctx)
		}
	}

	return nil
}

// SubscribeToEvents subscribes this agent to relevant events from the event system
func (a *AgentActor) SubscribeToEvents() error {
	// Subscribe to events that affect agent state
	err := event.SubscribeToEvent("activity_detected", func(ctx context.Context, ev event.Event) error {
		return a.HandleEvent(ctx, "activity_detected", ev.Data)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to activity_detected event: %w", err)
	}

	err = event.SubscribeToEvent("rate_limit_detected", func(ctx context.Context, ev event.Event) error {
		return a.HandleEvent(ctx, "rate_limit_detected", ev.Data)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to rate_limit_detected event: %w", err)
	}

	err = event.SubscribeToEvent("stuck_detected", func(ctx context.Context, ev event.Event) error {
		return a.HandleEvent(ctx, "stuck_detected", ev.Data)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to stuck_detected event: %w", err)
	}

	err = event.SubscribeToEvent("task_assigned", func(ctx context.Context, ev event.Event) error {
		return a.HandleEvent(ctx, "task_assigned", ev.Data)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to task_assigned event: %w", err)
	}

	err = event.SubscribeToEvent("task_completed", func(ctx context.Context, ev event.Event) error {
		return a.HandleEvent(ctx, "task_completed", ev.Data)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to task_completed event: %w", err)
	}

	return nil
}