package remote

import "testing"

func TestLoadOrCreatePeerID(t *testing.T) {
	dir := t.TempDir()
	id, err := LoadOrCreatePeerID(dir)
	if err != nil {
		t.Fatalf("load or create: %v", err)
	}
	id2, err := LoadOrCreatePeerID(dir)
	if err != nil {
		t.Fatalf("load peer id: %v", err)
	}
	if id.String() != id2.String() {
		t.Fatalf("expected same peer id")
	}
	if _, err := LoadOrCreatePeerID(""); err == nil {
		t.Fatalf("expected error for empty dir")
	}
}

