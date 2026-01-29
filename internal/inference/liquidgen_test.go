package inference

import (
	"context"
	"testing"
)

func TestLiquidgen_Link(t *testing.T) {
	// Just test if we can create the impl and call a method (which might fail if model missing, but shouldn't crash)
	impl := NewLiquidgenImpl()
	if impl == nil {
		t.Fatal("Failed to create LiquidgenImpl")
	}
	
	// We expect this to fail because models are missing, but it shouldn't be a linker error
	ctx := context.Background()
	_, err := impl.Generate(ctx, LiquidgenRequest{
		Model: "lfm2.5-thinking",
		Prompt: "Hello",
	})
	
	if err == nil {
		t.Error("Expected error for missing model, but got nil")
	} else {
		t.Logf("Got expected error: %v", err)
	}
}
