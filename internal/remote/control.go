// control.go defines control message types and payloads for the remote protocol per spec §5.5.7.2, §5.5.7.3.
package remote

import "encoding/json"

// ControlMessage is the top-level envelope for control requests and responses (spec §5.5.7.2).
type ControlMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// HandshakePayload is the handshake request/response payload (spec §5.5.7.3).
// Daemon sends to P.handshake.<host_id>; director replies with peer_id and role.
type HandshakePayload struct {
	Protocol int    `json:"protocol"`
	PeerID   string `json:"peer_id"`   // base-10 string
	Role     string `json:"role"`     // "daemon" or "director"
	HostID   string `json:"host_id"`
}

// SpawnPayloadRequest is the spawn request payload (spec §5.5.7.3).
type SpawnPayloadRequest struct {
	AgentID   string            `json:"agent_id"`
	AgentSlug string            `json:"agent_slug"`
	RepoPath  string            `json:"repo_path"`
	Command   []string          `json:"command"`
	Env       map[string]string `json:"env,omitempty"`
}

// SpawnPayloadResponse is the spawn response payload (spec §5.5.7.3).
type SpawnPayloadResponse struct {
	AgentID   string `json:"agent_id"`
	SessionID  string `json:"session_id"`
}

// KillPayloadRequest is the kill request payload (spec §5.5.7.3).
type KillPayloadRequest struct {
	SessionID string `json:"session_id"`
}

// KillPayloadResponse is the kill response payload (spec §5.5.7.3).
type KillPayloadResponse struct {
	SessionID string `json:"session_id"`
	Killed    bool   `json:"killed"`
}

// ReplayPayloadRequest is the replay request payload (spec §5.5.7.3).
type ReplayPayloadRequest struct {
	SessionID string `json:"session_id"`
}

// ReplayPayloadResponse is the replay response payload (spec §5.5.7.3).
type ReplayPayloadResponse struct {
	SessionID string `json:"session_id"`
	Accepted  bool   `json:"accepted"`
}

// ErrorPayload is the error response payload (spec §5.5.7.3).
// request_type is one of: "handshake", "spawn", "kill", "replay", "unknown".
type ErrorPayload struct {
	RequestType string `json:"request_type"`
	Code        string `json:"code"`
	Message     string `json:"message"`
}

// Control message type constants (spec §5.5.7.2).
const (
	ControlTypeHandshake = "handshake"
	ControlTypePing      = "ping"
	ControlTypePong      = "pong"
	ControlTypeSpawn     = "spawn"
	ControlTypeKill      = "kill"
	ControlTypeReplay    = "replay"
	ControlTypeError     = "error"
)

// Error codes (spec §5.5.7.3).
const (
	ErrorCodeNotReady       = "not_ready"
	ErrorCodeSessionConflict = "session_conflict"
	ErrorCodeInvalidRepo    = "invalid_repo"
)
