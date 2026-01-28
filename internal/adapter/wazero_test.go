package adapter

import (
	"context"
	"testing"

	"github.com/tetratelabs/wazero"
)

func TestWazeroInstantiate(t *testing.T) {
	ctx := context.Background()
	runtime := wazero.NewRuntime(ctx)
	wasm := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	mod, err := runtime.Instantiate(ctx, wasm)
	if err != nil {
		t.Fatalf("instantiate: %v", err)
	}
	if err := mod.Close(ctx); err != nil {
		t.Fatalf("close module: %v", err)
	}
	if err := runtime.Close(ctx); err != nil {
		t.Fatalf("close runtime: %v", err)
	}
}
