package protocol

import (
	"encoding/json"
	"testing"
)

func TestEventMessageBroadcast(t *testing.T) {
	data := &ProcessSpawnedEvent{
		PID:       12345,
		AgentID:   "42",
		ProcessID: "9002",
		Command:   "cargo",
		Args:      []string{"test"},
		WorkDir:   "/repo",
		ParentPID: 12000,
		StartedAt: "2026-01-18T10:30:00Z",
	}

	msg, err := NewBroadcastEvent("process.spawned", data)
	if err != nil {
		t.Fatalf("NewBroadcastEvent: %v", err)
	}

	if msg.Type != MsgBroadcast {
		t.Fatalf("Type = %d, want %d", msg.Type, MsgBroadcast)
	}

	if msg.Event.Name != "process.spawned" {
		t.Fatalf("Event.Name = %q, want %q", msg.Event.Name, "process.spawned")
	}

	// Verify JSON encoding matches spec example format
	jsonData, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	// Decode and verify structure
	var decoded EventMessage
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Type != MsgBroadcast {
		t.Fatalf("decoded Type = %d, want %d", decoded.Type, MsgBroadcast)
	}

	// Decode the event data
	var spawned ProcessSpawnedEvent
	if err := json.Unmarshal(decoded.Event.Data, &spawned); err != nil {
		t.Fatalf("Unmarshal event data: %v", err)
	}

	if spawned.PID != 12345 {
		t.Fatalf("PID = %d, want 12345", spawned.PID)
	}
	if spawned.AgentID != "42" {
		t.Fatalf("AgentID = %q, want %q", spawned.AgentID, "42")
	}
}

func TestEventMessageUnicast(t *testing.T) {
	data := map[string]string{"key": "value"}
	msg, err := NewUnicastEvent("custom.event", "42", data)
	if err != nil {
		t.Fatalf("NewUnicastEvent: %v", err)
	}

	if msg.Type != MsgUnicast {
		t.Fatalf("Type = %d, want %d", msg.Type, MsgUnicast)
	}
	if msg.Target != "42" {
		t.Fatalf("Target = %q, want %q", msg.Target, "42")
	}
}

func TestMessageTypeEncoding(t *testing.T) {
	// Per spec §9.1.3.1: type MUST be encoded as a JSON number
	msg := &EventMessage{
		Type: MsgBroadcast,
		Event: WireEvent{
			Name: "test",
			Data: json.RawMessage(`{}`),
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	// Verify type is a number in JSON
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal raw: %v", err)
	}

	typeStr := string(raw["type"])
	if typeStr != "1" {
		t.Fatalf("type JSON = %s, want 1", typeStr)
	}
}

func TestConnectionEstablishedEvent(t *testing.T) {
	evt := &ConnectionEstablishedEvent{
		PeerID:    "5678",
		Timestamp: "2026-01-18T10:30:00Z",
	}

	data, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	expected := `{"peer_id":"5678","timestamp":"2026-01-18T10:30:00Z"}`
	if string(data) != expected {
		t.Fatalf("JSON = %s\nwant = %s", data, expected)
	}
}

func TestConnectionLostEvent(t *testing.T) {
	evt := &ConnectionLostEvent{
		PeerID:    "5678",
		Timestamp: "2026-01-18T10:30:00Z",
		Reason:    "io_error",
	}

	data, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	expected := `{"peer_id":"5678","timestamp":"2026-01-18T10:30:00Z","reason":"io_error"}`
	if string(data) != expected {
		t.Fatalf("JSON = %s\nwant = %s", data, expected)
	}
}

func TestProcessCompletedEvent(t *testing.T) {
	signal := 9
	evt := &ProcessCompletedEvent{
		PID:       12345,
		AgentID:   "42",
		ProcessID: "9002",
		Command:   "cargo",
		ExitCode:  0,
		Signal:    &signal,
		StartedAt: "2026-01-18T10:30:00Z",
		EndedAt:   "2026-01-18T10:31:05Z",
		Duration:  "1m5s",
	}

	data, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded ProcessCompletedEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Signal == nil || *decoded.Signal != 9 {
		t.Fatalf("Signal = %v, want *9", decoded.Signal)
	}
}

func TestProcessCompletedEventNullSignal(t *testing.T) {
	evt := &ProcessCompletedEvent{
		PID:       12345,
		AgentID:   "42",
		ProcessID: "9002",
		Command:   "cargo",
		ExitCode:  0,
		Signal:    nil,
		StartedAt: "2026-01-18T10:30:00Z",
		EndedAt:   "2026-01-18T10:31:05Z",
		Duration:  "1m5s",
	}

	data, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	// signal should be null in JSON
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal raw: %v", err)
	}

	if string(raw["signal"]) != "null" {
		t.Fatalf("signal JSON = %s, want null", raw["signal"])
	}
}

func TestProcessIOEvent(t *testing.T) {
	evt := &ProcessIOEvent{
		PID:       12345,
		AgentID:   "42",
		ProcessID: "9002",
		Command:   "cargo",
		Stream:    "stderr",
		DataB64:   "dGVzdCBmYWlsZWQK",
		Timestamp: "2026-01-18T10:30:30Z",
	}

	data, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded ProcessIOEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Stream != "stderr" {
		t.Fatalf("Stream = %q, want %q", decoded.Stream, "stderr")
	}
	if decoded.DataB64 != "dGVzdCBmYWlsZWQK" {
		t.Fatalf("DataB64 = %q, want %q", decoded.DataB64, "dGVzdCBmYWlsZWQK")
	}
}
