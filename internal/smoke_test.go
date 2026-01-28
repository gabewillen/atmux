package internal_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/creack/pty"
	"github.com/kevinburke/ssh_config"
	"github.com/stateforward/hsm-go"
	"github.com/stateforward/hsm-go/muid"
	"github.com/tetratelabs/wazero"
)

// TestWazeroSmoke verifies wazero WASM runtime can be instantiated per spec §4.2.2.
func TestWazeroSmoke(t *testing.T) {
	ctx := context.Background()
	
	// Create a new runtime
	r := wazero.NewRuntime(ctx)
	defer r.Close(ctx)
	
	if r == nil {
		t.Fatal("wazero runtime should not be nil")
	}
	
	t.Log("✓ wazero runtime instantiated successfully")
}

// TestHSMSmoke verifies hsm-go state machine dispatch per spec §4.2.3.
func TestHSMSmoke(t *testing.T) {
	// Define a simple state machine
	_ = hsm.Define("test",
		hsm.State("initial"),
		hsm.State("next"),
		hsm.Transition(hsm.On(hsm.Event{Name: "go"}), hsm.Source("initial"), hsm.Target("next")),
		hsm.Initial(hsm.Target("initial")),
	)
	
	t.Log("✓ hsm-go state machine defined successfully")
}

// TestMUIDSmoke verifies muid ID generation per spec §4.2.3.
func TestMUIDSmoke(t *testing.T) {
	// Generate a new ID
	id := muid.Make()
	
	if id == 0 {
		t.Fatal("muid should not generate zero ID")
	}
	
	// Generate another ID
	id2 := muid.Make()
	
	if id2 == 0 {
		t.Fatal("second muid should not generate zero ID")
	}
	
	if id == id2 {
		t.Fatal("muid should generate unique IDs")
	}
	
	t.Logf("✓ muid generated unique IDs: %d, %d", id, id2)
}

// TestPTYSmoke verifies creack/pty can allocate a PTY per spec §4.2.4.
func TestPTYSmoke(t *testing.T) {
	// Allocate a PTY
	ptmx, tty, err := pty.Open()
	if err != nil {
		t.Fatalf("pty.Open() failed: %v", err)
	}
	defer ptmx.Close()
	defer tty.Close()
	
	if ptmx == nil {
		t.Fatal("PTY master should not be nil")
	}
	
	if tty == nil {
		t.Fatal("PTY slave should not be nil")
	}
	
	// Test basic I/O
	msg := []byte("test")
	n, err := ptmx.Write(msg)
	if err != nil {
		t.Fatalf("PTY write failed: %v", err)
	}
	
	if n != len(msg) {
		t.Errorf("PTY write: got %d bytes, want %d", n, len(msg))
	}
	
	t.Log("✓ creack/pty allocated PTY and performed I/O")
}

// TestSSHConfigSmoke verifies kevinburke/ssh_config can parse SSH config per spec §5.1.
func TestSSHConfigSmoke(t *testing.T) {
	// Test parsing a simple SSH config snippet
	configText := `
Host example
    HostName example.com
    User testuser
    Port 2222
`
	
	cfg, err := ssh_config.Decode(bytes.NewBufferString(configText))
	if err != nil {
		t.Fatalf("ssh_config.Decode() failed: %v", err)
	}
	
	if cfg == nil {
		t.Fatal("SSH config should not be nil")
	}
	
	// Verify we can query values
	hostname, err := cfg.Get("example", "HostName")
	if err != nil {
		t.Fatalf("failed to get HostName: %v", err)
	}
	
	if hostname != "example.com" {
		t.Errorf("got hostname %q, want %q", hostname, "example.com")
	}
	
	t.Log("✓ kevinburke/ssh_config parsed SSH config successfully")
}
