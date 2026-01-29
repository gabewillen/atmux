package inference

import (
	"context"
	"io"
	"testing"
)

func TestIsValidModel(t *testing.T) {
	tests := []struct {
		name  string
		model string
		want  bool
	}{
		{"lfm2.5-thinking is valid", ModelLFM25Thinking, true},
		{"lfm2.5-VL is valid", ModelLFM25VL, true},
		{"empty string is invalid", "", false},
		{"unknown model is invalid", "gpt-4", false},
		{"case sensitive - uppercase", "LFM2.5-THINKING", false},
		{"partial match is invalid", "lfm2.5", false},
		{"prefix match is invalid", "lfm2.5-thinking-extra", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidModel(tt.model)
			if got != tt.want {
				t.Errorf("IsValidModel(%q) = %v, want %v", tt.model, got, tt.want)
			}
		})
	}
}

func TestValidModels(t *testing.T) {
	if len(ValidModels) != 2 {
		t.Errorf("len(ValidModels) = %d, want 2", len(ValidModels))
	}

	found := make(map[string]bool)
	for _, m := range ValidModels {
		found[m] = true
	}

	if !found[ModelLFM25Thinking] {
		t.Errorf("ValidModels does not contain %q", ModelLFM25Thinking)
	}
	if !found[ModelLFM25VL] {
		t.Errorf("ValidModels does not contain %q", ModelLFM25VL)
	}
}

func TestModelConstants(t *testing.T) {
	if ModelLFM25Thinking != "lfm2.5-thinking" {
		t.Errorf("ModelLFM25Thinking = %q, want %q", ModelLFM25Thinking, "lfm2.5-thinking")
	}
	if ModelLFM25VL != "lfm2.5-VL" {
		t.Errorf("ModelLFM25VL = %q, want %q", ModelLFM25VL, "lfm2.5-VL")
	}
}

func TestNoopEngine_Generate_ReturnsEOFStream(t *testing.T) {
	engine := NewNoopEngine()

	stream, err := engine.Generate(context.Background(), Request{
		Model:  ModelLFM25Thinking,
		Prompt: "Hello",
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	defer stream.Close()

	// noopStream should immediately return EOF
	token, err := stream.Next()
	if err != io.EOF {
		t.Errorf("Next() error = %v, want io.EOF", err)
	}
	if token != "" {
		t.Errorf("Next() token = %q, want empty", token)
	}
}

func TestNoopEngine_Generate_InvalidModel(t *testing.T) {
	engine := NewNoopEngine()

	_, err := engine.Generate(context.Background(), Request{
		Model:  "invalid-model",
		Prompt: "Hello",
	})
	if err == nil {
		t.Fatal("Generate with invalid model should return error")
	}
	if err != ErrModelNotFound {
		t.Errorf("Generate error = %v, want %v", err, ErrModelNotFound)
	}
}

func TestNoopEngine_Generate_Unavailable(t *testing.T) {
	engine := NewNoopEngine()
	engine.SetAvailable(false)

	_, err := engine.Generate(context.Background(), Request{
		Model:  ModelLFM25Thinking,
		Prompt: "Hello",
	})
	if err == nil {
		t.Fatal("Generate when unavailable should return error")
	}
	if err != ErrEngineUnavailable {
		t.Errorf("Generate error = %v, want %v", err, ErrEngineUnavailable)
	}
}

func TestNoopEngine_Available(t *testing.T) {
	engine := NewNoopEngine()

	// Default should be available
	if !engine.Available() {
		t.Error("new NoopEngine should be available by default")
	}

	// Set unavailable
	engine.SetAvailable(false)
	if engine.Available() {
		t.Error("engine should be unavailable after SetAvailable(false)")
	}

	// Set available again
	engine.SetAvailable(true)
	if !engine.Available() {
		t.Error("engine should be available after SetAvailable(true)")
	}
}

func TestNoopEngine_Close(t *testing.T) {
	engine := NewNoopEngine()

	err := engine.Close()
	if err != nil {
		t.Errorf("NoopEngine.Close() = %v, want nil", err)
	}
}

func TestNoopEngine_Generate_BothModels(t *testing.T) {
	engine := NewNoopEngine()

	for _, model := range ValidModels {
		t.Run(model, func(t *testing.T) {
			stream, err := engine.Generate(context.Background(), Request{
				Model:  model,
				Prompt: "test prompt",
			})
			if err != nil {
				t.Fatalf("Generate(%q) failed: %v", model, err)
			}
			defer stream.Close()

			_, err = stream.Next()
			if err != io.EOF {
				t.Errorf("Next() error = %v, want io.EOF", err)
			}
		})
	}
}

func TestStringStream_ReturnsTokensThenEOF(t *testing.T) {
	tokens := []string{"Hello", " ", "world", "!"}
	stream := NewStringStream(tokens...)

	for i, expected := range tokens {
		token, err := stream.Next()
		if err != nil {
			t.Fatalf("Next() #%d error = %v, want nil", i, err)
		}
		if token != expected {
			t.Errorf("Next() #%d = %q, want %q", i, token, expected)
		}
	}

	// After all tokens, should return EOF
	token, err := stream.Next()
	if err != io.EOF {
		t.Errorf("Next() after tokens error = %v, want io.EOF", err)
	}
	if token != "" {
		t.Errorf("Next() after tokens token = %q, want empty", token)
	}

	// Subsequent calls should continue returning EOF
	token, err = stream.Next()
	if err != io.EOF {
		t.Errorf("Next() second call after EOF error = %v, want io.EOF", err)
	}
	if token != "" {
		t.Errorf("Next() second call after EOF token = %q, want empty", token)
	}
}

func TestStringStream_Empty(t *testing.T) {
	stream := NewStringStream()

	token, err := stream.Next()
	if err != io.EOF {
		t.Errorf("Next() error = %v, want io.EOF", err)
	}
	if token != "" {
		t.Errorf("Next() token = %q, want empty", token)
	}
}

func TestStringStream_SingleToken(t *testing.T) {
	stream := NewStringStream("only-one")

	token, err := stream.Next()
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if token != "only-one" {
		t.Errorf("Next() = %q, want %q", token, "only-one")
	}

	_, err = stream.Next()
	if err != io.EOF {
		t.Errorf("Next() after last token error = %v, want io.EOF", err)
	}
}

func TestStringStream_Close(t *testing.T) {
	stream := NewStringStream("a", "b", "c")

	err := stream.Close()
	if err != nil {
		t.Errorf("StringStream.Close() = %v, want nil", err)
	}
}

func TestLiquidgenEngine_Available_NoBinary(t *testing.T) {
	// With no binary path and no liquidgen in PATH, should be unavailable
	engine := NewLiquidgenEngine(&LiquidgenOptions{
		BinaryPath: "/nonexistent/path/to/liquidgen",
	})

	if engine.Available() {
		t.Error("LiquidgenEngine should be unavailable when binary does not exist")
	}
}

func TestLiquidgenEngine_Available_EmptyBinaryPath(t *testing.T) {
	// With empty binary path and assuming liquidgen is not in PATH
	engine := NewLiquidgenEngine(&LiquidgenOptions{
		BinaryPath: "",
	})

	// This will search PATH for "liquidgen", which is very unlikely to exist
	// in a test environment. If it does exist, the test still passes.
	_ = engine.Available()
}

func TestLiquidgenEngine_Generate_Unavailable(t *testing.T) {
	engine := NewLiquidgenEngine(&LiquidgenOptions{
		BinaryPath: "/nonexistent/path/to/liquidgen",
	})

	_, err := engine.Generate(context.Background(), Request{
		Model:  ModelLFM25Thinking,
		Prompt: "Hello",
	})
	if err == nil {
		t.Fatal("Generate when unavailable should return error")
	}
	if err != ErrEngineUnavailable {
		t.Errorf("Generate error = %v, want %v", err, ErrEngineUnavailable)
	}
}

func TestLiquidgenEngine_Generate_InvalidModel(t *testing.T) {
	// Create a "fake" binary using /bin/true so Available() is true
	engine := NewLiquidgenEngine(&LiquidgenOptions{
		BinaryPath: "/bin/true",
	})

	if !engine.Available() {
		t.Skip("skipping test: /bin/true not available")
	}

	_, err := engine.Generate(context.Background(), Request{
		Model:  "invalid-model",
		Prompt: "Hello",
	})
	if err == nil {
		t.Fatal("Generate with invalid model should return error")
	}
	if err != ErrModelNotFound {
		t.Errorf("Generate error = %v, want %v", err, ErrModelNotFound)
	}
}

func TestLiquidgenEngine_Close(t *testing.T) {
	engine := NewLiquidgenEngine(&LiquidgenOptions{
		BinaryPath: "/nonexistent/path",
	})

	err := engine.Close()
	if err != nil {
		t.Errorf("LiquidgenEngine.Close() = %v, want nil", err)
	}
}

func TestLiquidgenEngine_NilOptions(t *testing.T) {
	engine := NewLiquidgenEngine(nil)

	// Should not panic and should create a valid engine
	if engine == nil {
		t.Fatal("NewLiquidgenEngine(nil) returned nil")
	}
}

func TestSentinelErrors_NonNil(t *testing.T) {
	sentinels := []struct {
		name string
		err  error
	}{
		{"ErrModelNotFound", ErrModelNotFound},
		{"ErrModelLoadFailed", ErrModelLoadFailed},
		{"ErrEngineUnavailable", ErrEngineUnavailable},
		{"ErrGenerationFailed", ErrGenerationFailed},
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

func TestSetAndGetDefaultEngine(t *testing.T) {
	// Save original
	original := DefaultEngine()
	defer SetDefaultEngine(original)

	noop := NewNoopEngine()
	SetDefaultEngine(noop)

	got := DefaultEngine()
	if got != noop {
		t.Error("DefaultEngine should return the engine set by SetDefaultEngine")
	}
}

func TestPackageLevelGenerate(t *testing.T) {
	// Save original
	original := DefaultEngine()
	defer SetDefaultEngine(original)

	noop := NewNoopEngine()
	SetDefaultEngine(noop)

	stream, err := Generate(context.Background(), Request{
		Model:  ModelLFM25Thinking,
		Prompt: "test",
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	defer stream.Close()

	_, err = stream.Next()
	if err != io.EOF {
		t.Errorf("Next() error = %v, want io.EOF", err)
	}
}

func TestPackageLevelAvailable(t *testing.T) {
	// Save original
	original := DefaultEngine()
	defer SetDefaultEngine(original)

	noop := NewNoopEngine()
	SetDefaultEngine(noop)

	if !Available() {
		t.Error("Available() should be true with NoopEngine")
	}

	noop.SetAvailable(false)
	if Available() {
		t.Error("Available() should be false when engine is unavailable")
	}
}
