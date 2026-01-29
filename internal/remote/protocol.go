// Package remote implements Phase 3 remote agent orchestration via NATS + JetStream.
//
// This package provides:
// - Director role: hub NATS server + orchestration of remote hosts
// - Manager role: leaf NATS connection + local PTY session management
// - NATS protocol subjects, control messages, and handshake per spec §5.5.7
// - SSH bootstrap for remote daemon installation per spec §5.5.2
// - JetStream KV state for durable control-plane metadata per spec §5.5.6.3
// - Per-host credentials and subject authorization per spec §5.5.6.4
package remote

import (
	"encoding/json"
	"fmt"

	"github.com/stateforward/hsm-go/muid"
)

// ControlMessage is the top-level envelope for NATS control messages.
// Per spec §5.5.7.2, all control requests and responses use this shape.
type ControlMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// HandshakePayload represents a handshake request or response.
// Per spec §5.5.7.3, both daemon→director and director→daemon use this shape.
type HandshakePayload struct {
	Protocol int    `json:"protocol"`
	PeerID   string `json:"peer_id"`   // base-10 MUID
	Role     string `json:"role"`      // "director" or "daemon"
	HostID   string `json:"host_id"`
}

// ErrorPayload represents an error response to a control request.
// Per spec §5.5.7.3.
type ErrorPayload struct {
	RequestType string `json:"request_type"` // "handshake", "spawn", "kill", "replay", "unknown"
	Code        string `json:"code"`
	Message     string `json:"message"`
}

// SpawnRequestPayload represents a spawn control request from director to manager.
// Per spec §5.5.7.3.
type SpawnRequestPayload struct {
	AgentID   string            `json:"agent_id"`   // base-10 MUID
	AgentSlug string            `json:"agent_slug"`
	RepoPath  string            `json:"repo_path"`
	Command   []string          `json:"command"`
	Env       map[string]string `json:"env,omitempty"`
}

// SpawnResponsePayload represents the manager's response to a spawn request.
// Per spec §5.5.7.3.
type SpawnResponsePayload struct {
	AgentID   string `json:"agent_id"`   // base-10 MUID (echoed from request)
	SessionID string `json:"session_id"` // base-10 MUID
}

// KillRequestPayload represents a kill control request from director to manager.
// Per spec §5.5.7.3.
type KillRequestPayload struct {
	SessionID string `json:"session_id"` // base-10 MUID
}

// KillResponsePayload represents the manager's response to a kill request.
// Per spec §5.5.7.3.
type KillResponsePayload struct {
	SessionID string `json:"session_id"` // base-10 MUID
	Killed    bool   `json:"killed"`
}

// ReplayRequestPayload represents a replay control request from director to manager.
// Per spec §5.5.7.3.
type ReplayRequestPayload struct {
	SessionID string `json:"session_id"` // base-10 MUID
}

// ReplayResponsePayload represents the manager's response to a replay request.
// Per spec §5.5.7.3.
type ReplayResponsePayload struct {
	SessionID string `json:"session_id"` // base-10 MUID
	Accepted  bool   `json:"accepted"`
}

// PingPayload represents a ping message.
// Per spec §5.5.7.3.
type PingPayload struct {
	TsUnixMs int64 `json:"ts_unix_ms"`
}

// PongPayload represents a pong message.
// Per spec §5.5.7.3.
type PongPayload struct {
	TsUnixMs int64 `json:"ts_unix_ms"`
}

// SubjectBuilder constructs NATS subjects using the configured prefix.
// Per spec §5.5.7.1, the subject prefix is configurable (default "amux").
type SubjectBuilder struct {
	Prefix string
}

// Handshake returns the handshake request subject for the given host_id.
// Subject: P.handshake.<host_id>
func (sb SubjectBuilder) Handshake(hostID string) string {
	return fmt.Sprintf("%s.handshake.%s", sb.Prefix, hostID)
}

// Control returns the control request subject for the given host_id.
// Subject: P.ctl.<host_id>
func (sb SubjectBuilder) Control(hostID string) string {
	return fmt.Sprintf("%s.ctl.%s", sb.Prefix, hostID)
}

// Events returns the host events subject for the given host_id.
// Subject: P.events.<host_id>
func (sb SubjectBuilder) Events(hostID string) string {
	return fmt.Sprintf("%s.events.%s", sb.Prefix, hostID)
}

// PTYOut returns the PTY output subject for the given host_id and session_id.
// Subject: P.pty.<host_id>.<session_id>.out
func (sb SubjectBuilder) PTYOut(hostID string, sessionID muid.MUID) string {
	return fmt.Sprintf("%s.pty.%s.%s.out", sb.Prefix, hostID, FormatID(sessionID))
}

// PTYIn returns the PTY input subject for the given host_id and session_id.
// Subject: P.pty.<host_id>.<session_id>.in
func (sb SubjectBuilder) PTYIn(hostID string, sessionID muid.MUID) string {
	return fmt.Sprintf("%s.pty.%s.%s.in", sb.Prefix, hostID, FormatID(sessionID))
}

// CommDirector returns the director communication channel subject.
// Subject: P.comm.director
func (sb SubjectBuilder) CommDirector() string {
	return fmt.Sprintf("%s.comm.director", sb.Prefix)
}

// CommManager returns the manager communication channel subject for the given host_id.
// Subject: P.comm.manager.<host_id>
func (sb SubjectBuilder) CommManager(hostID string) string {
	return fmt.Sprintf("%s.comm.manager.%s", sb.Prefix, hostID)
}

// CommAgent returns the agent communication channel subject for the given host_id and agent_id.
// Subject: P.comm.agent.<host_id>.<agent_id>
func (sb SubjectBuilder) CommAgent(hostID string, agentID muid.MUID) string {
	return fmt.Sprintf("%s.comm.agent.%s.%s", sb.Prefix, hostID, FormatID(agentID))
}

// CommBroadcast returns the broadcast communication channel subject.
// Subject: P.comm.broadcast
func (sb SubjectBuilder) CommBroadcast() string {
	return fmt.Sprintf("%s.comm.broadcast", sb.Prefix)
}

// FormatID formats a muid.MUID as a base-10 string for JSON encoding.
// Per spec §9.1.3.1, IDs are encoded as base-10 unsigned integer strings.
func FormatID(id muid.MUID) string {
	return fmt.Sprintf("%d", uint64(id))
}

// ParseID parses a base-10 string into a muid.MUID.
func ParseID(s string) (muid.MUID, error) {
	var val uint64
	_, err := fmt.Sscanf(s, "%d", &val)
	if err != nil {
		return 0, fmt.Errorf("parse ID: %w", err)
	}
	return muid.MUID(val), nil
}

// MarshalControlMessage marshals a control message with the given type and payload.
func MarshalControlMessage(msgType string, payload any) ([]byte, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	msg := ControlMessage{
		Type:    msgType,
		Payload: payloadBytes,
	}

	return json.Marshal(msg)
}

// UnmarshalControlMessage unmarshals a control message and extracts the payload.
func UnmarshalControlMessage(data []byte, payload any) (string, error) {
	var msg ControlMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return "", fmt.Errorf("unmarshal control message: %w", err)
	}

	if len(msg.Payload) > 0 {
		if err := json.Unmarshal(msg.Payload, payload); err != nil {
			return msg.Type, fmt.Errorf("unmarshal payload: %w", err)
		}
	}

	return msg.Type, nil
}
