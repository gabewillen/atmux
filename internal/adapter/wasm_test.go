package adapter

import (
	"context"
	"testing"
)

func TestNewWasmRuntime_Failure(t *testing.T) {
	// Invalid WASM
	_, err := NewWasmRuntime(context.Background(), []byte("invalid"))
	if err == nil {
		t.Error("Expected error for invalid WASM")
	}
}
