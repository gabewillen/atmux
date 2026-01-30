package config

import "testing"

func TestParseStringValueWithComment(t *testing.T) {
	t.Parallel()
	value, err := parseStringValue("\"ok\" # trailing comment")
	if err != nil {
		t.Fatalf("parse string: %v", err)
	}
	if value != "ok" {
		t.Fatalf("unexpected string: %q", value)
	}
	if _, err := parseStringValue("123"); err == nil {
		t.Fatalf("expected invalid string error")
	}
}

