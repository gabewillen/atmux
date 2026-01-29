// Package protocol implements remote communication protocol (transports events)
package protocol

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

// ReplayBuffer manages the replay buffer for PTY output
type ReplayBuffer struct {
	buffer     *bytes.Buffer
	maxSize    int64
	mutex      sync.RWMutex
	disconnected bool
	bufferedPublications map[string][]*BufferedPublication
	maxBufferSize int64
}

// BufferedPublication represents a buffered NATS publication
type BufferedPublication struct {
	Subject string
	Data    []byte
	Size    int64
}

// NewReplayBuffer creates a new ReplayBuffer
func NewReplayBuffer(maxSize int64) *ReplayBuffer {
	return &ReplayBuffer{
		buffer:             &bytes.Buffer{},
		maxSize:            maxSize,
		bufferedPublications: make(map[string][]*BufferedPublication),
		maxBufferSize:      maxSize,
	}
}

// AddOutput adds PTY output to the replay buffer
func (rb *ReplayBuffer) AddOutput(data []byte) {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	if rb.maxSize == 0 {
		// Replay buffering is disabled
		return
	}

	// Calculate how much space we need to free
	currentSize := int64(rb.buffer.Len())
	dataSize := int64(len(data))
	
	if currentSize+dataSize > rb.maxSize {
		// Need to truncate the buffer
		excess := (currentSize + dataSize) - rb.maxSize
		newStart := int(excess)
		if newStart < rb.buffer.Len() {
			// Keep the tail of the buffer
			rb.buffer = bytes.NewBuffer(rb.buffer.Bytes()[newStart:])
		} else {
			// Clear the entire buffer
			rb.buffer.Reset()
		}
	}

	// Add the new data
	rb.buffer.Write(data)
}

// GetReplayData returns the current replay buffer content
func (rb *ReplayBuffer) GetReplayData() []byte {
	rb.mutex.RLock()
	defer rb.mutex.RUnlock()

	data := make([]byte, rb.buffer.Len())
	copy(data, rb.buffer.Bytes())
	return data
}

// Clear clears the replay buffer
func (rb *ReplayBuffer) Clear() {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()
	
	rb.buffer.Reset()
}

// SetDisconnected sets the disconnection state
func (rb *ReplayBuffer) SetDisconnected(disconnected bool) {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()
	
	rb.disconnected = disconnected
}

// IsDisconnected returns the disconnection state
func (rb *ReplayBuffer) IsDisconnected() bool {
	rb.mutex.RLock()
	defer rb.mutex.RUnlock()
	
	return rb.disconnected
}

// BufferPublication buffers a NATS publication during disconnection
func (rb *ReplayBuffer) BufferPublication(subject string, data []byte) error {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()
	
	if rb.maxSize == 0 {
		// Buffering is disabled
		return nil
	}

	publication := &BufferedPublication{
		Subject: subject,
		Data:    data,
		Size:    int64(len(data)),
	}

	// Check if we exceed the max buffer size
	totalSize := rb.getTotalBufferSize()
	if totalSize + publication.Size > rb.maxBufferSize {
		// Drop oldest publications first
		rb.dropOldestPublications(publication.Size)
	}

	// Add the publication to the buffer
	rb.bufferedPublications[subject] = append(rb.bufferedPublications[subject], publication)
	
	return nil
}

// getTotalBufferSize calculates the total size of buffered publications
func (rb *ReplayBuffer) getTotalBufferSize() int64 {
	var total int64
	for _, pubs := range rb.bufferedPublications {
		for _, pub := range pubs {
			total += pub.Size
		}
	}
	return total
}

// dropOldestPublications drops the oldest publications to make room for new ones
func (rb *ReplayBuffer) dropOldestPublications(requiredSpace int64) {
	// Simple implementation: drop from the beginning of each subject's queue
	// until we have enough space
	for subject, pubs := range rb.bufferedPublications {
		for len(pubs) > 0 && rb.getTotalBufferSize() > rb.maxBufferSize-requiredSpace {
			// Remove the first publication
			dropped := pubs[0]
			pubs = pubs[1:]
			
			// Update the map
			if len(pubs) == 0 {
				delete(rb.bufferedPublications, subject)
			} else {
				rb.bufferedPublications[subject] = pubs
			}
			
			// Adjust the required space
			requiredSpace -= dropped.Size
		}
	}
}

// FlushBuffers flushes all buffered publications
func (rb *ReplayBuffer) FlushBuffers(nc *nats.Conn) error {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()
	
	// Process each subject's buffered publications FIFO
	for subject, pubs := range rb.bufferedPublications {
		for _, pub := range pubs {
			if err := nc.Publish(subject, pub.Data); err != nil {
				return fmt.Errorf("failed to publish buffered message to %s: %w", subject, err)
			}
		}
		// Clear the buffered publications for this subject after flushing
		rb.bufferedPublications[subject] = nil
	}
	
	return nil
}

// ReconnectionManager handles reconnection logic
type ReconnectionManager struct {
	replayBuffers map[string]*ReplayBuffer  // Map of session_id -> replay buffer
	natsConn      *nats.Conn
	subjectBuilder *SubjectBuilder
	controlOps    *ControlOperations
	mutex         sync.RWMutex
}

// NewReconnectionManager creates a new ReconnectionManager
func NewReconnectionManager(nc *nats.Conn, sb *SubjectBuilder, co *ControlOperations) *ReconnectionManager {
	return &ReconnectionManager{
		replayBuffers: make(map[string]*ReplayBuffer),
		natsConn:      nc,
		subjectBuilder: sb,
		controlOps:    co,
	}
}

// AddSession adds a session to be managed for replay
func (rm *ReconnectionManager) AddSession(sessionID string, bufferSize int64) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	rm.replayBuffers[sessionID] = NewReplayBuffer(bufferSize)
}

// RemoveSession removes a session from replay management
func (rm *ReconnectionManager) RemoveSession(sessionID string) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	delete(rm.replayBuffers, sessionID)
}

// GetReplayBuffer returns the replay buffer for a session
func (rm *ReconnectionManager) GetReplayBuffer(sessionID string) *ReplayBuffer {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	return rm.replayBuffers[sessionID]
}

// HandleReconnection handles the reconnection process after hub connectivity is restored
func (rm *ReconnectionManager) HandleReconnection(ctx context.Context, hostID string, activeSessions []string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	fmt.Printf("Handling reconnection for host %s with %d active sessions\n", hostID, len(activeSessions))
	
	// For each active session, send replay request
	for _, sessionID := range activeSessions {
		replayPayload := ReplayPayload{
			SessionID: sessionID,
		}
		
		// Send replay request
		resp, err := rm.controlOps.Replay(ctx, hostID, replayPayload)
		if err != nil {
			fmt.Printf("Failed to send replay request for session %s: %v\n", sessionID, err)
			continue
		}
		
		if !resp.Accepted {
			fmt.Printf("Daemon declined replay for session %s\n", sessionID)
			continue
		}
		
		fmt.Printf("Replay accepted for session %s\n", sessionID)
	}
	
	// Flush any buffered publications
	for _, buffer := range rm.replayBuffers {
		if err := buffer.FlushBuffers(rm.natsConn); err != nil {
			return fmt.Errorf("failed to flush buffered publications: %w", err)
		}
	}
	
	return nil
}

// OnDisconnection marks all sessions as disconnected
func (rm *ReconnectionManager) OnDisconnection() {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	for _, buffer := range rm.replayBuffers {
		buffer.SetDisconnected(true)
	}
}

// OnReconnection marks all sessions as reconnected
func (rm *ReconnectionManager) OnReconnection() {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	for _, buffer := range rm.replayBuffers {
		buffer.SetDisconnected(false)
	}
}

// BufferPublicationDuringDisconnection buffers a publication if disconnected
func (rm *ReconnectionManager) BufferPublicationDuringDisconnection(sessionID, subject string, data []byte) error {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	buffer := rm.replayBuffers[sessionID]
	if buffer != nil && buffer.IsDisconnected() {
		return buffer.BufferPublication(subject, data)
	}
	
	return nil
}