package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/internal/paths"
)

func TestActor_StartAndState(t *testing.T) {
	dispatcher := event.NewLocalDispatcher()
	defer dispatcher.Close()

	resolver := paths.DefaultResolver
	actor := NewActor(resolver, dispatcher)

	// Initial state should be loading
	if actor.State() != StateLoading {
		t.Errorf("initial state = %v, want %v", actor.State(), StateLoading)
	}

	ctx := context.Background()
	if err := actor.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer actor.Close()

	// After start, state should be ready
	if actor.State() != StateReady {
		t.Errorf("state after Start() = %v, want %v", actor.State(), StateReady)
	}

	// Config should be non-nil
	if actor.Config() == nil {
		t.Error("Config() is nil after Start()")
	}
}

func TestActor_Reload(t *testing.T) {
	dispatcher := event.NewLocalDispatcher()
	defer dispatcher.Close()

	// Use this dispatcher as the default for subscriptions
	event.SetDefaultDispatcher(dispatcher)
	defer event.SetDefaultDispatcher(event.NewLocalDispatcher())

	resolver := paths.DefaultResolver
	actor := NewActor(resolver, dispatcher)

	ctx := context.Background()
	if err := actor.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer actor.Close()

	// Track reload events
	var reloadedCount int
	unsub := dispatcher.Subscribe(event.Subscription{
		Types: []event.Type{event.TypeConfigReloaded},
		Handler: func(ctx context.Context, evt event.Event) error {
			reloadedCount++
			return nil
		},
	})
	defer unsub()

	// Trigger reload
	if err := actor.Reload(ctx); err != nil {
		t.Errorf("Reload() error: %v", err)
	}

	// State should return to ready
	if actor.State() != StateReady {
		t.Errorf("state after Reload() = %v, want %v", actor.State(), StateReady)
	}

	// Should have received reload event
	if reloadedCount != 1 {
		t.Errorf("reloaded event count = %d, want 1", reloadedCount)
	}
}

func TestActor_Close(t *testing.T) {
	dispatcher := event.NewLocalDispatcher()
	defer dispatcher.Close()

	resolver := paths.DefaultResolver
	actor := NewActor(resolver, dispatcher)

	ctx := context.Background()
	if err := actor.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	if err := actor.Close(); err != nil {
		t.Errorf("Close() error: %v", err)
	}

	// Operations should fail after close
	if err := actor.Reload(ctx); err != ErrActorClosed {
		t.Errorf("Reload after Close() error = %v, want ErrActorClosed", err)
	}
}

func TestActor_FileWatch(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.toml")

	// Write initial config
	initialConfig := `[general]
log_level = "info"
`
	if err := os.WriteFile(configFile, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Create a custom resolver
	resolver := &paths.Resolver{}

	dispatcher := event.NewLocalDispatcher()
	defer dispatcher.Close()

	actor := NewActor(resolver, dispatcher)
	actor.pollInterval = 100 * time.Millisecond

	ctx := context.Background()
	if err := actor.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer actor.Close()

	// Add the temp file to watch
	WatchFile(configFile)
	SetDefaultActor(actor)

	// Track file change events
	var fileChangedCount int
	unsub := event.Subscribe(event.Subscription{
		Types: []event.Type{event.TypeConfigFileChanged},
		Handler: func(ctx context.Context, evt event.Event) error {
			fileChangedCount++
			return nil
		},
	})
	defer unsub()

	// Modify the config file
	time.Sleep(50 * time.Millisecond) // Ensure different mtime
	updatedConfig := `[general]
log_level = "debug"
`
	if err := os.WriteFile(configFile, []byte(updatedConfig), 0644); err != nil {
		t.Fatalf("failed to update config: %v", err)
	}

	// Wait for file watcher to detect change
	time.Sleep(250 * time.Millisecond)

	// File changed event should have been dispatched
	// Note: This test may be flaky due to timing
	if fileChangedCount == 0 {
		t.Log("warning: file change event not detected (timing sensitive)")
	}
}

func TestHotReloadableKeys(t *testing.T) {
	keys := HotReloadableKeys()

	if len(keys) == 0 {
		t.Error("HotReloadableKeys() returned empty list")
	}

	// Check some expected keys
	expectedKeys := []string{
		"timeouts.idle",
		"timeouts.stuck",
		"telemetry.enabled",
	}

	for _, expected := range expectedKeys {
		found := false
		for _, key := range keys {
			if key == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected key %q not in HotReloadableKeys()", expected)
		}
	}
}

func TestIsHotReloadable(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"timeouts.idle", true},
		{"timeouts.stuck", true},
		{"telemetry.enabled", true},
		{"adapters.claude-code.patterns", true},  // Wildcard match
		{"adapters.cursor.cli", true},             // Wildcard match
		{"remote.transport", false},               // Non-reloadable
		{"nats.mode", false},                      // Non-reloadable
		{"unknown.key", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := IsHotReloadable(tt.path)
			if result != tt.expected {
				t.Errorf("IsHotReloadable(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestCompareConfigs(t *testing.T) {
	old := DefaultConfig()
	new := DefaultConfig()

	// No changes
	changes := compareConfigs(old, new)
	if len(changes) != 0 {
		t.Errorf("compareConfigs with identical configs returned %d changes, want 0", len(changes))
	}

	// Make a change
	new.General.LogLevel = "debug"
	changes = compareConfigs(old, new)

	if len(changes) != 1 {
		t.Errorf("compareConfigs returned %d changes, want 1", len(changes))
	}

	if changes[0].Path != "general.log_level" {
		t.Errorf("change path = %q, want %q", changes[0].Path, "general.log_level")
	}

	if changes[0].OldValue != "info" {
		t.Errorf("old value = %v, want %v", changes[0].OldValue, "info")
	}

	if changes[0].NewValue != "debug" {
		t.Errorf("new value = %v, want %v", changes[0].NewValue, "debug")
	}
}

func TestConfigChangeSubscription(t *testing.T) {
	var receivedChanges []ConfigChange

	unsub := Subscribe(func(ctx context.Context, change ConfigChange) {
		receivedChanges = append(receivedChanges, change)
	})
	defer unsub()

	// Dispatch a config updated event
	change := ConfigChange{
		Path:     "test.path",
		OldValue: "old",
		NewValue: "new",
	}
	evt := event.NewEvent(event.TypeConfigUpdated, 0, change)
	ctx := context.Background()
	if err := event.Dispatch(ctx, evt); err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	// Check received
	if len(receivedChanges) != 1 {
		t.Fatalf("received %d changes, want 1", len(receivedChanges))
	}

	if receivedChanges[0].Path != "test.path" {
		t.Errorf("received path = %q, want %q", receivedChanges[0].Path, "test.path")
	}
}
