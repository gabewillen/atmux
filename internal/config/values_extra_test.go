package config

import (
	"testing"

	"github.com/agentflare-ai/amux/internal/paths"
)

func TestParseStringBoolInt(t *testing.T) {
	if value, ok := parseString(123); ok || value != "" {
		t.Fatalf("expected parseString failure")
	}
	if value, ok := parseString("ok"); !ok || value != "ok" {
		t.Fatalf("expected parseString success")
	}
	if value, ok := parseBool("no"); ok || value {
		t.Fatalf("expected parseBool failure")
	}
	if value, ok := parseBool(true); !ok || !value {
		t.Fatalf("expected parseBool success")
	}
	if value, ok := parseInt("no"); ok || value != 0 {
		t.Fatalf("expected parseInt failure")
	}
	if value, ok := parseInt(int64(7)); !ok || value != 7 {
		t.Fatalf("expected parseInt success")
	}
}

func TestExpandPath(t *testing.T) {
	resolver := &paths.Resolver{}
	value := expandPath(resolver, "/tmp/test")
	if value != "/tmp/test" {
		t.Fatalf("expected path unchanged")
	}
	if expandPath(nil, "/tmp/test") != "/tmp/test" {
		t.Fatalf("expected nil resolver to leave path")
	}
}
