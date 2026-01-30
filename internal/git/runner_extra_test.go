package git

import (
	"context"
	"testing"
)

func TestRunnerRunErrors(t *testing.T) {
	var runner *Runner
	if _, err := runner.run(context.Background(), ""); err == nil {
		t.Fatalf("expected runner error")
	}
	runner = &Runner{}
	if _, err := runner.run(context.Background(), ""); err == nil {
		t.Fatalf("expected runner exec error")
	}
	if _, err := defaultExec(context.Background(), "", "status"); err == nil {
		t.Fatalf("expected repo required error")
	}
}
