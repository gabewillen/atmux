package telemetry_test

import (
	"context"
	"testing"

	"github.com/stateforward/amux/internal/telemetry"
)

// TestInitAndStartSpan verifies that Init configures a tracer provider that
// can be used to start and end spans without error.
func TestInitAndStartSpan(t *testing.T) {
	ctx := context.Background()

	shutdown, err := telemetry.Init(ctx)
	if err != nil {
		t.Fatalf("Init error: %v", err)
	}
	if shutdown == nil {
		t.Fatalf("Init returned nil shutdown function")
	}

	ctx, span := telemetry.StartSpan(ctx, "test-span")
	if span == nil {
		t.Fatalf("StartSpan returned nil span")
	}
	span.End()

	if err := shutdown(ctx); err != nil {
		t.Fatalf("shutdown error: %v", err)
	}
}
