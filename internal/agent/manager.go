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
)

// Common sentinel errors for agent operations.
var (
	// ErrAgentNotFound indicates an agent with the given ID was not found.
	ErrAgentNotFound = errors.New("agent not found")

	// ErrInvalidState indicates an operation cannot be performed in the current agent state.
	ErrInvalidState = errors.New("invalid agent state")

	// ErrAdapterLoadFailed indicates the agent's WASM adapter failed to load.
	ErrAdapterLoadFailed = errors.New("adapter load failed")
)

// Manager orchestrates multiple agents in an agent-agnostic manner.
// It treats all agents uniformly through the adapter interface.
type Manager struct {
	// Implementation pending for Phase 1-3
}

// NewManager creates a new agent manager instance.
func NewManager() (*Manager, error) {
	return &Manager{}, nil
}

// Lifecycle methods will be implemented in subsequent phases
// following the HSM pattern: Pending → Starting → Running → Terminated/Errored

// Start initiates an agent by its adapter name and configuration.
func (m *Manager) Start(adapterName string, config map[string]interface{}) error {
	if adapterName == "" {
		return fmt.Errorf("adapter name required: %w", ErrAdapterLoadFailed)
	}
	// Implementation deferred to Phase 1
	return fmt.Errorf("agent start not implemented: %w", ErrInvalidState)
}