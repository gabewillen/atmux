package manager

import "testing"

func TestOutboundBufferBasic(t *testing.T) {
	buf := NewOutboundBuffer(100)

	buf.Enqueue("subj1", []byte("hello"))
	buf.Enqueue("subj2", []byte("world"))

	if buf.Len() != 2 {
		t.Fatalf("Len = %d, want 2", buf.Len())
	}
	if buf.TotalBytes() != 10 {
		t.Fatalf("TotalBytes = %d, want 10", buf.TotalBytes())
	}
}

func TestOutboundBufferFlush(t *testing.T) {
	buf := NewOutboundBuffer(100)

	buf.Enqueue("subj1", []byte("hello"))
	buf.Enqueue("subj2", []byte("world"))

	var flushed []string
	buf.FlushTo(func(subject string, data []byte) {
		flushed = append(flushed, subject+":"+string(data))
	})

	if len(flushed) != 2 {
		t.Fatalf("flushed len = %d, want 2", len(flushed))
	}
	if flushed[0] != "subj1:hello" {
		t.Fatalf("flushed[0] = %q, want %q", flushed[0], "subj1:hello")
	}
	if flushed[1] != "subj2:world" {
		t.Fatalf("flushed[1] = %q, want %q", flushed[1], "subj2:world")
	}

	// Buffer should be empty after flush
	if buf.Len() != 0 {
		t.Fatalf("after flush Len = %d, want 0", buf.Len())
	}
}

func TestOutboundBufferDropOldest(t *testing.T) {
	buf := NewOutboundBuffer(10)

	buf.Enqueue("subj1", []byte("12345")) // 5 bytes
	buf.Enqueue("subj2", []byte("67890")) // 5 bytes, total = 10
	buf.Enqueue("subj3", []byte("abc"))   // 3 bytes, must drop oldest to fit

	// Should have dropped subj1 to make room
	var flushed []string
	buf.FlushTo(func(subject string, data []byte) {
		flushed = append(flushed, subject)
	})

	if len(flushed) != 2 {
		t.Fatalf("flushed len = %d, want 2", len(flushed))
	}
	if flushed[0] != "subj2" {
		t.Fatalf("flushed[0] = %q, want subj2", flushed[0])
	}
	if flushed[1] != "subj3" {
		t.Fatalf("flushed[1] = %q, want subj3", flushed[1])
	}
}

func TestOutboundBufferOversizedEntry(t *testing.T) {
	buf := NewOutboundBuffer(5)

	// Entry larger than max capacity should be dropped
	buf.Enqueue("subj1", []byte("toolongforthemax"))

	if buf.Len() != 0 {
		t.Fatalf("oversized entry should be dropped, Len = %d", buf.Len())
	}
}

func TestOutboundBufferFIFOOrder(t *testing.T) {
	buf := NewOutboundBuffer(100)

	for i := 0; i < 5; i++ {
		buf.Enqueue("subj", []byte{byte('a' + i)})
	}

	var order []byte
	buf.FlushTo(func(subject string, data []byte) {
		order = append(order, data...)
	})

	if string(order) != "abcde" {
		t.Fatalf("flush order = %q, want %q", order, "abcde")
	}
}
