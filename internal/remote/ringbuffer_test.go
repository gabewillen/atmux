package remote

import (
	"bytes"
	"testing"
)

func TestRingBuffer_WriteAndSnapshot(t *testing.T) {
	rb := NewRingBuffer(10)
	rb.Write([]byte("hello"))
	snap := rb.Snapshot()
	if !bytes.Equal(snap, []byte("hello")) {
		t.Errorf("Snapshot() = %q, want hello", snap)
	}
	if rb.Len() != 5 {
		t.Errorf("Len() = %d, want 5", rb.Len())
	}
}

func TestRingBuffer_Overflow(t *testing.T) {
	rb := NewRingBuffer(5)
	rb.Write([]byte("12345"))
	rb.Write([]byte("678"))
	snap := rb.Snapshot()
	// Oldest to newest: after overflow we keep last 5 bytes 4,5,6,7,8
	want := []byte("45678")
	if !bytes.Equal(snap, want) {
		t.Errorf("Snapshot() = %q, want %q", snap, want)
	}
	if rb.Len() != 5 {
		t.Errorf("Len() = %d, want 5", rb.Len())
	}
}

func TestRingBuffer_Disabled(t *testing.T) {
	rb := NewRingBuffer(0)
	rb.Write([]byte("hello"))
	snap := rb.Snapshot()
	if snap != nil {
		t.Errorf("Snapshot() = %v, want nil", snap)
	}
	if rb.Cap() != 0 {
		t.Errorf("Cap() = %d, want 0", rb.Cap())
	}
}
