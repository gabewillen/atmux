// Package inference implements tests for the local inference integration
package inference

import (
	"context"
	"testing"
)

// TestNewLiquidGenEngine tests creating a new LiquidGenEngine
func TestNewLiquidGenEngine(t *testing.T) {
	engine := NewLiquidGenEngine()
	
	if engine == nil {
		t.Fatal("Expected engine to be created")
	}
	
	if engine.models == nil {
		t.Error("Expected models map to be initialized")
	}
	
	if len(engine.ListModels()) != 0 {
		t.Error("Expected no models initially")
	}
}

// TestRegisterModel tests registering a model with the engine
func TestRegisterModel(t *testing.T) {
	engine := NewLiquidGenEngine()
	
	modelInfo := ModelInfo{
		ID:          "test-model",
		Name:        "Test Model",
		Description: "A test model",
		SizeBytes:   1024,
		Loaded:      false,
		Version:     "1.0.0",
		URL:         "https://example.com/model",
	}
	
	engine.RegisterModel(modelInfo)
	
	models := engine.ListModels()
	if len(models) != 1 {
		t.Fatalf("Expected 1 model, got %d", len(models))
	}
	
	if models[0] != "test-model" {
		t.Errorf("Expected model ID 'test-model', got '%s'", models[0])
	}
}

// TestGetModelInfo tests retrieving model information
func TestGetModelInfo(t *testing.T) {
	engine := NewLiquidGenEngine()
	
	modelInfo := ModelInfo{
		ID:          "test-model",
		Name:        "Test Model",
		Description: "A test model",
		SizeBytes:   1024,
		Loaded:      false,
		Version:     "1.0.0",
		URL:         "https://example.com/model",
	}
	
	engine.RegisterModel(modelInfo)
	
	retrievedInfo, err := engine.GetModelInfo("test-model")
	if err != nil {
		t.Fatalf("Unexpected error getting model info: %v", err)
	}
	
	if retrievedInfo.ID != "test-model" {
		t.Errorf("Expected model ID 'test-model', got '%s'", retrievedInfo.ID)
	}
	
	if retrievedInfo.Name != "Test Model" {
		t.Errorf("Expected model name 'Test Model', got '%s'", retrievedInfo.Name)
	}
	
	if retrievedInfo.SizeBytes != 1024 {
		t.Errorf("Expected model size 1024, got %d", retrievedInfo.SizeBytes)
	}
	
	// Test with non-existent model
	_, err = engine.GetModelInfo("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent model")
	}
}

// TestIsModelLoaded tests checking if a model is loaded
func TestIsModelLoaded(t *testing.T) {
	engine := NewLiquidGenEngine()
	
	modelInfo := ModelInfo{
		ID:          "test-model",
		Name:        "Test Model",
		Description: "A test model",
		SizeBytes:   1024,
		Loaded:      false, // Initially not loaded
		Version:     "1.0.0",
		URL:         "https://example.com/model",
	}
	
	engine.RegisterModel(modelInfo)
	
	// Initially should not be loaded
	if engine.IsModelLoaded("test-model") {
		t.Error("Expected model to not be loaded initially")
	}
	
	// Load the model
	err := engine.LoadModel(context.Background(), "test-model")
	if err != nil {
		t.Fatalf("Unexpected error loading model: %v", err)
	}
	
	// Should now be loaded
	if !engine.IsModelLoaded("test-model") {
		t.Error("Expected model to be loaded after LoadModel call")
	}
	
	// Unload the model
	err = engine.UnloadModel("test-model")
	if err != nil {
		t.Fatalf("Unexpected error unloading model: %v", err)
	}
	
	// Should now be unloaded
	if engine.IsModelLoaded("test-model") {
		t.Error("Expected model to not be loaded after UnloadModel call")
	}
}

// TestLoadModel tests loading a model
func TestLoadModel(t *testing.T) {
	engine := NewLiquidGenEngine()
	
	modelInfo := ModelInfo{
		ID:          "test-model",
		Name:        "Test Model",
		Description: "A test model",
		SizeBytes:   1024,
		Loaded:      false,
		Version:     "1.0.0",
		URL:         "https://example.com/model",
	}
	
	engine.RegisterModel(modelInfo)
	
	// Load the model
	err := engine.LoadModel(context.Background(), "test-model")
	if err != nil {
		t.Fatalf("Unexpected error loading model: %v", err)
	}
	
	// Verify it's loaded
	info, err := engine.GetModelInfo("test-model")
	if err != nil {
		t.Fatalf("Unexpected error getting model info: %v", err)
	}
	
	if !info.Loaded {
		t.Error("Expected model to be marked as loaded")
	}
	
	// Try to load non-existent model
	err = engine.LoadModel(context.Background(), "non-existent")
	if err == nil {
		t.Error("Expected error for non-existent model")
	}
}

// TestUnloadModel tests unloading a model
func TestUnloadModel(t *testing.T) {
	engine := NewLiquidGenEngine()
	
	modelInfo := ModelInfo{
		ID:          "test-model",
		Name:        "Test Model",
		Description: "A test model",
		SizeBytes:   1024,
		Loaded:      true, // Initially loaded
		Version:     "1.0.0",
		URL:         "https://example.com/model",
	}
	
	engine.RegisterModel(modelInfo)
	
	// Unload the model
	err := engine.UnloadModel("test-model")
	if err != nil {
		t.Fatalf("Unexpected error unloading model: %v", err)
	}
	
	// Verify it's unloaded
	info, err := engine.GetModelInfo("test-model")
	if err != nil {
		t.Fatalf("Unexpected error getting model info: %v", err)
	}
	
	if info.Loaded {
		t.Error("Expected model to be marked as unloaded")
	}
	
	// Try to unload non-existent model
	err = engine.UnloadModel("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent model")
	}
}

// TestInfer tests performing inference
func TestInfer(t *testing.T) {
	engine := NewLiquidGenEngine()
	
	modelInfo := ModelInfo{
		ID:          "test-model",
		Name:        "Test Model",
		Description: "A test model",
		SizeBytes:   1024,
		Loaded:      true, // Must be loaded for inference
		Version:     "1.0.0",
		URL:         "https://example.com/model",
	}
	
	engine.RegisterModel(modelInfo)
	
	req := InferenceRequest{
		ModelID: "test-model",
		Prompt:  "Hello, world!",
		Options: map[string]interface{}{
			"temperature": 0.7,
		},
	}
	
	resp, err := engine.Infer(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error during inference: %v", err)
	}
	
	if resp.ModelID != "test-model" {
		t.Errorf("Expected model ID 'test-model', got '%s'", resp.ModelID)
	}
	
	if resp.Output == "" {
		t.Error("Expected non-empty output")
	}
	
	if resp.Stats.DurationMS <= 0 {
		t.Error("Expected positive duration")
	}
	
	if resp.Stats.TokenCount <= 0 {
		t.Error("Expected positive token count")
	}
	
	if resp.Stats.ModelSize != 1024 {
		t.Errorf("Expected model size 1024, got %d", resp.Stats.ModelSize)
	}
	
	// Test with non-existent model
	req.ModelID = "non-existent"
	_, err = engine.Infer(context.Background(), req)
	if err == nil {
		t.Error("Expected error for non-existent model")
	}
	
	// Test with unloaded model
	modelInfo.Loaded = false
	engine.RegisterModel(modelInfo)
	
	req.ModelID = "test-model"
	_, err = engine.Infer(context.Background(), req)
	if err == nil {
		t.Error("Expected error for unloaded model")
	}
}