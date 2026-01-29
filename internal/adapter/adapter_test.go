package adapter

import (
	"context"
	"errors"
	"sort"
	"testing"
)

func TestNoopAdapter_Name(t *testing.T) {
	a := NewNoopAdapter("test-adapter")
	if a.Name() != "test-adapter" {
		t.Errorf("Name() = %q, want %q", a.Name(), "test-adapter")
	}
}

func TestNoopAdapter_Name_Empty(t *testing.T) {
	a := NewNoopAdapter("")
	if a.Name() != "" {
		t.Errorf("Name() = %q, want empty", a.Name())
	}
}

func TestNoopAdapter_Manifest(t *testing.T) {
	a := NewNoopAdapter("claude-code")

	m, err := a.Manifest()
	if err != nil {
		t.Fatalf("Manifest() error = %v", err)
	}

	if m == nil {
		t.Fatal("Manifest() returned nil")
	}

	if m.Name != "claude-code" {
		t.Errorf("Manifest().Name = %q, want %q", m.Name, "claude-code")
	}

	if m.Version != "0.0.0" {
		t.Errorf("Manifest().Version = %q, want %q", m.Version, "0.0.0")
	}
}

func TestNoopAdapter_OnOutput(t *testing.T) {
	a := NewNoopAdapter("test")

	events, err := a.OnOutput(context.Background(), []byte("some output"))
	if err != nil {
		t.Fatalf("OnOutput() error = %v", err)
	}

	if events != nil {
		t.Errorf("OnOutput() = %v, want nil", events)
	}
}

func TestNoopAdapter_OnOutput_EmptyInput(t *testing.T) {
	a := NewNoopAdapter("test")

	events, err := a.OnOutput(context.Background(), []byte{})
	if err != nil {
		t.Fatalf("OnOutput() error = %v", err)
	}

	if events != nil {
		t.Errorf("OnOutput() = %v, want nil", events)
	}
}

func TestNoopAdapter_OnOutput_NilInput(t *testing.T) {
	a := NewNoopAdapter("test")

	events, err := a.OnOutput(context.Background(), nil)
	if err != nil {
		t.Fatalf("OnOutput() error = %v", err)
	}

	if events != nil {
		t.Errorf("OnOutput() = %v, want nil", events)
	}
}

func TestNoopAdapter_FormatInput(t *testing.T) {
	a := NewNoopAdapter("test")

	result, err := a.FormatInput(context.Background(), "user input text")
	if err != nil {
		t.Fatalf("FormatInput() error = %v", err)
	}

	// FormatInput should return the input unchanged
	if result != "user input text" {
		t.Errorf("FormatInput() = %q, want %q", result, "user input text")
	}
}

func TestNoopAdapter_FormatInput_Empty(t *testing.T) {
	a := NewNoopAdapter("test")

	result, err := a.FormatInput(context.Background(), "")
	if err != nil {
		t.Fatalf("FormatInput() error = %v", err)
	}

	if result != "" {
		t.Errorf("FormatInput() = %q, want empty", result)
	}
}

func TestNoopAdapter_OnEvent(t *testing.T) {
	a := NewNoopAdapter("test")

	err := a.OnEvent(context.Background(), []byte(`{"type":"test"}`))
	if err != nil {
		t.Errorf("OnEvent() = %v, want nil", err)
	}
}

func TestNoopAdapter_OnEvent_NilInput(t *testing.T) {
	a := NewNoopAdapter("test")

	err := a.OnEvent(context.Background(), nil)
	if err != nil {
		t.Errorf("OnEvent() = %v, want nil", err)
	}
}

func TestNoopAdapter_Close(t *testing.T) {
	a := NewNoopAdapter("test")

	err := a.Close()
	if err != nil {
		t.Errorf("Close() = %v, want nil", err)
	}
}

func TestNoopAdapter_ImplementsInterface(t *testing.T) {
	// Compile-time check that NoopAdapter implements Adapter
	var _ Adapter = (*NoopAdapter)(nil)
}

func TestRegistry_RegisterAndLoad(t *testing.T) {
	ctx := context.Background()
	reg, err := NewRegistry(ctx)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}
	defer reg.Close(ctx)

	a := NewNoopAdapter("test-adapter")

	if err := reg.Register(a); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	loaded, err := reg.Load("test-adapter")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.Name() != "test-adapter" {
		t.Errorf("loaded adapter Name() = %q, want %q", loaded.Name(), "test-adapter")
	}
}

func TestRegistry_Load_NotFound(t *testing.T) {
	ctx := context.Background()
	reg, err := NewRegistry(ctx)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}
	defer reg.Close(ctx)

	_, err = reg.Load("nonexistent")
	if err == nil {
		t.Fatal("Load() should return error for unregistered adapter")
	}

	if !errors.Is(err, ErrAdapterNotFound) {
		t.Errorf("Load() error = %v, want ErrAdapterNotFound", err)
	}
}

func TestRegistry_Register_Duplicate(t *testing.T) {
	ctx := context.Background()
	reg, err := NewRegistry(ctx)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}
	defer reg.Close(ctx)

	a1 := NewNoopAdapter("test-adapter")
	a2 := NewNoopAdapter("test-adapter")

	if err := reg.Register(a1); err != nil {
		t.Fatalf("first Register() error = %v", err)
	}

	err = reg.Register(a2)
	if err == nil {
		t.Fatal("second Register() with same name should return error")
	}

	if !errors.Is(err, ErrAdapterAlreadyExists) {
		t.Errorf("Register() error = %v, want ErrAdapterAlreadyExists", err)
	}
}

func TestRegistry_List_Empty(t *testing.T) {
	ctx := context.Background()
	reg, err := NewRegistry(ctx)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}
	defer reg.Close(ctx)

	names := reg.List()
	if len(names) != 0 {
		t.Errorf("List() on empty registry = %v, want empty", names)
	}
}

func TestRegistry_List(t *testing.T) {
	ctx := context.Background()
	reg, err := NewRegistry(ctx)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}
	defer reg.Close(ctx)

	adapters := []string{"claude-code", "cursor", "windsurf"}
	for _, name := range adapters {
		if err := reg.Register(NewNoopAdapter(name)); err != nil {
			t.Fatalf("Register(%q) error = %v", name, err)
		}
	}

	names := reg.List()
	if len(names) != 3 {
		t.Fatalf("List() length = %d, want 3", len(names))
	}

	// Sort for deterministic comparison (map iteration order is random)
	sort.Strings(names)
	sort.Strings(adapters)

	for i, name := range names {
		if name != adapters[i] {
			t.Errorf("List()[%d] = %q, want %q", i, name, adapters[i])
		}
	}
}

func TestRegistry_Unregister(t *testing.T) {
	ctx := context.Background()
	reg, err := NewRegistry(ctx)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}
	defer reg.Close(ctx)

	a := NewNoopAdapter("test-adapter")
	if err := reg.Register(a); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// Unregister
	if err := reg.Unregister("test-adapter"); err != nil {
		t.Fatalf("Unregister() error = %v", err)
	}

	// Should no longer be loadable
	_, err = reg.Load("test-adapter")
	if err == nil {
		t.Fatal("Load() after Unregister() should return error")
	}

	// List should be empty
	names := reg.List()
	if len(names) != 0 {
		t.Errorf("List() after Unregister() = %v, want empty", names)
	}
}

func TestRegistry_Unregister_NotFound(t *testing.T) {
	ctx := context.Background()
	reg, err := NewRegistry(ctx)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}
	defer reg.Close(ctx)

	err = reg.Unregister("nonexistent")
	if err == nil {
		t.Fatal("Unregister() should return error for unregistered adapter")
	}

	if !errors.Is(err, ErrAdapterNotFound) {
		t.Errorf("Unregister() error = %v, want ErrAdapterNotFound", err)
	}
}

func TestRegistry_Close(t *testing.T) {
	ctx := context.Background()
	reg, err := NewRegistry(ctx)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}

	// Register some adapters
	if err := reg.Register(NewNoopAdapter("a")); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if err := reg.Register(NewNoopAdapter("b")); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// Close should not error
	if err := reg.Close(ctx); err != nil {
		t.Errorf("Close() = %v, want nil", err)
	}
}

func TestRegistry_RegisterMultipleDifferent(t *testing.T) {
	ctx := context.Background()
	reg, err := NewRegistry(ctx)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}
	defer reg.Close(ctx)

	names := []string{"adapter-1", "adapter-2", "adapter-3"}
	for _, name := range names {
		if err := reg.Register(NewNoopAdapter(name)); err != nil {
			t.Fatalf("Register(%q) error = %v", name, err)
		}
	}

	// Each should be loadable
	for _, name := range names {
		loaded, err := reg.Load(name)
		if err != nil {
			t.Errorf("Load(%q) error = %v", name, err)
			continue
		}
		if loaded.Name() != name {
			t.Errorf("loaded adapter Name() = %q, want %q", loaded.Name(), name)
		}
	}
}

func TestNoopPatternMatcher_Match(t *testing.T) {
	m := NewNoopPatternMatcher()

	matches := m.Match([]byte("some output text"))
	if matches != nil {
		t.Errorf("NoopPatternMatcher.Match() = %v, want nil", matches)
	}
}

func TestNoopPatternMatcher_Match_EmptyInput(t *testing.T) {
	m := NewNoopPatternMatcher()

	matches := m.Match([]byte{})
	if matches != nil {
		t.Errorf("NoopPatternMatcher.Match() = %v, want nil", matches)
	}
}

func TestNoopPatternMatcher_Match_NilInput(t *testing.T) {
	m := NewNoopPatternMatcher()

	matches := m.Match(nil)
	if matches != nil {
		t.Errorf("NoopPatternMatcher.Match() = %v, want nil", matches)
	}
}

func TestNoopPatternMatcher_ImplementsInterface(t *testing.T) {
	// Compile-time check that NoopPatternMatcher implements PatternMatcher
	var _ PatternMatcher = (*NoopPatternMatcher)(nil)
}

func TestSentinelErrors_NonNil(t *testing.T) {
	sentinels := []struct {
		name string
		err  error
	}{
		{"ErrAdapterNotFound", ErrAdapterNotFound},
		{"ErrAdapterAlreadyExists", ErrAdapterAlreadyExists},
		{"ErrAdapterLoadFailed", ErrAdapterLoadFailed},
		{"ErrAdapterCallFailed", ErrAdapterCallFailed},
	}

	for _, tt := range sentinels {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("%s is nil", tt.name)
			}
			if tt.err.Error() == "" {
				t.Errorf("%s.Error() is empty", tt.name)
			}
		})
	}
}

func TestManifest_Fields(t *testing.T) {
	m := &Manifest{
		Name:    "test-adapter",
		Version: "1.2.3",
		CLI: CLIConstraint{
			Constraint: ">=1.0.0",
		},
		Patterns: map[string]string{
			"prompt": `\$\s*$`,
		},
	}

	if m.Name != "test-adapter" {
		t.Errorf("Name = %q, want %q", m.Name, "test-adapter")
	}
	if m.Version != "1.2.3" {
		t.Errorf("Version = %q, want %q", m.Version, "1.2.3")
	}
	if m.CLI.Constraint != ">=1.0.0" {
		t.Errorf("CLI.Constraint = %q, want %q", m.CLI.Constraint, ">=1.0.0")
	}
	if len(m.Patterns) != 1 {
		t.Errorf("len(Patterns) = %d, want 1", len(m.Patterns))
	}
}

func TestOutputEvent_Fields(t *testing.T) {
	oe := OutputEvent{
		Type: "process.spawned",
		Data: map[string]int{"pid": 1234},
	}

	if oe.Type != "process.spawned" {
		t.Errorf("Type = %q, want %q", oe.Type, "process.spawned")
	}
	if oe.Data == nil {
		t.Error("Data should not be nil")
	}
}

func TestPatternMatch_Fields(t *testing.T) {
	pm := PatternMatch{
		Pattern: "prompt",
		Match:   "$ ",
		Index:   42,
	}

	if pm.Pattern != "prompt" {
		t.Errorf("Pattern = %q, want %q", pm.Pattern, "prompt")
	}
	if pm.Match != "$ " {
		t.Errorf("Match = %q, want %q", pm.Match, "$ ")
	}
	if pm.Index != 42 {
		t.Errorf("Index = %d, want 42", pm.Index)
	}
}
