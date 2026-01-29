package buffer

import (
	"bytes"
	"testing"
)

func TestRingNewZeroCap(t *testing.T) {
	r := NewRing(0)
	if r.Enabled() {
		t.Fatal("zero-cap ring should not be enabled")
	}
	r.Write([]byte("hello"))
	if r.Len() != 0 {
		t.Fatalf("zero-cap ring Len = %d, want 0", r.Len())
	}
	if snap := r.Snapshot(); snap != nil {
		t.Fatalf("zero-cap ring Snapshot = %v, want nil", snap)
	}
}

func TestRingBasicWrite(t *testing.T) {
	r := NewRing(10)
	r.Write([]byte("hello"))
	if r.Len() != 5 {
		t.Fatalf("Len = %d, want 5", r.Len())
	}
	snap := r.Snapshot()
	if !bytes.Equal(snap, []byte("hello")) {
		t.Fatalf("Snapshot = %q, want %q", snap, "hello")
	}
}

func TestRingMultipleWrites(t *testing.T) {
	r := NewRing(10)
	r.Write([]byte("hel"))
	r.Write([]byte("lo"))
	if r.Len() != 5 {
		t.Fatalf("Len = %d, want 5", r.Len())
	}
	snap := r.Snapshot()
	if !bytes.Equal(snap, []byte("hello")) {
		t.Fatalf("Snapshot = %q, want %q", snap, "hello")
	}
}

func TestRingWrapAround(t *testing.T) {
	r := NewRing(5)
	r.Write([]byte("abc"))   // buf: [a,b,c,_,_] head=3 size=3
	r.Write([]byte("defg"))  // buf: [f,g,c,d,e] head=2 size=5 (wrapped)
	snap := r.Snapshot()
	if !bytes.Equal(snap, []byte("cdefg")) {
		t.Fatalf("Snapshot = %q, want %q", snap, "cdefg")
	}
}

func TestRingExactCapacity(t *testing.T) {
	r := NewRing(5)
	r.Write([]byte("abcde"))
	if r.Len() != 5 {
		t.Fatalf("Len = %d, want 5", r.Len())
	}
	snap := r.Snapshot()
	if !bytes.Equal(snap, []byte("abcde")) {
		t.Fatalf("Snapshot = %q, want %q", snap, "abcde")
	}
}

func TestRingOverflow(t *testing.T) {
	r := NewRing(5)
	r.Write([]byte("abcdefgh"))
	// Only last 5 bytes should be retained
	snap := r.Snapshot()
	if !bytes.Equal(snap, []byte("defgh")) {
		t.Fatalf("Snapshot = %q, want %q", snap, "defgh")
	}
}

func TestRingDropOldest(t *testing.T) {
	r := NewRing(4)
	r.Write([]byte("ab"))
	r.Write([]byte("cd"))
	// Full: [a,b,c,d]
	r.Write([]byte("ef"))
	// Should drop a,b: [e,f,c,d] → oldest-to-newest: c,d,e,f
	snap := r.Snapshot()
	if !bytes.Equal(snap, []byte("cdef")) {
		t.Fatalf("Snapshot = %q, want %q", snap, "cdef")
	}
}

func TestRingReset(t *testing.T) {
	r := NewRing(10)
	r.Write([]byte("hello"))
	r.Reset()
	if r.Len() != 0 {
		t.Fatalf("after Reset, Len = %d, want 0", r.Len())
	}
	if snap := r.Snapshot(); snap != nil {
		t.Fatalf("after Reset, Snapshot = %v, want nil", snap)
	}
}

func TestRingSnapshotIsACopy(t *testing.T) {
	r := NewRing(10)
	r.Write([]byte("hello"))
	snap := r.Snapshot()
	// Mutate the snapshot; ring should be unaffected
	snap[0] = 'X'
	snap2 := r.Snapshot()
	if !bytes.Equal(snap2, []byte("hello")) {
		t.Fatalf("ring mutated by snapshot modification: got %q", snap2)
	}
}

func TestRingCap(t *testing.T) {
	r := NewRing(42)
	if r.Cap() != 42 {
		t.Fatalf("Cap = %d, want 42", r.Cap())
	}
}

func TestRingEmptySnapshot(t *testing.T) {
	r := NewRing(10)
	if snap := r.Snapshot(); snap != nil {
		t.Fatalf("empty ring Snapshot = %v, want nil", snap)
	}
}

func TestRingNegativeCap(t *testing.T) {
	r := NewRing(-1)
	if r.Enabled() {
		t.Fatal("negative-cap ring should not be enabled")
	}
}
