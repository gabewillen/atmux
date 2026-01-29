package remote

import (
	"encoding/json"
	"fmt"
	"time"
)

// MessageType describes the remote event envelope type.
type MessageType uint8

const (
	// MsgBroadcast is a broadcast message type.
	MsgBroadcast MessageType = 1
	// MsgMulticast is a multicast message type.
	MsgMulticast MessageType = 2
	// MsgUnicast is a unicast message type.
	MsgUnicast MessageType = 3
)

// EventMessage wraps a remote event in a wire envelope.
type EventMessage struct {
	Type    MessageType `json:"type"`
	Target  string      `json:"target,omitempty"`
	Targets []string    `json:"targets,omitempty"`
	Event   WireEvent   `json:"event"`
}

// WireEvent describes an event payload.
type WireEvent struct {
	Name string          `json:"name"`
	Data json.RawMessage `json:"data"`
}

// ConnectionEstablishedPayload is the payload for connection.established.
type ConnectionEstablishedPayload struct {
	PeerID    string `json:"peer_id"`
	Timestamp string `json:"timestamp"`
}

// ConnectionLostPayload is the payload for connection.lost.
type ConnectionLostPayload struct {
	PeerID    string `json:"peer_id"`
	Timestamp string `json:"timestamp"`
	Reason    string `json:"reason"`
}

// ConnectionRecoveredPayload is the payload for connection.recovered.
type ConnectionRecoveredPayload struct {
	PeerID    string `json:"peer_id"`
	Timestamp string `json:"timestamp"`
}

// EncodeEventMessage builds a broadcast event envelope.
func EncodeEventMessage(name string, payload any) (EventMessage, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return EventMessage{}, fmt.Errorf("encode event: %w", err)
	}
	return EventMessage{
		Type: MsgBroadcast,
		Event: WireEvent{
			Name: name,
			Data: data,
		},
	}, nil
}

// EncodeEventMessageJSON marshals an event envelope to JSON.
func EncodeEventMessageJSON(msg EventMessage) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("encode event: %w", err)
	}
	return data, nil
}

// NowRFC3339 returns the current time in RFC3339 UTC format.
func NowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}
