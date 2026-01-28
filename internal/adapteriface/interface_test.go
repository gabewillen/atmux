// Package adapteriface implements tests for the adapter interface
package adapteriface

import (
	"context"
	"testing"
)

// TestNoopInterface tests the no-op adapter interface implementation
func TestNoopInterface(t *testing.T) {
	manifest := Manifest{
		Name:        "test-adapter",
		Version:     "v1.0.0",
		Description: "Test adapter",
		Patterns:    []string{"pattern1", "pattern2"},
		Actions:     []string{"action1", "action2"},
	}
	
	adapter := NewNoopInterface(manifest)
	
	if adapter == nil {
		t.Fatal("Expected adapter to be created")
	}
	
	// Test GetManifest
	returnedManifest := adapter.GetManifest()
	if returnedManifest.Name != "test-adapter" {
		t.Errorf("Expected manifest name 'test-adapter', got '%s'", returnedManifest.Name)
	}
	
	if returnedManifest.Version != "v1.0.0" {
		t.Errorf("Expected manifest version 'v1.0.0', got '%s'", returnedManifest.Version)
	}
	
	if len(returnedManifest.Patterns) != 2 {
		t.Errorf("Expected 2 patterns, got %d", len(returnedManifest.Patterns))
	}
	
	if len(returnedManifest.Actions) != 2 {
		t.Errorf("Expected 2 actions, got %d", len(returnedManifest.Actions))
	}
	
	// Test MatchPatterns
	ctx := context.Background()
	matches, err := adapter.MatchPatterns(ctx, "test input")
	if err != nil {
		t.Errorf("Expected no error when matching patterns, got: %v", err)
	}
	
	if len(matches) != 0 {
		t.Errorf("Expected 0 matches from noop implementation, got %d", len(matches))
	}
	
	// Test ExecuteAction
	action := Action{
		Type: "test.action",
		Data: map[string]interface{}{"key": "value"},
	}
	
	err = adapter.ExecuteAction(ctx, action)
	if err != nil {
		t.Errorf("Expected no error when executing action, got: %v", err)
	}
}

// TestGlobalInterface tests the global adapter interface instance
func TestGlobalInterface(t *testing.T) {
	ctx := context.Background()
	
	// Test matching patterns using the global interface
	matches, err := MatchPatterns(ctx, "test input")
	if err != nil {
		t.Errorf("Expected no error when matching patterns via global interface, got: %v", err)
	}
	
	if len(matches) != 0 {
		t.Errorf("Expected 0 matches from noop implementation, got %d", len(matches))
	}
	
	// Test executing action using the global interface
	action := Action{
		Type: "test.action",
		Data: map[string]interface{}{"key": "value"},
	}
	
	err = ExecuteAction(ctx, action)
	if err != nil {
		t.Errorf("Expected no error when executing action via global interface, got: %v", err)
	}
	
	// Test getting manifest using the global interface
	manifest := GetAdapterManifest()
	if manifest.Name != "noop-adapter" {
		t.Errorf("Expected global manifest name 'noop-adapter', got '%s'", manifest.Name)
	}
}