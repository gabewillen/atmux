package auth

import (
	"os"
	"testing"
)

func TestGenerateHostCredential(t *testing.T) {
	cred, err := GenerateHostCredential("test-host")
	if err != nil {
		t.Fatalf("GenerateHostCredential: %v", err)
	}

	if cred.HostID != "test-host" {
		t.Fatalf("HostID = %q, want %q", cred.HostID, "test-host")
	}

	if len(cred.Seed) == 0 {
		t.Fatal("Seed is empty")
	}

	if cred.PublicKey == "" {
		t.Fatal("PublicKey is empty")
	}

	// Public key should start with "U" (NATS user key prefix)
	if cred.PublicKey[0] != 'U' {
		t.Fatalf("PublicKey should start with 'U', got %q", cred.PublicKey[:1])
	}
}

func TestGenerateUniqueCredentials(t *testing.T) {
	cred1, err := GenerateHostCredential("host1")
	if err != nil {
		t.Fatalf("GenerateHostCredential(host1): %v", err)
	}

	cred2, err := GenerateHostCredential("host2")
	if err != nil {
		t.Fatalf("GenerateHostCredential(host2): %v", err)
	}

	// Credentials must be unique per host
	if cred1.PublicKey == cred2.PublicKey {
		t.Fatal("two hosts should have different public keys")
	}
}

func TestWriteCredsFile(t *testing.T) {
	cred, err := GenerateHostCredential("test-host")
	if err != nil {
		t.Fatalf("GenerateHostCredential: %v", err)
	}

	dir := t.TempDir()
	path, err := WriteCredsFile(cred, dir)
	if err != nil {
		t.Fatalf("WriteCredsFile: %v", err)
	}

	// Verify file exists
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}

	// Verify permissions are 0600
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Fatalf("file permissions = %o, want 0600", perm)
	}

	// Verify content contains the seed
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("creds file is empty")
	}
}

func TestHostSubjectPermissions(t *testing.T) {
	prefix := "amux"
	hostID := "devbox"

	pub, sub := HostSubjectPermissions(prefix, hostID)

	// Verify publish permissions per spec §5.5.6.4
	expectedPub := []string{
		"amux.handshake.devbox",
		"amux.events.devbox",
		"amux.pty.devbox.*.out",
		"amux.comm.director",
		"amux.comm.manager.*",
		"amux.comm.agent.*.>",
		"amux.comm.broadcast",
	}

	if len(pub) != len(expectedPub) {
		t.Fatalf("publish len = %d, want %d", len(pub), len(expectedPub))
	}
	for i, p := range pub {
		if p != expectedPub[i] {
			t.Fatalf("publish[%d] = %q, want %q", i, p, expectedPub[i])
		}
	}

	// Verify subscribe permissions per spec §5.5.6.4
	expectedSub := []string{
		"amux.ctl.devbox",
		"amux.pty.devbox.*.in",
		"amux.comm.manager.devbox",
		"amux.comm.agent.devbox.>",
		"amux.comm.broadcast",
		"_INBOX.>",
	}

	if len(sub) != len(expectedSub) {
		t.Fatalf("subscribe len = %d, want %d", len(sub), len(expectedSub))
	}
	for i, s := range sub {
		if s != expectedSub[i] {
			t.Fatalf("subscribe[%d] = %q, want %q", i, s, expectedSub[i])
		}
	}
}

func TestHostSubjectPermissionsCustomPrefix(t *testing.T) {
	pub, sub := HostSubjectPermissions("custom.prefix", "myhost")

	// Spot check with custom prefix
	if pub[0] != "custom.prefix.handshake.myhost" {
		t.Fatalf("pub[0] = %q, want custom.prefix.handshake.myhost", pub[0])
	}
	if sub[0] != "custom.prefix.ctl.myhost" {
		t.Fatalf("sub[0] = %q, want custom.prefix.ctl.myhost", sub[0])
	}
}
