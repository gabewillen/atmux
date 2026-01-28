package config

import (
	"os"
	"strings"
	"testing"
)

func TestSpecVersionGuard(t *testing.T) {
	data, err := os.ReadFile("../../docs/spec-v1.22.md")
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	if !strings.Contains(string(data), "**Version:** v1.22") {
		t.Fatalf("spec version marker missing")
	}
}
