// Package adapter provides WASM adapter runtime and loading functionality.
// This package loads conforming WASM adapters without any knowledge of
// specific agent implementations.
//
// The adapter system provides a pluggable interface enabling any coding agent
// to be integrated through a WASM adapter that implements the required ABI.
package adapter

import (
	"context"
	"errors"
	"fmt"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// Common sentinel errors for adapter operations.
var (
	// ErrAdapterNotFound indicates the requested adapter was not found.
	ErrAdapterNotFound = errors.New("adapter not found")

	// ErrInvalidABI indicates the adapter does not implement the required ABI.
	ErrInvalidABI = errors.New("invalid adapter ABI")

	// ErrRuntimeFailed indicates a WASM runtime failure.
	ErrRuntimeFailed = errors.New("WASM runtime failure")
)

// Runtime manages WASM adapter instances using wazero.
// One WASM instance per agent with 256MB memory cap.
type Runtime struct {
	ctx    context.Context
	engine wazero.Runtime
}

// NewRuntime creates a new WASM adapter runtime.
func NewRuntime(ctx context.Context) (*Runtime, error) {
	engine := wazero.NewRuntime(ctx)
	return &Runtime{
		ctx:    ctx,
		engine: engine,
	}, nil
}

// LoadAdapter loads a WASM adapter from the given path.
// Returns an error if the adapter doesn't implement required exports:
// amux_alloc, amux_free, manifest, on_output, format_input, on_event
func (r *Runtime) LoadAdapter(path string) (api.Module, error) {
	if path == "" {
		return nil, fmt.Errorf("adapter path required: %w", ErrAdapterNotFound)
	}

	// Implementation deferred to Phase 8
	return nil, fmt.Errorf("adapter loading not implemented: %w", ErrRuntimeFailed)
}

// Close releases runtime resources.
func (r *Runtime) Close() error {
	return r.engine.Close(r.ctx)
}