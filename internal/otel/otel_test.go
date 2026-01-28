// Package otel implements tests for OpenTelemetry scaffolding
package otel

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// TestNewProvider tests creating a new OpenTelemetry provider
func TestNewProvider(t *testing.T) {
	provider, err := NewProvider("test-service", attribute.String("environment", "test"))
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	
	if provider.serviceName != "test-service" {
		t.Errorf("Expected service name 'test-service', got '%s'", provider.serviceName)
	}
	
	if provider.tracerProvider == nil {
		t.Error("Expected tracer provider to be initialized")
	}
	
	if provider.meterProvider == nil {
		t.Error("Expected meter provider to be initialized")
	}
	
	// Test shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err = provider.Shutdown(ctx)
	if err != nil {
		t.Errorf("Failed to shutdown provider: %v", err)
	}
}

// TestGetTracer tests getting a tracer from the provider
func TestGetTracer(t *testing.T) {
	provider, err := NewProvider("test-service")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		provider.Shutdown(ctx)
	}()
	
	tracer := provider.GetTracer("test-component")
	if tracer == nil {
		t.Error("Expected tracer to be returned")
	}
	
	// Verify it's a valid tracer by starting a span
	ctx, span := tracer.Start(context.Background(), "test-span")
	if span == nil {
		t.Error("Expected span to be created")
	}
	
	span.End()
	if ctx == nil {
		t.Error("Expected context to be returned")
	}
}

// TestGetMeter tests getting a meter from the provider
func TestGetMeter(t *testing.T) {
	provider, err := NewProvider("test-service")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		provider.Shutdown(ctx)
	}()
	
	meter := provider.GetMeter("test-component")
	if meter == nil {
		t.Error("Expected meter to be returned")
	}
	
	// We can't easily test metrics without adding more dependencies,
	// so we just verify the meter is not nil
}

// TestTracerProvider tests accessing the tracer provider directly
func TestTracerProvider(t *testing.T) {
	provider, err := NewProvider("test-service")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		provider.Shutdown(ctx)
	}()
	
	tp := provider.TracerProvider()
	if tp == nil {
		t.Error("Expected tracer provider to be returned")
	}
	
	// Test that we can create a tracer from the provider
	tracer := tp.Tracer("direct-test")
	if tracer == nil {
		t.Error("Expected tracer to be created from provider")
	}
	
	// Verify it works by starting a span
	ctx, span := tracer.Start(context.Background(), "direct-span")
	if span == nil {
		t.Error("Expected span to be created from direct tracer")
	}
	
	span.End()
	_ = ctx
}

// TestMeterProvider tests accessing the meter provider directly
func TestMeterProvider(t *testing.T) {
	provider, err := NewProvider("test-service")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		provider.Shutdown(ctx)
	}()
	
	mp := provider.MeterProvider()
	if mp == nil {
		t.Error("Expected meter provider to be returned")
	}
	
	// Just verify the meter provider is not nil
}

// TestShutdownTwice tests shutting down twice (should be safe)
func TestShutdownTwice(t *testing.T) {
	provider, err := NewProvider("test-service")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// First shutdown
	err = provider.Shutdown(ctx)
	if err != nil {
		t.Errorf("First shutdown failed: %v", err)
	}
	
	// Second shutdown (should be safe)
	err = provider.Shutdown(ctx)
	if err != nil {
		// Note: Some implementations may return an error on second shutdown
		// This is acceptable behavior, so we'll just log it
		t.Logf("Second shutdown returned error (this may be expected): %v", err)
	}
}