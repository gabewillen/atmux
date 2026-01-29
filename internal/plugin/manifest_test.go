package plugin

import (
	"testing"
)

func TestParseManifest(t *testing.T) {
	valid := `
name = "test-plugin"
version = "0.1.0"
description = "A test plugin"
permissions = ["agent.list", "agent.create"]
entrypoint = "plugin.wasm"
`
	m, err := ParseManifest([]byte(valid))
	if err != nil {
		t.Fatalf("ParseManifest failed: %v", err)
	}
	
	if m.Name != "test-plugin" {
		t.Errorf("Name mismatch")
	}
	if len(m.Permissions) != 2 {
		t.Errorf("Permissions len mismatch")
	}
}

func TestManifest_Validate(t *testing.T) {
	invalid := `
name = "bad"
version = "1.0"
# missing entrypoint
`
	_, err := ParseManifest([]byte(invalid))
	if err == nil {
		t.Error("Expected error for missing entrypoint")
	}
}
