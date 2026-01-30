package integrationtest

import (
	"context"
	"testing"
)

func TestHarnessContextDefaults(t *testing.T) {
	var h *Harness
	if h.Context() == nil {
		t.Fatalf("expected background context")
	}
	if err := h.Close(); err != nil {
		t.Fatalf("expected nil close")
	}
	type ctxKey string
	ctx := context.WithValue(context.Background(), ctxKey("k"), "v")
	if got := h.contextOrDefault(ctx); got != ctx {
		t.Fatalf("expected passed context")
	}
}
