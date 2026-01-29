package plugin

import (
	"testing"
)

func TestPluginLifecycle(t *testing.T) {
	mgr := NewManager()
	manifest := Manifest{
		Name:       "test-plugin",
		Version:    "1.0",
		Entrypoint: "main.wasm",
	}

	// Install
	if err := mgr.Install(manifest, "/path/to/plugin"); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// List
	list := mgr.List()
	if len(list) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(list))
	}
	if list[0].Manifest.Name != "test-plugin" {
		t.Errorf("Name mismatch")
	}
	if !list[0].Enabled {
		t.Error("Plugin should be enabled by default")
	}

	// Disable
	if err := mgr.Disable("test-plugin"); err != nil {
		t.Fatalf("Disable failed: %v", err)
	}
	if mgr.registry["test-plugin"].Enabled {
		t.Error("Plugin should be disabled")
	}

	// Enable
	if err := mgr.Enable("test-plugin"); err != nil {
		t.Fatalf("Enable failed: %v", err)
	}
	if !mgr.registry["test-plugin"].Enabled {
		t.Error("Plugin should be enabled")
	}

	// Remove
	if err := mgr.Remove("test-plugin"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	if len(mgr.List()) != 0 {
		t.Error("Registry should be empty")
	}
}
