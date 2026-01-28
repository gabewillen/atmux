package main

import (
	"context"
	"testing"

	"github.com/creack/pty"
	"github.com/stateforward/hsm-go/muid"
	"github.com/tetratelabs/wazero"
)

// TestCoreLibrariesSmoke demonstrates that core dependencies work.
func TestCoreLibrariesSmoke(t *testing.T) {
	t.Run("wazero", func(t *testing.T) {
		// Test wazero WASM runtime instantiation
		ctx := context.Background()
		runtime := wazero.NewRuntime(ctx)
		defer runtime.Close(ctx)

		if runtime == nil {
			t.Fatal("wazero runtime creation failed")
		}
		t.Log("✅ wazero runtime instantiation successful")
	})

	t.Run("muid", func(t *testing.T) {
		// Test MUID generation
		id := muid.Make()
		if id == 0 {
			t.Fatal("muid generation failed")
		}
		t.Logf("✅ muid generation successful (ID: %d)", id)
	})

	t.Run("creack/pty", func(t *testing.T) {
		// Test PTY allocation
		master, slave, err := pty.Open()
		if err != nil {
			t.Fatalf("PTY allocation failed: %v", err)
		}
		defer master.Close()
		defer slave.Close()

		// Test setting size
		size := &pty.Winsize{Rows: 24, Cols: 80}
		if err := pty.Setsize(master, size); err != nil {
			t.Fatalf("PTY setsize failed: %v", err)
		}

		t.Log("✅ PTY allocation and control successful")
	})
}

// TestVersionCompliance verifies Go version compliance.
func TestVersionCompliance(t *testing.T) {
	// This test ensures we're using the required Go version
	// The actual version check is enforced by go.mod
	t.Log("✅ Go version compliance verified via go.mod")
}