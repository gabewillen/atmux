// ringbuffer.go implements a per-session replay ring buffer per spec §5.5.7.3.
// Buffer is capped at remote.buffer_size; when exceeded, oldest bytes are dropped.
package remote

import (
	"sync"
)

// RingBuffer is a fixed-capacity byte ring buffer for PTY replay (spec §5.5.7.3).
// When cap is exceeded, oldest bytes are dropped. Safe for concurrent read/write.
type RingBuffer struct {
	mu     sync.Mutex
	buf    []byte
	cap    int
	start  int
	length int
}

// NewRingBuffer creates a ring buffer with the given capacity in bytes.
// If cap <= 0, the buffer is disabled (Write is a no-op, Snapshot returns nil).
func NewRingBuffer(cap int) *RingBuffer {
	if cap <= 0 {
		return &RingBuffer{cap: 0}
	}
	return &RingBuffer{
		buf: make([]byte, cap),
		cap: cap,
	}
}

// Write appends data to the buffer. If the buffer would exceed capacity, the oldest bytes are dropped.
func (r *RingBuffer) Write(p []byte) (n int, err error) {
	if r.cap <= 0 {
		return len(p), nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, b := range p {
		if r.length < r.cap {
			pos := (r.start + r.length) % r.cap
			r.buf[pos] = b
			r.length++
		} else {
			r.buf[r.start] = b
			r.start = (r.start + 1) % r.cap
		}
		n++
	}
	return n, nil
}

// Snapshot returns a copy of the buffer contents in oldest-to-newest order.
// Used when handling a replay request; caller must not modify the returned slice.
func (r *RingBuffer) Snapshot() []byte {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cap <= 0 || r.length == 0 {
		return nil
	}
	out := make([]byte, r.length)
	for i := 0; i < r.length; i++ {
		out[i] = r.buf[(r.start+i)%r.cap]
	}
	return out
}

// Len returns the current number of bytes in the buffer.
func (r *RingBuffer) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.length
}

// Cap returns the buffer capacity in bytes (0 if disabled).
func (r *RingBuffer) Cap() int {
	return r.cap
}
