package remote

import (
	"sync"
)

// ReplayBuffer is a thread-safe ring buffer for PTY output.
type ReplayBuffer struct {
	mu       sync.RWMutex
	buffer   []byte
	capacity int
	head     int // Points to the next write position
	full     bool
}

// NewReplayBuffer creates a new replay buffer with the given capacity.
func NewReplayBuffer(capacity int) *ReplayBuffer {
	if capacity <= 0 {
		return nil // Disabled
	}
	return &ReplayBuffer{
		buffer:   make([]byte, capacity),
		capacity: capacity,
	}
}

// Write appends data to the buffer, overwriting old data if full.
func (rb *ReplayBuffer) Write(p []byte) (n int, err error) {
	if rb == nil || rb.capacity == 0 {
		return len(p), nil
	}

	rb.mu.Lock()
	defer rb.mu.Unlock()

	n = len(p)
	if n > rb.capacity {
		// If input is larger than capacity, we only keep the last capacity bytes
		p = p[n-rb.capacity:]
		n = len(p)
	}

	// Simple ring buffer write
	// If head + n fits in remainder
	if rb.head+n <= rb.capacity {
		copy(rb.buffer[rb.head:], p)
		rb.head += n
		if rb.head == rb.capacity {
			rb.head = 0
			rb.full = true
		}
	} else {
		// Split write
		firstChunk := rb.capacity - rb.head
		copy(rb.buffer[rb.head:], p[:firstChunk])
		
		secondChunk := n - firstChunk
		copy(rb.buffer[0:], p[firstChunk:])
		
		rb.head = secondChunk
		rb.full = true
	}

	return len(p), nil // Return original length to satisfy io.Writer
}

// Bytes returns the current content of the buffer ordered from oldest to newest.
func (rb *ReplayBuffer) Bytes() []byte {
	if rb == nil || rb.capacity == 0 {
		return nil
	}

	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if !rb.full {
		// Buffer not full, valid data is [0 : head]
		out := make([]byte, rb.head)
		copy(out, rb.buffer[:rb.head])
		return out
	}

	// Buffer is full, valid data is [head : capacity] + [0 : head]
	out := make([]byte, rb.capacity)
	copy(out, rb.buffer[rb.head:])
	copy(out[rb.capacity-rb.head:], rb.buffer[:rb.head])
	return out
}
