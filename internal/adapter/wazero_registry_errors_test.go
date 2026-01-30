package adapter

import (
	"context"
	"testing"

	"github.com/tetratelabs/wazero"
)

func TestNewWasmAdapterMissingExports(t *testing.T) {
	ctx := context.Background()
	runtime := wazero.NewRuntime(ctx)
	defer runtime.Close(ctx)
	wasm := append([]byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}, []byte{0x05, 0x03, 0x01, 0x00, 0x01}...)
	mod, err := runtime.Instantiate(ctx, wasm)
	if err != nil {
		t.Fatalf("instantiate: %v", err)
	}
	if _, err := newWasmAdapter("test", mod); err == nil {
		t.Fatalf("expected missing export error")
	}
}

func TestNewWasmAdapterMissingMemory(t *testing.T) {
	ctx := context.Background()
	runtime := wazero.NewRuntime(ctx)
	defer runtime.Close(ctx)
	mod, err := runtime.Instantiate(ctx, []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00})
	if err != nil {
		t.Fatalf("instantiate: %v", err)
	}
	if _, err := newWasmAdapter("test", mod); err == nil {
		t.Fatalf("expected missing memory error")
	}
}

