package remote

import (
	"testing"
)

func TestBootstrapConfig_SSHTarget(t *testing.T) {
	cfg := &BootstrapConfig{Host: "devbox", User: "alice"}
	got := cfg.SSHTarget()
	want := "alice@devbox"
	if got != want {
		t.Errorf("SSHTarget = %q, want %q", got, want)
	}
	cfg.User = ""
	got = cfg.SSHTarget()
	if got != "devbox" {
		t.Errorf("SSHTarget (no user) = %q, want devbox", got)
	}
}

func TestBootstrapConfig_SSHArgs(t *testing.T) {
	cfg := &BootstrapConfig{Host: "devbox", User: "alice", Port: 2222}
	args := cfg.SSHArgs()
	if len(args) < 2 {
		t.Fatalf("SSHArgs = %v", args)
	}
	if args[0] != "-p" || args[1] != "2222" {
		t.Errorf("SSHArgs = %v, want -p 2222 ...", args)
	}
}

func TestQuoteRemotePath(t *testing.T) {
	if got := quoteRemotePath("/tmp/foo"); got != "/tmp/foo" {
		t.Errorf("quoteRemotePath = %q", got)
	}
	if got := quoteRemotePath("/path with spaces"); got != "'/path with spaces'" {
		t.Errorf("quoteRemotePath(space) = %q", got)
	}
}
