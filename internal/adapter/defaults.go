package adapter

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// DefaultsProvider loads adapter default configuration from WASM modules.
type DefaultsProvider struct {
	resolver *paths.Resolver
}

// NewDefaultsProvider constructs a DefaultsProvider.
func NewDefaultsProvider(resolver *paths.Resolver) *DefaultsProvider {
	return &DefaultsProvider{resolver: resolver}
}

// AdapterDefaults returns default TOML blocks for discovered adapters.
func (p *DefaultsProvider) AdapterDefaults() ([]config.AdapterDefault, error) {
	if p == nil || p.resolver == nil {
		return nil, fmt.Errorf("adapter defaults: resolver is required")
	}
	ctx := context.Background()
	runtime := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().WithMemoryLimitPages(wasmMemoryLimitPages))
	defer func() {
		_ = runtime.Close(ctx)
	}()
	modules, err := discoverAdapterModules(p.resolver)
	if err != nil {
		return nil, err
	}
	defaults := make([]config.AdapterDefault, 0, len(modules))
	for _, module := range modules {
		data, source, err := loadAdapterDefaults(ctx, runtime, module)
		if err != nil {
			return nil, err
		}
		if len(data) == 0 {
			continue
		}
		if !utf8.Valid(data) {
			return nil, fmt.Errorf("adapter defaults: %s: invalid utf-8", module.name)
		}
		defaults = append(defaults, config.AdapterDefault{
			Name:   module.name,
			Source: source,
			Data:   data,
		})
	}
	return defaults, nil
}

type adapterModule struct {
	name   string
	path   string
	source string
}

func discoverAdapterModules(resolver *paths.Resolver) ([]adapterModule, error) {
	if resolver == nil {
		return nil, fmt.Errorf("discover adapters: resolver is nil")
	}
	seen := make(map[string]struct{})
	var modules []adapterModule
	projectDir := resolver.ProjectAdaptersDir()
	projectModules, err := scanAdapterDir(projectDir, "project", seen)
	if err != nil {
		return nil, err
	}
	modules = append(modules, projectModules...)
	userDir := resolver.UserAdaptersDir()
	userModules, err := scanAdapterDir(userDir, "user", seen)
	if err != nil {
		return nil, err
	}
	modules = append(modules, userModules...)
	return modules, nil
}

func scanAdapterDir(dir string, label string, seen map[string]struct{}) ([]adapterModule, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("discover adapters: %w", err)
	}
	modules := make([]adapterModule, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if _, ok := seen[name]; ok {
			continue
		}
		wasmPath := filepath.Join(dir, name, name+".wasm")
		if _, err := os.Stat(wasmPath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("discover adapters: %w", err)
		}
		seen[name] = struct{}{}
		modules = append(modules, adapterModule{name: name, path: wasmPath, source: label})
	}
	return modules, nil
}

func loadAdapterDefaults(ctx context.Context, runtime wazero.Runtime, module adapterModule) ([]byte, string, error) {
	if runtime == nil {
		return nil, "", fmt.Errorf("adapter defaults: runtime is nil")
	}
	wasmBytes, err := os.ReadFile(module.path)
	if err != nil {
		return nil, "", fmt.Errorf("adapter defaults: %s: %w", module.name, err)
	}
	compiled, err := runtime.CompileModule(ctx, wasmBytes)
	if err != nil {
		return nil, "", fmt.Errorf("adapter defaults: %s: %w", module.name, err)
	}
	instance, err := runtime.InstantiateModule(ctx, compiled, wazero.NewModuleConfig().WithName(module.name))
	if err != nil {
		return nil, "", fmt.Errorf("adapter defaults: %s: %w", module.name, err)
	}
	defer func() {
		_ = instance.Close(ctx)
	}()
	configFn := instance.ExportedFunction("config_default")
	if configFn != nil {
		freeFn := instance.ExportedFunction("amux_free")
		data, err := callConfigDefault(ctx, instance, freeFn, configFn)
		if err != nil {
			return nil, "", fmt.Errorf("adapter defaults: %s: %w", module.name, err)
		}
		return data, "config_default", nil
	}
	fallback := filepath.Join(filepath.Dir(module.path), "config.default.toml")
	data, err := os.ReadFile(fallback)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, "", nil
		}
		return nil, "", fmt.Errorf("adapter defaults: %s: %w", module.name, err)
	}
	return data, fallback, nil
}

func callConfigDefault(ctx context.Context, module api.Module, freeFn api.Function, configFn api.Function) ([]byte, error) {
	if module == nil || configFn == nil {
		return nil, fmt.Errorf("config_default: missing function")
	}
	results, err := configFn.Call(ctx)
	if err != nil {
		return nil, fmt.Errorf("config_default: %w", err)
	}
	packed, ok := firstResult(results)
	if !ok || packed == 0 {
		return nil, fmt.Errorf("config_default: %w", ErrAdapterExecutionFailed)
	}
	ptr := uint32(packed >> 32)
	length := uint32(packed)
	if length == 0 {
		return nil, nil
	}
	memory := module.Memory()
	if memory == nil {
		return nil, fmt.Errorf("config_default: %w", ErrAdapterExecutionFailed)
	}
	buf, ok := memory.Read(ptr, length)
	if !ok {
		return nil, fmt.Errorf("config_default: %w", ErrAdapterExecutionFailed)
	}
	out := append([]byte(nil), buf...)
	if freeFn == nil {
		return nil, fmt.Errorf("config_default: %w", ErrAdapterMissingExport)
	}
	if _, err := freeFn.Call(ctx, uint64(ptr), uint64(length)); err != nil {
		return nil, fmt.Errorf("config_default: %w", err)
	}
	return out, nil
}

func firstResult(results []uint64) (uint64, bool) {
	if len(results) == 0 {
		return 0, false
	}
	return results[0], true
}
