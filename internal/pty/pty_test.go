package pty

import (
	"os/exec"
	"testing"
	"time"
)

func TestPTYLifecycle(t *testing.T) {
	cmd := exec.Command("cat") // cat echoes input
	p, err := Start(cmd)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer p.Close()

	if err := p.Resize(24, 80); err != nil {
		t.Errorf("Resize failed: %v", err)
	}

	// Write something
	input := []byte("hello\n")
	if _, err := p.File.Write(input); err != nil {
		t.Errorf("Write failed: %v", err)
	}

	// Read back (with timeout)
	readCh := make(chan []byte)
	go func() {
		buf := make([]byte, 1024)
		n, _ := p.File.Read(buf)
		if n > 0 {
			readCh <- buf[:n]
		}
	}()

	select {
	case output := <-readCh:
		// cat should echo "hello\r\n" or similar (echo is on by default)
		if len(output) == 0 {
			t.Error("Read returned empty")
		}
	case <-time.After(1 * time.Second):
		t.Error("Read timeout")
	}
}
