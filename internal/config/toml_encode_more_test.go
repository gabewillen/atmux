package config

import (
	"strings"
	"testing"
)

func TestFormatValueUnsupportedType(t *testing.T) {
	if _, err := formatValue(struct{}{}); err == nil {
		t.Fatalf("expected unsupported type error")
	}
}

func TestFormatArrayRejectsTables(t *testing.T) {
	if _, err := formatArray([]any{map[string]any{"a": 1}}); err == nil {
		t.Fatalf("expected array table error")
	}
}

func TestFlattenEntryRejectsNestedArrayTables(t *testing.T) {
	data := map[string]any{
		"nested": []any{
			map[string]any{"name": "alpha"},
		},
	}
	out := make(map[string]any)
	if err := flattenEntry("", data, out); err == nil {
		t.Fatalf("expected nested array table error")
	}
}

func TestWriteTableWithNilMap(t *testing.T) {
	var b strings.Builder
	if err := writeTable(&b, nil, nil); err != nil {
		t.Fatalf("expected nil table to be ignored: %v", err)
	}
}
