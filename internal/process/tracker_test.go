package process

import (
	"context"
	"testing"
)

func TestTrackerStart(t *testing.T) {
	tracker := &Tracker{}
	if err := tracker.Start(context.Background(), 0); err == nil {
		t.Fatalf("expected error for zero agent id")
	}
	if err := tracker.Start(context.Background(), 42); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
