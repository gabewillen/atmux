package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestSentinelErrors_NonNil(t *testing.T) {
	sentinels := []struct {
		name string
		err  error
	}{
		{"ErrNotFound", ErrNotFound},
		{"ErrAlreadyExists", ErrAlreadyExists},
		{"ErrInvalidInput", ErrInvalidInput},
		{"ErrNotReady", ErrNotReady},
		{"ErrTimeout", ErrTimeout},
		{"ErrClosed", ErrClosed},
		{"ErrPermissionDenied", ErrPermissionDenied},
		{"ErrNotImplemented", ErrNotImplemented},
		{"ErrConfigNotFound", ErrConfigNotFound},
		{"ErrConfigInvalid", ErrConfigInvalid},
		{"ErrConfigParseError", ErrConfigParseError},
		{"ErrAgentNotFound", ErrAgentNotFound},
		{"ErrAgentAlreadyExists", ErrAgentAlreadyExists},
		{"ErrAgentNotRunning", ErrAgentNotRunning},
		{"ErrAgentSlugCollision", ErrAgentSlugCollision},
		{"ErrAdapterNotFound", ErrAdapterNotFound},
		{"ErrAdapterLoadFailed", ErrAdapterLoadFailed},
		{"ErrAdapterCallFailed", ErrAdapterCallFailed},
		{"ErrAdapterManifestInvalid", ErrAdapterManifestInvalid},
		{"ErrAdapterVersionIncompatible", ErrAdapterVersionIncompatible},
		{"ErrNotInRepository", ErrNotInRepository},
		{"ErrWorktreeCreateFailed", ErrWorktreeCreateFailed},
		{"ErrWorktreeRemoveFailed", ErrWorktreeRemoveFailed},
		{"ErrMergeConflict", ErrMergeConflict},
		{"ErrMergeFailed", ErrMergeFailed},
		{"ErrDirtyWorktree", ErrDirtyWorktree},
		{"ErrBranchNotFound", ErrBranchNotFound},
		{"ErrDetachedHead", ErrDetachedHead},
		{"ErrInvalidStrategy", ErrInvalidStrategy},
		{"ErrShutdownInProgress", ErrShutdownInProgress},
		{"ErrDrainTimeout", ErrDrainTimeout},
		{"ErrSessionNotFound", ErrSessionNotFound},
		{"ErrAgentAlreadyRunning", ErrAgentAlreadyRunning},
		{"ErrConnectionFailed", ErrConnectionFailed},
		{"ErrHandshakeFailed", ErrHandshakeFailed},
		{"ErrHostNotConnected", ErrHostNotConnected},
		{"ErrSessionConflict", ErrSessionConflict},
		{"ErrModelNotFound", ErrModelNotFound},
		{"ErrModelLoadFailed", ErrModelLoadFailed},
		{"ErrInferenceUnavailable", ErrInferenceUnavailable},
	}

	for _, tt := range sentinels {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("%s is nil, expected non-nil sentinel error", tt.name)
			}
		})
	}
}

func TestSentinelErrors_HaveMessages(t *testing.T) {
	sentinels := []struct {
		name string
		err  error
	}{
		{"ErrNotFound", ErrNotFound},
		{"ErrAlreadyExists", ErrAlreadyExists},
		{"ErrTimeout", ErrTimeout},
		{"ErrClosed", ErrClosed},
	}

	for _, tt := range sentinels {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.err.Error()
			if msg == "" {
				t.Errorf("%s.Error() returned empty string", tt.name)
			}
		})
	}
}

func TestWrap_NilError(t *testing.T) {
	result := Wrap("context", nil)
	if result != nil {
		t.Errorf("Wrap(context, nil) = %v, want nil", result)
	}
}

func TestWrap_NonNilError(t *testing.T) {
	original := errors.New("original error")
	wrapped := Wrap("operation failed", original)

	if wrapped == nil {
		t.Fatal("Wrap returned nil for non-nil error")
	}

	expected := "operation failed: original error"
	if wrapped.Error() != expected {
		t.Errorf("Wrap() = %q, want %q", wrapped.Error(), expected)
	}

	// Verify the chain preserves the original error
	if !errors.Is(wrapped, original) {
		t.Error("Wrap() should preserve original error in chain")
	}
}

func TestWrap_SentinelChain(t *testing.T) {
	wrapped := Wrap("loading config", ErrConfigNotFound)

	if !Is(wrapped, ErrConfigNotFound) {
		t.Error("Wrap with sentinel: errors.Is should find ErrConfigNotFound in chain")
	}

	expected := "loading config: configuration not found"
	if wrapped.Error() != expected {
		t.Errorf("Wrap() = %q, want %q", wrapped.Error(), expected)
	}
}

func TestWrapf_NilError(t *testing.T) {
	result := Wrapf(nil, "operation %s", "test")
	if result != nil {
		t.Errorf("Wrapf(nil, ...) = %v, want nil", result)
	}
}

func TestWrapf_FormattedContext(t *testing.T) {
	original := errors.New("disk full")
	wrapped := Wrapf(original, "writing file %q to %s", "data.txt", "/tmp")

	if wrapped == nil {
		t.Fatal("Wrapf returned nil for non-nil error")
	}

	expected := `writing file "data.txt" to /tmp: disk full`
	if wrapped.Error() != expected {
		t.Errorf("Wrapf() = %q, want %q", wrapped.Error(), expected)
	}

	if !errors.Is(wrapped, original) {
		t.Error("Wrapf() should preserve original error in chain")
	}
}

func TestIs_DirectMatch(t *testing.T) {
	if !Is(ErrNotFound, ErrNotFound) {
		t.Error("Is(ErrNotFound, ErrNotFound) should be true")
	}
}

func TestIs_WrappedMatch(t *testing.T) {
	wrapped := fmt.Errorf("lookup failed: %w", ErrNotFound)
	if !Is(wrapped, ErrNotFound) {
		t.Error("Is should find ErrNotFound through wrapping")
	}
}

func TestIs_NoMatch(t *testing.T) {
	if Is(ErrNotFound, ErrTimeout) {
		t.Error("Is(ErrNotFound, ErrTimeout) should be false")
	}
}

func TestIs_NilTarget(t *testing.T) {
	if Is(ErrNotFound, nil) {
		t.Error("Is(ErrNotFound, nil) should be false")
	}
}

func TestIs_NilErr(t *testing.T) {
	if Is(nil, ErrNotFound) {
		t.Error("Is(nil, ErrNotFound) should be false")
	}
}

func TestIs_BothNil(t *testing.T) {
	if !Is(nil, nil) {
		t.Error("Is(nil, nil) should be true")
	}
}

func TestAs_WithTypedError(t *testing.T) {
	custom := &customError{Code: 42}

	wrapped := fmt.Errorf("wrapped: %w", custom)

	var target *customError
	if !As(wrapped, &target) {
		t.Fatal("As should find *customError in chain")
	}
	if target.Code != 42 {
		t.Errorf("As target.Code = %d, want 42", target.Code)
	}
}

func (e *customError) Error() string {
	return fmt.Sprintf("custom error code %d", e.Code)
}

type customError struct {
	Code int
}

func TestAs_NoMatch(t *testing.T) {
	var target *customError
	if As(ErrNotFound, &target) {
		t.Error("As should return false for non-matching error type")
	}
}

func TestNew(t *testing.T) {
	err := New("test error")
	if err == nil {
		t.Fatal("New returned nil")
	}
	if err.Error() != "test error" {
		t.Errorf("New().Error() = %q, want %q", err.Error(), "test error")
	}
}

func TestJoin(t *testing.T) {
	err1 := New("error one")
	err2 := New("error two")

	joined := Join(err1, err2)
	if joined == nil {
		t.Fatal("Join returned nil")
	}

	if !Is(joined, err1) {
		t.Error("Join result should contain err1")
	}
	if !Is(joined, err2) {
		t.Error("Join result should contain err2")
	}
}

func TestJoin_AllNil(t *testing.T) {
	joined := Join(nil, nil)
	if joined != nil {
		t.Errorf("Join(nil, nil) = %v, want nil", joined)
	}
}

func TestAgentNotFound_IsNotFound(t *testing.T) {
	// ErrAgentNotFound wraps ErrNotFound
	if !Is(ErrAgentNotFound, ErrNotFound) {
		t.Error("ErrAgentNotFound should match ErrNotFound via Is()")
	}
}

func TestAdapterNotFound_IsNotFound(t *testing.T) {
	// ErrAdapterNotFound wraps ErrNotFound
	if !Is(ErrAdapterNotFound, ErrNotFound) {
		t.Error("ErrAdapterNotFound should match ErrNotFound via Is()")
	}
}

func TestSessionNotFound_IsNotFound(t *testing.T) {
	// ErrSessionNotFound wraps ErrNotFound
	if !Is(ErrSessionNotFound, ErrNotFound) {
		t.Error("ErrSessionNotFound should match ErrNotFound via Is()")
	}
}

func TestAgentAlreadyExists_IsAlreadyExists(t *testing.T) {
	// ErrAgentAlreadyExists wraps ErrAlreadyExists
	if !Is(ErrAgentAlreadyExists, ErrAlreadyExists) {
		t.Error("ErrAgentAlreadyExists should match ErrAlreadyExists via Is()")
	}
}

func TestWrap_DoubleWrap(t *testing.T) {
	base := ErrNotFound
	first := Wrap("lookup", base)
	second := Wrap("service", first)

	if !Is(second, ErrNotFound) {
		t.Error("double-wrapped error should still match ErrNotFound")
	}

	expected := "service: lookup: not found"
	if second.Error() != expected {
		t.Errorf("double-wrapped error = %q, want %q", second.Error(), expected)
	}
}
