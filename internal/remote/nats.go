// Package remote implements NATS-based communication for remote agents.
package remote

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stateforward/hsm-go/muid"
)

// Common sentinel errors for NATS operations.
var (
	// ErrNATSConnectionFailed indicates NATS connection failed.
	ErrNATSConnectionFailed = fmt.Errorf("nats connection failed")
	
	// ErrJetStreamFailed indicates JetStream operation failed.
	ErrJetStreamFailed = fmt.Errorf("jetstream failed")
	
	// ErrHandshakeFailed indicates handshake exchange failed.
	ErrHandshakeFailed = fmt.Errorf("handshake failed")
	
	// ErrNotReady indicates daemon is not ready for operations.
	ErrNotReady = fmt.Errorf("not ready")
)

// NATSConfig holds NATS connection configuration.
type NATSConfig struct {
	URL           string        // NATS server URL
	CredsFile     string        // Path to NATS credentials file
	SubjectPrefix string        // Subject prefix (default "amux")
	KVBucket      string        // JetStream KV bucket name
	Timeout       time.Duration // Request timeout
}

// NATSManager manages NATS connectivity and protocol operations.
type NATSManager struct {
	conn     *nats.Conn
	js       nats.JetStreamContext
	kv       nats.KeyValue
	config   NATSConfig
	hostID   string
	peerID   muid.MUID
	role     string // "director" or "manager"
	ready    bool   // handshake completed
}

// NewNATSManager creates a new NATS manager instance.
func NewNATSManager(hostID, role string, config NATSConfig) (*NATSManager, error) {
	if hostID == "" {
		return nil, fmt.Errorf("hostID required: %w", ErrNATSConnectionFailed)
	}
	
	if role != "director" && role != "manager" {
		return nil, fmt.Errorf("role must be 'director' or 'manager': %w", ErrNATSConnectionFailed)
	}
	
	if config.SubjectPrefix == "" {
		config.SubjectPrefix = "amux"
	}
	
	if config.KVBucket == "" {
		config.KVBucket = "AMUX_KV"
	}
	
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	
	return &NATSManager{
		config: config,
		hostID: hostID,
		peerID: muid.Make(),
		role:   role,
	}, nil
}

// Connect establishes NATS connection and sets up JetStream.
func (nm *NATSManager) Connect() error {
	opts := []nats.Option{
		nats.Name(fmt.Sprintf("amux-%s-%s", nm.role, nm.hostID)),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(-1), // Infinite reconnects
	}
	
	if nm.config.CredsFile != "" {
		opts = append(opts, nats.UserCredentials(nm.config.CredsFile))
	}
	
	conn, err := nats.Connect(nm.config.URL, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS at %s: %w", nm.config.URL, ErrNATSConnectionFailed)
	}
	
	nm.conn = conn
	
	// Set up JetStream
	js, err := conn.JetStream()
	if err != nil {
		nm.conn.Close()
		return fmt.Errorf("failed to get JetStream context: %w", ErrJetStreamFailed)
	}
	nm.js = js
	
	// Set up KV bucket (director role only)
	if nm.role == "director" {
		if err := nm.ensureKVBucket(); err != nil {
			nm.conn.Close()
			return fmt.Errorf("failed to set up KV bucket: %w", err)
		}
	}
	
	return nil
}

// ensureKVBucket creates the KV bucket if it doesn't exist.
func (nm *NATSManager) ensureKVBucket() error {
	kv, err := nm.js.KeyValue(nm.config.KVBucket)
	if err != nil {
		// Bucket doesn't exist, create it
		kv, err = nm.js.CreateKeyValue(&nats.KeyValueConfig{
			Bucket:      nm.config.KVBucket,
			Description: "amux remote control-plane state",
			TTL:         24 * time.Hour, // Default TTL for state
		})
		if err != nil {
			return fmt.Errorf("failed to create KV bucket %s: %w", nm.config.KVBucket, ErrJetStreamFailed)
		}
	}
	
	nm.kv = kv
	return nil
}

// getKVBucket gets the KV bucket (manager role).
func (nm *NATSManager) getKVBucket() error {
	if nm.kv != nil {
		return nil
	}
	
	kv, err := nm.js.KeyValue(nm.config.KVBucket)
	if err != nil {
		return fmt.Errorf("failed to get KV bucket %s: %w", nm.config.KVBucket, ErrJetStreamFailed)
	}
	
	nm.kv = kv
	return nil
}

// Close closes the NATS connection.
func (nm *NATSManager) Close() {
	if nm.conn != nil {
		nm.conn.Close()
		nm.conn = nil
	}
}

// Subject returns a fully qualified subject name.
func (nm *NATSManager) Subject(parts ...string) string {
	return nm.config.SubjectPrefix + "." + fmt.Sprintf("%s", parts[0]) + 
		func() string {
			if len(parts) > 1 {
				for _, part := range parts[1:] {
					return "." + part
				}
			}
			return ""
		}()
}

// ControlMessage represents a control protocol message.
type ControlMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// HandshakePayload represents handshake message payload.
type HandshakePayload struct {
	Protocol int    `json:"protocol"`
	PeerID   string `json:"peer_id"`  // base-10 string
	Role     string `json:"role"`
	HostID   string `json:"host_id"`
}

// ErrorPayload represents error message payload.
type ErrorPayload struct {
	RequestType string `json:"request_type"`
	Code        string `json:"code"`
	Message     string `json:"message"`
}

// SpawnPayload represents spawn request/response payload.
type SpawnPayload struct {
	AgentID    string            `json:"agent_id"`    // base-10 string
	AgentSlug  string            `json:"agent_slug,omitempty"`
	RepoPath   string            `json:"repo_path,omitempty"`
	Command    []string          `json:"command,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
	SessionID  string            `json:"session_id,omitempty"` // base-10 string  
}

// Handshake performs the initial handshake exchange.
func (nm *NATSManager) Handshake() error {
	if nm.role == "manager" {
		return nm.performManagerHandshake()
	}
	
	// Director role handles handshakes by subscribing
	return nm.subscribeToHandshakes()
}

// performManagerHandshake sends a handshake request to the director.
func (nm *NATSManager) performManagerHandshake() error {
	payload := HandshakePayload{
		Protocol: 1,
		PeerID:   nm.peerID.String(),
		Role:     "daemon",  // Protocol specifies "daemon" for manager role
		HostID:   nm.hostID,
	}
	
	payloadData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal handshake payload: %w", ErrHandshakeFailed)
	}
	
	msg := ControlMessage{
		Type:    "handshake",
		Payload: payloadData,
	}
	
	msgData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal handshake message: %w", ErrHandshakeFailed)
	}
	
	// Send handshake request
	subject := nm.Subject("handshake", nm.hostID)
	resp, err := nm.conn.Request(subject, msgData, nm.config.Timeout)
	if err != nil {
		return fmt.Errorf("handshake request failed: %w", ErrHandshakeFailed)
	}
	
	// Parse response
	var respMsg ControlMessage
	if err := json.Unmarshal(resp.Data, &respMsg); err != nil {
		return fmt.Errorf("failed to parse handshake response: %w", ErrHandshakeFailed)
	}
	
	if respMsg.Type == "error" {
		var errPayload ErrorPayload
		if err := json.Unmarshal(respMsg.Payload, &errPayload); err == nil {
			return fmt.Errorf("handshake rejected: %s - %s", errPayload.Code, errPayload.Message)
		}
		return fmt.Errorf("handshake rejected: %w", ErrHandshakeFailed)
	}
	
	if respMsg.Type != "handshake" {
		return fmt.Errorf("unexpected handshake response type: %s", respMsg.Type)
	}
	
	nm.ready = true
	return nil
}

// subscribeToHandshakes sets up handshake subscription for director role.
func (nm *NATSManager) subscribeToHandshakes() error {
	subject := nm.Subject("handshake", "*")
	_, err := nm.conn.Subscribe(subject, nm.handleHandshakeRequest)
	if err != nil {
		return fmt.Errorf("failed to subscribe to handshakes: %w", ErrHandshakeFailed)
	}
	return nil
}

// handleHandshakeRequest handles incoming handshake requests (director role).
func (nm *NATSManager) handleHandshakeRequest(msg *nats.Msg) {
	var reqMsg ControlMessage
	if err := json.Unmarshal(msg.Data, &reqMsg); err != nil {
		nm.sendErrorResponse(msg.Reply, "handshake", "invalid_payload", "failed to parse handshake request")
		return
	}
	
	if reqMsg.Type != "handshake" {
		nm.sendErrorResponse(msg.Reply, "handshake", "invalid_type", "expected handshake message")
		return
	}
	
	var payload HandshakePayload
	if err := json.Unmarshal(reqMsg.Payload, &payload); err != nil {
		nm.sendErrorResponse(msg.Reply, "handshake", "invalid_payload", "failed to parse handshake payload")
		return
	}
	
	// Extract hostID from subject
	subjectParts := strings.Split(msg.Subject, ".")
	if len(subjectParts) < 2 {
		nm.sendErrorResponse(msg.Reply, "handshake", "invalid_subject", "malformed subject")
		return
	}
	subjectHostID := subjectParts[len(subjectParts)-1]
	
	// Verify hostID matches
	if payload.HostID != subjectHostID {
		nm.sendErrorResponse(msg.Reply, "handshake", "host_id_mismatch", "hostID in payload must match subject")
		return
	}
	
	// Send successful response
	respPayload := HandshakePayload{
		Protocol: 1,
		PeerID:   nm.peerID.String(),
		Role:     "director",
		HostID:   nm.hostID,
	}
	
	nm.sendControlResponse(msg.Reply, "handshake", respPayload)
	
	// Update KV store with host info
	if err := nm.updateHostInfo(payload.HostID, payload.PeerID); err != nil {
		// Log error but don't fail handshake
	}
}

// sendErrorResponse sends an error response message.
func (nm *NATSManager) sendErrorResponse(replyTo, requestType, code, message string) {
	errPayload := ErrorPayload{
		RequestType: requestType,
		Code:        code,
		Message:     message,
	}
	
	payloadData, _ := json.Marshal(errPayload)
	
	msg := ControlMessage{
		Type:    "error",
		Payload: payloadData,
	}
	
	msgData, _ := json.Marshal(msg)
	nm.conn.Publish(replyTo, msgData)
}

// sendControlResponse sends a successful control response.
func (nm *NATSManager) sendControlResponse(replyTo, msgType string, payload interface{}) {
	payloadData, _ := json.Marshal(payload)
	
	msg := ControlMessage{
		Type:    msgType,
		Payload: payloadData,
	}
	
	msgData, _ := json.Marshal(msg)
	nm.conn.Publish(replyTo, msgData)
}

// updateHostInfo updates host information in the KV store.
func (nm *NATSManager) updateHostInfo(hostID, peerID string) error {
	if nm.kv == nil {
		if err := nm.getKVBucket(); err != nil {
			return err
		}
	}
	
	info := map[string]interface{}{
		"peer_id":   peerID,
		"connected": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0.0", // TODO: get from build info
	}
	
	infoData, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal host info: %w", err)
	}
	
	key := fmt.Sprintf("hosts/%s/info", hostID)
	_, err = nm.kv.Put(key, infoData)
	if err != nil {
		return fmt.Errorf("failed to update host info: %w", err)
	}
	
	return nil
}

// IsReady returns true if handshake has been completed.
func (nm *NATSManager) IsReady() bool {
	return nm.ready
}

// GetHostID returns the host ID.
func (nm *NATSManager) GetHostID() string {
	return nm.hostID
}

// GetPeerID returns the peer ID.
func (nm *NATSManager) GetPeerID() muid.MUID {
	return nm.peerID
}