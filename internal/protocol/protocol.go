// Package protocol provides the remote communication protocol for amux.
//
// This package implements the NATS-based protocol for communication
// between the director and remote host managers. All protocol operations
// are agent-agnostic.
//
// See spec §5.5 and §9 for remote protocol requirements.
package protocol

import (
	"encoding/json"
	"time"

	"github.com/stateforward/hsm-go/muid"
)

// MessageType represents a protocol message type.
type MessageType string

const (
	// Handshake messages
	TypeHandshakeRequest  MessageType = "handshake.request"
	TypeHandshakeResponse MessageType = "handshake.response"

	// Control messages
	TypeSpawnRequest  MessageType = "spawn.request"
	TypeSpawnResponse MessageType = "spawn.response"
	TypeKillRequest   MessageType = "kill.request"
	TypeKillResponse  MessageType = "kill.response"
	TypeReplayRequest MessageType = "replay.request"
	TypeReplayResponse MessageType = "replay.response"

	// PTY messages
	TypePTYInput  MessageType = "pty.input"
	TypePTYOutput MessageType = "pty.output"

	// Event messages
	TypeEvent MessageType = "event"
)

// Message is the base protocol message envelope.
type Message struct {
	// ID is the unique message identifier.
	ID string `json:"id"`

	// Type is the message type.
	Type MessageType `json:"type"`

	// Timestamp is when the message was created (RFC 3339).
	Timestamp string `json:"timestamp"`

	// HostID is the sending host identifier.
	HostID string `json:"host_id,omitempty"`

	// AgentID is the target agent identifier (base-10 string).
	AgentID string `json:"agent_id,omitempty"`

	// SessionID is the session identifier (base-10 string).
	SessionID string `json:"session_id,omitempty"`

	// Data is the message payload.
	Data json.RawMessage `json:"data,omitempty"`

	// TraceID is the optional trace context.
	TraceID string `json:"trace_id,omitempty"`
}

// NewMessage creates a new protocol message.
func NewMessage(msgType MessageType, data any) (*Message, error) {
	var rawData json.RawMessage
	if data != nil {
		var err error
		rawData, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
	}

	return &Message{
		ID:        muid.Make().String(),
		Type:      msgType,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Data:      rawData,
	}, nil
}

// WithHost sets the host ID.
func (m *Message) WithHost(hostID string) *Message {
	m.HostID = hostID
	return m
}

// WithAgent sets the agent ID.
func (m *Message) WithAgent(agentID muid.MUID) *Message {
	m.AgentID = agentID.String()
	return m
}

// WithSession sets the session ID.
func (m *Message) WithSession(sessionID muid.MUID) *Message {
	m.SessionID = sessionID.String()
	return m
}

// WithTrace sets the trace ID.
func (m *Message) WithTrace(traceID string) *Message {
	m.TraceID = traceID
	return m
}

// HandshakeRequest is the handshake request payload.
type HandshakeRequest struct {
	HostID      string `json:"host_id"`
	Version     string `json:"version"`
	SpecVersion string `json:"spec_version"`
}

// HandshakeResponse is the handshake response payload.
type HandshakeResponse struct {
	Accepted    bool   `json:"accepted"`
	Error       string `json:"error,omitempty"`
	Version     string `json:"version"`
	SpecVersion string `json:"spec_version"`
}

// SpawnRequest is the spawn request payload.
type SpawnRequest struct {
	AgentID   string `json:"agent_id"`
	RepoPath  string `json:"repo_path"`
	AgentSlug string `json:"agent_slug"`
	Adapter   string `json:"adapter"`
}

// SpawnResponse is the spawn response payload.
type SpawnResponse struct {
	SessionID string `json:"session_id"`
	Error     string `json:"error,omitempty"`
	Code      string `json:"code,omitempty"`
}

// KillRequest is the kill request payload.
type KillRequest struct {
	SessionID string `json:"session_id"`
	Force     bool   `json:"force,omitempty"`
}

// KillResponse is the kill response payload.
type KillResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// ReplayRequest is the replay request payload.
type ReplayRequest struct {
	SessionID string `json:"session_id"`
}

// ReplayResponse is the replay response payload.
type ReplayResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// PTYData is the PTY I/O payload.
type PTYData struct {
	Data string `json:"data_b64"` // Base64-encoded PTY data
}

// Subject returns the NATS subject for a message type.
func Subject(prefix string, hostID string, msgType MessageType) string {
	switch msgType {
	case TypeHandshakeRequest, TypeHandshakeResponse:
		return prefix + ".C." + hostID + ".handshake"
	case TypeSpawnRequest, TypeSpawnResponse:
		return prefix + ".C." + hostID + ".spawn"
	case TypeKillRequest, TypeKillResponse:
		return prefix + ".C." + hostID + ".kill"
	case TypeReplayRequest, TypeReplayResponse:
		return prefix + ".C." + hostID + ".replay"
	default:
		return prefix + ".E." + hostID
	}
}

// PTYSubject returns the NATS subject for PTY I/O.
func PTYSubject(prefix, hostID, sessionID string, isOutput bool) string {
	direction := "in"
	if isOutput {
		direction = "out"
	}
	return prefix + ".P.pty." + hostID + "." + sessionID + "." + direction
}
