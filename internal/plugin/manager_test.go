package plugin

import (
	"testing"
)

func TestManager_Lifecycle(t *testing.T) {
	m := NewManager()
	
	manifest := Manifest{
		Name:       "test-plugin",
		Version:    "1.0.0",
		Entrypoint: "test.wasm",
	}
	
	// Install
	if err := m.Install(manifest, "/path/to/plugin"); err != nil {
		t.Fatalf("Install failed: %v", err)
	}
	
	// List
	list := m.List()
	if len(list) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(list))
	}
	if list[0].Manifest.Name != "test-plugin" {
		t.Errorf("Name mismatch")
	}
	if !list[0].Enabled {
		t.Error("Should be enabled by default")
	}
	
	// Disable
	if err := m.Disable("test-plugin"); err != nil {
		t.Errorf("Disable failed: %v", err)
	}
	if m.registry["test-plugin"].Enabled {
		t.Error("Should be disabled")
	}
	
	// Enable
	if err := m.Enable("test-plugin"); err != nil {
		t.Errorf("Enable failed: %v", err)
	}
	if !m.registry["test-plugin"].Enabled {
		t.Error("Should be enabled")
	}
	
	// Remove
	if err := m.Remove("test-plugin"); err != nil {
		t.Errorf("Remove failed: %v", err)
	}
	if len(m.List()) != 0 {
		t.Error("Should be empty after remove")
	}
}