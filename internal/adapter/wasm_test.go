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

// Since we cannot easily generate valid WASM with imports/exports in pure Go test without TinyGo installed/running,
// we rely on integration tests (conformance) for the full flow.
// However, we can test that Match returns error if functions are missing (which they are in a real WASM if not exported).
// But instantiating "invalid" fails.
// We'd need a valid minimal WASM module.
// minimal empty wasm: \x00\x61\x73\x6d\x01\x00\x00\x00
func TestWasmRuntime_MissingExports(t *testing.T) {
	minimalWasm := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	r, err := NewWasmRuntime(context.Background(), minimalWasm)
	if err != nil {
		t.Fatalf("Failed to instantiate minimal wasm: %v", err)
	}
	defer r.Stop()

	_, err = r.Match([]byte("input"))
	if err == nil {
		t.Error("Expected error due to missing exports")
	}
	// Error should be about amux_alloc not exported
}