package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

const wasmMemoryLimitPages = 4096

var (
	// ErrAdapterInvalid is returned when adapter inputs are invalid.
	ErrAdapterInvalid = errors.New("adapter invalid")
	// ErrAdapterMissingExport is returned when required exports are missing.
	ErrAdapterMissingExport = errors.New("adapter missing export")
	// ErrAdapterManifestMismatch is returned when the manifest name mismatches the requested name.
	ErrAdapterManifestMismatch = errors.New("adapter manifest mismatch")
	// ErrAdapterExecutionFailed is returned when a WASM call fails.
	ErrAdapterExecutionFailed = errors.New("adapter execution failed")
)

// Manifest describes the minimal adapter manifest fields needed by the runtime.
type Manifest struct {
	Name string `json:"name"`
}

// WazeroRegistry loads adapters from WASM modules using wazero.
type WazeroRegistry struct {
	resolver *paths.Resolver
	runtime  wazero.Runtime
	mu       sync.Mutex
	compiled map[string]wazero.CompiledModule
}

// NewWazeroRegistry constructs a registry with a wazero runtime.
func NewWazeroRegistry(ctx context.Context, resolver *paths.Resolver) (*WazeroRegistry, error) {
	if resolver == nil {
		return nil, fmt.Errorf("new wazero registry: %w", ErrAdapterInvalid)
	}
	config := wazero.NewRuntimeConfig().WithMemoryLimitPages(wasmMemoryLimitPages)
	runtime := wazero.NewRuntimeWithConfig(ctx, config)
	return &WazeroRegistry{
		resolver: resolver,
		runtime:  runtime,
		compiled: make(map[string]wazero.CompiledModule),
	}, nil
}

// Close releases the wazero runtime.
func (r *WazeroRegistry) Close(ctx context.Context) error {
	if r == nil || r.runtime == nil {
		return nil
	}
	if err := r.runtime.Close(ctx); err != nil {
		return fmt.Errorf("close wazero registry: %w", err)
	}
	return nil
}

// Load loads an adapter by name from the WASM registry.
func (r *WazeroRegistry) Load(ctx context.Context, name string) (Adapter, error) {
	if r == nil || r.runtime == nil {
		return nil, fmt.Errorf("load adapter: %w", ErrAdapterInvalid)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("load adapter: %w", ErrAdapterInvalid)
	}
	wasmPath, wasmBytes, err := r.findModule(name)
	if err != nil {
		return nil, fmt.Errorf("load adapter %s: %w", name, err)
	}
	compiled, err := r.compile(ctx, wasmPath, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("load adapter %s: %w", name, err)
	}
	module, err := r.runtime.InstantiateModule(ctx, compiled, wazero.NewModuleConfig().WithName(name))
	if err != nil {
		return nil, fmt.Errorf("load adapter %s: %w", name, err)
	}
	adapter, err := newWasmAdapter(name, module)
	if err != nil {
		_ = module.Close(ctx)
		return nil, fmt.Errorf("load adapter %s: %w", name, err)
	}
	manifest, err := adapter.manifest(ctx)
	if err != nil {
		_ = module.Close(ctx)
		return nil, fmt.Errorf("load adapter %s: %w", name, err)
	}
	if manifest.Name != "" && manifest.Name != name {
		_ = module.Close(ctx)
		return nil, fmt.Errorf("load adapter %s: %w", name, ErrAdapterManifestMismatch)
	}
	return adapter, nil
}

func (r *WazeroRegistry) compile(ctx context.Context, path string, wasmBytes []byte) (wazero.CompiledModule, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if compiled, ok := r.compiled[path]; ok {
		return compiled, nil
	}
	compiled, err := r.runtime.CompileModule(ctx, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("compile module: %w", err)
	}
	r.compiled[path] = compiled
	return compiled, nil
}

func (r *WazeroRegistry) findModule(name string) (string, []byte, error) {
	pathsToCheck := []string{
		r.resolver.ProjectAdapterWasmPath(name),
		r.resolver.UserAdapterWasmPath(name),
	}
	for _, path := range pathsToCheck {
		data, err := os.ReadFile(path)
		if err == nil {
			return path, data, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", nil, fmt.Errorf("read wasm: %w", err)
		}
	}
	return "", nil, ErrAdapterNotFound
}

type wasmAdapter struct {
	name       string
	module     api.Module
	memory     api.Memory
	alloc      api.Function
	free       api.Function
	manifestFn api.Function
	onOutputFn api.Function
	formatFn   api.Function
	onEventFn  api.Function
	mu         sync.Mutex
}

func newWasmAdapter(name string, module api.Module) (*wasmAdapter, error) {
	if module == nil {
		return nil, fmt.Errorf("new wasm adapter: %w", ErrAdapterInvalid)
	}
	memory := module.Memory()
	if memory == nil {
		return nil, fmt.Errorf("new wasm adapter: %w", ErrAdapterMissingExport)
	}
	alloc := module.ExportedFunction("amux_alloc")
	free := module.ExportedFunction("amux_free")
	manifestFn := module.ExportedFunction("manifest")
	onOutputFn := module.ExportedFunction("on_output")
	formatFn := module.ExportedFunction("format_input")
	onEventFn := module.ExportedFunction("on_event")
	if alloc == nil || free == nil || manifestFn == nil || onOutputFn == nil || formatFn == nil || onEventFn == nil {
		return nil, fmt.Errorf("new wasm adapter: %w", ErrAdapterMissingExport)
	}
	return &wasmAdapter{
		name:       name,
		module:     module,
		memory:     memory,
		alloc:      alloc,
		free:       free,
		manifestFn: manifestFn,
		onOutputFn: onOutputFn,
		formatFn:   formatFn,
		onEventFn:  onEventFn,
	}, nil
}

func (w *wasmAdapter) Name() string {
	return w.name
}

func (w *wasmAdapter) Matcher() PatternMatcher {
	return &wasmMatcher{adapter: w}
}

func (w *wasmAdapter) Formatter() ActionFormatter {
	return &wasmFormatter{adapter: w}
}

func (w *wasmAdapter) manifest(ctx context.Context) (Manifest, error) {
	raw, err := w.callNoInput(ctx, w.manifestFn)
	if err != nil {
		return Manifest{}, fmt.Errorf("manifest: %w", err)
	}
	var manifest Manifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("manifest: %w", err)
	}
	return manifest, nil
}

type wasmMatcher struct {
	adapter *wasmAdapter
}

func (m *wasmMatcher) Match(ctx context.Context, output []byte) ([]PatternMatch, error) {
	if m == nil || m.adapter == nil {
		return nil, fmt.Errorf("match output: %w", ErrAdapterInvalid)
	}
	raw, err := m.adapter.callWithInput(ctx, m.adapter.onOutputFn, output)
	if err != nil {
		return nil, fmt.Errorf("match output: %w", err)
	}
	if len(raw) == 0 {
		return nil, nil
	}
	var matches []PatternMatch
	if err := json.Unmarshal(raw, &matches); err != nil {
		return nil, fmt.Errorf("match output: %w", err)
	}
	return matches, nil
}

type wasmFormatter struct {
	adapter *wasmAdapter
}

func (f *wasmFormatter) Format(ctx context.Context, input string) (string, error) {
	if f == nil || f.adapter == nil {
		return "", fmt.Errorf("format input: %w", ErrAdapterInvalid)
	}
	raw, err := f.adapter.callWithInput(ctx, f.adapter.formatFn, []byte(input))
	if err != nil {
		return "", fmt.Errorf("format input: %w", err)
	}
	return string(raw), nil
}

func (w *wasmAdapter) callNoInput(ctx context.Context, fn api.Function) ([]byte, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	results, err := fn.Call(ctx)
	if err != nil {
		return nil, fmt.Errorf("call wasm: %w", err)
	}
	return w.readPacked(ctx, results)
}

func (w *wasmAdapter) callWithInput(ctx context.Context, fn api.Function, input []byte) ([]byte, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	ptr, length, err := w.writeInput(ctx, input)
	if err != nil {
		return nil, err
	}
	results, callErr := fn.Call(ctx, uint64(ptr), uint64(length))
	freeErr := w.freeBuffer(ctx, ptr, length)
	if callErr != nil {
		if freeErr != nil {
			return nil, fmt.Errorf("call wasm: %w", errors.Join(callErr, freeErr))
		}
		return nil, fmt.Errorf("call wasm: %w", callErr)
	}
	if freeErr != nil {
		return nil, fmt.Errorf("call wasm: %w", freeErr)
	}
	return w.readPacked(ctx, results)
}

func (w *wasmAdapter) readPacked(ctx context.Context, results []uint64) ([]byte, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("call wasm: %w", ErrAdapterExecutionFailed)
	}
	packed := results[0]
	if packed == 0 {
		return nil, fmt.Errorf("call wasm: %w", ErrAdapterExecutionFailed)
	}
	ptr := uint32(packed >> 32)
	length := uint32(packed)
	if length == 0 {
		return nil, nil
	}
	buf, ok := w.memory.Read(ptr, length)
	if !ok {
		return nil, fmt.Errorf("call wasm: %w", ErrAdapterExecutionFailed)
	}
	out := append([]byte(nil), buf...)
	if err := w.freeBuffer(ctx, ptr, length); err != nil {
		return nil, fmt.Errorf("call wasm: %w", err)
	}
	return out, nil
}

func (w *wasmAdapter) writeInput(ctx context.Context, input []byte) (uint32, uint32, error) {
	if len(input) == 0 {
		return 0, 0, nil
	}
	results, err := w.alloc.Call(ctx, uint64(len(input)))
	if err != nil {
		return 0, 0, fmt.Errorf("alloc buffer: %w", err)
	}
	if len(results) == 0 {
		return 0, 0, fmt.Errorf("alloc buffer: %w", ErrAdapterExecutionFailed)
	}
	ptr := uint32(results[0])
	if ptr == 0 {
		return 0, 0, fmt.Errorf("alloc buffer: %w", ErrAdapterExecutionFailed)
	}
	buf, ok := w.memory.Read(ptr, uint32(len(input)))
	if !ok {
		return 0, 0, fmt.Errorf("alloc buffer: %w", ErrAdapterExecutionFailed)
	}
	copy(buf, input)
	return ptr, uint32(len(input)), nil
}

func (w *wasmAdapter) freeBuffer(ctx context.Context, ptr uint32, length uint32) error {
	if ptr == 0 && length == 0 {
		return nil
	}
	_, err := w.free.Call(ctx, uint64(ptr), uint64(length))
	if err != nil {
		return fmt.Errorf("free buffer: %w", err)
	}
	return nil
}
