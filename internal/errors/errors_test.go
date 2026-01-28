// Package errors implements error handling conventions for the amux project
package errors

import (
	"errors"
	"fmt"
	"testing"
)

// TestErrorWrapping verifies the error wrapping convention: fmt.Errorf("context: %w", err)
func TestErrorWrapping(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := fmt.Errorf("additional context: %w", originalErr)

	if wrappedErr == nil {
		t.Error("Expected wrapped error, got nil")
	}

	// Check if the error is properly wrapped
	if !errors.Is(wrappedErr, originalErr) {
		t.Error("Error was not properly wrapped")
	}

	// Test our Wrapf function
	testErr := errors.New("test error")
	wrappedByFunc := Wrapf(testErr, "wrapped context")
	if wrappedByFunc == nil {
		t.Error("Expected wrapped error from function, got nil")
	}

	if !errors.Is(wrappedByFunc, testErr) {
		t.Error("Error was not properly wrapped by Wrapf function")
	}
}

// TestSentinelErrors verifies that sentinel errors are defined as package-level variables
func TestSentinelErrors(t *testing.T) {
	if ErrNotImplemented == nil {
		t.Error("ErrNotImplemented should be defined as a sentinel error")
	}

	if ErrInvalidInput == nil {
		t.Error("ErrInvalidInput should be defined as a sentinel error")
	}

	if ErrInternal == nil {
		t.Error("ErrInternal should be defined as a sentinel error")
	}
}