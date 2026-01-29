package adapter

import (
	"testing"
)

func TestParseManifest(t *testing.T) {
	valid := `
name = "test-adapter"
version = "1.0.0"
description = "A test adapter"

[cli]
min_version = "0.1.0"
`
	m, err := ParseManifest([]byte(valid))
	if err != nil {
		t.Fatalf("ParseManifest failed: %v", err)
	}
	
	if m.Name != "test-adapter" {
		t.Errorf("Name mismatch")
	}
	if m.CLI.MinVersion != "0.1.0" {
		t.Errorf("CLI MinVersion mismatch")
	}
}

func TestManifest_Validate(t *testing.T) {
	invalid := `
name = "bad"
# missing version
[cli]
min_version = "1.0"
`
	_, err := ParseManifest([]byte(invalid))
	if err == nil {
		t.Error("Expected error for missing version")
	}
}
