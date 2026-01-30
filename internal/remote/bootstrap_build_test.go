package remote

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"runtime"
	"testing"

	"github.com/agentflare-ai/amux/pkg/api"
)

func TestBuildManagerBinary(t *testing.T) {
	path, err := buildManagerBinary(context.Background(), runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Fatalf("build manager: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(path)
	})
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected binary at %s: %v", path, err)
	}
}

func TestBuildBootstrapZipIncludesBinary(t *testing.T) {
	runner := stubSSHRunner{
		outputs: map[string][]byte{
			"uname -s": []byte(runtime.GOOS),
			"uname -m": []byte(runtime.GOARCH),
		},
	}
	req := BootstrapRequest{
		HostID: api.MustParseHostID("host"),
		Location: api.Location{
			Type: api.LocationSSH,
			Host: "example.com",
		},
		LeafURL:       "nats://leaf",
		CredsPath:     "/tmp/creds",
		SubjectPrefix: "amux",
		KVBucket:      "kv",
		ManagerModel:  "model",
	}
	zipBytes, err := buildBootstrapZip(context.Background(), req, runner)
	if err != nil {
		t.Fatalf("build zip: %v", err)
	}
	reader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatalf("zip reader: %v", err)
	}
	found := false
	for _, file := range reader.File {
		if file.Name == ".local/bin/amux-manager" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected manager binary in zip")
	}
}
