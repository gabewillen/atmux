package remote

import "sync"

type queuedMessage struct {
	subject string
	payload []byte
}

// Outbox buffers outbound publications while disconnected.
type Outbox struct {
	mu       sync.Mutex
	maxBytes int
	total    int
	queue    []queuedMessage
}

// NewOutbox constructs an outbox with a max payload size.
func NewOutbox(maxBytes int) *Outbox {
	if maxBytes < 0 {
		maxBytes = 0
	}
	return &Outbox{maxBytes: maxBytes}
}

// Enqueue buffers a subject payload pair, dropping oldest entries when full.
func (o *Outbox) Enqueue(subject string, payload []byte) {
	if o == nil || len(payload) == 0 {
		return
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.maxBytes == 0 {
		return
	}
	entry := queuedMessage{subject: subject, payload: append([]byte(nil), payload...)}
	o.queue = append(o.queue, entry)
	o.total += len(entry.payload)
	for o.total > o.maxBytes && len(o.queue) > 0 {
		o.total -= len(o.queue[0].payload)
		o.queue = o.queue[1:]
	}
}

// Drain returns buffered messages in enqueue order and clears the outbox.
func (o *Outbox) Drain() []queuedMessage {
	if o == nil {
		return nil
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if len(o.queue) == 0 {
		return nil
	}
	items := make([]queuedMessage, len(o.queue))
	copy(items, o.queue)
	o.queue = nil
	o.total = 0
	return items
}
