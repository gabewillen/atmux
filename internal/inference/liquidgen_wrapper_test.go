// Package inference implements tests for the liquidgen wrapper
package inference

import (
	"context"
	"testing"
)

// TestNewLiquidGenWrapper tests creating a new liquidgen wrapper
func TestNewLiquidGenWrapper(t *testing.T) {
	wrapper, err := NewLiquidGenWrapper("/path/to/model")
	if err != nil {
		t.Fatalf("Unexpected error creating wrapper: %v", err)
	}
	
	if wrapper == nil {
		t.Fatal("Expected wrapper to be created")
	}
	
	if wrapper.modelPath != "/path/to/model" {
		t.Errorf("Expected model path '/path/to/model', got '%s'", wrapper.modelPath)
	}
	
	if !wrapper.loaded {
		t.Error("Expected wrapper to be loaded initially")
	}
	
	expectedVersion := GetLiquidGenVersion()
	if wrapper.version != expectedVersion {
		t.Errorf("Expected version '%s', got '%s'", expectedVersion, wrapper.version)
	}
}

// TestLiquidGenWrapperInfer tests inference with the liquidgen wrapper
func TestLiquidGenWrapperInfer(t *testing.T) {
	wrapper, err := NewLiquidGenWrapper("/path/to/model")
	if err != nil {
		t.Fatalf("Unexpected error creating wrapper: %v", err)
	}
	
	req := InferenceRequest{
		ModelID: "test-model",
		Prompt:  "Hello, world!",
	}
	
	resp, err := wrapper.Infer(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error during inference: %v", err)
	}
	
	if resp.ModelID != "test-model" {
		t.Errorf("Expected model ID 'test-model', got '%s'", resp.ModelID)
	}
	
	if resp.Output == "" {
		t.Error("Expected non-empty output")
	}
	
	if resp.Output != "LiquidGen output for prompt: Hello, world!" {
		t.Errorf("Expected specific output format, got '%s'", resp.Output)
	}
	
	if resp.Stats.DurationMS <= 0 {
		t.Error("Expected positive duration")
	}
	
	if resp.Stats.TokenCount <= 0 {
		t.Error("Expected positive token count")
	}
	
	if resp.Stats.ModelSize <= 0 {
		t.Error("Expected positive model size")
	}
}

// TestLiquidGenWrapperInferWhenNotLoaded tests inference when wrapper is not loaded
func TestLiquidGenWrapperInferWhenNotLoaded(t *testing.T) {
	wrapper := &LiquidGenWrapper{
		modelPath: "/path/to/model",
		loaded:    false,
	}
	
	req := InferenceRequest{
		ModelID: "test-model",
		Prompt:  "Hello, world!",
	}
	
	_, err := wrapper.Infer(context.Background(), req)
	if err == nil {
		t.Error("Expected error when wrapper is not loaded")
	}
	
	expectedErrMsg := "liquidgen model not loaded"
	if err.Error() != expectedErrMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

// TestLiquidGenWrapperClose tests closing the wrapper
func TestLiquidGenWrapperClose(t *testing.T) {
	wrapper, err := NewLiquidGenWrapper("/path/to/model")
	if err != nil {
		t.Fatalf("Unexpected error creating wrapper: %v", err)
	}
	
	if !wrapper.loaded {
		t.Error("Expected wrapper to be loaded initially")
	}
	
	wrapper.Close()
	
	if wrapper.loaded {
		t.Error("Expected wrapper to be unloaded after Close()")
	}
}

// TestGetLiquidGenVersion tests getting the liquidgen version
func TestGetLiquidGenVersion(t *testing.T) {
	version := GetLiquidGenVersion()
	
	if version == "" {
		t.Error("Expected non-empty version string")
	}
	
	if version == "liquidgen-v0.1.0-placeholder" {
		t.Error("Expected updated version string, not placeholder")
	}
}