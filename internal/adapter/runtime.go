package adapter

import (
	"context"
)

// Runtime defines the interface for the WASM adapter runtime.
type Runtime interface {
	// Load loads an adapter from the registry.
	Load(ctx context.Context, name string) (Instance, error)
}

// Instance represents a running adapter instance.
type Instance interface {
	// ExecuteAction checks for patterns and executes actions.
	ExecuteAction(ctx context.Context, input []byte) ([]byte, error)
}

// NoopRuntime implements a no-op runtime.
type NoopRuntime struct{}

func NewNoopRuntime() *NoopRuntime {
	return &NoopRuntime{}
}

func (r *NoopRuntime) Load(ctx context.Context, name string) (Instance, error) {
	return &NoopInstance{}, nil
}

type NoopInstance struct{}

func (i *NoopInstance) ExecuteAction(ctx context.Context, input []byte) ([]byte, error) {
	return []byte{}, nil
}
