// Package protocol implements remote communication protocol (transports events)
package protocol

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// MessageEnvelope wraps messages sent over NATS
type MessageEnvelope struct {
	ID        string      `json:"id"`
	Timestamp time.Time   `json:"timestamp"`
	Source    string      `json:"source"`
	Payload   interface{} `json:"payload"`
}

// ControlMessageType represents the type of control message
type ControlMessageType string

const (
	HandshakeType ControlMessageType = "handshake"
	PingType      ControlMessageType = "ping"
	PongType      ControlMessageType = "pong"
	SpawnType     ControlMessageType = "spawn"
	KillType      ControlMessageType = "kill"
	ReplayType    ControlMessageType = "replay"
	ErrorType     ControlMessageType = "error"
)

// ControlMessage represents a control message exchanged between director and daemon
type ControlMessage struct {
	Type    ControlMessageType `json:"type"`
	Payload json.RawMessage    `json:"payload"`
}

// HandshakePayload represents the payload for handshake messages
type HandshakePayload struct {
	Protocol int    `json:"protocol"`
	PeerID   string `json:"peer_id"`
	Role     string `json:"role"`     // "director" or "daemon"
	HostID   string `json:"host_id"`  // The host ID
	Version  string `json:"version"`  // Version of the daemon
}

// SpawnPayload represents the payload for spawn messages
type SpawnPayload struct {
	AgentID   string   `json:"agent_id"`
	AgentSlug string   `json:"agent_slug"`
	RepoPath  string   `json:"repo_path"`
	Command   []string `json:"command"`
	Env       map[string]string `json:"env"`
}

// SpawnResponsePayload represents the response payload for spawn messages
type SpawnResponsePayload struct {
	AgentID   string `json:"agent_id"`
	SessionID string `json:"session_id"`
}

// KillPayload represents the payload for kill messages
type KillPayload struct {
	SessionID string `json:"session_id"`
}

// KillResponsePayload represents the response payload for kill messages
type KillResponsePayload struct {
	SessionID string `json:"session_id"`
	Killed    bool   `json:"killed"`
}

// ReplayPayload represents the payload for replay messages
type ReplayPayload struct {
	SessionID string `json:"session_id"`
}

// ReplayResponsePayload represents the response payload for replay messages
type ReplayResponsePayload struct {
	SessionID string `json:"session_id"`
	Accepted  bool   `json:"accepted"`
}

// ErrorPayload represents the payload for error messages
type ErrorPayload struct {
	RequestType string `json:"request_type"` // The type of request that caused the error
	Code        string `json:"code"`         // Short machine-readable error code
	Message     string `json:"message"`      // Human-readable error message
}

// PingPayload represents the payload for ping messages
type PingPayload struct {
	TimestampUnixMs int64 `json:"ts_unix_ms"`
}

// PongPayload represents the payload for pong messages
type PongPayload struct {
	TimestampUnixMs int64 `json:"ts_unix_ms"`
}

// SubjectBuilder builds NATS subjects according to the specification
type SubjectBuilder struct {
	prefix string
}

// NewSubjectBuilder creates a new SubjectBuilder
func NewSubjectBuilder(prefix string) *SubjectBuilder {
	if prefix == "" {
		prefix = "amux" // Default prefix
	}
	return &SubjectBuilder{
		prefix: prefix,
	}
}

// HandshakeSubject returns the subject for handshake messages
func (sb *SubjectBuilder) HandshakeSubject(hostID string) string {
	return fmt.Sprintf("%s.handshake.%s", sb.prefix, hostID)
}

// ControlSubject returns the subject for control messages
func (sb *SubjectBuilder) ControlSubject(hostID string) string {
	return fmt.Sprintf("%s.ctl.%s", sb.prefix, hostID)
}

// EventsSubject returns the subject for host events
func (sb *SubjectBuilder) EventsSubject(hostID string) string {
	return fmt.Sprintf("%s.events.%s", sb.prefix, hostID)
}

// PTYOutputSubject returns the subject for PTY output
func (sb *SubjectBuilder) PTYOutputSubject(hostID, sessionID string) string {
	return fmt.Sprintf("%s.pty.%s.%s.out", sb.prefix, hostID, sessionID)
}

// PTYInputSubject returns the subject for PTY input
func (sb *SubjectBuilder) PTYInputSubject(hostID, sessionID string) string {
	return fmt.Sprintf("%s.pty.%s.%s.in", sb.prefix, hostID, sessionID)
}

// DirectorCommSubject returns the subject for director communication
func (sb *SubjectBuilder) DirectorCommSubject() string {
	return fmt.Sprintf("%s.comm.director", sb.prefix)
}

// ManagerCommSubject returns the subject for manager communication
func (sb *SubjectBuilder) ManagerCommSubject(hostID string) string {
	return fmt.Sprintf("%s.comm.manager.%s", sb.prefix, hostID)
}

// AgentCommSubject returns the subject for agent communication
func (sb *SubjectBuilder) AgentCommSubject(hostID, agentID string) string {
	return fmt.Sprintf("%s.comm.agent.%s.%s", sb.prefix, hostID, agentID)
}

// BroadcastCommSubject returns the subject for broadcast communication
func (sb *SubjectBuilder) BroadcastCommSubject() string {
	return fmt.Sprintf("%s.comm.broadcast", sb.prefix)
}

// Publish publishes a message to a NATS subject
func (sb *SubjectBuilder) Publish(nc *nats.Conn, subject string, msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := nc.Publish(subject, data); err != nil {
		return fmt.Errorf("failed to publish message to %s: %w", subject, err)
	}

	return nil
}

// PublishControlMessage publishes a control message to a NATS subject
func (sb *SubjectBuilder) PublishControlMessage(nc *nats.Conn, subject string, msgType ControlMessageType, payload interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal control message payload: %w", err)
	}

	controlMsg := ControlMessage{
		Type:    msgType,
		Payload: payloadBytes,
	}

	return sb.Publish(nc, subject, controlMsg)
}

// Request sends a request-reply message to NATS
func (sb *SubjectBuilder) Request(nc *nats.Conn, subject string, msg interface{}, timeout time.Duration) (*nats.Msg, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	resp, err := nc.Request(subject, data, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to %s: %w", subject, err)
	}

	return resp, nil
}

// RequestControlMessage sends a control message using request-reply
func (sb *SubjectBuilder) RequestControlMessage(nc *nats.Conn, subject string, msgType ControlMessageType, payload interface{}, timeout time.Duration) (*nats.Msg, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal control message payload: %w", err)
	}

	controlMsg := ControlMessage{
		Type:    msgType,
		Payload: payloadBytes,
	}

	return sb.Request(nc, subject, controlMsg, timeout)
}