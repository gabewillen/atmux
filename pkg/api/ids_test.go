package api

import (
	"encoding/json"
	"testing"
)

func TestNormalizeAgentSlug(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase", "Frontend-Dev", "frontend-dev"},
		{"alphanumeric and dash", "backend-dev", "backend-dev"},
		{"empty", "", "agent"},
		{"only non-alphanumeric", "!!!", "agent"},
		{"trim and collapse", "  a  -  b  ", "a-b"},
		{"truncate 63", string(make([]byte, 70)), "agent"}, // all zeros -> trimmed to empty -> agent
		{"underscore to dash", "agent_name", "agent-name"},
		{"spec example frontend-dev", "Frontend Dev", "frontend-dev"},
		{"spec example backend-dev", "Backend Dev", "backend-dev"},
		{"spec example test-runner", "Test Runner", "test-runner"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeAgentSlug(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizeAgentSlug(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestNormalizeAgentSlugMaxLen(t *testing.T) {
	long := ""
	for i := 0; i < 70; i++ {
		long += "a"
	}
	got := NormalizeAgentSlug(long)
	if len(got) > MaxAgentSlugLen {
		t.Errorf("len(NormalizeAgentSlug(long)) = %d, want <= %d", len(got), MaxAgentSlugLen)
	}
}

func TestUniquifyAgentSlug(t *testing.T) {
	existing := map[string]struct{}{"frontend-dev": {}, "frontend-dev-2": {}}
	got := UniquifyAgentSlug("frontend-dev", existing)
	if got != "frontend-dev-3" {
		t.Errorf("UniquifyAgentSlug = %q, want frontend-dev-3", got)
	}
	got = UniquifyAgentSlug("backend-dev", existing)
	if got != "backend-dev" {
		t.Errorf("UniquifyAgentSlug(backend-dev) = %q, want backend-dev", got)
	}
}

func TestValidRuntimeID(t *testing.T) {
	if ValidRuntimeID(BroadcastID) {
		t.Error("ValidRuntimeID(0) should be false")
	}
	if !ValidRuntimeID(1) {
		t.Error("ValidRuntimeID(1) should be true")
	}
}

func TestEncodeDecodeID(t *testing.T) {
	id := ID(42)
	s := EncodeID(id)
	if s != "42" {
		t.Errorf("EncodeID(42) = %q, want \"42\"", s)
	}
	dec, err := DecodeID(s)
	if err != nil {
		t.Fatalf("DecodeID: %v", err)
	}
	if dec != id {
		t.Errorf("DecodeID(%q) = %v, want %v", s, dec, id)
	}
}

func TestDecodeIDErrors(t *testing.T) {
	_, err := DecodeID("")
	if err == nil {
		t.Error("DecodeID(\"\") should error")
	}
	_, err = DecodeID("x")
	if err == nil {
		t.Error("DecodeID(\"x\") should error")
	}
}

func TestIDJSONBase10(t *testing.T) {
	type W struct {
		ID ID `json:"id"`
	}
	w := W{ID: 12345}
	data, err := json.Marshal(w)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(data) != `{"id":"12345"}` {
		t.Errorf("Marshal = %s, want base-10 string", data)
	}
	var w2 W
	if err := json.Unmarshal(data, &w2); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if w2.ID != w.ID {
		t.Errorf("Unmarshal ID = %v, want %v", w2.ID, w.ID)
	}
}
