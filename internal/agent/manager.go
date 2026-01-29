// Package agent provides agent-agnostic orchestration functionality.
// This package manages agent lifecycle, presence, and messaging without
// any knowledge of specific agent implementations.
//
// All agent-specific behavior is delegated to WASM adapters loaded
// via the adapter package.
package agent

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/stateforward/hsm-go/muid"
	
	"github.com/copilot-claude-sonnet-4/amux/pkg/api"
	"github.com/copilot-claude-sonnet-4/amux/internal/ids"
)

// Common sentinel errors for agent operations.
var (
	// ErrAgentNotFound indicates an agent with the given ID was not found.
	ErrAgentNotFound = errors.New("agent not found")

	// ErrInvalidState indicates an operation cannot be performed in the current agent state.
	ErrInvalidState = errors.New("invalid agent state")

	// ErrAdapterLoadFailed indicates the agent's WASM adapter failed to load.
	ErrAdapterLoadFailed = errors.New("adapter load failed")

	// ErrInvalidTransition indicates an invalid state transition was attempted.
	ErrInvalidTransition = errors.New("invalid state transition")
)

// LifecycleEvent represents events that can trigger agent lifecycle transitions.
type LifecycleEvent string

const (
	// EventStart triggers transition from Pending to Starting.
	EventStart LifecycleEvent = "start"

	// EventStartupComplete triggers transition from Starting to Running.
	EventStartupComplete LifecycleEvent = "startup_complete"

	// EventTerminate triggers transition to Terminated.
	EventTerminate LifecycleEvent = "terminate"

	// EventError triggers transition to Errored.
	EventError LifecycleEvent = "error"

	// EventRestart triggers transition from Terminated/Errored back to Pending.
	EventRestart LifecycleEvent = "restart"
)

// PresenceEvent represents events that can trigger presence transitions.
type PresenceEvent string

const (
	// EventGoOnline triggers transition to Online.
	EventGoOnline PresenceEvent = "go_online"

	// EventGoBusy triggers transition to Busy.
	EventGoBusy PresenceEvent = "go_busy"

	// EventGoOffline triggers transition to Offline.
	EventGoOffline PresenceEvent = "go_offline"

	// EventGoAway triggers transition to Away.
	EventGoAway PresenceEvent = "go_away"

	// EventActivity triggers transition from Away back to Online.
	EventActivity PresenceEvent = "activity"
)

// AgentActor wraps an Agent with simple state machines for lifecycle and presence.
// This is a simplified implementation until full HSM integration is complete.
type AgentActor struct {
	// Agent is the underlying agent data.
	Agent *api.Agent

	// mu protects concurrent access to the actor.
	mu sync.RWMutex

	// eventHandlers contains callbacks for state transitions.
	eventHandlers map[string][]func(*api.Agent, interface{})
}

// NewAgentActor creates a new agent actor with initialized state machines.
func NewAgentActor(name, adapter, repoRoot string, config map[string]interface{}) (*AgentActor, error) {
	// Validate inputs
	if !ids.IsValidIdentifierName(name) {
		return nil, fmt.Errorf("invalid agent name: %q", name)
	}

	agentID := ids.NewAgentID()
	slug := ids.AgentSlugFromName(name)
	
	if err := ids.ValidateAgentSlug(slug); err != nil {
		return nil, fmt.Errorf("generated invalid slug %q from name %q: %w", slug, name, err)
	}

	canonicalRoot, err := ids.CanonicalizeRepoRoot(repoRoot, false)
	if err != nil {
		return nil, fmt.Errorf("invalid repo root: %w", err)
	}

	now := time.Now().UTC()
	agent := &api.Agent{
		ID:        agentID,
		Slug:      slug,
		Adapter:   adapter,
		Name:      name,
		State:     api.AgentStatePending,
		Presence:  api.PresenceOffline,
		RepoRoot:  canonicalRoot,
		CreatedAt: now,
		UpdatedAt: now,
		Config:    config,
	}

	actor := &AgentActor{
		Agent:         agent,
		eventHandlers: make(map[string][]func(*api.Agent, interface{})),
	}

	return actor, nil
}

// SendLifecycleEvent dispatches a lifecycle event to trigger state transitions.
func (a *AgentActor) SendLifecycleEvent(event LifecycleEvent) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	oldState := a.Agent.State
	
	// Simple state machine logic - Per spec: Pending → Starting → Running → Terminated/Errored
	newState := oldState
	
	switch oldState {
	case api.AgentStatePending:
		if event == EventStart {
			newState = api.AgentStateStarting
		} else {
			return fmt.Errorf("invalid transition from %s with event %s: %w", oldState, event, ErrInvalidTransition)
		}
	case api.AgentStateStarting:
		switch event {
		case EventStartupComplete:
			newState = api.AgentStateRunning
		case EventError:
			newState = api.AgentStateErrored
		default:
			return fmt.Errorf("invalid transition from %s with event %s: %w", oldState, event, ErrInvalidTransition)
		}
	case api.AgentStateRunning:
		switch event {
		case EventTerminate:
			newState = api.AgentStateTerminated
		case EventError:
			newState = api.AgentStateErrored
		default:
			return fmt.Errorf("invalid transition from %s with event %s: %w", oldState, event, ErrInvalidTransition)
		}
	case api.AgentStateTerminated, api.AgentStateErrored:
		if event == EventRestart {
			newState = api.AgentStatePending
		} else {
			return fmt.Errorf("invalid transition from %s with event %s: %w", oldState, event, ErrInvalidTransition)
		}
	}

	// Update agent state
	a.Agent.State = newState
	a.Agent.UpdatedAt = time.Now().UTC()

	// Trigger event handlers
	a.triggerEventHandlers("lifecycle", map[string]interface{}{
		"old_state": oldState,
		"new_state": newState,
		"event":     event,
	})

	return nil
}

// SendPresenceEvent dispatches a presence event to trigger state transitions.
func (a *AgentActor) SendPresenceEvent(event PresenceEvent) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	oldPresence := a.Agent.Presence
	
	// Simple state machine logic - Per spec: Online ↔ Busy ↔ Offline ↔ Away
	newPresence := oldPresence
	
	switch oldPresence {
	case api.PresenceOffline:
		if event == EventGoOnline {
			newPresence = api.PresenceOnline
		} else {
			return fmt.Errorf("invalid transition from %s with event %s: %w", oldPresence, event, ErrInvalidTransition)
		}
	case api.PresenceOnline:
		switch event {
		case EventGoBusy:
			newPresence = api.PresenceBusy
		case EventGoAway:
			newPresence = api.PresenceAway
		case EventGoOffline:
			newPresence = api.PresenceOffline
		default:
			return fmt.Errorf("invalid transition from %s with event %s: %w", oldPresence, event, ErrInvalidTransition)
		}
	case api.PresenceBusy:
		switch event {
		case EventGoOnline:
			newPresence = api.PresenceOnline
		case EventGoOffline:
			newPresence = api.PresenceOffline
		default:
			return fmt.Errorf("invalid transition from %s with event %s: %w", oldPresence, event, ErrInvalidTransition)
		}
	case api.PresenceAway:
		switch event {
		case EventActivity:
			newPresence = api.PresenceOnline
		case EventGoOffline:
			newPresence = api.PresenceOffline
		default:
			return fmt.Errorf("invalid transition from %s with event %s: %w", oldPresence, event, ErrInvalidTransition)
		}
	}

	// Update agent presence
	a.Agent.Presence = newPresence
	a.Agent.UpdatedAt = time.Now().UTC()

	// Trigger event handlers
	a.triggerEventHandlers("presence", map[string]interface{}{
		"old_presence": oldPresence,
		"new_presence": newPresence,
		"event":        event,
	})

	return nil
}

// GetState returns the current agent state safely.
func (a *AgentActor) GetState() api.AgentState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Agent.State
}

// GetPresence returns the current agent presence safely.
func (a *AgentActor) GetPresence() api.PresenceState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Agent.Presence
}

// GetAgent returns a copy of the agent data safely.
func (a *AgentActor) GetAgent() api.Agent {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return *a.Agent
}

// OnEvent registers an event handler for state changes.
func (a *AgentActor) OnEvent(eventType string, handler func(*api.Agent, interface{})) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.eventHandlers[eventType] = append(a.eventHandlers[eventType], handler)
}

// triggerEventHandlers invokes registered event handlers.
func (a *AgentActor) triggerEventHandlers(eventType string, data interface{}) {
	handlers, exists := a.eventHandlers[eventType]
	if !exists {
		return
	}

	for _, handler := range handlers {
		handler(a.Agent, data)
	}
}

// Manager orchestrates multiple agents in an agent-agnostic manner.
// It treats all agents uniformly through the adapter interface.
type Manager struct {
	// agents maps agent IDs to their actors.
	agents map[muid.MUID]*AgentActor

	// mu protects concurrent access to the manager.
	mu sync.RWMutex
}

// NewManager creates a new agent manager instance.
func NewManager() (*Manager, error) {
	return &Manager{
		agents: make(map[muid.MUID]*AgentActor),
	}, nil
}

// AddAgent creates and adds a new agent to the manager.
func (m *Manager) AddAgent(name, adapter, repoRoot string, config map[string]interface{}) (*api.Agent, error) {
	actor, err := NewAgentActor(name, adapter, repoRoot, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent actor: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.agents[actor.Agent.ID] = actor

	agent := actor.GetAgent()
	return &agent, nil
}

// GetAgent returns an agent by ID.
func (m *Manager) GetAgent(id muid.MUID) (*api.Agent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	actor, exists := m.agents[id]
	if !exists {
		return nil, ErrAgentNotFound
	}

	agent := actor.GetAgent()
	return &agent, nil
}

// ListAgents returns all agents managed by this manager.
func (m *Manager) ListAgents() []api.Agent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agents := make([]api.Agent, 0, len(m.agents))
	for _, actor := range m.agents {
		agents = append(agents, actor.GetAgent())
	}

	return agents
}

// StartAgent starts an agent by sending a start lifecycle event.
func (m *Manager) StartAgent(id muid.MUID) error {
	m.mu.RLock()
	actor, exists := m.agents[id]
	m.mu.RUnlock()

	if !exists {
		return ErrAgentNotFound
	}

	return actor.SendLifecycleEvent(EventStart)
}

// TerminateAgent terminates an agent by sending a terminate lifecycle event.
func (m *Manager) TerminateAgent(id muid.MUID) error {
	m.mu.RLock()
	actor, exists := m.agents[id]
	m.mu.RUnlock()

	if !exists {
		return ErrAgentNotFound
	}

	return actor.SendLifecycleEvent(EventTerminate)
}

// UpdatePresence updates an agent's presence state.
func (m *Manager) UpdatePresence(id muid.MUID, event PresenceEvent) error {
	m.mu.RLock()
	actor, exists := m.agents[id]
	m.mu.RUnlock()

	if !exists {
		return ErrAgentNotFound
	}

	return actor.SendPresenceEvent(event)
}