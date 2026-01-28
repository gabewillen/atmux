// Package hsm provides a minimal hierarchical state machine implementation for Phase 0.
package hsm

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ID represents a state machine identifier (using muid-like interface for now).
type ID uint64

// Event represents a state machine event.
type Event struct {
	Type   string      `json:"type"`
	Source ID          `json:"source"`
	Data   interface{} `json:"data,omitempty"`
}

// State represents a state in the HSM.
type State interface {
	Name() string
	Entry(ctx context.Context, ev Event)
	Exit(ctx context.Context, ev Event)
}

// Transition represents a state transition.
type Transition struct {
	On     string                                   `json:"on"`
	Source string                                   `json:"source"`
	Target string                                   `json:"target"`
	Effect func(ctx context.Context, ev Event)      `json:"-"`
	Guard  func(ctx context.Context, ev Event) bool `json:"-"`
}

// Machine represents a hierarchical state machine.
type Machine struct {
	mu       sync.RWMutex
	id       ID
	current  State
	states   map[string]State
	transits []Transition
	handlers map[string][]func(context.Context, Event)
}

// NewMachine creates a new state machine.
func NewMachine(id ID, initialState State) *Machine {
	return &Machine{
		id:       id,
		current:  initialState,
		states:   make(map[string]State),
		transits: make([]Transition, 0),
		handlers: make(map[string][]func(context.Context, Event)),
	}
}

// AddState adds a state to the machine.
func (m *Machine) AddState(state State) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.states[state.Name()] = state
}

// AddTransition adds a transition to the machine.
func (m *Machine) AddTransition(transit Transition) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.transits = append(m.transits, transit)
}

// Subscribe subscribes to events of a specific type.
func (m *Machine) Subscribe(eventType string, handler func(context.Context, Event)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.handlers[eventType] = append(m.handlers[eventType], handler)
}

// Dispatch dispatches an event to the state machine.
func (m *Machine) Dispatch(ctx context.Context, ev Event) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Call subscribed handlers
	if handlers, exists := m.handlers[ev.Type]; exists {
		for _, handler := range handlers {
			go handler(ctx, ev)
		}
	}

	// Process transitions
	for _, transit := range m.transits {
		if transit.On == ev.Type && m.current.Name() == transit.Source {
			if transit.Guard == nil || transit.Guard(ctx, ev) {
				// Exit current state
				m.current.Exit(ctx, ev)

				// Execute effect
				if transit.Effect != nil {
					transit.Effect(ctx, ev)
				}

				// Enter new state
				if newState, exists := m.states[transit.Target]; exists {
					m.current = newState
					m.current.Entry(ctx, ev)
				}
				break
			}
		}
	}
}

// CurrentState returns the current state.
func (m *Machine) CurrentState() State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current
}

// ID returns the machine ID.
func (m *Machine) ID() ID {
	return m.id
}

// StateWrapper creates a State from a simple struct.
func StateWrapper(s interface{}) State {
	if state, ok := s.(interface {
		Name() string
		Entry(ctx context.Context, ev Event)
		Exit(ctx context.Context, ev Event)
	}); ok {
		return state
	}

	// Fallback for simple structs with name field
	return &simpleState{name: fmt.Sprintf("%v", s)}
}

// simpleState implements State interface for simple structs.
type simpleState struct {
	name string
}

func (s *simpleState) Name() string {
	return s.name
}

func (s *simpleState) Entry(ctx context.Context, ev Event) {
	fmt.Printf("State %s: Entry\n", s.name)
}

func (s *simpleState) Exit(ctx context.Context, ev Event) {
	fmt.Printf("State %s: Exit\n", s.name)
}

// GenerateID creates a new unique ID (simplified muid for Phase 0).
func GenerateID() ID {
	// This is a simplified ID generator - in full implementation this would be muid
	return ID(uint64(time.Now().UnixNano()))
}
