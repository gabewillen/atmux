package inference

import (
	"context"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
)

func TestNewManager(t *testing.T) {
	// Test disabled inference manager
	inferenceConfig := &config.InferenceConfig{
		Enabled: false,
		Engine:  "liquidgen",
		Models:  make(map[string]config.ModelConfig),
	}

	manager, err := NewManager(inferenceConfig)
	if err != nil {
		t.Fatalf("Failed to create disabled manager: %v", err)
	}

	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}

	// Test with disabled manager, should not have engines
	if len(manager.engines) != 0 {
		t.Errorf("Expected 0 engines, got %d", len(manager.engines))
	}

	// Test shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := manager.Shutdown(ctx); err != nil {
		t.Errorf("Unexpected shutdown error: %v", err)
	}
}

func TestNewManagerEnabled(t *testing.T) {
	// Test enabled inference manager
	inferenceConfig := &config.InferenceConfig{
		Enabled: true,
		Engine:  "liquidgen",
		Models: map[string]config.ModelConfig{
			"test-model": {
				Type: "generation",
				Path: "/path/to/model",
			},
		},
	}

	manager, err := NewManager(inferenceConfig)
	if err != nil {
		t.Fatalf("Failed to create enabled manager: %v", err)
	}

	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}

	// Should have one engine
	if len(manager.engines) != 1 {
		t.Errorf("Expected 1 engine, got %d", len(manager.engines))
	}

	// Test getting engine
	engine, err := manager.GetEngine("liquidgen")
	if err != nil {
		t.Errorf("Failed to get engine: %v", err)
	}

	if engine == nil {
		t.Error("Expected non-nil engine")
	}

	// Test getting default engine
	defaultEngine, err := manager.GetDefaultEngine()
	if err != nil {
		t.Errorf("Failed to get default engine: %v", err)
	}

	if defaultEngine == nil {
		t.Error("Expected non-nil default engine")
	}

	// Test listing engines
	engines := manager.ListEngines()
	if len(engines) != 1 {
		t.Errorf("Expected 1 engine in list, got %d", len(engines))
	}

	if _, exists := engines["liquidgen"]; !exists {
		t.Error("Expected liquidgen engine in list")
	}

	// Test getting non-existent engine
	_, err = manager.GetEngine("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent engine")
	}

	// Test shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := manager.Shutdown(ctx); err != nil {
		t.Errorf("Unexpected shutdown error: %v", err)
	}
}

func TestLiquidgenEngine(t *testing.T) {
	models := map[string]config.ModelConfig{
		"test-model": {
			Type: "generation",
			Path: "/path/to/model",
			Parameters: map[string]interface{}{
				"temperature": 0.7,
			},
		},
	}

	engine, err := NewLiquidgenEngine(models)
	if err != nil {
		t.Fatalf("Failed to create liquidgen engine: %v", err)
	}

	if engine == nil {
		t.Fatal("Expected non-nil engine")
	}

	// Test engine info
	info := engine.Info()
	if info.Name != "liquidgen" {
		t.Errorf("Expected name 'liquidgen', got %q", info.Name)
	}

	if info.Type != "local" {
		t.Errorf("Expected type 'local', got %q", info.Type)
	}

	// Test model info
	modelInfo, exists := info.Models["test-model"]
	if !exists {
		t.Error("Expected test-model in models")
	}

	if modelInfo.Type != "generation" {
		t.Errorf("Expected model type 'generation', got %q", modelInfo.Type)
	}

	if modelInfo.Path != "/path/to/model" {
		t.Errorf("Expected model path '/path/to/model', got %q", modelInfo.Path)
	}

	// Test operations (should return not ready)
	ctx := context.Background()

	// Test initialize
	model := models["test-model"]
	err = engine.Initialize(ctx, &model)
	if err == nil {
		t.Error("Expected initialization to return not ready error")
	}

	// Test generate
	_, err = engine.Generate(ctx, "test prompt", &GenerateOptions{})
	if err == nil {
		t.Error("Expected generate to return not ready error")
	}

	// Test embed
	_, err = engine.Embed(ctx, []string{"test text"}, &EmbedOptions{})
	if err == nil {
		t.Error("Expected embed to return not ready error")
	}

	// Test shutdown
	err = engine.Shutdown(ctx)
	if err != nil {
		t.Errorf("Unexpected shutdown error: %v", err)
	}
}
