package errors_test

import (
	"errors"
	"testing"

	amerrors "github.com/stateforward/amux/internal/errors"
)

func TestSentinelErrors(t *testing.T) {
	// Verify sentinel errors are defined
	if amerrors.ErrNotFound == nil {
		t.Error("ErrNotFound should be defined")
	}
	if amerrors.ErrInvalidInput == nil {
		t.Error("ErrInvalidInput should be defined")
	}
}

func TestWrap(t *testing.T) {
	// Test wrapping nil returns nil
	if err := amerrors.Wrap(nil, "context"); err != nil {
		t.Errorf("Wrap(nil, ...) should return nil, got %v", err)
	}

	// Test wrapping an error adds context
	baseErr := errors.New("base error")
	wrapped := amerrors.Wrap(baseErr, "operation failed")

	if wrapped == nil {
		t.Fatal("Wrap should return non-nil for non-nil error")
	}

	// Verify the error can be unwrapped
	if !errors.Is(wrapped, baseErr) {
		t.Error("wrapped error should unwrap to base error")
	}

	// Verify context is in the message
	msg := wrapped.Error()
	if msg != "operation failed: base error" {
		t.Errorf("unexpected error message: %s", msg)
	}
}

func TestWrapf(t *testing.T) {
	// Test wrapping nil returns nil
	if err := amerrors.Wrapf(nil, "context %d", 42); err != nil {
		t.Errorf("Wrapf(nil, ...) should return nil, got %v", err)
	}

	// Test formatted wrapping
	baseErr := errors.New("base error")
	wrapped := amerrors.Wrapf(baseErr, "operation %s failed with code %d", "test", 42)

	if wrapped == nil {
		t.Fatal("Wrapf should return non-nil for non-nil error")
	}

	// Verify the error can be unwrapped
	if !errors.Is(wrapped, baseErr) {
		t.Error("wrapped error should unwrap to base error")
	}

	// Verify formatted context is in the message
	msg := wrapped.Error()
	expected := "operation test failed with code 42: base error"
	if msg != expected {
		t.Errorf("unexpected error message: got %s, want %s", msg, expected)
	}
}
