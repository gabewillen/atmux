package remote

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

// TestGenerateHostID verifies host ID generation.
func TestGenerateHostID(t *testing.T) {
	hostID := GenerateHostID()
	if hostID == "" {
		t.Fatal("GenerateHostID returned empty string")
	}
	
	// Generate multiple IDs and verify they're unique
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := GenerateHostID()
		if ids[id] {
			t.Fatalf("GenerateHostID produced duplicate: %s", id)
		}
		ids[id] = true
	}
}

// TestNATSManager verifies NATS manager creation.
func TestNATSManager(t *testing.T) {
	config := NATSConfig{
		URL:           "nats://localhost:4222",
		SubjectPrefix: "amux",
		KVBucket:      "TEST_KV",
		Timeout:       30 * time.Second,
	}
	
	nm, err := NewNATSManager("test-host", "director", config)
	if err != nil {
		t.Fatalf("NewNATSManager failed: %v", err)
	}
	
	if nm.GetHostID() != "test-host" {
		t.Errorf("Expected hostID 'test-host', got %s", nm.GetHostID())
	}
	
	if nm.role != "director" {
		t.Errorf("Expected role 'director', got %s", nm.role)
	}
	
	if nm.IsReady() {
		t.Error("Manager should not be ready before handshake")
	}
}

// TestControlMessage verifies control message marshaling.
func TestControlMessage(t *testing.T) {
	tests := []struct {
		name string
		msg  ControlMessage
	}{
		{
			name: "handshake",
			msg: ControlMessage{
				Type:    "handshake",
				Payload: []byte(`{"protocol":1,"peer_id":"123","role":"director","host_id":"test"}`),
			},
		},
		{
			name: "spawn",
			msg: ControlMessage{
				Type:    "spawn",
				Payload: []byte(`{"agent_id":"42","agent_slug":"test-agent"}`),
			},
		},
		{
			name: "error",
			msg: ControlMessage{
				Type:    "error",
				Payload: []byte(`{"request_type":"spawn","code":"not_ready","message":"handshake not completed"}`),
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON round-trip
			data, err := json.Marshal(tt.msg)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}
			
			var unmarshaled ControlMessage
			if err := json.Unmarshal(data, &unmarshaled); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}
			
			if unmarshaled.Type != tt.msg.Type {
				t.Errorf("Expected type %s, got %s", tt.msg.Type, unmarshaled.Type)
			}
		})
	}
}

// TestRemoteManager verifies remote manager creation and configuration.
func TestRemoteManager(t *testing.T) {
	config := &RemoteConfig{
		Role:           "director",
		HostID:         "test-director",
		NATSURL:        "nats://localhost:4222",
		SubjectPrefix:  "amux",
		KVBucket:       "TEST_KV",
		RequestTimeout: 30 * time.Second,
		BufferSize:     8192,
	}
	
	rm, err := NewRemoteManager(config)
	if err != nil {
		t.Fatalf("NewRemoteManager failed: %v", err)
	}
	
	if rm.GetRole() != "director" {
		t.Errorf("Expected role 'director', got %s", rm.GetRole())
	}
	
	if rm.GetHostID() != "test-director" {
		t.Errorf("Expected hostID 'test-director', got %s", rm.GetHostID())
	}
	
	if rm.IsReady() {
		t.Error("Manager should not be ready before start")
	}
	
	// Test invalid role
	config.Role = "invalid"
	_, err = NewRemoteManager(config)
	if err == nil {
		t.Error("Expected error for invalid role")
	}
}

// TestPTYData verifies PTY data marshaling.
func TestPTYData(t *testing.T) {
	ptyData := PTYData{
		SessionID: "test-session",
		Data:      []byte("test data"),
		Timestamp: time.Now().UTC(),
		Sequence:  42,
	}
	
	data, err := json.Marshal(ptyData)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	
	var unmarshaled PTYData
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	
	if unmarshaled.SessionID != ptyData.SessionID {
		t.Errorf("Expected session_id %s, got %s", ptyData.SessionID, unmarshaled.SessionID)
	}
	
	if string(unmarshaled.Data) != string(ptyData.Data) {
		t.Errorf("Expected data %s, got %s", ptyData.Data, unmarshaled.Data)
	}
	
	if unmarshaled.Sequence != ptyData.Sequence {
		t.Errorf("Expected sequence %d, got %d", ptyData.Sequence, unmarshaled.Sequence)
	}
}

// TestRingBuffer verifies ring buffer functionality.
func TestRingBuffer(t *testing.T) {
	rb := NewRingBuffer(3)
	
	if rb.Size() != 0 {
		t.Errorf("Expected size 0, got %d", rb.Size())
	}
	
	// Add items
	rb.Add("item1")
	rb.Add("item2")
	rb.Add("item3")
	
	if rb.Size() != 3 {
		t.Errorf("Expected size 3, got %d", rb.Size())
	}
	
	// Verify items in order
	items := make([]string, 0)
	rb.ForEach(func(item interface{}) {
		items = append(items, item.(string))
	})
	
	expected := []string{"item1", "item2", "item3"}
	if len(items) != len(expected) {
		t.Fatalf("Expected %d items, got %d", len(expected), len(items))
	}
	
	for i, item := range items {
		if item != expected[i] {
			t.Errorf("Expected item %s at index %d, got %s", expected[i], i, item)
		}
	}
	
	// Add one more to trigger wrap-around
	rb.Add("item4")
	
	if rb.Size() != 3 {
		t.Errorf("Expected size to remain 3, got %d", rb.Size())
	}
	
	// Verify oldest item was dropped
	items = make([]string, 0)
	rb.ForEach(func(item interface{}) {
		items = append(items, item.(string))
	})
	
	expected = []string{"item2", "item3", "item4"}
	for i, item := range items {
		if item != expected[i] {
			t.Errorf("After wrap-around, expected item %s at index %d, got %s", expected[i], i, item)
		}
	}
	
	// Test clear
	rb.Clear()
	if rb.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", rb.Size())
	}
}

// TestBootstrapConfig verifies bootstrap configuration validation.
func TestBootstrapConfig(t *testing.T) {
	config := BootstrapConfig{
		SSHHost:         "user@example.com",
		BinaryPath:      "/path/to/amux",
		AdapterPaths:    []string{"/path/to/adapter1.wasm", "/path/to/adapter2.wasm"},
		CredsPath:       "/path/to/creds.json",
		RemoteCredsPath: "~/.amux/nats.creds",
		HubURL:          "nats://hub.example.com:4222",
		Timeout:         60 * time.Second,
	}
	
	ctx := context.Background()
	
	// Test bootstrap with empty hostID
	err := Bootstrap(ctx, "", config)
	if err == nil {
		t.Error("Expected error for empty hostID")
	}
	
	// Verify error wrapping
	if !errors.Is(err, ErrBootstrapFailed) {
		t.Errorf("Expected error to wrap ErrBootstrapFailed, got %v", err)
	}
}