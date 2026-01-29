// Package protocol implements remote communication protocol (transports events)
package protocol

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// HandshakeHandler handles the handshake protocol between director and daemon
type HandshakeHandler struct {
	nc           *nats.Conn
	subjectBuilder *SubjectBuilder
	connectedHosts map[string]*PeerInfo
}

// PeerInfo holds information about a connected peer
type PeerInfo struct {
	HostID   string
	PeerID   string
	Role     string
	Version  string
	LastSeen time.Time
}

// NewHandshakeHandler creates a new HandshakeHandler
func NewHandshakeHandler(nc *nats.Conn, subjectBuilder *SubjectBuilder) *HandshakeHandler {
	return &HandshakeHandler{
		nc:             nc,
		subjectBuilder: subjectBuilder,
		connectedHosts: make(map[string]*PeerInfo),
	}
}

// StartListening starts listening for handshake requests
func (hh *HandshakeHandler) StartListening(hostID string) error {
	subject := hh.subjectBuilder.HandshakeSubject(hostID)
	
	_, err := hh.nc.Subscribe(subject, func(msg *nats.Msg) {
		hh.handleHandshakeRequest(msg, hostID)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to handshake subject %s: %w", subject, err)
	}

	return nil
}

// handleHandshakeRequest handles an incoming handshake request from a daemon
func (hh *HandshakeHandler) handleHandshakeRequest(msg *nats.Msg, expectedHostID string) {
	var req ControlMessage
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		hh.sendHandshakeError(msg.Reply, "handshake", "invalid_payload", "Failed to unmarshal handshake request")
		return
	}

	if req.Type != HandshakeType {
		hh.sendHandshakeError(msg.Reply, "handshake", "invalid_request_type", "Expected handshake type")
		return
	}

	var payload HandshakePayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		hh.sendHandshakeError(msg.Reply, "handshake", "invalid_payload", "Failed to unmarshal handshake payload")
		return
	}

	// Validate the host_id in the payload matches the expected host_id from the subject
	if payload.HostID != expectedHostID {
		hh.sendHandshakeError(msg.Reply, "handshake", "invalid_host_id", 
			fmt.Sprintf("Host ID mismatch: expected %s, got %s", expectedHostID, payload.HostID))
		return
	}

	// Check if protocol version is supported
	if payload.Protocol != 1 {
		hh.sendHandshakeError(msg.Reply, "handshake", "unsupported_protocol", 
			fmt.Sprintf("Unsupported protocol version: %d", payload.Protocol))
		return
	}

	// Check for host ID collision
	if _, exists := hh.connectedHosts[payload.HostID]; exists {
		hh.sendHandshakeError(msg.Reply, "handshake", "host_collision", 
			fmt.Sprintf("Host ID %s already connected", payload.HostID))
		return
	}

	// Create the response payload
	responsePayload := HandshakePayload{
		Protocol: 1,
		PeerID:   "1234", // This would be the director's peer ID in a real implementation
		Role:     "director",
		HostID:   expectedHostID, // Echo back the host ID
		Version:  "1.0.0", // This would be the actual version
	}

	// Send the handshake response
	response := ControlMessage{
		Type:    HandshakeType,
		Payload: mustMarshal(responsePayload),
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		hh.sendHandshakeError(msg.Reply, "handshake", "marshal_error", "Failed to marshal handshake response")
		return
	}

	if err := hh.nc.Publish(msg.Reply, responseData); err != nil {
		// Log error but can't send response if publish fails
		fmt.Printf("Failed to send handshake response: %v\n", err)
		return
	}

	// Register the connected host
	hh.connectedHosts[payload.HostID] = &PeerInfo{
		HostID:   payload.HostID,
		PeerID:   payload.PeerID,
		Role:     payload.Role,
		Version:  payload.Version,
		LastSeen: time.Now(),
	}

	fmt.Printf("Successfully completed handshake with host %s (peer %s)\n", payload.HostID, payload.PeerID)
}

// sendHandshakeError sends an error response for handshake
func (hh *HandshakeHandler) sendHandshakeError(replySubject, requestType, code, message string) {
	errorPayload := ErrorPayload{
		RequestType: requestType,
		Code:        code,
		Message:     message,
	}

	errorMsg := ControlMessage{
		Type:    ErrorType,
		Payload: mustMarshal(errorPayload),
	}

	errorData, err := json.Marshal(errorMsg)
	if err != nil {
		fmt.Printf("Failed to marshal handshake error: %v\n", err)
		return
	}

	if err := hh.nc.Publish(replySubject, errorData); err != nil {
		fmt.Printf("Failed to send handshake error: %v\n", err)
	}
}

// PerformHandshake performs the handshake from the daemon side
func (hh *HandshakeHandler) PerformHandshake(hostID, peerID string) error {
	subject := hh.subjectBuilder.HandshakeSubject(hostID)

	// Prepare the handshake request payload
	requestPayload := HandshakePayload{
		Protocol: 1,
		PeerID:   peerID,
		Role:     "daemon",
		HostID:   hostID,
		Version:  "1.0.0", // This would be the actual version
	}

	request := ControlMessage{
		Type:    HandshakeType,
		Payload: mustMarshal(requestPayload),
	}

	// Send the request and wait for response
	resp, err := hh.subjectBuilder.Request(hh.nc, subject, request, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to send handshake request: %w", err)
	}

	// Parse the response
	var response ControlMessage
	if err := json.Unmarshal(resp.Data, &response); err != nil {
		return fmt.Errorf("failed to unmarshal handshake response: %w", err)
	}

	if response.Type == ErrorType {
		var errorPayload ErrorPayload
		if err := json.Unmarshal(response.Payload, &errorPayload); err != nil {
			return fmt.Errorf("failed to unmarshal handshake error payload: %w", err)
		}
		return fmt.Errorf("handshake failed: %s (%s)", errorPayload.Message, errorPayload.Code)
	}

	if response.Type != HandshakeType {
		return fmt.Errorf("unexpected response type: %s", response.Type)
	}

	var responsePayload HandshakePayload
	if err := json.Unmarshal(response.Payload, &responsePayload); err != nil {
		return fmt.Errorf("failed to unmarshal handshake response payload: %w", err)
	}

	// Validate the response
	if responsePayload.Role != "director" {
		return fmt.Errorf("unexpected role in handshake response: %s", responsePayload.Role)
	}

	if responsePayload.HostID != hostID {
		return fmt.Errorf("host ID mismatch in response: expected %s, got %s", hostID, responsePayload.HostID)
	}

	// Register the connected host
	hh.connectedHosts[hostID] = &PeerInfo{
		HostID:   hostID,
		PeerID:   responsePayload.PeerID,
		Role:     responsePayload.Role,
		Version:  responsePayload.Version,
		LastSeen: time.Now(),
	}

	fmt.Printf("Handshake completed with director (peer %s)\n", responsePayload.PeerID)
	return nil
}

// IsConnected checks if a host is connected
func (hh *HandshakeHandler) IsConnected(hostID string) bool {
	_, exists := hh.connectedHosts[hostID]
	return exists
}

// GetConnectedHosts returns all connected hosts
func (hh *HandshakeHandler) GetConnectedHosts() []string {
	hosts := make([]string, 0, len(hh.connectedHosts))
	for hostID := range hh.connectedHosts {
		hosts = append(hosts, hostID)
	}
	return hosts
}

// mustMarshal marshals JSON and panics on error (for use in cases where error is unexpected)
func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal JSON: %v", err))
	}
	return data
}