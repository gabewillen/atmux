package event

import (
	"context"
	"testing"
	"time"
)

func TestNewHandlerFunc(t *testing.T) {
	calls := 0

	handler := NewHandlerFunc("test-handler", func(ctx context.Context, event *Event) error {
		calls++
		return nil
	})

	if handler.HandlerID() != "test-handler" {
		t.Errorf("Expected handler ID 'test-handler', got %q", handler.HandlerID())
	}

	// Test handling
	err := handler.Handle(context.Background(), &Event{
		ID:        "test-event",
		Type:      EventAgentSpawned,
		Source:    "test",
		Timestamp: time.Now().Unix(),
		Data:      map[string]interface{}{"test": true},
	})

	if err != nil {
		t.Errorf("Unexpected handling error: %v", err)
	}

	if calls != 1 {
		t.Errorf("Expected 1 call, got %d", calls)
	}
}

func TestNoopDispatcher(t *testing.T) {
	dispatcher := NewNoopDispatcher()

	if dispatcher == nil {
		t.Fatal("Expected non-nil dispatcher")
	}

	ctx := context.Background()

	// Test emit
	event := &Event{
		ID:        "test-event",
		Type:      EventAgentSpawned,
		Source:    "test",
		Timestamp: time.Now().Unix(),
		Data:      map[string]interface{}{},
	}

	err := dispatcher.Emit(ctx, event)
	if err != nil {
		t.Errorf("Unexpected emit error: %v", err)
	}

	// Test subscribe
	handler := NewHandlerFunc("test-handler", func(ctx context.Context, event *Event) error {
		return nil
	})

	err = dispatcher.Subscribe(ctx, []EventType{EventAgentSpawned}, handler)
	if err != nil {
		t.Errorf("Unexpected subscribe error: %v", err)
	}

	// Test unsubscribe
	err = dispatcher.Unsubscribe(handler)
	if err != nil {
		t.Errorf("Unexpected unsubscribe error: %v", err)
	}

	// Test shutdown
	err = dispatcher.Shutdown(ctx)
	if err != nil {
		t.Errorf("Unexpected shutdown error: %v", err)
	}
}
