package remote

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/agentflare-ai/amux/pkg/api"
)

type stubSSHRunner struct {
	outputs map[string][]byte
}

func (s stubSSHRunner) Run(ctx context.Context, target string, options []string, command string, stdin []byte) error {
	_ = ctx
	_ = target
	_ = options
	_ = command
	_ = stdin
	return nil
}

func (s stubSSHRunner) RunOutput(ctx context.Context, target string, options []string, command string, stdin []byte) ([]byte, error) {
	_ = ctx
	_ = target
	_ = options
	_ = stdin
	if out, ok := s.outputs[command]; ok {
		return out, nil
	}
	return []byte(""), nil
}

func TestBootstrapHelpers(t *testing.T) {
	location := api.Location{Type: api.LocationSSH, Host: "example.com", User: "root", Port: 2222}
	if sshTarget(location) != "root@example.com" {
		t.Fatalf("unexpected ssh target")
	}
	opts := sshOptions(location)
	if len(opts) != 2 || opts[0] != "-p" {
		t.Fatalf("unexpected ssh options: %#v", opts)
	}
	if shellEscape("a'b") != "'a'\\''b'" {
		t.Fatalf("unexpected shell escape")
	}
	if _, err := mapGOOS("linux"); err != nil {
		t.Fatalf("mapGOOS: %v", err)
	}
	if _, err := mapGOARCH("x86_64"); err != nil {
		t.Fatalf("mapGOARCH: %v", err)
	}
	if _, err := mapGOOS("plan9"); err == nil {
		t.Fatalf("expected unsupported os")
	}
	if _, err := mapGOARCH("sparc"); err == nil {
		t.Fatalf("expected unsupported arch")
	}
	runner := stubSSHRunner{
		outputs: map[string][]byte{
			"uname -s": []byte("Linux"),
			"uname -m": []byte("x86_64"),
		},
	}
	goos, goarch, err := detectRemoteArch(context.Background(), location, runner)
	if err != nil || goos != "linux" || goarch != "amd64" {
		t.Fatalf("detect arch: %v %s %s", err, goos, goarch)
	}
	badRunner := stubSSHRunner{
		outputs: map[string][]byte{
			"uname -s": []byte("Plan9"),
			"uname -m": []byte("mips"),
		},
	}
	if _, _, err := detectRemoteArch(context.Background(), location, badRunner); err == nil {
		t.Fatalf("expected detect arch error")
	}
	if !isHubConnected([]byte("hub_connected=true")) {
		t.Fatalf("expected hub connected")
	}
}

func TestBootstrapConfigAndZipHelpers(t *testing.T) {
	req := BootstrapRequest{
		HostID:        api.MustParseHostID("host"),
		LeafURL:       "nats://leaf",
		CredsPath:     "/tmp/creds",
		SubjectPrefix: "amux",
		KVBucket:      "kv",
		ManagerModel:  "model",
		HubClientURL:  "nats://hub",
	}
	encoded, err := bootstrapConfig(req)
	if err != nil || len(encoded) == 0 {
		t.Fatalf("bootstrapConfig: %v", err)
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	if err := addZipBytes(zw, []byte("data"), "path/file.txt", 0o644); err != nil {
		t.Fatalf("addZipBytes: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	if err := addZipBytes(nil, []byte("data"), "path/file.txt", 0o644); err == nil {
		t.Fatalf("expected nil writer error")
	}
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.txt")
	if err := os.WriteFile(src, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}
	zw = zip.NewWriter(&buf)
	if err := addZipFile(zw, src, "dest.txt", 0o644); err != nil {
		t.Fatalf("addZipFile: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
}

func TestFindModuleRootFailure(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if _, err := findModuleRoot(); err == nil {
		t.Fatalf("expected module root error")
	}
}
