package config

import "testing"

func TestSetKeyAndSetPath(t *testing.T) {
	root := map[string]any{}
	if err := setKey(root, "alpha", 1); err != nil {
		t.Fatalf("set key: %v", err)
	}
	if err := setKey(root, "alpha", 2); err == nil {
		t.Fatalf("expected duplicate key error")
	}
	root = map[string]any{"a": "not-map"}
	if err := setKey(root, "a.b", 1); err == nil {
		t.Fatalf("expected table conflict")
	}
	root = map[string]any{}
	if err := setPath(root, nil, 1); err == nil {
		t.Fatalf("expected empty path error")
	}
	root = map[string]any{"a": "bad"}
	if err := setPath(root, []string{"a", "b"}, 1); err == nil {
		t.Fatalf("expected path conflict")
	}
	root = map[string]any{}
	if err := setPath(root, []string{"a", "b"}, 1); err != nil {
		t.Fatalf("set path: %v", err)
	}
}
