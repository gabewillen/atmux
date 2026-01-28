package agent

import "testing"

func TestNewManager(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}
	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}
}

func TestManagerStart(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	// Should fail with empty adapter name
	err = manager.Start("", nil)
	if err == nil {
		t.Fatal("Start() should fail with empty adapter name")
	}

	// Should fail with unimplemented functionality
	err = manager.Start("claude-code", nil)
	if err == nil {
		t.Fatal("Start() should fail as not yet implemented")
	}
}