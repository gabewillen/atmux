package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/agentflare-ai/amux/internal/paths"
)

func TestWazeroRegistryLoadAndAdapterCalls(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repoRoot := t.TempDir()
	if err := os.Mkdir(filepath.Join(repoRoot, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	resolver, err := paths.NewResolver(repoRoot)
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}
	if err := os.MkdirAll(resolver.ProjectAdaptersDir(), 0o755); err != nil {
		t.Fatalf("mkdir adapters: %v", err)
	}
	manifest := Manifest{Name: "test"}
	matches := []PatternMatch{{Pattern: "prompt", Text: "ready"}}
	actions := []Action{{Type: "notify"}}
	format := "formatted"
	wasm := buildAdapterWasm(t, manifest, matches, format, actions)
	path := resolver.ProjectAdapterWasmPath("test")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir adapter dir: %v", err)
	}
	if err := os.WriteFile(path, wasm, 0o644); err != nil {
		t.Fatalf("write wasm: %v", err)
	}
	registry, err := NewWazeroRegistry(ctx, resolver)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	t.Cleanup(func() {
		_ = registry.Close(ctx)
	})
	adapter, err := registry.Load(ctx, "test")
	if err != nil {
		t.Fatalf("load adapter: %v", err)
	}
	if adapter.Name() != "test" {
		t.Fatalf("unexpected adapter name")
	}
	if got := adapter.Manifest().Name; got != "test" {
		t.Fatalf("unexpected manifest name: %s", got)
	}
	found, err := adapter.Matcher().Match(ctx, []byte("ready?"))
	if err != nil {
		t.Fatalf("match: %v", err)
	}
	if len(found) != 1 || found[0].Pattern != "prompt" {
		t.Fatalf("unexpected match: %#v", found)
	}
	out, err := adapter.Formatter().Format(ctx, "ping")
	if err != nil {
		t.Fatalf("format: %v", err)
	}
	if out != format {
		t.Fatalf("unexpected format: %s", out)
	}
	actionsOut, err := adapter.OnEvent(ctx, Event{Type: "ping"})
	if err != nil {
		t.Fatalf("event: %v", err)
	}
	if len(actionsOut) != 1 || actionsOut[0].Type != "notify" {
		t.Fatalf("unexpected actions: %#v", actionsOut)
	}
}

func TestWazeroRegistryManifestMismatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repoRoot := t.TempDir()
	if err := os.Mkdir(filepath.Join(repoRoot, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	resolver, err := paths.NewResolver(repoRoot)
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}
	if err := os.MkdirAll(resolver.ProjectAdaptersDir(), 0o755); err != nil {
		t.Fatalf("mkdir adapters: %v", err)
	}
	wasm := buildAdapterWasm(t, Manifest{Name: "wrong"}, nil, "", nil)
	path := resolver.ProjectAdapterWasmPath("test")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir adapter dir: %v", err)
	}
	if err := os.WriteFile(path, wasm, 0o644); err != nil {
		t.Fatalf("write wasm: %v", err)
	}
	registry, err := NewWazeroRegistry(ctx, resolver)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	t.Cleanup(func() {
		_ = registry.Close(ctx)
	})
	if _, err := registry.Load(ctx, "test"); err == nil || !errors.Is(err, ErrAdapterManifestMismatch) {
		t.Fatalf("expected manifest mismatch, got %v", err)
	}
}

func TestWazeroRegistryLoadInvalid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repoRoot := t.TempDir()
	if err := os.Mkdir(filepath.Join(repoRoot, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	resolver, err := paths.NewResolver(repoRoot)
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}
	registry, err := NewWazeroRegistry(ctx, resolver)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	t.Cleanup(func() {
		_ = registry.Close(ctx)
	})
	if _, err := registry.Load(ctx, ""); err == nil {
		t.Fatalf("expected error for empty name")
	}
	if err := registry.Close(ctx); err != nil {
		t.Fatalf("close: %v", err)
	}
}

func buildAdapterWasm(t *testing.T, manifest Manifest, matches []PatternMatch, format string, actions []Action) []byte {
	t.Helper()
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	matchesBytes, err := json.Marshal(matches)
	if err != nil {
		t.Fatalf("marshal matches: %v", err)
	}
	actionsBytes, err := json.Marshal(actions)
	if err != nil {
		t.Fatalf("marshal actions: %v", err)
	}
	formatBytes := []byte(format)
	const (
		manifestOffset = 0
		matchesOffset  = 256
		formatOffset   = 512
		actionsOffset  = 768
		allocPtr       = 1024
	)
	return buildWasmModule(
		manifestBytes, manifestOffset,
		matchesBytes, matchesOffset,
		formatBytes, formatOffset,
		actionsBytes, actionsOffset,
		allocPtr,
	)
}

func buildWasmModule(
	manifest []byte, manifestOffset uint32,
	matches []byte, matchesOffset uint32,
	format []byte, formatOffset uint32,
	actions []byte, actionsOffset uint32,
	allocPtr uint32,
) []byte {
	packedManifest := packPtrLen(manifestOffset, uint32(len(manifest)))
	packedMatches := packPtrLen(matchesOffset, uint32(len(matches)))
	packedFormat := packPtrLen(formatOffset, uint32(len(format)))
	packedActions := packPtrLen(actionsOffset, uint32(len(actions)))

	types := appendU32(nil, 4)
	types = append(types, funcType([]byte{0x7f}, []byte{0x7f})...)           // (i32) -> i32
	types = append(types, funcType([]byte{0x7f, 0x7f}, nil)...)              // (i32,i32) -> ()
	types = append(types, funcType(nil, []byte{0x7e})...)                    // () -> i64
	types = append(types, funcType([]byte{0x7f, 0x7f}, []byte{0x7e})...)      // (i32,i32) -> i64
	functions := appendU32(nil, 6)
	functions = append(functions, appendU32(nil, 0)...)
	functions = append(functions, appendU32(nil, 1)...)
	functions = append(functions, appendU32(nil, 2)...)
	functions = append(functions, appendU32(nil, 3)...)
	functions = append(functions, appendU32(nil, 3)...)
	functions = append(functions, appendU32(nil, 3)...)

	mem := []byte{0x01, 0x00, 0x01}

	exports := appendU32(nil, 7)
	exports = append(exports, exportEntry("memory", 0x02, 0)...)
	exports = append(exports, exportEntry("amux_alloc", 0x00, 0)...)
	exports = append(exports, exportEntry("amux_free", 0x00, 1)...)
	exports = append(exports, exportEntry("manifest", 0x00, 2)...)
	exports = append(exports, exportEntry("on_output", 0x00, 3)...)
	exports = append(exports, exportEntry("format_input", 0x00, 4)...)
	exports = append(exports, exportEntry("on_event", 0x00, 5)...)

	code := appendU32(nil, 6)
	code = append(code, funcBody(i32Const(allocPtr), 0)...)
	code = append(code, funcBody([]byte{0x0b}, 0)...)
	code = append(code, funcBody(i64Const(packedManifest), 0)...)
	code = append(code, funcBody(i64Const(packedMatches), 0)...)
	code = append(code, funcBody(i64Const(packedFormat), 0)...)
	code = append(code, funcBody(i64Const(packedActions), 0)...)

	data := appendU32(nil, 4)
	data = append(data, dataEntry(manifestOffset, manifest)...)
	data = append(data, dataEntry(matchesOffset, matches)...)
	data = append(data, dataEntry(formatOffset, format)...)
	data = append(data, dataEntry(actionsOffset, actions)...)

	out := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	out = append(out, section(1, types)...)
	out = append(out, section(3, functions)...)
	out = append(out, section(5, mem)...)
	out = append(out, section(7, exports)...)
	out = append(out, section(10, code)...)
	out = append(out, section(11, data)...)
	return out
}

func packPtrLen(ptr uint32, length uint32) uint64 {
	return (uint64(ptr) << 32) | uint64(length)
}

func funcType(params, results []byte) []byte {
	out := []byte{0x60}
	out = appendU32(out, uint32(len(params)))
	out = append(out, params...)
	out = appendU32(out, uint32(len(results)))
	out = append(out, results...)
	return out
}

func exportEntry(name string, kind byte, index uint32) []byte {
	out := appendU32(nil, uint32(len(name)))
	out = append(out, name...)
	out = append(out, kind)
	out = appendU32(out, index)
	return out
}

func funcBody(code []byte, localCount byte) []byte {
	body := []byte{localCount}
	body = append(body, code...)
	out := appendU32(nil, uint32(len(body)))
	out = append(out, body...)
	return out
}

func i64Const(value uint64) []byte {
	out := []byte{0x42}
	out = append(out, appendS64(nil, int64(value))...)
	out = append(out, 0x0b)
	return out
}

func i32Const(value uint32) []byte {
	out := []byte{0x41}
	out = append(out, appendU32(nil, value)...)
	out = append(out, 0x0b)
	return out
}

func dataEntry(offset uint32, payload []byte) []byte {
	out := []byte{0x00, 0x41}
	out = append(out, appendU32(nil, offset)...)
	out = append(out, 0x0b)
	out = append(out, appendU32(nil, uint32(len(payload)))...)
	out = append(out, payload...)
	return out
}

func section(id byte, payload []byte) []byte {
	out := []byte{id}
	out = append(out, appendU32(nil, uint32(len(payload)))...)
	out = append(out, payload...)
	return out
}

func appendU32(dst []byte, v uint32) []byte {
	for {
		b := byte(v & 0x7f)
		v >>= 7
		if v != 0 {
			b |= 0x80
		}
		dst = append(dst, b)
		if v == 0 {
			break
		}
	}
	return dst
}

func appendS64(dst []byte, v int64) []byte {
	for {
		b := byte(v & 0x7f)
		sign := b & 0x40
		v >>= 7
		done := (v == 0 && sign == 0) || (v == -1 && sign != 0)
		if !done {
			b |= 0x80
		}
		dst = append(dst, b)
		if done {
			break
		}
	}
	return dst
}
