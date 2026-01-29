package pty

import (
	"os/exec"
	"testing"
)

func TestPTY_Lifecycle(t *testing.T) {
	cmd := exec.Command("echo", "hello")
	p, err := Start(cmd)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer p.Close()

	if err := p.Resize(24, 80); err != nil {
		t.Errorf("Resize failed: %v", err)
	}
	
	// Wait for command
	if err := cmd.Wait(); err != nil {
		t.Errorf("Command wait failed: %v", err)
	}
}