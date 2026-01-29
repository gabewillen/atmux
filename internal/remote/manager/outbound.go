// Package manager - outbound.go provides buffering for cross-host publications
// during hub disconnection.
//
// Per spec §5.5.8: the manager-role node SHOULD buffer outbound publications
// while disconnected, up to a maximum queued payload size of remote.buffer_size
// bytes total across all buffered publications. Oldest publications are dropped
// first when the limit is exceeded. Per-subject publish order MUST be preserved.
package manager

import (
	"fmt"
	"os"
	"sync"
)

// outboundEntry holds a single buffered publication.
type outboundEntry struct {
	subject string
	data    []byte
}

// OutboundBuffer buffers cross-host NATS publications during hub disconnection.
//
// The buffer has a maximum total payload size. When exceeded, the oldest
// entries are dropped first (FIFO eviction). Per-subject order is preserved
// because entries are stored in global FIFO order.
type OutboundBuffer struct {
	mu       sync.Mutex
	entries  []outboundEntry
	totalLen int64
	maxLen   int64
}

// NewOutboundBuffer creates a new OutboundBuffer with the given capacity.
func NewOutboundBuffer(maxBytes int64) *OutboundBuffer {
	return &OutboundBuffer{
		maxLen: maxBytes,
	}
}

// Enqueue adds a publication to the buffer.
// If the total size exceeds maxLen, the oldest entries are dropped.
//
// Per spec §5.5.8: "MUST account queued size as the sum of NATS message
// payload lengths in bytes (excluding subject names and headers)."
func (b *OutboundBuffer) Enqueue(subject string, data []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	payloadLen := int64(len(data))

	// Drop oldest entries until we have room
	for b.totalLen+payloadLen > b.maxLen && len(b.entries) > 0 {
		dropped := b.entries[0]
		b.entries = b.entries[1:]
		b.totalLen -= int64(len(dropped.data))
	}

	// If a single entry exceeds max, drop it with a warning
	if payloadLen > b.maxLen {
		fmt.Fprintf(os.Stderr, "outbound buffer: dropping oversized entry (%d bytes > %d max) on subject %s\n",
			payloadLen, b.maxLen, subject)
		return
	}

	b.entries = append(b.entries, outboundEntry{
		subject: subject,
		data:    data,
	})
	b.totalLen += payloadLen
}

// FlushTo drains all buffered entries to the given publish function.
//
// Per spec §5.5.8: "Flush MUST be FIFO per subject. New publications
// generated while a flush is in progress MUST be appended after older
// buffered publications for that same subject."
func (b *OutboundBuffer) FlushTo(publish func(subject string, data []byte)) {
	b.mu.Lock()
	entries := b.entries
	b.entries = nil
	b.totalLen = 0
	b.mu.Unlock()

	for _, e := range entries {
		publish(e.subject, e.data)
	}
}

// Len returns the number of buffered entries.
func (b *OutboundBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.entries)
}

// TotalBytes returns the total buffered payload size.
func (b *OutboundBuffer) TotalBytes() int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.totalLen
}
