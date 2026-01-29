package protocol

import (
	"time"
)

// Subject prefixes and templates.
const (
	// DefaultSubjectPrefix is the default prefix for all Amux NATS subjects.
	DefaultSubjectPrefix = "amux"

	// ControlSubjectTemplate is the template for request/reply control messages.
	// Format: <prefix>.control.<host_id>.<op>
	ControlSubjectTemplate = "%s.control.%s.%s"

	// EventsSubjectTemplate is the template for event publishing.
	// Format: <prefix>.events.<host_id>
	EventsSubjectTemplate = "%s.events.%s"

	// PTYInSubjectTemplate is the template for PTY input streaming.
	// Format: <prefix>.pty.<host_id>.<session_id>.in
	PTYInSubjectTemplate = "%s.pty.%s.%s.in"

	// PTYOutSubjectTemplate is the template for PTY output streaming.
	// Format: <prefix>.pty.<host_id>.<session_id>.out
	PTYOutSubjectTemplate = "%s.pty.%s.%s.out"

	// KVBucketDefault is the default name for the JetStream KV bucket.
	KVBucketDefault = "AMUX_KV"

	// KVHostInfoTemplate is the key for host metadata.
	// Format: hosts/<host_id>/info
	KVHostInfoTemplate = "hosts/%s/info"

	// KVHostHeartbeatTemplate is the key for host heartbeats.
	// Format: hosts/<host_id>/heartbeat
	KVHostHeartbeatTemplate = "hosts/%s/heartbeat"

	// KVSessionTemplate is the key for session metadata.
	// Format: sessions/<host_id>/<session_id>
	KVSessionTemplate = "sessions/%s/%s"
)

// Control Operations.
const (
	OpHandshake = "handshake"
	OpSpawn     = "spawn"
	OpSignal    = "signal" // kill/stop
	OpResize    = "resize"
	OpReplay    = "replay"
)

// Message Envelopes

// ControlRequest is the generic envelope for control requests.
type ControlRequest struct {
	Op        string    `json:"op"`
	RequestID string    `json:"req_id"`
	Payload   []byte    `json:"payload"` // JSON encoded specific payload
	CreatedAt time.Time `json:"created_at"`
}

// ControlResponse is the generic envelope for control responses.
type ControlResponse struct {
	RequestID string    `json:"req_id"`
	Status    string    `json:"status"` // "ok", "error"
	Error     *Error    `json:"error,omitempty"`
	Payload   []byte    `json:"payload,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Error represents a protocol error.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Specific Payloads

// HandshakePayload used for OpHandshake.
type HandshakePayload struct {
	HostID       string   `json:"host_id"`
	Version      string   `json:"version"`
	Capabilities []string `json:"capabilities"`
}

// SpawnPayload used for OpSpawn.
type SpawnPayload struct {
	AgentID     string            `json:"agent_id"` // muid
	AgentSlug   string            `json:"agent_slug"`
	RepoPath    string            `json:"repo_path"`
	Env         map[string]string `json:"env,omitempty"`
	Command     []string          `json:"command,omitempty"`
	Interactive bool              `json:"interactive"`
}

// SpawnResponsePayload (inside ControlResponse.Payload).
type SpawnResponsePayload struct {
	SessionID string `json:"session_id"` // muid
}

// SignalPayload used for OpSignal.
type SignalPayload struct {
	SessionID string `json:"session_id"`
	Signal    string `json:"signal"` // "kill", "term", "int" or syscall name
}

// ResizePayload used for OpResize.
type ResizePayload struct {
	SessionID string `json:"session_id"`
	Rows      int    `json:"rows"`
	Cols      int    `json:"cols"`
}

// ReplayPayload used for OpReplay.
type ReplayPayload struct {
	SessionID string `json:"session_id"`
	// If 0, replay all buffered.
	SinceSequence uint64 `json:"since_seq,omitempty"`
}

// PTYIO represents a chunk of PTY data.
type PTYIO struct {
	SessionID string    `json:"sid"`
	Data      []byte    `json:"d"`             // base64 encoded by json/standard, but we use []byte here
	Seq       uint64    `json:"seq,omitempty"` // For output ordering
	Timestamp time.Time `json:"ts"`
}

// KV Payloads

// HostInfo represents static host metadata stored in KV.
type HostInfo struct {
	ID           string    `json:"id"`
	Hostname     string    `json:"hostname"`
	Platform     string    `json:"platform"`
	Arch         string    `json:"arch"`
	Version      string    `json:"version"`
	Capabilities []string  `json:"capabilities"`
	FirstSeenAt  time.Time `json:"first_seen_at"`
}

// Heartbeat represents dynamic host status.
type Heartbeat struct {
	HostID    string    `json:"host_id"`
	Timestamp time.Time `json:"ts"`
	Load      float64   `json:"load"`
	MemUsage  uint64    `json:"mem_usage"`
	Sessions  int       `json:"sessions"`
}

// SessionInfo represents session metadata in KV.
type SessionInfo struct {
	SessionID string    `json:"session_id"`
	AgentID   string    `json:"agent_id"`
	HostID    string    `json:"host_id"`
	State     string    `json:"state"` // "running", "ended"
	CreatedAt time.Time `json:"created_at"`
}
