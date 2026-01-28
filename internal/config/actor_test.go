// Package config implements tests for the configuration actor
package config

import (
	"testing"
	"time"
)

// TestConfigActor tests the configuration actor functionality
func TestConfigActor(t *testing.T) {
	initialConfig := &Config{
		Core: CoreConfig{
			RepoRoot: "/initial/path",
			Debug:    false,
		},
		Server: ServerConfig{
			SocketPath: "/tmp/initial.sock",
		},
	}

	actor := NewActor(initialConfig)

	// Test getting the initial config
	currentConfig := actor.Get()
	if currentConfig.Core.RepoRoot != "/initial/path" {
		t.Errorf("Expected repo root '/initial/path', got '%s'", currentConfig.Core.RepoRoot)
	}

	// Test updating the config
	newConfig := &Config{
		Core: CoreConfig{
			RepoRoot: "/new/path",
			Debug:    true,
		},
		Server: ServerConfig{
			SocketPath: "/tmp/new.sock",
		},
	}

	err := actor.Update(newConfig)
	if err != nil {
		t.Fatalf("Unexpected error updating config: %v", err)
	}

	updatedConfig := actor.Get()
	if updatedConfig.Core.RepoRoot != "/new/path" {
		t.Errorf("Expected repo root '/new/path', got '%s'", updatedConfig.Core.RepoRoot)
	}
	if !updatedConfig.Core.Debug {
		t.Error("Expected debug to be true")
	}
}

// TestConfigActorSubscription tests the subscription functionality
func TestConfigActorSubscription(t *testing.T) {
	initialConfig := &Config{
		Core: CoreConfig{
			RepoRoot: "/initial/path",
			Debug:    false,
		},
	}

	actor := NewActor(initialConfig)

	// Subscribe to config updates
	ch, unsubscribe := actor.Subscribe()
	defer unsubscribe()

	// Should receive the initial config immediately
	select {
	case config := <-ch:
		if config.Core.RepoRoot != "/initial/path" {
			t.Errorf("Expected initial config with repo root '/initial/path', got '%s'", config.Core.RepoRoot)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive initial config immediately")
	}

	// Update the config
	newConfig := &Config{
		Core: CoreConfig{
			RepoRoot: "/updated/path",
			Debug:    true,
		},
	}

	err := actor.Update(newConfig)
	if err != nil {
		t.Fatalf("Unexpected error updating config: %v", err)
	}

	// Should receive the updated config
	select {
	case config := <-ch:
		if config.Core.RepoRoot != "/updated/path" {
			t.Errorf("Expected updated config with repo root '/updated/path', got '%s'", config.Core.RepoRoot)
		}
		if !config.Core.Debug {
			t.Error("Expected debug to be true in updated config")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive updated config")
	}
}

// TestConfigActorMultipleSubscribers tests multiple subscribers
func TestConfigActorMultipleSubscribers(t *testing.T) {
	initialConfig := &Config{
		Core: CoreConfig{
			RepoRoot: "/initial/path",
		},
	}

	actor := NewActor(initialConfig)

	// Subscribe multiple times
	ch1, unsub1 := actor.Subscribe()
	ch2, unsub2 := actor.Subscribe()
	defer unsub1()
	defer unsub2()

	// Update the config
	newConfig := &Config{
		Core: CoreConfig{
			RepoRoot: "/new/path",
		},
	}

	err := actor.Update(newConfig)
	if err != nil {
		t.Fatalf("Unexpected error updating config: %v", err)
	}

	// Both subscribers should receive the update
	select {
	case config1 := <-ch1:
		if config1.Core.RepoRoot != "/new/path" {
			t.Errorf("Subscriber 1 expected repo root '/new/path', got '%s'", config1.Core.RepoRoot)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Subscriber 1 expected to receive updated config")
	}

	select {
	case config2 := <-ch2:
		if config2.Core.RepoRoot != "/new/path" {
			t.Errorf("Subscriber 2 expected repo root '/new/path', got '%s'", config2.Core.RepoRoot)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Subscriber 2 expected to receive updated config")
	}
}

// TestConfigActorUnsubscribe tests unsubscribing
func TestConfigActorUnsubscribe(t *testing.T) {
	initialConfig := &Config{
		Core: CoreConfig{
			RepoRoot: "/initial/path",
		},
	}

	actor := NewActor(initialConfig)

	// Subscribe and then unsubscribe
	ch, unsub := actor.Subscribe()
	unsub()

	// Update the config - the unsubscribed channel should not receive it
	newConfig := &Config{
		Core: CoreConfig{
			RepoRoot: "/new/path",
		},
	}

	err := actor.Update(newConfig)
	if err != nil {
		t.Fatalf("Unexpected error updating config: %v", err)
	}

	// Try to read from the unsubscribed channel - should not receive anything
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("Expected channel to be closed after unsubscribe")
		}
	case <-time.After(100 * time.Millisecond):
		// Channel should be closed, so this timeout is expected
	}
}

// TestConfigActorValidation tests config validation during updates
func TestConfigActorValidation(t *testing.T) {
	initialConfig := &Config{
		Core: CoreConfig{
			RepoRoot: "/initial/path",
		},
	}

	actor := NewActor(initialConfig)

	// Try to update with invalid config (empty repo root)
	invalidConfig := &Config{
		Core: CoreConfig{
			RepoRoot: "", // Invalid
		},
	}

	err := actor.Update(invalidConfig)
	if err == nil {
		t.Error("Expected validation error for invalid config")
	}

	// Config should remain unchanged
	currentConfig := actor.Get()
	if currentConfig.Core.RepoRoot != "/initial/path" {
		t.Errorf("Expected config to remain unchanged, got repo root '%s'", currentConfig.Core.RepoRoot)
	}
}

// TestConfigActorClose tests closing the actor
func TestConfigActorClose(t *testing.T) {
	initialConfig := &Config{
		Core: CoreConfig{
			RepoRoot: "/initial/path",
		},
	}

	actor := NewActor(initialConfig)

	// Subscribe
	ch, _ := actor.Subscribe()

	// Close the actor
	actor.Close()

	// Channel should be closed
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("Expected channel to be closed after actor close")
		}
	case <-time.After(100 * time.Millisecond):
		// Channel should be closed, so this timeout is expected
	}
}