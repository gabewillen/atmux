package adapter

import (
	"context"
	"testing"

	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/experimental/wazerotest"
)

func TestCallConfigDefaultSuccess(t *testing.T) {
	memory := wazerotest.NewMemory(64)
	ptr := uint32(8)
	payload := []byte("hello")
	copy(memory.Bytes[ptr:], payload)
	packed := (uint64(ptr) << 32) | uint64(len(payload))

	var freedPtr uint64
	var freedLen uint64
	freeFn := wazerotest.NewFunction(func(ctx context.Context, m api.Module, p uint64, l uint64) {
		freedPtr = p
		freedLen = l
	})
	freeFn.ExportNames = []string{"amux_free"}
	configFn := wazerotest.NewFunction(func(ctx context.Context, m api.Module) uint64 {
		return packed
	})
	configFn.ExportNames = []string{"config_default"}

	module := wazerotest.NewModule(memory, configFn, freeFn)

	exportedFree := module.ExportedFunction("amux_free")
	exportedConfig := module.ExportedFunction("config_default")
	out, err := callConfigDefault(context.Background(), module, exportedFree, exportedConfig)
	if err != nil {
		t.Fatalf("call config default: %v", err)
	}
	if string(out) != "hello" {
		t.Fatalf("unexpected output: %q", string(out))
	}
	if freedPtr != uint64(ptr) || freedLen != uint64(len(payload)) {
		t.Fatalf("expected free to be called with ptr %d len %d", ptr, len(payload))
	}
}

func TestCallConfigDefaultMissingFree(t *testing.T) {
	memory := wazerotest.NewMemory(64)
	ptr := uint32(4)
	payload := []byte("data")
	copy(memory.Bytes[ptr:], payload)
	packed := (uint64(ptr) << 32) | uint64(len(payload))

	configFn := wazerotest.NewFunction(func(ctx context.Context, m api.Module) uint64 {
		return packed
	})
	configFn.ExportNames = []string{"config_default"}
	module := wazerotest.NewModule(memory, configFn)

	exportedConfig := module.ExportedFunction("config_default")
	if _, err := callConfigDefault(context.Background(), module, nil, exportedConfig); err == nil {
		t.Fatalf("expected missing export error")
	}
}

func TestCallConfigDefaultNoMemory(t *testing.T) {
	configFn := wazerotest.NewFunction(func(ctx context.Context, m api.Module) uint64 {
		return (uint64(1) << 32) | 1
	})
	configFn.ExportNames = []string{"config_default"}
	module := wazerotest.NewModule(nil, configFn)

	exportedConfig := module.ExportedFunction("config_default")
	if _, err := callConfigDefault(context.Background(), module, nil, exportedConfig); err == nil {
		t.Fatalf("expected memory error")
	}
}

func TestCallConfigDefaultEmptyLength(t *testing.T) {
	memory := wazerotest.NewMemory(64)
	configFn := wazerotest.NewFunction(func(ctx context.Context, m api.Module) uint64 {
		return uint64(1) << 32
	})
	configFn.ExportNames = []string{"config_default"}
	module := wazerotest.NewModule(memory, configFn)

	exportedConfig := module.ExportedFunction("config_default")
	out, err := callConfigDefault(context.Background(), module, nil, exportedConfig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil output")
	}
}

func TestCallConfigDefaultPackedZero(t *testing.T) {
	configFn := wazerotest.NewFunction(func(ctx context.Context, m api.Module) uint64 {
		return 0
	})
	configFn.ExportNames = []string{"config_default"}
	module := wazerotest.NewModule(nil, configFn)

	exportedConfig := module.ExportedFunction("config_default")
	if _, err := callConfigDefault(context.Background(), module, nil, exportedConfig); err == nil {
		t.Fatalf("expected packed zero error")
	}
}

func TestCallConfigDefaultReadFailure(t *testing.T) {
	memory := wazerotest.NewMemory(16)
	ptr := uint32(20)
	packed := (uint64(ptr) << 32) | 4

	configFn := wazerotest.NewFunction(func(ctx context.Context, m api.Module) uint64 {
		return packed
	})
	configFn.ExportNames = []string{"config_default"}
	module := wazerotest.NewModule(memory, configFn)

	exportedConfig := module.ExportedFunction("config_default")
	if _, err := callConfigDefault(context.Background(), module, nil, exportedConfig); err == nil {
		t.Fatalf("expected read failure")
	}
}

func TestCallConfigDefaultCallError(t *testing.T) {
	module := wazerotest.NewModule(nil)
	configFn := &wazerotest.Function{}

	if _, err := callConfigDefault(context.Background(), module, nil, configFn); err == nil {
		t.Fatalf("expected call error")
	}
}
