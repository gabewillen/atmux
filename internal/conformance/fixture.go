package conformance

import "context"

// NoopFixture is a placeholder fixture.
type NoopFixture struct{}

// Start is a no-op.
func (n *NoopFixture) Start(ctx context.Context) error {
	return nil
}

// Stop is a no-op.
func (n *NoopFixture) Stop(ctx context.Context) error {
	return nil
}
