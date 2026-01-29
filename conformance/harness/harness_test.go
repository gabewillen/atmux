package harness

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestHarnessSmoke runs the minimal conformance harness against locally built
// amux binaries. It is skipped automatically when the expected binaries are
// not present, so it is safe to run as part of the default test suite.
func TestHarnessSmoke(t *testing.T) {
	binDir := filepath.Join("..", "..", "bin")

	if _, err := os.Stat(filepath.Join(binDir, "amux")); err != nil {
		t.Skipf("amux binary not found in %s; run `make build` first", binDir)
	}
	if _, err := os.Stat(filepath.Join(binDir, "amux-node")); err != nil {
		t.Skipf("amux-node binary not found in %s; run `make build` first", binDir)
	}

	t.Setenv("AMUX_BIN_DIR", binDir)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	res, err := Run(ctx)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if res.SpecVersion != "v1.22" {
		t.Fatalf("unexpected SpecVersion: got %q, want %q", res.SpecVersion, "v1.22")
	}

	outPath := filepath.Join(t.TempDir(), "conformance.json")
	if err := WriteResults(res, outPath); err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}
}
