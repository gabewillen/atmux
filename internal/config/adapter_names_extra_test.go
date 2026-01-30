package config

import "testing"

func TestAdapterNamesFromMap(t *testing.T) {
	root := map[string]any{
		"adapters": map[string]any{
			"alpha": map[string]any{},
		},
		"agents": []any{
			map[string]any{"adapter": "beta"},
			map[string]any{"adapter": 123},
		},
	}
	names := adapterNamesFromMap(root)
	if len(names) != 2 {
		t.Fatalf("expected 2 adapter names")
	}
}
