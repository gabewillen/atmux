// Package adapter provides the WASM adapter runtime interface per spec §10.
//
// Phase 0: Provides stable interfaces with noop implementations.
// Phase 8 will add full WASM runtime with wazero.
package adapter

import (
	"context"
)

// Adapter is the interface for WASM adapters.
type Adapter interface {
	// Name returns the adapter name.
	Name() string
	
	// OnOutput processes PTY output and returns pattern matches.
	OnOutput(ctx context.Context, output []byte) ([]Pattern, error)
	
	// FormatInput formats input for the agent.
	FormatInput(ctx context.Context, input string) (string, error)
	
	// OnEvent processes an event and returns actions.
	OnEvent(ctx context.Context, event any) ([]Action, error)
	
	// Close releases adapter resources.
	Close() error
}

// Pattern represents a matched pattern in PTY output.
type Pattern struct {
	Name    string // Pattern name (e.g., "prompt", "error")
	Matched string // Matched text
}

// Action represents an action to be taken by the core.
type Action struct {
	Type string // Action type (e.g., "send_input", "notify")
	Data any    // Action-specific data
}

// Runtime manages adapter loading and lifecycle.
type Runtime interface {
	// LoadAdapter loads an adapter by name.
	LoadAdapter(ctx context.Context, name string) (Adapter, error)
	
	// Close releases all adapter resources.
	Close() error
}

// NewRuntime creates a new adapter runtime.
// Phase 0: Returns a stub that will be implemented with wazero in Phase 8.
func NewRuntime() Runtime {
	return &stubRuntime{}
}

// stubRuntime is a placeholder implementation for Phase 0.
type stubRuntime struct{}

func (r *stubRuntime) LoadAdapter(ctx context.Context, name string) (Adapter, error) {
	return &stubAdapter{name: name}, nil
}

func (r *stubRuntime) Close() error {
	return nil
}

// stubAdapter is a placeholder adapter for Phase 0.
type stubAdapter struct {
	name string
}

func (a *stubAdapter) Name() string {
	return a.name
}

func (a *stubAdapter) OnOutput(ctx context.Context, output []byte) ([]Pattern, error) {
	// Phase 0: No patterns matched
	return nil, nil
}

func (a *stubAdapter) FormatInput(ctx context.Context, input string) (string, error) {
	// Phase 0: Pass through
	return input, nil
}

func (a *stubAdapter) OnEvent(ctx context.Context, event any) ([]Action, error) {
	// Phase 0: No actions
	return nil, nil
}

func (a *stubAdapter) Close() error {
	return nil
}
