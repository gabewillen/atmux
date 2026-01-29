// Package protocol provides the remote communication protocol for amux.
//
// This package implements the NATS-based protocol for communication
// between the director and remote host managers. All protocol operations
// are agent-agnostic.
//
// Control messages use the ControlMessage envelope with a type discriminator
// and a JSON payload. Subject namespaces follow spec §5.5.7.1.
//
// See spec §5.5.7 for the communication protocol and §9.1.3 for wire format.
package protocol

import (
	"encoding/json"
	"fmt"
	"time"
)

// ProtocolVersion is the current protocol version used in handshake exchanges.
const ProtocolVersion = 1

// ControlMessage is the top-level envelope for all control messages exchanged
// between the director and manager-role daemons over NATS request-reply.
//
// See spec §5.5.7.2.
type ControlMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// NewControlMessage creates a ControlMessage with the given type and payload.
func NewControlMessage(msgType string, payload any) (*ControlMessage, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal control payload: %w", err)
	}
	return &ControlMessage{
		Type:    msgType,
		Payload: json.RawMessage(data),
	}, nil
}

// DecodePayload decodes the Payload field into the given target.
func (m *ControlMessage) DecodePayload(target any) error {
	return json.Unmarshal(m.Payload, target)
}

// Control message type constants per spec §5.5.7.2.
const (
	TypeHandshake = "handshake"
	TypePing      = "ping"
	TypePong      = "pong"
	TypeSpawn     = "spawn"
	TypeKill      = "kill"
	TypeReplay    = "replay"
	TypeError     = "error"
)

// HandshakePayload is the handshake request/response payload.
//
// See spec §5.5.7.3.
type HandshakePayload struct {
	// Protocol is the protocol version (must be ProtocolVersion).
	Protocol int `json:"protocol"`

	// PeerID is the unique peer identifier (base-10 unsigned integer string).
	PeerID string `json:"peer_id"`

	// Role is the peer role: "daemon" for manager, "director" for director.
	Role string `json:"role"`

	// HostID is the host identifier.
	HostID string `json:"host_id"`
}

// ErrorPayload is the error response payload.
//
// See spec §5.5.7.3.
type ErrorPayload struct {
	// RequestType is the type of request that caused the error.
	// One of: "handshake", "spawn", "kill", "replay", "unknown".
	RequestType string `json:"request_type"`

	// Code is a short machine-readable error code.
	Code string `json:"code"`

	// Message is a human-readable diagnostic.
	Message string `json:"message"`
}

// SpawnRequest is the spawn request payload.
//
// See spec §5.5.7.3.
type SpawnRequest struct {
	// AgentID is the agent identifier (base-10 unsigned integer string).
	AgentID string `json:"agent_id"`

	// AgentSlug is the normalized agent slug per §5.3.1.
	AgentSlug string `json:"agent_slug"`

	// RepoPath is the git repository root on the remote host.
	// May begin with ~/ and is expanded by the remote daemon.
	RepoPath string `json:"repo_path"`

	// Command is the argv vector for the agent CLI.
	Command []string `json:"command"`

	// Env is optional environment variables for the spawned process.
	Env map[string]string `json:"env,omitempty"`
}

// SpawnResponse is the spawn response payload.
//
// See spec §5.5.7.3.
type SpawnResponse struct {
	// AgentID is echoed from the request (base-10 unsigned integer string).
	AgentID string `json:"agent_id"`

	// SessionID is the session identifier (base-10 unsigned integer string, non-zero).
	SessionID string `json:"session_id"`
}

// KillRequest is the kill request payload.
//
// See spec §5.5.7.3.
type KillRequest struct {
	// SessionID is the session to terminate (base-10 unsigned integer string).
	SessionID string `json:"session_id"`
}

// KillResponse is the kill response payload.
//
// See spec §5.5.7.3.
type KillResponse struct {
	// SessionID is echoed from the request.
	SessionID string `json:"session_id"`

	// Killed is true if a session was found and termination was initiated.
	Killed bool `json:"killed"`
}

// ReplayRequest is the replay request payload.
//
// See spec §5.5.7.3.
type ReplayRequest struct {
	// SessionID is the session to replay (base-10 unsigned integer string).
	SessionID string `json:"session_id"`
}

// ReplayResponse is the replay response payload.
//
// See spec §5.5.7.3.
type ReplayResponse struct {
	// SessionID is echoed from the request.
	SessionID string `json:"session_id"`

	// Accepted is true if the daemon will replay buffered PTY output.
	Accepted bool `json:"accepted"`
}

// PingPayload is the ping request payload.
//
// See spec §5.5.7.3.
type PingPayload struct {
	// TSUnixMs is the timestamp in milliseconds since Unix epoch.
	TSUnixMs int64 `json:"ts_unix_ms"`
}

// PongPayload is the pong response payload.
//
// See spec §5.5.7.3.
type PongPayload struct {
	// TSUnixMs is echoed from the ping request.
	TSUnixMs int64 `json:"ts_unix_ms"`
}

// NewPingPayload creates a PingPayload with the current timestamp.
func NewPingPayload() *PingPayload {
	return &PingPayload{TSUnixMs: time.Now().UnixMilli()}
}

// --- Subject namespace functions per spec §5.5.7.1 ---

// HandshakeSubject returns the handshake subject for a host.
// Format: P.handshake.<host_id>
func HandshakeSubject(prefix, hostID string) string {
	return prefix + ".handshake." + hostID
}

// ControlSubject returns the control request subject for a host.
// Format: P.ctl.<host_id>
func ControlSubject(prefix, hostID string) string {
	return prefix + ".ctl." + hostID
}

// EventsSubject returns the host events subject.
// Format: P.events.<host_id>
func EventsSubject(prefix, hostID string) string {
	return prefix + ".events." + hostID
}

// PTYOutputSubject returns the PTY output subject for a session.
// Format: P.pty.<host_id>.<session_id>.out
func PTYOutputSubject(prefix, hostID, sessionID string) string {
	return prefix + ".pty." + hostID + "." + sessionID + ".out"
}

// PTYInputSubject returns the PTY input subject for a session.
// Format: P.pty.<host_id>.<session_id>.in
func PTYInputSubject(prefix, hostID, sessionID string) string {
	return prefix + ".pty." + hostID + "." + sessionID + ".in"
}

// PTYInputWildcard returns the wildcard subscription for all PTY input on a host.
// Format: P.pty.<host_id>.*.in
func PTYInputWildcard(prefix, hostID string) string {
	return prefix + ".pty." + hostID + ".*.in"
}

// --- Communication channel subjects per spec §5.5.7.1 ---

// DirectorChannelSubject returns the director communication channel subject.
// Format: P.comm.director
func DirectorChannelSubject(prefix string) string {
	return prefix + ".comm.director"
}

// ManagerChannelSubject returns the host manager communication channel subject.
// Format: P.comm.manager.<host_id>
func ManagerChannelSubject(prefix, hostID string) string {
	return prefix + ".comm.manager." + hostID
}

// AgentChannelSubject returns the agent communication channel subject.
// Format: P.comm.agent.<host_id>.<agent_id>
func AgentChannelSubject(prefix, hostID, agentID string) string {
	return prefix + ".comm.agent." + hostID + "." + agentID
}

// BroadcastChannelSubject returns the broadcast communication channel subject.
// Format: P.comm.broadcast
func BroadcastChannelSubject(prefix string) string {
	return prefix + ".comm.broadcast"
}

// --- Error code constants ---

// Error codes used in ErrorPayload.Code.
const (
	CodeNotReady        = "not_ready"
	CodeSessionConflict = "session_conflict"
	CodeInvalidRepo     = "invalid_repo"
	CodeInvalidAgent    = "invalid_agent"
	CodeSessionNotFound = "session_not_found"
	CodeProtocolError   = "protocol_error"
	CodeInternalError   = "internal_error"
	CodeHostIDMismatch  = "host_id_mismatch"
	CodePeerCollision   = "peer_collision"
	CodeHostCollision   = "host_collision"
)

// NewErrorMessage creates a ControlMessage of type "error".
func NewErrorMessage(requestType, code, message string) (*ControlMessage, error) {
	return NewControlMessage(TypeError, &ErrorPayload{
		RequestType: requestType,
		Code:        code,
		Message:     message,
	})
}
