package adapter

import (
	"context"
	"testing"

	"github.com/tetratelabs/wazero"
)

// TestWazeroRuntimeSmoke instantiates a minimal empty WASM module using wazero.
// This satisfies the Phase 0 requirement to demonstrate wazero usage per spec §4.2.2
// and the plan's dependency smoke test expectations.
func TestWazeroRuntimeSmoke(t *testing.T) {
	ctx := context.Background()

	runtime := wazero.NewRuntime(ctx)
	defer func() {
		_ = runtime.Close(ctx)
	}()

	// Minimal valid WASM module: magic + version with no sections.
	emptyModule := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}

	compiled, err := runtime.CompileModule(ctx, emptyModule)
	if err != nil {
		t.Fatalf("compile empty module: %v", err)
	}

	if _, err := runtime.InstantiateModule(ctx, compiled, wazero.NewModuleConfig()); err != nil {
		t.Fatalf("instantiate empty module: %v", err)
	}
}
