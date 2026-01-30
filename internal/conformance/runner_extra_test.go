package conformance

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type stubFixture struct {
	startErr error
	stopErr  error
}

func (s stubFixture) Start(ctx context.Context) error {
	_ = ctx
	return s.startErr
}

func (s stubFixture) Stop(ctx context.Context) error {
	_ = ctx
	return s.stopErr
}

func TestRunnerFlowErrors(t *testing.T) {
	runner := &Runner{
		SpecVersion: "v1",
		OutputPath:  filepath.Join(t.TempDir(), "results.json"),
		Daemon:      stubFixture{startErr: errors.New("boom")},
		Clock:       func() time.Time { return time.Unix(0, 0) },
	}
	if _, err := runner.Run(context.Background()); err != nil {
		t.Fatalf("run error: %v", err)
	}
	runner = &Runner{
		SpecVersion: "v1",
		OutputPath:  filepath.Join(t.TempDir(), "results.json"),
		Daemon:      stubFixture{stopErr: errors.New("boom")},
	}
	if _, err := runner.Run(context.Background()); err != nil {
		t.Fatalf("run error: %v", err)
	}
	runner = &Runner{
		SpecVersion: "v1",
		OutputPath:  filepath.Join(t.TempDir(), "results.json"),
		CLI:         stubFixture{startErr: errors.New("boom")},
	}
	if _, err := runner.Run(context.Background()); err != nil {
		t.Fatalf("run error: %v", err)
	}
}

func TestWriteResultsErrors(t *testing.T) {
	if err := writeResults("", Results{}); err == nil {
		t.Fatalf("expected path error")
	}
	path := filepath.Join(t.TempDir(), "dir")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := writeResults(filepath.Join(path, "results.json"), Results{}); err == nil {
		t.Fatalf("expected mkdir error")
	}
}
