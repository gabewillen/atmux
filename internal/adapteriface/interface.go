// Package adapteriface implements a stable interface for adapter-provided pattern matching and actions
// that can be used by other packages during Phase 0 before the full WASM implementation is complete in Phase 8.
package adapteriface

import (
	"context"
	"sync"
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

// WASMInterface is a WASM-based implementation of the Interface
type WASMInterface struct {
	manifest   Manifest
	mutex      sync.RWMutex
}

// NewWASMInterface creates a new WASM-based adapter interface
func NewWASMInterface(manifest Manifest) *WASMInterface {
	return &WASMInterface{
		manifest: manifest,
	}
}

// MatchPatterns implements the Interface
func (wi *WASMInterface) MatchPatterns(ctx context.Context, input string) ([]Match, error) {
	wi.mutex.RLock()
	defer wi.mutex.RUnlock()

	// In a real implementation, this would call the WASM module's on_output function
	// For now, we'll simulate the behavior with basic pattern matching
	
	var matches []Match
	
	// Check each pattern in the manifest
	for _, pattern := range wi.manifest.Patterns {
		if contains(input, pattern) {
			match := Match{
				PatternID: pattern,
				Action:    "default_action", // Would come from WASM in real implementation
				Data:      make(map[string]interface{}),
				Score:     0.8, // Default confidence score
			}
			matches = append(matches, match)
		}
	}
	
	return matches, nil
}

// ExecuteAction implements the Interface
func (wi *WASMInterface) ExecuteAction(ctx context.Context, action Action) error {
	wi.mutex.RLock()
	defer wi.mutex.RUnlock()

	// In a real implementation, this would call the appropriate WASM function
	// For now, we'll just validate that the action is in the manifest
	for _, allowedAction := range wi.manifest.Actions {
		if action.Type == allowedAction {
			// Action is valid, perform it
			return nil
		}
	}
	
	return nil // For now, just return nil
}

// GetManifest implements the Interface
func (wi *WASMInterface) GetManifest() Manifest {
	wi.mutex.RLock()
	defer wi.mutex.RUnlock()
	
	return wi.manifest
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

// SetGlobalInterface sets the global adapter interface
func SetGlobalInterface(iface Interface) {
	GlobalInterface = iface
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && find(s, substr)
}

// Helper function to find a substring
func find(s, substr string) bool {
	sLen := len(s)
	substrLen := len(substr)
	
	if substrLen == 0 {
		return true
	}
	
	for i := 0; i <= sLen-substrLen; i++ {
		match := true
		for j := 0; j < substrLen; j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	
	return false
}