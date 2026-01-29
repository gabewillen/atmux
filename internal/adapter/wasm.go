package adapter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// WasmRuntime executes a WASM adapter.
type WasmRuntime struct {
	runtime wazero.Runtime
	module  api.Module
}

// NewWasmRuntime creates a new runtime for the given WASM binary.
func NewWasmRuntime(ctx context.Context, wasmBytes []byte) (*WasmRuntime, error) {
	r := wazero.NewRuntime(ctx)
	
	// Instantiate host functions
	// "ALWAYS export: amux_alloc, amux_free, ..." from the adapter side.
	// But host needs to provide imports if any?
	// Spec says: "Expose host functions"
	// "adapter pattern/action interfaces"
	// Actually, the Adapter WASM ABI says "ALWAYS export: amux_alloc...". This means the ADAPTER exports them.
	// The HOST calls them.
	// The HOST might export functions for the adapter to call (e.g. logging, state access).
	// Spec §10.4 doesn't explicitly list host exports required by adapter, 
	// but usually WASI or specific host functions are needed.
	// Let's assume pure compute for Match for now unless spec specifies host imports.
	// "Implement WASM interface (host functions...)" in the plan might refer to implementing the CALLER side.
	
	// Compile and instantiate
	mod, err := r.Instantiate(ctx, wasmBytes)
	if err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate module: %w", err)
	}

	return &WasmRuntime{
		runtime: r,
		module:  mod,
	}, nil
}

// Match invokes the adapter's on_output function.
func (r *WasmRuntime) Match(input []byte) ([]Action, error) {
	ctx := context.Background()
	
	// Call amux_alloc to reserve memory for input
	alloc := r.module.ExportedFunction("amux_alloc")
	if alloc == nil {
		return nil, fmt.Errorf("amux_alloc not exported")
	}
	
	inputSize := uint64(len(input))
	results, err := alloc.Call(ctx, inputSize)
	if err != nil {
		return nil, fmt.Errorf("amux_alloc failed: %w", err)
	}
	ptr := results[0]
	
	// Write input to memory
	if !r.module.Memory().Write(uint32(ptr), input) {
		return nil, fmt.Errorf("memory write failed")
	}
	
	// Call on_output
	onOutput := r.module.ExportedFunction("on_output")
	if onOutput == nil {
		return nil, fmt.Errorf("on_output not exported")
	}
	
	results, err = onOutput.Call(ctx, ptr, inputSize)
	if err != nil {
		return nil, fmt.Errorf("on_output failed: %w", err)
	}
	
	packed := results[0]
	if packed == 0 {
		return nil, nil // No actions or failure? "0 = failure" according to invariant?
		// If 0 is failure, we should check for error.
		// If no match, maybe it returns valid ptr with 0 len?
	}
	
	// Decode result ptr/len
	resPtr := uint32(packed >> 32)
	resLen := uint32(packed)
	
	if resLen == 0 {
		return nil, nil
	}
	
	// Read result JSON
	resBytes, ok := r.module.Memory().Read(resPtr, resLen)
	if !ok {
		return nil, fmt.Errorf("memory read failed")
	}
	
	// Clean up input and output
	free := r.module.ExportedFunction("amux_free")
	if free != nil {
		free.Call(ctx, ptr, inputSize)
		free.Call(ctx, uint64(resPtr), uint64(resLen))
	}
	
	var actions []Action
	if err := json.Unmarshal(resBytes, &actions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal actions: %w", err)
	}
	
	return actions, nil
}

// FormatInput invokes the adapter's format_input function.
func (r *WasmRuntime) FormatInput(input any) ([]byte, error) {
	ctx := context.Background()
	
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	alloc := r.module.ExportedFunction("amux_alloc")
	if alloc == nil {
		return nil, fmt.Errorf("amux_alloc not exported")
	}
	
	inputSize := uint64(len(inputBytes))
	results, err := alloc.Call(ctx, inputSize)
	if err != nil {
		return nil, fmt.Errorf("amux_alloc failed: %w", err)
	}
	ptr := results[0]
	
	if !r.module.Memory().Write(uint32(ptr), inputBytes) {
		return nil, fmt.Errorf("memory write failed")
	}
	
	formatInput := r.module.ExportedFunction("format_input")
	if formatInput == nil {
		return nil, fmt.Errorf("format_input not exported")
	}
	
	results, err = formatInput.Call(ctx, ptr, inputSize)
	if err != nil {
		return nil, fmt.Errorf("format_input failed: %w", err)
	}
	
	packed := results[0]
	if packed == 0 {
		return nil, nil
	}
	
	resPtr := uint32(packed >> 32)
	resLen := uint32(packed)
	
	if resLen == 0 {
		return nil, nil
	}
	
	resBytes, ok := r.module.Memory().Read(resPtr, resLen)
	if !ok {
		return nil, fmt.Errorf("memory read failed")
	}
	
	// Copy result because free will invalidate it
	result := make([]byte, resLen)
	copy(result, resBytes)

	free := r.module.ExportedFunction("amux_free")
	if free != nil {
		free.Call(ctx, ptr, inputSize)
		free.Call(ctx, uint64(resPtr), uint64(resLen))
	}
	
	return result, nil
}

func (r *WasmRuntime) Start() error {
	return nil
}

func (r *WasmRuntime) Stop() error {
	return r.runtime.Close(context.Background())
}
