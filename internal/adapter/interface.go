// Package adapter provides the adapter interface for pattern matching and actions.
// Phase 0 introduces stable interfaces with noop implementations.
// Phase 8 will provide the full WASM-backed runtime.
package adapter

import (
	"context"
)

// Adapter provides pattern matching and action capabilities.
// Phase 0: Noop implementation that returns no matches
// Phase 8: WASM-backed implementation
type Adapter interface {
	// MatchPatterns checks PTY output for matching patterns.
	MatchPatterns(ctx context.Context, output []byte) ([]Match, error)

	// FormatInput formats input for the agent CLI.
	FormatInput(ctx context.Context, input string) ([]byte, error)

	// OnEvent handles events from the system.
	OnEvent(ctx context.Context, event interface{}) error
}

// Match represents a pattern match result.
type Match struct {
	Pattern string
	Data    interface{}
}

// Registry manages adapter instances.
type Registry interface {
	// Load loads an adapter by name.
	Load(name string) (Adapter, error)
}

// NewRegistry creates a new adapter registry.
// Phase 0: Returns a noop registry
func NewRegistry() Registry {
	return &noopRegistry{}
}

// noopRegistry is a Phase 0 noop adapter registry.
type noopRegistry struct{}

func (r *noopRegistry) Load(name string) (Adapter, error) {
	return &noopAdapter{}, nil
}

// noopAdapter is a Phase 0 noop adapter implementation.
type noopAdapter struct{}

func (a *noopAdapter) MatchPatterns(ctx context.Context, output []byte) ([]Match, error) {
	return nil, nil
}

func (a *noopAdapter) FormatInput(ctx context.Context, input string) ([]byte, error) {
	return []byte(input), nil
}

func (a *noopAdapter) OnEvent(ctx context.Context, event interface{}) error {
	return nil
}
