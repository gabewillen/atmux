// Package adapteriface implements a stable interface for adapter-provided pattern matching and actions
// that can be used by other packages during Phase 0 before the full WASM implementation is complete in Phase 8.
package adapteriface

import (
	"context"
)

// Match represents a pattern match result
type Match struct {
	PatternID string                 `json:"pattern_id"`
	Action    string                 `json:"action"`
	Data      map[string]interface{} `json:"data"`
	Score     float64                `json:"score"` // Confidence score between 0 and 1
}

// Action represents an action to be taken based on a match
type Action struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// Interface defines the interface for adapter-provided pattern matching and actions
type Interface interface {
	// MatchPatterns attempts to match patterns against the provided input
	MatchPatterns(ctx context.Context, input string) ([]Match, error)

	// ExecuteAction executes an action returned by MatchPatterns
	ExecuteAction(ctx context.Context, action Action) error

	// GetManifest returns the adapter manifest
	GetManifest() Manifest
}

// Manifest describes the adapter's capabilities
type Manifest struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Patterns    []string `json:"patterns"`
	Actions     []string `json:"actions"`
}

// NoopInterface is a no-op implementation of the Interface
type NoopInterface struct {
	manifest Manifest
}

// NewNoopInterface creates a new no-op adapter interface
func NewNoopInterface(manifest Manifest) *NoopInterface {
	return &NoopInterface{
		manifest: manifest,
	}
}

// MatchPatterns implements the Interface
func (ni *NoopInterface) MatchPatterns(ctx context.Context, input string) ([]Match, error) {
	// In the noop implementation, we return no matches
	return []Match{}, nil
}

// ExecuteAction implements the Interface
func (ni *NoopInterface) ExecuteAction(ctx context.Context, action Action) error {
	// In the noop implementation, we just return nil (success) without doing anything
	return nil
}

// GetManifest implements the Interface
func (ni *NoopInterface) GetManifest() Manifest {
	return ni.manifest
}

// GlobalInterface is a global instance of the adapter interface that can be used by other packages
var GlobalInterface Interface = NewNoopInterface(Manifest{
	Name:        "noop-adapter",
	Version:     "v0.0.0",
	Description: "No-op adapter for Phase 0",
	Patterns:    []string{},
	Actions:     []string{},
})

// MatchPatterns is a convenience function to match patterns using the global interface
func MatchPatterns(ctx context.Context, input string) ([]Match, error) {
	return GlobalInterface.MatchPatterns(ctx, input)
}

// ExecuteAction is a convenience function to execute an action using the global interface
func ExecuteAction(ctx context.Context, action Action) error {
	return GlobalInterface.ExecuteAction(ctx, action)
}

// GetAdapterManifest is a convenience function to get the adapter manifest
func GetAdapterManifest() Manifest {
	return GlobalInterface.GetManifest()
}