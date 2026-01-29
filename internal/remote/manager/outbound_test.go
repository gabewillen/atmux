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

func TestNewOutboundBuffer(t *testing.T) {
	tests := []struct {
		name     string
		maxBytes int64
	}{
		{"small", 100},
		{"medium", 1024 * 1024},
		{"large", 10 * 1024 * 1024},
		{"zero", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := NewOutboundBuffer(tt.maxBytes)
			if buf == nil {
				t.Fatal("NewOutboundBuffer returned nil")
			}
			if buf.Len() != 0 {
				t.Errorf("initial Len = %d, want 0", buf.Len())
			}
			if buf.TotalBytes() != 0 {
				t.Errorf("initial TotalBytes = %d, want 0", buf.TotalBytes())
			}
		})
	}
}

func TestOutboundBufferTotalBytesReflectsSize(t *testing.T) {
	buf := NewOutboundBuffer(1000)

	// Add entries and check TotalBytes accumulates correctly
	buf.Enqueue("a", []byte("123"))      // 3 bytes
	if buf.TotalBytes() != 3 {
		t.Errorf("after 1st enqueue: TotalBytes = %d, want 3", buf.TotalBytes())
	}

	buf.Enqueue("b", []byte("4567"))     // 4 bytes
	if buf.TotalBytes() != 7 {
		t.Errorf("after 2nd enqueue: TotalBytes = %d, want 7", buf.TotalBytes())
	}

	buf.Enqueue("c", []byte("89"))       // 2 bytes
	if buf.TotalBytes() != 9 {
		t.Errorf("after 3rd enqueue: TotalBytes = %d, want 9", buf.TotalBytes())
	}
}

func TestOutboundBufferFIFOEvictionMultiple(t *testing.T) {
	// Buffer can hold 10 bytes total
	buf := NewOutboundBuffer(10)

	buf.Enqueue("s1", []byte("aaa"))  // 3 bytes, total=3
	buf.Enqueue("s2", []byte("bbb"))  // 3 bytes, total=6
	buf.Enqueue("s3", []byte("ccc"))  // 3 bytes, total=9

	// This 5-byte entry needs 5 bytes free, but only 1 free.
	// Must drop s1 (3 bytes, now 4 free) and s2 (3 bytes, now 7 free).
	buf.Enqueue("s4", []byte("ddddd")) // 5 bytes

	if buf.Len() != 2 {
		t.Fatalf("Len = %d, want 2", buf.Len())
	}

	var subjects []string
	buf.FlushTo(func(subject string, data []byte) {
		subjects = append(subjects, subject)
	})

	if len(subjects) != 2 {
		t.Fatalf("flushed len = %d, want 2", len(subjects))
	}
	if subjects[0] != "s3" {
		t.Errorf("subjects[0] = %q, want s3", subjects[0])
	}
	if subjects[1] != "s4" {
		t.Errorf("subjects[1] = %q, want s4", subjects[1])
	}
}

func TestOutboundBufferOversizedEntryEvictsExistingThenDrops(t *testing.T) {
	// Per the implementation, the eviction loop runs before the oversized check.
	// When an oversized entry is enqueued, existing entries are evicted first
	// as the loop tries to make room, and then the oversized entry is dropped.
	buf := NewOutboundBuffer(10)

	buf.Enqueue("s1", []byte("hello")) // 5 bytes

	// Try to add an entry that exceeds max capacity.
	// The eviction loop will drop s1 trying to make room,
	// then the oversized check will drop the new entry.
	buf.Enqueue("s2", []byte("this is way too long for the buffer"))

	// Both s1 was evicted and s2 was dropped as oversized
	if buf.Len() != 0 {
		t.Fatalf("Len = %d, want 0", buf.Len())
	}
	if buf.TotalBytes() != 0 {
		t.Errorf("TotalBytes = %d, want 0", buf.TotalBytes())
	}
}

func TestOutboundBufferFlushDrainsClearsState(t *testing.T) {
	buf := NewOutboundBuffer(100)

	buf.Enqueue("s1", []byte("data1"))
	buf.Enqueue("s2", []byte("data2"))
	buf.Enqueue("s3", []byte("data3"))

	// Flush everything
	var count int
	buf.FlushTo(func(subject string, data []byte) {
		count++
	})

	if count != 3 {
		t.Errorf("flushed count = %d, want 3", count)
	}

	// After flush, buffer should be clean
	if buf.Len() != 0 {
		t.Errorf("Len after flush = %d, want 0", buf.Len())
	}
	if buf.TotalBytes() != 0 {
		t.Errorf("TotalBytes after flush = %d, want 0", buf.TotalBytes())
	}

	// Flushing again should yield nothing
	var count2 int
	buf.FlushTo(func(subject string, data []byte) {
		count2++
	})
	if count2 != 0 {
		t.Errorf("second flush count = %d, want 0", count2)
	}
}

func TestOutboundBufferLenConcurrency(t *testing.T) {
	buf := NewOutboundBuffer(10000)

	// Verify Len is safe to call (mutex protected)
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			buf.Enqueue("s", []byte("x"))
		}
		close(done)
	}()

	// Call Len concurrently
	for i := 0; i < 100; i++ {
		_ = buf.Len()
		_ = buf.TotalBytes()
	}

	<-done

	if buf.Len() != 100 {
		t.Errorf("final Len = %d, want 100", buf.Len())
	}
}

func TestOutboundBufferEmptyData(t *testing.T) {
	buf := NewOutboundBuffer(100)

	// Enqueue entries with empty data
	buf.Enqueue("s1", []byte{})
	buf.Enqueue("s2", nil)

	if buf.Len() != 2 {
		t.Errorf("Len = %d, want 2", buf.Len())
	}
	if buf.TotalBytes() != 0 {
		t.Errorf("TotalBytes = %d, want 0", buf.TotalBytes())
	}
}

func TestOutboundBufferExactCapacity(t *testing.T) {
	buf := NewOutboundBuffer(10)

	// Fill exactly to capacity
	buf.Enqueue("s1", []byte("12345"))  // 5 bytes
	buf.Enqueue("s2", []byte("67890"))  // 5 bytes, total = 10

	if buf.Len() != 2 {
		t.Errorf("Len = %d, want 2", buf.Len())
	}
	if buf.TotalBytes() != 10 {
		t.Errorf("TotalBytes = %d, want 10", buf.TotalBytes())
	}

	// Verify both entries are present in FIFO order
	var subjects []string
	buf.FlushTo(func(subject string, data []byte) {
		subjects = append(subjects, subject)
	})
	if len(subjects) != 2 {
		t.Fatalf("flushed len = %d, want 2", len(subjects))
	}
	if subjects[0] != "s1" {
		t.Errorf("subjects[0] = %q, want s1", subjects[0])
	}
	if subjects[1] != "s2" {
		t.Errorf("subjects[1] = %q, want s2", subjects[1])
	}
}

func TestOutboundBufferPerSubjectOrder(t *testing.T) {
	buf := NewOutboundBuffer(1000)

	// Interleave messages on two subjects
	buf.Enqueue("alpha", []byte("a1"))
	buf.Enqueue("beta", []byte("b1"))
	buf.Enqueue("alpha", []byte("a2"))
	buf.Enqueue("beta", []byte("b2"))
	buf.Enqueue("alpha", []byte("a3"))

	// Per spec: per-subject publish order MUST be preserved
	var alphaOrder []string
	var betaOrder []string
	buf.FlushTo(func(subject string, data []byte) {
		switch subject {
		case "alpha":
			alphaOrder = append(alphaOrder, string(data))
		case "beta":
			betaOrder = append(betaOrder, string(data))
		}
	})

	// Verify alpha order
	expectedAlpha := []string{"a1", "a2", "a3"}
	if len(alphaOrder) != len(expectedAlpha) {
		t.Fatalf("alpha entries = %d, want %d", len(alphaOrder), len(expectedAlpha))
	}
	for i, v := range expectedAlpha {
		if alphaOrder[i] != v {
			t.Errorf("alpha[%d] = %q, want %q", i, alphaOrder[i], v)
		}
	}

	// Verify beta order
	expectedBeta := []string{"b1", "b2"}
	if len(betaOrder) != len(expectedBeta) {
		t.Fatalf("beta entries = %d, want %d", len(betaOrder), len(expectedBeta))
	}
	for i, v := range expectedBeta {
		if betaOrder[i] != v {
			t.Errorf("beta[%d] = %q, want %q", i, betaOrder[i], v)
		}
	}
}
