package event

import (
	"context"
)

// Dispatcher defines the interface for event dispatching.
// Phase 0: Interface and Noop implementation.
type Dispatcher interface {
	Dispatch(ctx context.Context, event any) error
	Subscribe(pattern string) (<-chan any, func())
}

// NoopDispatcher is a dispatcher that does nothing.
type NoopDispatcher struct{}

func NewNoopDispatcher() *NoopDispatcher {
	return &NoopDispatcher{}
}

func (d *NoopDispatcher) Dispatch(ctx context.Context, event any) error {
	return nil
}

func (d *NoopDispatcher) Subscribe(pattern string) (<-chan any, func()) {
	// Return a closed channel or empty channel
	ch := make(chan any)
	close(ch)
	return ch, func() {}
}
