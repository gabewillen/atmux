// Package adapter provides WASM adapter runtime and loading functionality.
// This package loads conforming WASM adapters without any knowledge of
// specific agent implementations.
//
// The adapter system provides a pluggable interface enabling any coding agent
// to be integrated through a WASM adapter that implements the required ABI.
package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// Common sentinel errors for adapter operations.
var (
	// ErrAdapterNotFound indicates the requested adapter was not found.
	ErrAdapterNotFound = errors.New("adapter not found")

	// ErrInvalidABI indicates the adapter does not implement the required ABI.
	ErrInvalidABI = errors.New("invalid adapter ABI")

	// ErrRuntimeFailed indicates a WASM runtime failure.
	ErrRuntimeFailed = errors.New("WASM runtime failure")
)

// AdapterManifest represents adapter metadata per spec §10.2
type AdapterManifest struct {
	Name        string          `json:"name"`
	Version     string          `json:"version"`
	Description string          `json:"description,omitempty"`
	CLI         CLIRequirement  `json:"cli"`
	Patterns    AdapterPatterns `json:"patterns"`
	Commands    AdapterCommands `json:"commands"`
}

// CLIRequirement defines CLI version constraints per spec §10.3
type CLIRequirement struct {
	Binary     string `json:"binary"`
	VersionCmd string `json:"version_cmd"`
	VersionRe  string `json:"version_re"`
	Constraint string `json:"constraint"`
}

// AdapterPatterns defines output patterns for monitoring
type AdapterPatterns struct {
	Ready    string `json:"ready"`
	Error    string `json:"error"`
	Complete string `json:"complete"`
}

// AdapterCommands defines commands to interact with the agent
type AdapterCommands struct {
	Start       []string `json:"start"`
	SendMessage string   `json:"send_message"`
}

// AdapterInstance represents a loaded WASM adapter instance
type AdapterInstance struct {
	module   api.Module
	manifest AdapterManifest
	name     string
}

// Runtime manages WASM adapter instances using wazero.
// One WASM instance per agent with 256MB memory cap.
type Runtime struct {
	ctx       context.Context
	engine    wazero.Runtime
	instances map[string]*AdapterInstance
}

// NewRuntime creates a new WASM adapter runtime.
func NewRuntime(ctx context.Context) (*Runtime, error) {
	engine := wazero.NewRuntime(ctx)
	
	// Instantiate WASI for adapters that need it
	_, err := wasi_snapshot_preview1.Instantiate(ctx, engine)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate WASI: %w", err)
	}
	
	return &Runtime{
		ctx:       ctx,
		engine:    engine,
		instances: make(map[string]*AdapterInstance),
	}, nil
}

// LoadAdapter loads a WASM adapter from the given path.
// Returns an error if the adapter doesn't implement required exports:
// amux_alloc, amux_free, manifest, on_output, format_input, on_event
func (r *Runtime) LoadAdapter(name, path string) (*AdapterInstance, error) {
	if path == "" {
		return nil, fmt.Errorf("adapter path required: %w", ErrAdapterNotFound)
	}
	
	// Check if already loaded
	if instance, exists := r.instances[name]; exists {
		return instance, nil
	}
	
	// Read WASM file
	wasmBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read adapter %s: %w", path, ErrAdapterNotFound)
	}
	
	// Compile WASM module
	compiledModule, err := r.engine.CompileModule(r.ctx, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to compile adapter %s: %w", path, ErrRuntimeFailed)
	}
	
	// Configure module with memory limits per spec (256MB cap)
	config := wazero.NewModuleConfig().WithName(name)
	
	// Instantiate module
	module, err := r.engine.InstantiateModule(r.ctx, compiledModule, config)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate adapter %s: %w", path, ErrRuntimeFailed)
	}
	
	// Verify required exports per spec §10.4.2
	requiredExports := []string{"amux_alloc", "amux_free", "manifest", "on_output", "format_input", "on_event"}
	for _, export := range requiredExports {
		if module.ExportedFunction(export) == nil {
			return nil, fmt.Errorf("adapter %s missing required export %s: %w", path, export, ErrInvalidABI)
		}
	}
	
	// Get manifest from adapter
	manifestFunc := module.ExportedFunction("manifest")
	if manifestFunc == nil {
		return nil, fmt.Errorf("adapter %s missing manifest function: %w", path, ErrInvalidABI)
	}
	
	// Call manifest function
	results, err := manifestFunc.Call(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to call manifest function: %w", err)
	}
	
	if len(results) != 1 {
		return nil, fmt.Errorf("manifest function returned unexpected results: %w", ErrInvalidABI)
	}
	
	// Unpack result per spec §10.4.1
	packed := results[0]
	ptr := uint32(packed >> 32)
	length := uint32(packed & 0xFFFFFFFF)
	
	var manifest AdapterManifest
	if length > 0 {
		// Read manifest JSON from WASM memory
		memory := module.Memory()
		manifestBytes, ok := memory.Read(ptr, length)
		if !ok {
			return nil, fmt.Errorf("failed to read manifest from WASM memory: %w", ErrRuntimeFailed)
		}
		
		// Free the allocated memory
		freeFunc := module.ExportedFunction("amux_free")
		if freeFunc != nil {
			freeFunc.Call(r.ctx, api.EncodeU32(ptr), api.EncodeU32(length))
		}
		
		if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
			return nil, fmt.Errorf("failed to parse adapter manifest: %w", err)
		}
	}
	
	instance := &AdapterInstance{
		module:   module,
		manifest: manifest,
		name:     name,
	}
	
	r.instances[name] = instance
	return instance, nil
}

// GetInstance returns a loaded adapter instance by name
func (r *Runtime) GetInstance(name string) (*AdapterInstance, error) {
	instance, exists := r.instances[name]
	if !exists {
		return nil, fmt.Errorf("adapter %s not loaded: %w", name, ErrAdapterNotFound)
	}
	return instance, nil
}

// ListInstances returns all loaded adapter instances
func (r *Runtime) ListInstances() []*AdapterInstance {
	instances := make([]*AdapterInstance, 0, len(r.instances))
	for _, instance := range r.instances {
		instances = append(instances, instance)
	}
	return instances
}

// LoadAdaptersFromDirectory loads all WASM adapters from the given directory
func (r *Runtime) LoadAdaptersFromDirectory(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read adapter directory %s: %w", dir, err)
	}
	
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".wasm" {
			continue
		}
		
		name := strings.TrimSuffix(entry.Name(), ".wasm") // Remove .wasm extension
		path := filepath.Join(dir, entry.Name())
		
		if _, err := r.LoadAdapter(name, path); err != nil {
			// Log error but continue loading other adapters
			fmt.Printf("Warning: failed to load adapter %s: %v\n", name, err)
		}
	}
	
	return nil
}

// Close releases runtime resources.
func (r *Runtime) Close() error {
	// Close all instances
	for _, instance := range r.instances {
		if err := instance.module.Close(r.ctx); err != nil {
			return fmt.Errorf("failed to close adapter instance: %w", err)
		}
	}
	
	return r.engine.Close(r.ctx)
}

// AdapterInstance methods

// GetManifest returns the adapter's manifest
func (i *AdapterInstance) GetManifest() AdapterManifest {
	return i.manifest
}

// GetName returns the adapter's name
func (i *AdapterInstance) GetName() string {
	return i.name
}

// ProcessOutput processes PTY output through the adapter's on_output function
func (i *AdapterInstance) ProcessOutput(ctx context.Context, output []byte) ([]byte, error) {
	if len(output) == 0 {
		return nil, nil
	}
	
	// Allocate memory in WASM for input
	allocFunc := i.module.ExportedFunction("amux_alloc")
	if allocFunc == nil {
		return nil, fmt.Errorf("adapter missing amux_alloc function: %w", ErrInvalidABI)
	}
	
	results, err := allocFunc.Call(ctx, api.EncodeU32(uint32(len(output))))
	if err != nil {
		return nil, fmt.Errorf("failed to allocate memory in adapter: %w", err)
	}
	
	ptr := uint32(results[0])
	if ptr == 0 {
		return nil, fmt.Errorf("adapter failed to allocate memory: %w", ErrRuntimeFailed)
	}
	
	// Write input to WASM memory
	memory := i.module.Memory()
	if !memory.Write(ptr, output) {
		return nil, fmt.Errorf("failed to write input to WASM memory: %w", ErrRuntimeFailed)
	}
	
	// Call on_output function
	onOutputFunc := i.module.ExportedFunction("on_output")
	if onOutputFunc == nil {
		return nil, fmt.Errorf("adapter missing on_output function: %w", ErrInvalidABI)
	}
	
	results, err = onOutputFunc.Call(ctx, api.EncodeU32(ptr), api.EncodeU32(uint32(len(output))))
	if err != nil {
		return nil, fmt.Errorf("failed to call on_output: %w", err)
	}
	
	// Free input memory
	freeFunc := i.module.ExportedFunction("amux_free")
	if freeFunc != nil {
		freeFunc.Call(ctx, api.EncodeU32(ptr), api.EncodeU32(uint32(len(output))))
	}
	
	// Unpack result
	packed := results[0]
	resultPtr := uint32(packed >> 32)
	resultLen := uint32(packed & 0xFFFFFFFF)
	
	if resultLen == 0 {
		return nil, nil
	}
	
	// Read result from WASM memory
	resultBytes, ok := memory.Read(resultPtr, resultLen)
	if !ok {
		return nil, fmt.Errorf("failed to read result from WASM memory: %w", ErrRuntimeFailed)
	}
	
	// Free result memory
	if freeFunc != nil {
		freeFunc.Call(ctx, api.EncodeU32(resultPtr), api.EncodeU32(resultLen))
	}
	
	return resultBytes, nil
}

// FormatInput formats input through the adapter's format_input function
func (i *AdapterInstance) FormatInput(ctx context.Context, input []byte) ([]byte, error) {
	if len(input) == 0 {
		return input, nil
	}
	
	// Similar implementation to ProcessOutput but calls format_input
	allocFunc := i.module.ExportedFunction("amux_alloc")
	results, err := allocFunc.Call(ctx, api.EncodeU32(uint32(len(input))))
	if err != nil {
		return nil, fmt.Errorf("failed to allocate memory for format_input: %w", err)
	}
	
	ptr := uint32(results[0])
	if ptr == 0 {
		return nil, fmt.Errorf("adapter failed to allocate memory for format_input: %w", ErrRuntimeFailed)
	}
	
	memory := i.module.Memory()
	if !memory.Write(ptr, input) {
		return nil, fmt.Errorf("failed to write input to WASM memory for format_input: %w", ErrRuntimeFailed)
	}
	
	formatInputFunc := i.module.ExportedFunction("format_input")
	if formatInputFunc == nil {
		return input, nil // Optional function
	}
	
	results, err = formatInputFunc.Call(ctx, api.EncodeU32(ptr), api.EncodeU32(uint32(len(input))))
	if err != nil {
		return nil, fmt.Errorf("failed to call format_input: %w", err)
	}
	
	// Free input memory
	freeFunc := i.module.ExportedFunction("amux_free")
	if freeFunc != nil {
		freeFunc.Call(ctx, api.EncodeU32(ptr), api.EncodeU32(uint32(len(input))))
	}
	
	// Unpack and read result
	packed := results[0]
	resultPtr := uint32(packed >> 32)
	resultLen := uint32(packed & 0xFFFFFFFF)
	
	if resultLen == 0 {
		return input, nil // Return original input if no formatting
	}
	
	resultBytes, ok := memory.Read(resultPtr, resultLen)
	if !ok {
		return nil, fmt.Errorf("failed to read formatted input from WASM memory: %w", ErrRuntimeFailed)
	}
	
	if freeFunc != nil {
		freeFunc.Call(ctx, api.EncodeU32(resultPtr), api.EncodeU32(resultLen))
	}
	
	return resultBytes, nil
}