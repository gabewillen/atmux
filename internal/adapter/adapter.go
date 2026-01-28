// Package adapter provides the WASM adapter runtime for amux.
//
// Adapters are WebAssembly modules that implement agent-specific behavior.
// The core amux system loads adapters by name and interacts with them
// through a standardized WASM interface.
//
// This package is agent-agnostic; all agent-specific code resides in the
// adapter WASM modules themselves.
//
// See spec §10 for the full adapter interface specification.
package adapter

import (
	"context"
	"fmt"
	"sync"

	"github.com/tetratelabs/wazero"
)

// Adapter is the interface for WASM adapters.
type Adapter interface {
	// Name returns the adapter name.
	Name() string

	// Manifest returns the adapter manifest.
	Manifest() (*Manifest, error)

	// OnOutput processes PTY output and returns detected events.
	OnOutput(ctx context.Context, output []byte) ([]OutputEvent, error)

	// FormatInput formats input for the agent.
	FormatInput(ctx context.Context, input string) (string, error)

	// OnEvent handles an incoming event.
	OnEvent(ctx context.Context, event []byte) error

	// Close releases adapter resources.
	Close() error
}

// Manifest represents an adapter manifest.
type Manifest struct {
	// Name is the adapter name.
	Name string `json:"name"`

	// Version is the adapter version (semver).
	Version string `json:"version"`

	// CLI contains CLI version constraints.
	CLI CLIConstraint `json:"cli"`

	// Patterns contains pattern definitions.
	Patterns map[string]string `json:"patterns,omitempty"`
}

// CLIConstraint defines version constraints for the agent CLI.
type CLIConstraint struct {
	// Constraint is a semver constraint string.
	Constraint string `json:"constraint"`
}

// OutputEvent represents an event detected from PTY output.
type OutputEvent struct {
	// Type is the event type.
	Type string `json:"type"`

	// Data is the event-specific data.
	Data any `json:"data,omitempty"`
}

// Registry manages adapter discovery and loading.
type Registry struct {
	mu       sync.RWMutex
	adapters map[string]Adapter
	runtime  wazero.Runtime
}

// NewRegistry creates a new adapter registry.
func NewRegistry(ctx context.Context) (*Registry, error) {
	rt := wazero.NewRuntime(ctx)

	return &Registry{
		adapters: make(map[string]Adapter),
		runtime:  rt,
	}, nil
}

// Load loads an adapter by name.
// Returns ErrAdapterNotFound if the adapter is not registered.
func (r *Registry) Load(name string) (Adapter, error) {
	r.mu.RLock()
	adapter, ok := r.adapters[name]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("adapter %q: %w", name, ErrAdapterNotFound)
	}

	return adapter, nil
}

// Register registers an adapter.
func (r *Registry) Register(adapter Adapter) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := adapter.Name()
	if _, exists := r.adapters[name]; exists {
		return fmt.Errorf("adapter %q: %w", name, ErrAdapterAlreadyExists)
	}

	r.adapters[name] = adapter
	return nil
}

// Unregister removes an adapter from the registry.
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	adapter, ok := r.adapters[name]
	if !ok {
		return fmt.Errorf("adapter %q: %w", name, ErrAdapterNotFound)
	}

	if err := adapter.Close(); err != nil {
		return fmt.Errorf("close adapter %q: %w", name, err)
	}

	delete(r.adapters, name)
	return nil
}

// List returns the names of all registered adapters.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.adapters))
	for name := range r.adapters {
		names = append(names, name)
	}
	return names
}

// Close closes the registry and all adapters.
func (r *Registry) Close(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var lastErr error
	for name, adapter := range r.adapters {
		if err := adapter.Close(); err != nil {
			lastErr = fmt.Errorf("close adapter %q: %w", name, err)
		}
	}

	r.adapters = nil

	if err := r.runtime.Close(ctx); err != nil {
		return fmt.Errorf("close wazero runtime: %w", err)
	}

	return lastErr
}

// Adapter errors
var (
	// ErrAdapterNotFound indicates the adapter was not found.
	ErrAdapterNotFound = fmt.Errorf("not found")

	// ErrAdapterAlreadyExists indicates the adapter already exists.
	ErrAdapterAlreadyExists = fmt.Errorf("already exists")

	// ErrAdapterLoadFailed indicates the adapter failed to load.
	ErrAdapterLoadFailed = fmt.Errorf("load failed")

	// ErrAdapterCallFailed indicates an adapter call failed.
	ErrAdapterCallFailed = fmt.Errorf("call failed")
)

// NoopAdapter is a no-op adapter for testing.
type NoopAdapter struct {
	name string
}

// NewNoopAdapter creates a new no-op adapter.
func NewNoopAdapter(name string) *NoopAdapter {
	return &NoopAdapter{name: name}
}

// Name returns the adapter name.
func (a *NoopAdapter) Name() string {
	return a.name
}

// Manifest returns a minimal manifest.
func (a *NoopAdapter) Manifest() (*Manifest, error) {
	return &Manifest{
		Name:    a.name,
		Version: "0.0.0",
	}, nil
}

// OnOutput returns no events.
func (a *NoopAdapter) OnOutput(ctx context.Context, output []byte) ([]OutputEvent, error) {
	return nil, nil
}

// FormatInput returns the input unchanged.
func (a *NoopAdapter) FormatInput(ctx context.Context, input string) (string, error) {
	return input, nil
}

// OnEvent is a no-op.
func (a *NoopAdapter) OnEvent(ctx context.Context, event []byte) error {
	return nil
}

// Close is a no-op.
func (a *NoopAdapter) Close() error {
	return nil
}

// PatternMatcher is the interface for pattern matching.
// During Phase 0, this uses a noop implementation.
// Phase 8 will provide the full WASM-backed implementation.
type PatternMatcher interface {
	// Match checks if output matches any configured patterns.
	Match(output []byte) []PatternMatch
}

// PatternMatch represents a pattern match result.
type PatternMatch struct {
	// Pattern is the matched pattern name.
	Pattern string

	// Match is the matched text.
	Match string

	// Index is the byte offset of the match.
	Index int
}

// NoopPatternMatcher is a no-op pattern matcher.
type NoopPatternMatcher struct{}

// NewNoopPatternMatcher creates a new no-op pattern matcher.
func NewNoopPatternMatcher() *NoopPatternMatcher {
	return &NoopPatternMatcher{}
}

// Match returns no matches.
func (m *NoopPatternMatcher) Match(output []byte) []PatternMatch {
	return nil
}
