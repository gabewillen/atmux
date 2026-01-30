package remote

import (
	"context"
	"os/exec"
	"testing"
)

func TestExecSSHRunner(t *testing.T) {
	if _, err := exec.LookPath("ssh"); err != nil {
		t.Skip("ssh not available")
	}
	runner := ExecSSHRunner{}
	if err := runner.Run(context.Background(), "invalid-host", nil, "true", nil); err == nil {
		t.Fatalf("expected ssh run error")
	}
	if _, err := runner.RunOutput(context.Background(), "invalid-host", nil, "true", nil); err == nil {
		t.Fatalf("expected ssh run output error")
	}
}
