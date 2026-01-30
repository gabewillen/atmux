package remote

import "testing"

func TestReplayBuffer(t *testing.T) {
	buf := NewReplayBuffer(5)
	if !buf.Enabled() {
		t.Fatalf("expected enabled")
	}
	buf.Add([]byte("abc"))
	buf.Add([]byte("def"))
	snap := buf.Snapshot()
	if string(snap) != "bcdef" {
		t.Fatalf("unexpected snapshot: %s", snap)
	}
	buf.Add([]byte("123456"))
	snap = buf.Snapshot()
	if string(snap) != "23456" {
		t.Fatalf("unexpected snapshot after overflow: %s", snap)
	}
}

func TestReplayBufferDisabled(t *testing.T) {
	var buf *ReplayBuffer
	if buf.Enabled() {
		t.Fatalf("expected nil buffer disabled")
	}
	buf = NewReplayBuffer(0)
	buf.Add([]byte("abc"))
	if buf.Snapshot() != nil {
		t.Fatalf("expected no snapshot")
	}
}

