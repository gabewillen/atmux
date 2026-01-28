// Package wasm provides WASM runtime management for adapters.
package wasm

import (
	"context"
	"fmt"
	"os"

	amuxerrors "github.com/agentflare-ai/amux/internal/errors"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// Runtime represents a WASM runtime for adapters.
type Runtime struct {
	ctx     context.Context
	runtime wazero.Runtime
	module  api.Module
	config  *Config
}

// Config contains WASM runtime configuration.
type Config struct {
	// Path to WASM module
	ModulePath string `json:"module_path"`

	// Memory limit in bytes
	MemoryLimit uint64 `json:"memory_limit"`

	// Enable debugging
	Debug bool `json:"debug"`
}

// New creates a new WASM runtime.
func New(ctx context.Context, config *Config) (*Runtime, error) {
	if config == nil {
		return nil, amuxerrors.Wrap("creating WASM runtime", amuxerrors.ErrInvalidConfig)
	}

	// Create wazero runtime
	runtime := wazero.NewRuntime(ctx)

	// Configure memory limit
	if config.MemoryLimit > 0 {
		// This will be used when instantiating modules
	}

	return &Runtime{
		ctx:     ctx,
		runtime: runtime,
		config:  config,
	}, nil
}

// LoadModule loads a WASM module.
func (r *Runtime) LoadModule() error {
	if r.config.ModulePath == "" {
		return amuxerrors.Wrap("loading WASM module", amuxerrors.ErrInvalidConfig)
	}

	// Read WASM file
	wasmBytes, err := os.ReadFile(r.config.ModulePath)
	if err != nil {
		return amuxerrors.Wrap("reading WASM module", err)
	}

	// Instantiate module
	module, err := r.runtime.InstantiateWithConfig(r.ctx, wasmBytes, wazero.NewModuleConfig().WithName("adapter"))
	if err != nil {
		return amuxerrors.Wrap("instantiating WASM module", err)
	}

	r.module = module
	return nil
}

// CallFunction calls a function in the loaded WASM module.
func (r *Runtime) CallFunction(name string, args ...uint64) (uint64, error) {
	if r.module == nil {
		return 0, amuxerrors.Wrap("calling WASM function", amuxerrors.ErrNotReady)
	}

	fn := r.module.ExportedFunction(name)
	if fn == nil {
		return 0, amuxerrors.Wrap(fmt.Sprintf("function %s not found", name), amuxerrors.ErrNotFound)
	}

	result, err := fn.Call(r.ctx, args...)
	if err != nil {
		return 0, amuxerrors.Wrap(fmt.Sprintf("calling function %s", name), err)
	}

	if len(result) == 0 {
		return 0, nil
	}

	return result[0], nil
}

// Memory returns the memory exports from the WASM module.
func (r *Runtime) Memory() api.Memory {
	if r.module == nil {
		return nil
	}

	return r.module.ExportedMemory("memory")
}

// Close closes the WASM runtime.
func (r *Runtime) Close() error {
	if r.module != nil {
		if err := r.module.Close(r.ctx); err != nil {
			return amuxerrors.Wrap("closing WASM module", err)
		}
	}

	return nil
}

// Demo demonstrates WASM functionality for Phase 0.
func Demo(ctx context.Context) error {
	// Create a simple WASM module in memory for demo
	wasmCode := []byte{
		// Wat code: (module (memory 1) (func (export "demo") (param i32) (result i32) local.get 0)
		0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, // module header
		0x01,                         // type section
		0x60, 0x01, 0x7f, 0x01, 0x7f, // func type (i32) -> i32
		0x03,             // function section
		0x02, 0x01, 0x00, // one function at index 0
		0x05,             // memory section
		0x01, 0x00, 0x01, // one memory page
		0x07,                                                 // export section
		0x07, 0x01, 0x06, 0x64, 0x65, 0x6d, 0x6f, 0x00, 0x00, // export "demo" function
		0x0a,                                                       // code section
		0x09, 0x01, 0x06, 0x00, 0x20, 0x00, 0x20, 0x00, 0x6a, 0x0b, // function body: local.get 0, local.get 0, i32.add
	}

	config := &Config{
		MemoryLimit: 65536, // 64KB
		Debug:       true,
	}

	runtime, err := New(ctx, config)
	if err != nil {
		return amuxerrors.Wrap("creating demo WASM runtime", err)
	}
	defer runtime.Close()

	// Load demo module directly from bytes
	module, err := runtime.runtime.InstantiateWithConfig(ctx, wasmCode, wazero.NewModuleConfig().WithName("demo"))
	if err != nil {
		return amuxerrors.Wrap("instantiating demo WASM", err)
	}

	// Call the demo function
	fn := module.ExportedFunction("demo")
	if fn == nil {
		return amuxerrors.New("demo function not found in WASM module")
	}

	result, err := fn.Call(ctx, uint64(42))
	if err != nil {
		return amuxerrors.Wrap("calling demo function", err)
	}

	if len(result) > 0 {
		fmt.Printf("WASM Demo: 42 + 42 = %d\n", result[0])
	}

	return nil
}
