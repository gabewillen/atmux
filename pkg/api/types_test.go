package api

import "testing"

func TestParseIDRejectsZero(t *testing.T) {
	if _, err := ParseID("0"); err == nil {
		t.Fatalf("expected error for zero id")
	}
}

func TestNewIDNonZero(t *testing.T) {
	id := NewID()
	if id.Value() == 0 {
		t.Fatalf("expected non-zero id")
	}
}
