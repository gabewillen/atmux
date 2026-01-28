package adapter

import (
	"context"
	"testing"

	"github.com/agentflare-ai/amux/internal/event"
)

func TestNoopManager(t *testing.T) {
	manager := NewNoopManager()

	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}

	// Test initial state
	adapters := manager.GetAdapters()
	if len(adapters) != 0 {
		t.Errorf("Expected 0 adapters, got %d", len(adapters))
	}

	// Test info
	info := manager.Info()
	if info.LoadedAdapters != 0 {
		t.Errorf("Expected 0 loaded adapters, got %d", info.LoadedAdapters)
	}

	if info.ActiveMatchers != 0 {
		t.Errorf("Expected 0 active matchers, got %d", info.ActiveMatchers)
	}

	if info.ActiveActions != 0 {
		t.Errorf("Expected 0 active actions, got %d", info.ActiveActions)
	}

	ctx := context.Background()

	// Test load (no-op)
	err := manager.Load(ctx, "test-adapter", map[string]interface{}{})
	if err != nil {
		t.Errorf("Unexpected load error: %v", err)
	}

	// Test unload
	err = manager.Unload(ctx, "test-adapter")
	if err != nil {
		t.Errorf("Unexpected unload error: %v", err)
	}

	// Test match
	matches, err := manager.Match(ctx, "test content")
	if err != nil {
		t.Errorf("Unexpected match error: %v", err)
	}

	if len(matches) != 0 {
		t.Errorf("Expected 0 matches, got %d", len(matches))
	}

	// Test execute
	action := &Action{
		ID:   "test-action",
		Type: "input",
		Parameters: map[string]interface{}{
			"text": "test input",
		},
	}

	err = manager.Execute(ctx, action)
	if err != nil {
		t.Errorf("Unexpected execute error: %v", err)
	}

	// Test process event
	ev := &event.Event{
		ID:     "test-event",
		Type:   "test.type",
		Source: "test",
		Data:   map[string]interface{}{},
	}

	actions, err := manager.ProcessEvent(ctx, ev)
	if err != nil {
		t.Errorf("Unexpected process event error: %v", err)
	}

	if len(actions) != 0 {
		t.Errorf("Expected 0 actions, got %d", len(actions))
	}

	// Test shutdown
	err = manager.Shutdown(ctx)
	if err != nil {
		t.Errorf("Unexpected shutdown error: %v", err)
	}
}
