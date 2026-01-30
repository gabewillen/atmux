package api

import (
	"encoding/json"
	"testing"
)

func TestTargetIDValueAndBroadcast(t *testing.T) {
	broadcast, err := ParseTargetID("0")
	if err != nil {
		t.Fatalf("parse target id: %v", err)
	}
	if !broadcast.IsBroadcast() {
		t.Fatalf("expected broadcast id")
	}
	runtime := NewRuntimeID()
	target := TargetIDFromRuntime(runtime)
	if target.IsBroadcast() {
		t.Fatalf("expected non-broadcast id")
	}
	if target.Value() != runtime.Value() {
		t.Fatalf("expected target value to match runtime")
	}
	data, err := json.Marshal(target)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded TargetID
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Value() != runtime.Value() {
		t.Fatalf("expected target round trip")
	}
}
