package api

import (
	"encoding/json"
	"testing"
)

func TestParseRuntimeIDRejectsZero(t *testing.T) {
	if _, err := ParseRuntimeID("0"); err == nil {
		t.Fatalf("expected error for zero id")
	}
}

func TestNewRuntimeIDNonZero(t *testing.T) {
	id := NewRuntimeID()
	if id.Value() == 0 {
		t.Fatalf("expected non-zero id")
	}
}

func TestMarshalRuntimeIDRejectsZero(t *testing.T) {
	var id RuntimeID
	if _, err := json.Marshal(id); err == nil {
		t.Fatalf("expected error when marshaling zero id")
	}
}

func TestAgentIDJSONRoundTrip(t *testing.T) {
	id := NewAgentID()
	data, err := json.Marshal(id)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded AgentID
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Value() != id.Value() {
		t.Fatalf("expected %s, got %s", id.String(), decoded.String())
	}
}

func TestParseHostIDRejectsEmpty(t *testing.T) {
	if _, err := ParseHostID(""); err == nil {
		t.Fatalf("expected error for empty host id")
	}
}
