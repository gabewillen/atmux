package config

import (
	"strings"
	"testing"
)

func TestParseTOMLMultiLineString(t *testing.T) {
	doc := "key = \"\"\"line1\nline2\"\"\"\n"
	parsed, err := ParseTOML([]byte(doc))
	if err != nil {
		t.Fatalf("parse toml: %v", err)
	}
	value, ok := parsed["key"].(string)
	if !ok {
		t.Fatalf("expected string value")
	}
	if value != "line1\nline2" {
		t.Fatalf("unexpected multiline value: %q", value)
	}
}

func TestParseTOMLUnicodeEscape(t *testing.T) {
	doc := "key = \"\\u0041\"\n"
	parsed, err := ParseTOML([]byte(doc))
	if err != nil {
		t.Fatalf("parse toml: %v", err)
	}
	value, ok := parsed["key"].(string)
	if !ok || value != "A" {
		t.Fatalf("unexpected unicode value: %v", parsed["key"])
	}
}

func TestParseTOMLArrayErrors(t *testing.T) {
	doc := "items = [1 2]\n"
	if _, err := ParseTOML([]byte(doc)); err == nil {
		t.Fatalf("expected array separator error")
	}
}

func TestParseTOMLInlineTable(t *testing.T) {
	doc := "obj = { a = 1, b = \"two\" }\n"
	parsed, err := ParseTOML([]byte(doc))
	if err != nil {
		t.Fatalf("parse toml: %v", err)
	}
	obj, ok := parsed["obj"].(map[string]any)
	if !ok {
		t.Fatalf("expected inline table")
	}
	if obj["a"].(int64) != 1 || obj["b"].(string) != "two" {
		t.Fatalf("unexpected inline table: %#v", obj)
	}
}

func TestEncodeTOMLArrayTable(t *testing.T) {
	data := map[string]any{
		"items": []any{
			map[string]any{"name": "alpha", "meta": map[string]any{"rank": 1}},
			map[string]any{"name": "beta"},
		},
		"flag": true,
	}
	encoded, err := EncodeTOML(data)
	if err != nil {
		t.Fatalf("encode toml: %v", err)
	}
	text := string(encoded)
	if !strings.Contains(text, "[[items]]") {
		t.Fatalf("expected array table, got: %s", text)
	}
	if !strings.Contains(text, "meta.rank = 1") {
		t.Fatalf("expected flattened entry, got: %s", text)
	}
}

func TestEncodeTOMLRejectNestedArrayTable(t *testing.T) {
	data := map[string]any{
		"items": []any{
			map[string]any{"nested": []any{map[string]any{"name": "alpha"}}},
		},
	}
	if _, err := EncodeTOML(data); err == nil {
		t.Fatalf("expected nested array table error")
	}
}

func TestEncodeTOMLRejectArrayWithTable(t *testing.T) {
	data := map[string]any{
		"values": []any{1, map[string]any{"x": 1}},
	}
	if _, err := EncodeTOML(data); err == nil {
		t.Fatalf("expected array table in array error")
	}
}
