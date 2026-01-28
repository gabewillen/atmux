package adapter

import (
	"context"
	"errors"
)

// ErrAdapterNotFound is returned when a named adapter cannot be loaded.
var ErrAdapterNotFound = errors.New("adapter not found")

// PatternMatch describes a detected pattern match.
type PatternMatch struct {
	Pattern string
	Text    string
}

// PatternMatcher scans output and returns matches.
type PatternMatcher interface {
	Match(ctx context.Context, output []byte) ([]PatternMatch, error)
}

// ActionFormatter converts a high-level action into agent input.
type ActionFormatter interface {
	Format(ctx context.Context, input string) (string, error)
}

// Adapter is the runtime-facing interface to a loaded adapter.
type Adapter interface {
	Name() string
	Matcher() PatternMatcher
	Formatter() ActionFormatter
}

// Registry loads adapters by name.
type Registry interface {
	Load(ctx context.Context, name string) (Adapter, error)
}

// NoopAdapter returns no matches and echoes input.
type NoopAdapter struct {
	name string
}

// NewNoopAdapter constructs a noop adapter.
func NewNoopAdapter(name string) *NoopAdapter {
	return &NoopAdapter{name: name}
}

// Name returns the adapter name.
func (n *NoopAdapter) Name() string {
	return n.name
}

// Matcher returns a noop matcher.
func (n *NoopAdapter) Matcher() PatternMatcher {
	return &NoopMatcher{}
}

// Formatter returns a noop formatter.
func (n *NoopAdapter) Formatter() ActionFormatter {
	return &NoopFormatter{}
}

// NoopMatcher returns no matches.
type NoopMatcher struct{}

// Match returns no matches.
func (m *NoopMatcher) Match(ctx context.Context, output []byte) ([]PatternMatch, error) {
	return nil, nil
}

// NoopFormatter returns the input unchanged.
type NoopFormatter struct{}

// Format returns the input unchanged.
func (f *NoopFormatter) Format(ctx context.Context, input string) (string, error) {
	return input, nil
}
