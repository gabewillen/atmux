package event

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stateforward/hsm-go/muid"
)

func TestNewEvent_Fields(t *testing.T) {
	source := muid.Make()
	data := map[string]string{"key": "value"}

	before := time.Now().UTC()
	evt := NewEvent(TypeAgentStarted, source, data)
	after := time.Now().UTC()

	if evt.ID == 0 {
		t.Error("NewEvent ID should be non-zero")
	}

	if evt.Type != TypeAgentStarted {
		t.Errorf("NewEvent Type = %q, want %q", evt.Type, TypeAgentStarted)
	}

	if evt.Source != source {
		t.Errorf("NewEvent Source = %v, want %v", evt.Source, source)
	}

	if evt.Target != 0 {
		t.Errorf("NewEvent Target = %v, want 0 (broadcast)", evt.Target)
	}

	if evt.Timestamp.Before(before) || evt.Timestamp.After(after) {
		t.Errorf("NewEvent Timestamp %v not between %v and %v", evt.Timestamp, before, after)
	}

	if evt.Data == nil {
		t.Error("NewEvent Data should not be nil")
	}

	if evt.TraceID != "" {
		t.Errorf("NewEvent TraceID = %q, want empty", evt.TraceID)
	}
}

func TestNewEvent_UniqueIDs(t *testing.T) {
	source := muid.Make()
	evt1 := NewEvent(TypeAgentStarted, source, nil)
	evt2 := NewEvent(TypeAgentStarted, source, nil)

	if evt1.ID == evt2.ID {
		t.Error("two NewEvent calls should produce different IDs")
	}
}

func TestEvent_WithTarget(t *testing.T) {
	source := muid.Make()
	target := muid.Make()
	evt := NewEvent(TypeAgentStarted, source, nil)

	targeted := evt.WithTarget(target)

	// Original should be unchanged
	if evt.Target != 0 {
		t.Error("WithTarget should not modify the original event")
	}

	if targeted.Target != target {
		t.Errorf("WithTarget Target = %v, want %v", targeted.Target, target)
	}

	// Other fields should be preserved
	if targeted.ID != evt.ID {
		t.Error("WithTarget should preserve ID")
	}
	if targeted.Type != evt.Type {
		t.Error("WithTarget should preserve Type")
	}
	if targeted.Source != evt.Source {
		t.Error("WithTarget should preserve Source")
	}
}

func TestEvent_WithTraceID(t *testing.T) {
	source := muid.Make()
	evt := NewEvent(TypeAgentStarted, source, nil)
	traceID := "trace-abc-123"

	traced := evt.WithTraceID(traceID)

	if evt.TraceID != "" {
		t.Error("WithTraceID should not modify the original event")
	}

	if traced.TraceID != traceID {
		t.Errorf("WithTraceID TraceID = %q, want %q", traced.TraceID, traceID)
	}

	if traced.ID != evt.ID {
		t.Error("WithTraceID should preserve ID")
	}
}

func TestEvent_MarshalJSON(t *testing.T) {
	source := muid.Make()
	id := muid.Make()
	ts := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	evt := Event{
		ID:        id,
		Type:      TypeAgentStarted,
		Source:    source,
		Timestamp: ts,
		Data:      map[string]string{"status": "ok"},
	}

	data, err := evt.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Check ID is encoded as a string (matching MUID.String() output)
	idStr, ok := result["id"].(string)
	if !ok {
		t.Fatalf("id is not a string, got %T", result["id"])
	}
	wantID := strconv.FormatUint(uint64(id), 10)
	if idStr != wantID {
		t.Errorf("id = %q, want %q", idStr, wantID)
	}

	// Check type
	typeStr, ok := result["type"].(string)
	if !ok {
		t.Fatalf("type is not a string, got %T", result["type"])
	}
	if typeStr != string(TypeAgentStarted) {
		t.Errorf("type = %q, want %q", typeStr, TypeAgentStarted)
	}

	// Check source is a string
	sourceStr, ok := result["source"].(string)
	if !ok {
		t.Fatalf("source is not a string, got %T", result["source"])
	}
	wantSource := strconv.FormatUint(uint64(source), 10)
	if sourceStr != wantSource {
		t.Errorf("source = %q, want %q", sourceStr, wantSource)
	}

	// Check timestamp is RFC3339
	tsStr, ok := result["timestamp"].(string)
	if !ok {
		t.Fatalf("timestamp is not a string, got %T", result["timestamp"])
	}
	parsedTS, err := time.Parse(time.RFC3339Nano, tsStr)
	if err != nil {
		t.Fatalf("timestamp is not RFC3339: %v", err)
	}
	if !parsedTS.Equal(ts) {
		t.Errorf("timestamp = %v, want %v", parsedTS, ts)
	}

	// Check data is present
	if result["data"] == nil {
		t.Error("data should be present")
	}
}

func TestEvent_MarshalJSON_WithTarget(t *testing.T) {
	target := muid.Make()
	evt := Event{
		ID:        muid.Make(),
		Type:      TypePTYOutput,
		Source:    muid.Make(),
		Target:    target,
		Timestamp: time.Now().UTC(),
	}

	data, err := evt.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	targetStr, ok := result["target"].(string)
	if !ok {
		t.Fatalf("target is not a string, got %T", result["target"])
	}
	wantTarget := strconv.FormatUint(uint64(target), 10)
	if targetStr != wantTarget {
		t.Errorf("target = %q, want %q", targetStr, wantTarget)
	}
}

func TestEvent_MarshalJSON_ZeroTarget_Omitted(t *testing.T) {
	evt := Event{
		ID:        muid.MUID(1),
		Type:      TypeAgentStopped,
		Source:    muid.MUID(2),
		Timestamp: time.Now().UTC(),
	}

	data, err := evt.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Target should be omitted or empty when zero
	if val, exists := result["target"]; exists && val != nil && val != "" {
		t.Errorf("target should be omitted when zero, got %v", val)
	}
}

func TestEvent_MarshalJSON_NoData(t *testing.T) {
	evt := Event{
		ID:        muid.MUID(1),
		Type:      TypeAgentStopped,
		Source:    muid.MUID(2),
		Timestamp: time.Now().UTC(),
	}

	data, err := evt.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// data should be omitted (omitempty)
	if val, exists := result["data"]; exists && val != nil {
		t.Errorf("data should be omitted when nil, got %v", val)
	}
}

func TestEvent_MarshalJSON_WithTraceID(t *testing.T) {
	evt := Event{
		ID:        muid.MUID(1),
		Type:      TypeAgentStarted,
		Source:    muid.MUID(2),
		Timestamp: time.Now().UTC(),
		TraceID:   "trace-xyz-789",
	}

	data, err := evt.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	traceStr, ok := result["trace_id"].(string)
	if !ok {
		t.Fatalf("trace_id is not a string, got %T", result["trace_id"])
	}
	if traceStr != "trace-xyz-789" {
		t.Errorf("trace_id = %q, want %q", traceStr, "trace-xyz-789")
	}
}

func TestEvent_UnmarshalJSON_RoundTrip(t *testing.T) {
	source := muid.Make()
	target := muid.Make()
	ts := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	original := Event{
		ID:        muid.Make(),
		Type:      TypeAgentStarted,
		Source:    source,
		Target:    target,
		Timestamp: ts,
		Data:      map[string]string{"key": "value"},
		TraceID:   "trace-123",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID = %d, want %d", decoded.ID, original.ID)
	}
	if decoded.Type != original.Type {
		t.Errorf("Type = %q, want %q", decoded.Type, original.Type)
	}
	if decoded.Source != original.Source {
		t.Errorf("Source = %d, want %d", decoded.Source, original.Source)
	}
	if decoded.Target != original.Target {
		t.Errorf("Target = %d, want %d", decoded.Target, original.Target)
	}
	if !decoded.Timestamp.Equal(original.Timestamp) {
		t.Errorf("Timestamp = %v, want %v", decoded.Timestamp, original.Timestamp)
	}
	if decoded.TraceID != original.TraceID {
		t.Errorf("TraceID = %q, want %q", decoded.TraceID, original.TraceID)
	}
	// Data is json.RawMessage after round-trip
	if decoded.Data == nil {
		t.Error("Data should not be nil")
	}
}

func TestEvent_UnmarshalJSON_NoTarget(t *testing.T) {
	original := Event{
		ID:        muid.MUID(42),
		Type:      TypePTYOutput,
		Source:    muid.MUID(7),
		Timestamp: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Target != 0 {
		t.Errorf("Target = %d, want 0", decoded.Target)
	}
	if decoded.ID != muid.MUID(42) {
		t.Errorf("ID = %d, want 42", decoded.ID)
	}
}

func TestEvent_UnmarshalJSON_InvalidID(t *testing.T) {
	raw := `{"id":"notanumber","type":"test","source":"1","timestamp":"2025-01-01T00:00:00Z"}`
	var evt Event
	if err := json.Unmarshal([]byte(raw), &evt); err == nil {
		t.Error("should fail for non-numeric ID")
	}
}

func TestLocalDispatcher_DispatchAndSubscribe(t *testing.T) {
	d := NewLocalDispatcher()
	defer d.Close()

	var received []Event
	var mu sync.Mutex

	unsub := d.Subscribe(Subscription{
		Handler: func(ctx context.Context, evt Event) error {
			mu.Lock()
			received = append(received, evt)
			mu.Unlock()
			return nil
		},
	})
	defer unsub()

	source := muid.Make()
	evt := NewEvent(TypeAgentStarted, source, "test-data")

	if err := d.Dispatch(context.Background(), evt); err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if len(received) != 1 {
		t.Fatalf("expected 1 event, got %d", len(received))
	}
	if received[0].Type != TypeAgentStarted {
		t.Errorf("received type = %q, want %q", received[0].Type, TypeAgentStarted)
	}
}

func TestLocalDispatcher_TypeFiltering(t *testing.T) {
	d := NewLocalDispatcher()
	defer d.Close()

	var received []Event
	var mu sync.Mutex

	// Subscribe only to agent.started events
	unsub := d.Subscribe(Subscription{
		Types: []Type{TypeAgentStarted},
		Handler: func(ctx context.Context, evt Event) error {
			mu.Lock()
			received = append(received, evt)
			mu.Unlock()
			return nil
		},
	})
	defer unsub()

	source := muid.Make()

	// Dispatch matching event
	if err := d.Dispatch(context.Background(), NewEvent(TypeAgentStarted, source, nil)); err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	// Dispatch non-matching event
	if err := d.Dispatch(context.Background(), NewEvent(TypeAgentStopped, source, nil)); err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if len(received) != 1 {
		t.Fatalf("expected 1 event (filtered), got %d", len(received))
	}
	if received[0].Type != TypeAgentStarted {
		t.Errorf("received wrong type: got %q, want %q", received[0].Type, TypeAgentStarted)
	}
}

func TestLocalDispatcher_MultipleTypeFilters(t *testing.T) {
	d := NewLocalDispatcher()
	defer d.Close()

	var received []Event

	unsub := d.Subscribe(Subscription{
		Types: []Type{TypeAgentStarted, TypeAgentStopped},
		Handler: func(ctx context.Context, evt Event) error {
			received = append(received, evt)
			return nil
		},
	})
	defer unsub()

	source := muid.Make()

	if err := d.Dispatch(context.Background(), NewEvent(TypeAgentStarted, source, nil)); err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}
	if err := d.Dispatch(context.Background(), NewEvent(TypeAgentStopped, source, nil)); err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}
	if err := d.Dispatch(context.Background(), NewEvent(TypePTYOutput, source, nil)); err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if len(received) != 2 {
		t.Fatalf("expected 2 events, got %d", len(received))
	}
}

func TestLocalDispatcher_EmptyTypesReceivesAll(t *testing.T) {
	d := NewLocalDispatcher()
	defer d.Close()

	var received []Event

	// Subscribe with no type filter (should receive all)
	unsub := d.Subscribe(Subscription{
		Handler: func(ctx context.Context, evt Event) error {
			received = append(received, evt)
			return nil
		},
	})
	defer unsub()

	source := muid.Make()

	if err := d.Dispatch(context.Background(), NewEvent(TypeAgentStarted, source, nil)); err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}
	if err := d.Dispatch(context.Background(), NewEvent(TypePTYOutput, source, nil)); err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}
	if err := d.Dispatch(context.Background(), NewEvent(TypeProcessSpawned, source, nil)); err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if len(received) != 3 {
		t.Errorf("expected 3 events with no type filter, got %d", len(received))
	}
}

func TestLocalDispatcher_Unsubscribe(t *testing.T) {
	d := NewLocalDispatcher()
	defer d.Close()

	callCount := 0
	unsub := d.Subscribe(Subscription{
		Handler: func(ctx context.Context, evt Event) error {
			callCount++
			return nil
		},
	})

	source := muid.Make()

	// Dispatch before unsubscribe
	if err := d.Dispatch(context.Background(), NewEvent(TypeAgentStarted, source, nil)); err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}
	if callCount != 1 {
		t.Fatalf("expected callCount=1 before unsub, got %d", callCount)
	}

	// Unsubscribe
	unsub()

	// Dispatch after unsubscribe
	if err := d.Dispatch(context.Background(), NewEvent(TypeAgentStarted, source, nil)); err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected callCount=1 after unsub, got %d", callCount)
	}
}

func TestLocalDispatcher_MultipleSubscribers(t *testing.T) {
	d := NewLocalDispatcher()
	defer d.Close()

	count1 := 0
	count2 := 0

	unsub1 := d.Subscribe(Subscription{
		Handler: func(ctx context.Context, evt Event) error {
			count1++
			return nil
		},
	})
	defer unsub1()

	unsub2 := d.Subscribe(Subscription{
		Handler: func(ctx context.Context, evt Event) error {
			count2++
			return nil
		},
	})
	defer unsub2()

	source := muid.Make()
	if err := d.Dispatch(context.Background(), NewEvent(TypeAgentStarted, source, nil)); err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if count1 != 1 {
		t.Errorf("subscriber 1 count = %d, want 1", count1)
	}
	if count2 != 1 {
		t.Errorf("subscriber 2 count = %d, want 1", count2)
	}
}

func TestLocalDispatcher_ClosedDispatch(t *testing.T) {
	d := NewLocalDispatcher()

	if err := d.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	source := muid.Make()
	err := d.Dispatch(context.Background(), NewEvent(TypeAgentStarted, source, nil))
	if err == nil {
		t.Fatal("Dispatch on closed dispatcher should return error")
	}

	var dce *DispatcherClosedError
	if !errors.As(err, &dce) {
		t.Errorf("expected DispatcherClosedError, got %T: %v", err, err)
	}
}

func TestLocalDispatcher_HandlerError_ContinuesDispatching(t *testing.T) {
	d := NewLocalDispatcher()
	defer d.Close()

	count := 0

	// First subscriber that always fails
	unsub1 := d.Subscribe(Subscription{
		Handler: func(ctx context.Context, evt Event) error {
			return fmt.Errorf("handler error")
		},
	})
	defer unsub1()

	// Second subscriber that should still be called
	unsub2 := d.Subscribe(Subscription{
		Handler: func(ctx context.Context, evt Event) error {
			count++
			return nil
		},
	})
	defer unsub2()

	source := muid.Make()
	err := d.Dispatch(context.Background(), NewEvent(TypeAgentStarted, source, nil))
	if err != nil {
		t.Fatalf("Dispatch should not return error when handler fails: %v", err)
	}

	if count != 1 {
		t.Errorf("second subscriber should have been called, count = %d", count)
	}
}

func TestLocalDispatcher_TargetedEvent(t *testing.T) {
	d := NewLocalDispatcher()
	defer d.Close()

	subID := muid.Make()
	otherID := muid.Make()

	received := false
	otherReceived := false

	unsub1 := d.Subscribe(Subscription{
		ID: subID,
		Handler: func(ctx context.Context, evt Event) error {
			received = true
			return nil
		},
	})
	defer unsub1()

	unsub2 := d.Subscribe(Subscription{
		ID: otherID,
		Handler: func(ctx context.Context, evt Event) error {
			otherReceived = true
			return nil
		},
	})
	defer unsub2()

	source := muid.Make()
	evt := NewEvent(TypeAgentStarted, source, nil).WithTarget(subID)

	if err := d.Dispatch(context.Background(), evt); err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if !received {
		t.Error("targeted subscriber should have received the event")
	}
	if otherReceived {
		t.Error("non-targeted subscriber should not have received the event")
	}
}

func TestLocalDispatcher_SubscribeAutoGeneratesID(t *testing.T) {
	d := NewLocalDispatcher()
	defer d.Close()

	called := false
	unsub := d.Subscribe(Subscription{
		// ID is zero, should be auto-generated
		Handler: func(ctx context.Context, evt Event) error {
			called = true
			return nil
		},
	})
	defer unsub()

	source := muid.Make()
	if err := d.Dispatch(context.Background(), NewEvent(TypeAgentStarted, source, nil)); err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if !called {
		t.Error("subscriber with auto-generated ID should have been called")
	}
}

func TestNoopDispatcher_Dispatch(t *testing.T) {
	d := NewNoopDispatcher()

	source := muid.Make()
	err := d.Dispatch(context.Background(), NewEvent(TypeAgentStarted, source, nil))
	if err != nil {
		t.Errorf("NoopDispatcher.Dispatch() = %v, want nil", err)
	}
}

func TestNoopDispatcher_Subscribe(t *testing.T) {
	d := NewNoopDispatcher()

	unsub := d.Subscribe(Subscription{
		Handler: func(ctx context.Context, evt Event) error {
			t.Error("NoopDispatcher should never call handler")
			return nil
		},
	})

	// Unsubscribe should not panic
	unsub()
}

func TestNoopDispatcher_Close(t *testing.T) {
	d := NewNoopDispatcher()

	err := d.Close()
	if err != nil {
		t.Errorf("NoopDispatcher.Close() = %v, want nil", err)
	}
}

func TestDispatcherClosedError_Message(t *testing.T) {
	err := &DispatcherClosedError{}
	msg := err.Error()
	if msg == "" {
		t.Error("DispatcherClosedError.Error() should not be empty")
	}
	if msg != "event dispatcher is closed" {
		t.Errorf("DispatcherClosedError.Error() = %q, want %q", msg, "event dispatcher is closed")
	}
}

func TestEventTypes_Constants(t *testing.T) {
	types := []struct {
		name  string
		value Type
		want  string
	}{
		{"TypeAgentAdded", TypeAgentAdded, "agent.added"},
		{"TypeAgentStarting", TypeAgentStarting, "agent.starting"},
		{"TypeAgentStarted", TypeAgentStarted, "agent.started"},
		{"TypeAgentStopping", TypeAgentStopping, "agent.stopping"},
		{"TypeAgentStopped", TypeAgentStopped, "agent.stopped"},
		{"TypeAgentTerminated", TypeAgentTerminated, "agent.terminated"},
		{"TypeAgentErrored", TypeAgentErrored, "agent.errored"},
		{"TypePresenceChanged", TypePresenceChanged, "presence.changed"},
		{"TypePTYOutput", TypePTYOutput, "pty.output"},
		{"TypePTYActivity", TypePTYActivity, "pty.activity"},
		{"TypePTYIdle", TypePTYIdle, "pty.idle"},
		{"TypePTYStuck", TypePTYStuck, "pty.stuck"},
		{"TypeProcessSpawned", TypeProcessSpawned, "process.spawned"},
		{"TypeProcessCompleted", TypeProcessCompleted, "process.completed"},
		{"TypeProcessIO", TypeProcessIO, "process.io"},
		{"TypeShutdownInitiated", TypeShutdownInitiated, "shutdown.initiated"},
		{"TypeGitMergeRequested", TypeGitMergeRequested, "git.merge.requested"},
		{"TypeGitMergeConflict", TypeGitMergeConflict, "git.merge.conflict"},
		{"TypeConnectionEstablished", TypeConnectionEstablished, "connection.established"},
		{"TypeAdapterLoaded", TypeAdapterLoaded, "adapter.loaded"},
		{"TypeAdapterUnloaded", TypeAdapterUnloaded, "adapter.unloaded"},
		{"TypeTaskCancel", TypeTaskCancel, "task.cancel"},
	}

	for _, tt := range types {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.value, tt.want)
			}
		})
	}
}

func TestSetAndGetDefaultDispatcher(t *testing.T) {
	// Save original
	original := GetDefaultDispatcher()
	defer SetDefaultDispatcher(original)

	noop := NewNoopDispatcher()
	SetDefaultDispatcher(noop)

	got := GetDefaultDispatcher()
	if got != noop {
		t.Error("GetDefaultDispatcher should return the dispatcher set by SetDefaultDispatcher")
	}
}

func TestDispatchPackageLevel(t *testing.T) {
	// Save and restore original
	original := GetDefaultDispatcher()
	defer SetDefaultDispatcher(original)

	d := NewLocalDispatcher()
	SetDefaultDispatcher(d)
	defer d.Close()

	called := false
	unsub := Subscribe(Subscription{
		Handler: func(ctx context.Context, evt Event) error {
			called = true
			return nil
		},
	})
	defer unsub()

	source := muid.Make()
	if err := Dispatch(context.Background(), NewEvent(TypeAgentStarted, source, nil)); err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if !called {
		t.Error("package-level Dispatch should route through default dispatcher")
	}
}

func TestSubscribePackageLevel(t *testing.T) {
	// Save and restore original
	original := GetDefaultDispatcher()
	defer SetDefaultDispatcher(original)

	d := NewLocalDispatcher()
	SetDefaultDispatcher(d)
	defer d.Close()

	count := 0
	unsub := Subscribe(Subscription{
		Handler: func(ctx context.Context, evt Event) error {
			count++
			return nil
		},
	})

	source := muid.Make()
	if err := Dispatch(context.Background(), NewEvent(TypeAgentStarted, source, nil)); err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}

	unsub()

	if err := Dispatch(context.Background(), NewEvent(TypeAgentStarted, source, nil)); err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if count != 1 {
		t.Errorf("after unsub count = %d, want 1", count)
	}
}

func TestErrDispatcherClosed_IsPointer(t *testing.T) {
	if ErrDispatcherClosed == nil {
		t.Fatal("ErrDispatcherClosed should not be nil")
	}

	var dce *DispatcherClosedError
	if !errors.As(ErrDispatcherClosed, &dce) {
		t.Errorf("ErrDispatcherClosed should be a *DispatcherClosedError, got %T", ErrDispatcherClosed)
	}
}
