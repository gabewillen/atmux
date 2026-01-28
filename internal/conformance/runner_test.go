package conformance

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRunnerWritesResults(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "results.json")
	runner := Runner{
		SpecVersion: "v1.22",
		OutputPath:  path,
		Daemon:      &NoopFixture{},
		CLI:         &NoopFixture{},
	}
	if _, err := runner.Run(context.Background()); err != nil {
		t.Fatalf("run: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("results missing: %v", err)
	}
}
