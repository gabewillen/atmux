package remote

import (
	"encoding/json"
	"fmt"
)

// ControlMessage is the envelope for remote control requests.
type ControlMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// HandshakePayload is the handshake request/response payload.
type HandshakePayload struct {
	Protocol int    `json:"protocol"`
	PeerID   string `json:"peer_id"`
	Role     string `json:"role"`
	HostID   string `json:"host_id"`
}

// ErrorPayload describes a control error response.
type ErrorPayload struct {
	RequestType string `json:"request_type"`
	Code        string `json:"code"`
	Message     string `json:"message"`
}

// SpawnRequest describes a spawn request payload.
type SpawnRequest struct {
	Name           string            `json:"name,omitempty"`
	About          string            `json:"about,omitempty"`
	AgentID        string            `json:"agent_id"`
	AgentSlug      string            `json:"agent_slug"`
	RepoPath       string            `json:"repo_path"`
	Adapter        string            `json:"adapter"`
	Command        []string          `json:"command"`
	Env            map[string]string `json:"env,omitempty"`
	ListenChannels []string          `json:"listen_channels,omitempty"`
}

// SpawnResponse describes a spawn response payload.
type SpawnResponse struct {
	AgentID   string `json:"agent_id"`
	SessionID string `json:"session_id"`
}

// KillRequest describes a kill request payload.
type KillRequest struct {
	SessionID string `json:"session_id"`
}

// KillResponse describes a kill response payload.
type KillResponse struct {
	SessionID string `json:"session_id"`
	Killed    bool   `json:"killed"`
}

// ReplayRequest describes a replay request payload.
type ReplayRequest struct {
	SessionID string `json:"session_id"`
}

// ReplayResponse describes a replay response payload.
type ReplayResponse struct {
	SessionID string `json:"session_id"`
	Accepted  bool   `json:"accepted"`
}

// PingPayload describes ping/pong payloads.
type PingPayload struct {
	UnixMS int64 `json:"ts_unix_ms"`
}

// EncodeControlMessage marshals a control message to JSON.
func EncodeControlMessage(msg ControlMessage) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("encode control: %w", err)
	}
	return data, nil
}

// DecodeControlMessage decodes a control message from JSON.
func DecodeControlMessage(data []byte) (ControlMessage, error) {
	var msg ControlMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return ControlMessage{}, fmt.Errorf("decode control: %w", err)
	}
	if msg.Type == "" {
		return ControlMessage{}, fmt.Errorf("decode control: %w", ErrInvalidMessage)
	}
	return msg, nil
}

// NewErrorMessage constructs a control error response.
func NewErrorMessage(requestType, code, message string) (ControlMessage, error) {
	payload, err := json.Marshal(ErrorPayload{
		RequestType: requestType,
		Code:        code,
		Message:     message,
	})
	if err != nil {
		return ControlMessage{}, fmt.Errorf("encode error: %w", err)
	}
	return ControlMessage{Type: "error", Payload: payload}, nil
}

// EncodePayload marshals a payload into a control message.
func EncodePayload(msgType string, payload any) (ControlMessage, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return ControlMessage{}, fmt.Errorf("encode payload: %w", err)
	}
	return ControlMessage{Type: msgType, Payload: data}, nil
}

// DecodePayload decodes a control payload into the provided struct.
func DecodePayload(msg ControlMessage, dest any) error {
	if dest == nil {
		return fmt.Errorf("decode payload: %w", ErrInvalidMessage)
	}
	if len(msg.Payload) == 0 {
		return fmt.Errorf("decode payload: %w", ErrInvalidMessage)
	}
	if err := json.Unmarshal(msg.Payload, dest); err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}
	return nil
}
