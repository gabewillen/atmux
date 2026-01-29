// Package remote implements PTY streaming and buffering for remote sessions.
package remote

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

// PTYData represents PTY input/output data.
type PTYData struct {
	SessionID string    `json:"session_id"`
	Data      []byte    `json:"data"`
	Timestamp time.Time `json:"timestamp"`
	Sequence  uint64    `json:"sequence"`
}

// PTYStreamer manages PTY streaming for remote sessions.
type PTYStreamer struct {
	nm         *NATSManager
	hostID     string
	buffers    map[string]*RingBuffer // sessionID -> buffer
	sequences  map[string]uint64      // sessionID -> sequence counter
	subs       map[string]*nats.Subscription
	mutex      sync.RWMutex
	bufferSize int
}

// NewPTYStreamer creates a new PTY streamer.
func NewPTYStreamer(nm *NATSManager, bufferSize int) *PTYStreamer {
	if bufferSize <= 0 {
		bufferSize = 16384 // Default 16KB
	}
	
	return &PTYStreamer{
		nm:         nm,
		hostID:     nm.GetHostID(),
		buffers:    make(map[string]*RingBuffer),
		sequences:  make(map[string]uint64),
		subs:       make(map[string]*nats.Subscription),
		bufferSize: bufferSize,
	}
}

// StartPTYStreaming starts PTY streaming for a session.
func (ps *PTYStreamer) StartPTYStreaming(sessionID string) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	
	// Create ring buffer for replay
	ps.buffers[sessionID] = NewRingBuffer(ps.bufferSize)
	ps.sequences[sessionID] = 0
	
	// Subscribe to PTY input for this session (manager role)
	if ps.nm.role == "manager" {
		inputSubject := ps.nm.Subject("pty", ps.hostID, sessionID, "in")
		inputSub, err := ps.nm.conn.Subscribe(inputSubject, ps.handlePTYInput)
		if err != nil {
			delete(ps.buffers, sessionID)
			delete(ps.sequences, sessionID)
			return fmt.Errorf("failed to subscribe to PTY input: %w", err)
		}
		ps.subs[sessionID+":in"] = inputSub
	}
	
	return nil
}

// StopPTYStreaming stops PTY streaming for a session.
func (ps *PTYStreamer) StopPTYStreaming(sessionID string) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	
	// Unsubscribe
	if sub, exists := ps.subs[sessionID+":in"]; exists {
		sub.Unsubscribe()
		delete(ps.subs, sessionID+":in")
	}
	
	if sub, exists := ps.subs[sessionID+":out"]; exists {
		sub.Unsubscribe()
		delete(ps.subs, sessionID+":out")
	}
	
	// Clean up buffers
	delete(ps.buffers, sessionID)
	delete(ps.sequences, sessionID)
}

// PublishPTYOutput publishes PTY output data to NATS.
func (ps *PTYStreamer) PublishPTYOutput(sessionID string, data []byte) error {
	if len(data) == 0 {
		return nil
	}
	
	ps.mutex.Lock()
	
	// Get or create sequence counter
	seq := ps.sequences[sessionID]
	ps.sequences[sessionID] = seq + 1
	
	// Add to ring buffer for replay
	if buffer, exists := ps.buffers[sessionID]; exists {
		ptyData := PTYData{
			SessionID: sessionID,
			Data:      data,
			Timestamp: time.Now().UTC(),
			Sequence:  seq,
		}
		buffer.Add(ptyData)
	}
	
	ps.mutex.Unlock()
	
	// Publish to NATS
	ptyData := PTYData{
		SessionID: sessionID,
		Data:      data,
		Timestamp: time.Now().UTC(),
		Sequence:  seq,
	}
	
	dataBytes, err := json.Marshal(ptyData)
	if err != nil {
		return fmt.Errorf("failed to marshal PTY data: %w", err)
	}
	
	// Check NATS payload size limits (typically 1MB)
	if len(dataBytes) > 1024*1024 {
		return ps.publishChunked(sessionID, data, seq)
	}
	
	subject := ps.nm.Subject("pty", ps.hostID, sessionID, "out")
	return ps.nm.conn.Publish(subject, dataBytes)
}

// publishChunked publishes large PTY data in chunks.
func (ps *PTYStreamer) publishChunked(sessionID string, data []byte, baseSeq uint64) error {
	chunkSize := 32 * 1024 // 32KB chunks
	subject := ps.nm.Subject("pty", ps.hostID, sessionID, "out")
	
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		
		chunk := data[i:end]
		ptyData := PTYData{
			SessionID: sessionID,
			Data:      chunk,
			Timestamp: time.Now().UTC(),
			Sequence:  baseSeq + uint64(i/chunkSize),
		}
		
		dataBytes, err := json.Marshal(ptyData)
		if err != nil {
			return fmt.Errorf("failed to marshal PTY chunk: %w", err)
		}
		
		if err := ps.nm.conn.Publish(subject, dataBytes); err != nil {
			return fmt.Errorf("failed to publish PTY chunk: %w", err)
		}
	}
	
	return nil
}

// handlePTYInput handles incoming PTY input from director.
func (ps *PTYStreamer) handlePTYInput(msg *nats.Msg) {
	var ptyData PTYData
	if err := json.Unmarshal(msg.Data, &ptyData); err != nil {
		return // Ignore malformed data
	}
	
	// TODO: Write to actual PTY session
	// For now, this is a placeholder
}

// SubscribeToPTYOutput subscribes to PTY output from a remote session (director role).
func (ps *PTYStreamer) SubscribeToPTYOutput(hostID, sessionID string, handler func(PTYData)) error {
	if ps.nm.role != "director" {
		return fmt.Errorf("PTY output subscription only for director role")
	}
	
	subject := ps.nm.Subject("pty", hostID, sessionID, "out")
	sub, err := ps.nm.conn.Subscribe(subject, func(msg *nats.Msg) {
		var ptyData PTYData
		if err := json.Unmarshal(msg.Data, &ptyData); err != nil {
			return // Ignore malformed data
		}
		handler(ptyData)
	})
	
	if err != nil {
		return fmt.Errorf("failed to subscribe to PTY output: %w", err)
	}
	
	ps.mutex.Lock()
	ps.subs[hostID+":"+sessionID+":out"] = sub
	ps.mutex.Unlock()
	
	return nil
}

// SendPTYInput sends input to a remote PTY session (director role).
func (ps *PTYStreamer) SendPTYInput(hostID, sessionID string, data []byte) error {
	if ps.nm.role != "director" {
		return fmt.Errorf("PTY input sending only for director role")
	}
	
	ptyData := PTYData{
		SessionID: sessionID,
		Data:      data,
		Timestamp: time.Now().UTC(),
		Sequence:  0, // Input doesn't need sequencing
	}
	
	dataBytes, err := json.Marshal(ptyData)
	if err != nil {
		return fmt.Errorf("failed to marshal PTY input: %w", err)
	}
	
	subject := ps.nm.Subject("pty", hostID, sessionID, "in")
	return ps.nm.conn.Publish(subject, dataBytes)
}

// ReplayPTYOutput replays buffered PTY output for a session.
func (ps *PTYStreamer) ReplayPTYOutput(sessionID string, handler func(PTYData)) error {
	ps.mutex.RLock()
	buffer, exists := ps.buffers[sessionID]
	ps.mutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("no buffer found for session %s", sessionID)
	}
	
	// Replay buffered data oldest to newest
	buffer.ForEach(func(item interface{}) {
		if ptyData, ok := item.(PTYData); ok {
			handler(ptyData)
		}
	})
	
	return nil
}

// Close cleans up PTY streaming resources.
func (ps *PTYStreamer) Close() {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	
	for _, sub := range ps.subs {
		if sub != nil {
			sub.Unsubscribe()
		}
	}
	
	ps.subs = make(map[string]*nats.Subscription)
	ps.buffers = make(map[string]*RingBuffer)
	ps.sequences = make(map[string]uint64)
}

// RingBuffer implements a thread-safe ring buffer for PTY data replay.
type RingBuffer struct {
	buffer []interface{}
	head   int
	tail   int
	size   int
	maxSize int
	mutex  sync.RWMutex
}

// NewRingBuffer creates a new ring buffer with the specified capacity.
func NewRingBuffer(capacity int) *RingBuffer {
	return &RingBuffer{
		buffer:  make([]interface{}, capacity),
		maxSize: capacity,
	}
}

// Add adds an item to the ring buffer.
func (rb *RingBuffer) Add(item interface{}) {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()
	
	rb.buffer[rb.tail] = item
	rb.tail = (rb.tail + 1) % rb.maxSize
	
	if rb.size < rb.maxSize {
		rb.size++
	} else {
		// Buffer is full, move head forward (drop oldest)
		rb.head = (rb.head + 1) % rb.maxSize
	}
}

// ForEach iterates over all items in the buffer from oldest to newest.
func (rb *RingBuffer) ForEach(fn func(interface{})) {
	rb.mutex.RLock()
	defer rb.mutex.RUnlock()
	
	if rb.size == 0 {
		return
	}
	
	current := rb.head
	for i := 0; i < rb.size; i++ {
		fn(rb.buffer[current])
		current = (current + 1) % rb.maxSize
	}
}

// Size returns the current number of items in the buffer.
func (rb *RingBuffer) Size() int {
	rb.mutex.RLock()
	defer rb.mutex.RUnlock()
	return rb.size
}

// Clear empties the ring buffer.
func (rb *RingBuffer) Clear() {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()
	rb.head = 0
	rb.tail = 0
	rb.size = 0
}