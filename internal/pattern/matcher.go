// Package pattern provides pattern matching and action interfaces.
// This package provides stable interfaces for pattern matching that adapters
// will implement, with noop implementations to unblock phased work.
package pattern

import (
	"errors"
	"fmt"
)

// Common sentinel errors for pattern operations.
var (
	// ErrNoMatches indicates no patterns matched the input.
	ErrNoMatches = errors.New("no pattern matches")

	// ErrMatcherNotAvailable indicates pattern matching is not available.
	ErrMatcherNotAvailable = errors.New("pattern matcher not available")

	// ErrActionFailed indicates a pattern action failed to execute.
	ErrActionFailed = errors.New("action execution failed")
)

// Match represents a pattern match result.
type Match struct {
	// Pattern is the pattern that matched.
	Pattern string

	// Confidence is the match confidence score (0.0-1.0).
	Confidence float64

	// Data contains match-specific data.
	Data map[string]interface{}
}

// Action represents an action to be taken based on a pattern match.
type Action struct {
	// Type is the action type (e.g., "respond", "notify", "execute").
	Type string

	// Payload contains action-specific data.
	Payload map[string]interface{}
}

// Matcher provides pattern matching functionality.
// Phase 0 provides a noop implementation that returns no matches.
type Matcher struct {
	available bool
}

// NewMatcher creates a new pattern matcher.
func NewMatcher() *Matcher {
	return &Matcher{
		available: false, // Will be implemented in Phase 5
	}
}

// IsAvailable returns whether pattern matching is available.
func (m *Matcher) IsAvailable() bool {
	return m.available
}

// Match attempts to match the given input against available patterns.
// Phase 0: Returns no matches to unblock later development.
func (m *Matcher) Match(input string) ([]Match, error) {
	if !m.available {
		return nil, fmt.Errorf("pattern matching not available: %w", ErrMatcherNotAvailable)
	}

	// Phase 0: Return no matches by default
	// Real pattern matching will be implemented in Phase 5
	return []Match{}, nil
}

// ExecuteAction executes the specified action.
// Phase 0: Noop implementation.
func (m *Matcher) ExecuteAction(action Action) error {
	if !m.available {
		return fmt.Errorf("pattern matcher not available: %w", ErrMatcherNotAvailable)
	}

	// Phase 0: Accept actions but don't execute them
	// Real implementation will be added in Phase 5
	return nil
}