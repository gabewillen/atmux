package manager

import (
	"sync"
	"time"

	"github.com/agentflare-ai/amux/internal/remote/protocol"
)

// ReplayBuffer is a thread-safe ring buffer for PTY output replay.
type ReplayBuffer struct {
	capacity int64
	current  int64
	items    []*protocol.PTYIO
	mu       sync.RWMutex
	nextSeq  uint64
}

// NewReplayBuffer creates a new buffer with the given capacity in bytes.
func NewReplayBuffer(capacityBytes int64) *ReplayBuffer {
	return &ReplayBuffer{
		capacity: capacityBytes,
		items:    make([]*protocol.PTYIO, 0),
		nextSeq:  1,
	}
}

// Append adds data to the buffer, dropping oldest items if needed.
func (rb *ReplayBuffer) Append(sessionID string, data []byte) *protocol.PTYIO {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	item := &protocol.PTYIO{
		SessionID: sessionID,
		Data:      data, // In-memory reference, ensure copy if needed upstream
		Seq:       rb.nextSeq,
		Timestamp: time.Now().UTC(),
	}
	rb.nextSeq++

	itemSize := int64(len(data) + 64) // approx overhead
	rb.items = append(rb.items, item)
	rb.current += itemSize

	// Evict old items
	for rb.current > rb.capacity && len(rb.items) > 0 {
		oldest := rb.items[0]
		oldestSize := int64(len(oldest.Data) + 64)
		rb.current -= oldestSize
		rb.items = rb.items[1:]
	}

	return item
}

// Replay returns all items since the given sequence number (exclusive).
func (rb *ReplayBuffer) Replay(sinceSeq uint64) []*protocol.PTYIO {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	var result []*protocol.PTYIO
	for _, item := range rb.items {
		if item.Seq > sinceSeq {
			result = append(result, item)
		}
	}
	return result
}
