// Package event implements tests for the event dispatcher
package event

import (
	"context"
	"testing"
)

// TestNoopDispatcher tests the no-op dispatcher implementation
func TestNoopDispatcher(t *testing.T) {
	dispatcher := NewNoopDispatcher()
	
	if dispatcher == nil {
		t.Fatal("Expected dispatcher to be created")
	}
	
	ctx := context.Background()
	
	// Test emitting an event
	event := Event{
		Type:      "test.event",
		Timestamp: 1234567890,
		Data:      map[string]interface{}{"key": "value"},
	}
	
	err := dispatcher.Emit(ctx, event)
	if err != nil {
		t.Errorf("Expected no error when emitting event, got: %v", err)
	}
	
	// Test subscribing to an event
	handler := func(ctx context.Context, event Event) error {
		return nil
	}
	
	err = dispatcher.Subscribe("test.event", handler)
	if err != nil {
		t.Errorf("Expected no error when subscribing, got: %v", err)
	}
	
	// Test unsubscribing from an event
	err = dispatcher.Unsubscribe("test.event", handler)
	if err != nil {
		t.Errorf("Expected no error when unsubscribing, got: %v", err)
	}
}

// TestGlobalDispatcher tests the global dispatcher instance
func TestGlobalDispatcher(t *testing.T) {
	ctx := context.Background()
	
	// Test emitting an event using the global dispatcher
	err := EmitEvent(ctx, "test.event", map[string]interface{}{"key": "value"})
	if err != nil {
		t.Errorf("Expected no error when emitting event via global dispatcher, got: %v", err)
	}
	
	// Test subscribing to an event using the global dispatcher
	handler := func(ctx context.Context, event Event) error {
		return nil
	}
	
	err = SubscribeToEvent("test.event", handler)
	if err != nil {
		t.Errorf("Expected no error when subscribing via global dispatcher, got: %v", err)
	}
	
	// Test unsubscribing from an event using the global dispatcher
	err = UnsubscribeFromEvent("test.event", handler)
	if err != nil {
		t.Errorf("Expected no error when unsubscribing via global dispatcher, got: %v", err)
	}
}