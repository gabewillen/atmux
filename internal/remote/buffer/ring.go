// Package buffer provides a ring buffer for PTY output replay.
//
// The ring buffer retains the most recent N bytes of PTY output,
// dropping the oldest bytes when the capacity is exceeded.
// It is used by the manager-role daemon to support the replay
// request-reply protocol.
//
// See spec §5.5.7.3 for replay buffer requirements.
package buffer

import "sync"

// Ring is a thread-safe ring buffer that retains the most recent cap bytes.
//
// When the buffer is full, writing additional bytes drops the oldest
// bytes (ring-buffer semantics). The buffer supports taking a snapshot
// of the current contents in oldest-to-newest byte order.
//
// A zero-capacity Ring disables buffering: writes are discarded and
// Snapshot returns nil.
type Ring struct {
	mu   sync.Mutex
	buf  []byte
	cap  int64
	head int64 // write position (circular)
	size int64 // current bytes stored
}

// NewRing creates a Ring with the given capacity in bytes.
// If cap is 0, buffering is disabled.
func NewRing(cap int64) *Ring {
	if cap <= 0 {
		return &Ring{cap: 0}
	}
	return &Ring{
		buf: make([]byte, cap),
		cap: cap,
	}
}

// Write appends data to the ring buffer.
// If the buffer is disabled (cap == 0), the data is discarded.
// If data exceeds the buffer capacity, only the last cap bytes are retained.
func (r *Ring) Write(data []byte) {
	if r.cap == 0 || len(data) == 0 {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	d := data
	// If data is larger than capacity, only keep the tail
	if int64(len(d)) >= r.cap {
		d = d[int64(len(d))-r.cap:]
		// Overwrite the entire buffer
		copy(r.buf, d)
		r.head = 0
		r.size = r.cap
		return
	}

	n := int64(len(d))
	// Write in one or two segments depending on wrap-around
	firstLen := r.cap - r.head
	if firstLen >= n {
		// No wrap
		copy(r.buf[r.head:], d)
	} else {
		// Wrap around
		copy(r.buf[r.head:], d[:firstLen])
		copy(r.buf, d[firstLen:])
	}

	r.head = (r.head + n) % r.cap
	r.size += n
	if r.size > r.cap {
		r.size = r.cap
	}
}

// Snapshot returns a copy of the current buffer contents in oldest-to-newest
// byte order. Returns nil if the buffer is disabled or empty.
//
// This is used to fulfill replay requests per spec §5.5.7.3:
// "The replayed bytes MUST correspond to a snapshot of the replay buffer
// taken at the moment the daemon receives the replay request."
func (r *Ring) Snapshot() []byte {
	if r.cap == 0 {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.size == 0 {
		return nil
	}

	result := make([]byte, r.size)

	if r.size < r.cap {
		// Buffer hasn't wrapped yet; data starts at 0
		copy(result, r.buf[:r.size])
	} else {
		// Buffer has wrapped; oldest data starts at head
		firstLen := r.cap - r.head
		copy(result, r.buf[r.head:])
		copy(result[firstLen:], r.buf[:r.head])
	}

	return result
}

// Len returns the current number of bytes stored.
func (r *Ring) Len() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.size
}

// Cap returns the buffer capacity in bytes.
func (r *Ring) Cap() int64 {
	return r.cap
}

// Enabled returns true if the buffer has a non-zero capacity.
func (r *Ring) Enabled() bool {
	return r.cap > 0
}

// Reset clears the buffer contents without changing capacity.
func (r *Ring) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.head = 0
	r.size = 0
}
