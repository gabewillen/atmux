package remote

import "sync"

// ReplayBuffer stores a bounded history of PTY output.
type ReplayBuffer struct {
	mu   sync.Mutex
	max  int
	data []byte
}

// NewReplayBuffer constructs a replay buffer with a maximum size.
func NewReplayBuffer(maxBytes int) *ReplayBuffer {
	if maxBytes < 0 {
		maxBytes = 0
	}
	return &ReplayBuffer{max: maxBytes}
}

// Enabled reports whether replay buffering is enabled.
func (b *ReplayBuffer) Enabled() bool {
	if b == nil {
		return false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.max > 0
}

// Add appends bytes to the replay buffer with ring semantics.
func (b *ReplayBuffer) Add(chunk []byte) {
	if b == nil || len(chunk) == 0 {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.max <= 0 {
		return
	}
	if len(chunk) >= b.max {
		b.data = append([]byte(nil), chunk[len(chunk)-b.max:]...)
		return
	}
	if len(b.data)+len(chunk) > b.max {
		over := len(b.data) + len(chunk) - b.max
		if over < len(b.data) {
			b.data = b.data[over:]
		} else {
			b.data = b.data[:0]
		}
	}
	b.data = append(b.data, chunk...)
}

// Snapshot returns a copy of the buffered bytes.
func (b *ReplayBuffer) Snapshot() []byte {
	if b == nil {
		return nil
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.data) == 0 {
		return nil
	}
	copyBuf := make([]byte, len(b.data))
	copy(copyBuf, b.data)
	return copyBuf
}
