package protocol

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/agentflare-ai/amux/pkg/api"
)

// Subject prefixes
const (
	SubjectHandshake = "handshake"
	SubjectCtl       = "ctl"
	SubjectEvents    = "events"
	SubjectPTY       = "pty"
)

// HandshakeRequest is the payload for the handshake request.
type HandshakeRequest struct {
	Protocol int        `json:"protocol"`
	PeerID   api.PeerID `json:"peer_id"`
	Role     string     `json:"role"`
	HostID   api.HostID `json:"host_id"`
}

// HandshakeResponse is the payload for the handshake response.
type HandshakeResponse struct {
	Protocol int        `json:"protocol"`
	PeerID   api.PeerID `json:"peer_id"`
	Role     string     `json:"role"`
	HostID   api.HostID `json:"host_id"`
	Error    *Error     `json:"error,omitempty"`
}

// ControlRequest is the payload for control operations (spawn/kill/replay).
type ControlRequest struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// ControlResponse is the payload for control operation responses.
type ControlResponse struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	Error   *Error          `json:"error,omitempty"`
}

// SpawnPayload is the payload for the spawn command.
type SpawnPayload struct {
	AgentID  api.AgentID   `json:"agent_id"`
	Slug     api.AgentSlug `json:"agent_slug"`
	RepoPath string        `json:"repo_path"`
	Command  []string      `json:"command,omitempty"`
	Env      map[string]string `json:"env,omitempty"`
}

// SpawnResponsePayload is the payload for the spawn response.
type SpawnResponsePayload struct {
	AgentID   api.AgentID   `json:"agent_id"`
	SessionID api.SessionID `json:"session_id"`
}

// Error represents a protocol error.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// EventMessage represents an event envelope over NATS.
// It matches the spec for hsmnet wire format.
type EventMessage struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Source    string      `json:"source"` // peer_id
	Payload   interface{} `json:"payload"`
}

// SubjectHandshake returns the handshake subject for a host.
// P.handshake.<host_id>
func SubjectForHandshake(prefix string, hostID api.HostID) string {
	return fmt.Sprintf("%s.%s.%s", prefix, SubjectHandshake, hostID)
}

// SubjectForCtl returns the control subject for a host.
// P.ctl.<host_id>
func SubjectForCtl(prefix string, hostID api.HostID) string {
	return fmt.Sprintf("%s.%s.%s", prefix, SubjectCtl, hostID)
}

// SubjectForEvents returns the events subject for a host.
// P.events.<host_id>
func SubjectForEvents(prefix string, hostID api.HostID) string {
	return fmt.Sprintf("%s.%s.%s", prefix, SubjectEvents, hostID)
}

// SubjectForPTYOut returns the PTY output subject.
// P.pty.<host_id>.<session_id>.out
func SubjectForPTYOut(prefix string, hostID api.HostID, sessionID api.SessionID) string {
	return fmt.Sprintf("%s.%s.%s.%s.out", prefix, SubjectPTY, hostID, sessionID)
}

// SubjectForPTYIn returns the PTY input subject.
// P.pty.<host_id>.<session_id>.in
func SubjectForPTYIn(prefix string, hostID api.HostID, sessionID api.SessionID) string {
	return fmt.Sprintf("%s.%s.%s.%s.in", prefix, SubjectPTY, hostID, sessionID)
}
