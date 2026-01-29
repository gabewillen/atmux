// Package protocol - wire.go provides EventMessage and WireEvent types
// for the hsmnet wire format as specified in §9.1.3.
//
// EventMessage envelopes are published on the host events subject
// (P.events.<host_id>) and carry structured event data.
package protocol

import "encoding/json"

// MessageType represents the event message routing type.
// Encoded as a JSON number per spec §9.1.3.1.
type MessageType uint8

const (
	// MsgBroadcast routes the event to all peers.
	MsgBroadcast MessageType = 1

	// MsgMulticast routes the event to specified targets.
	MsgMulticast MessageType = 2

	// MsgUnicast routes the event to a single target.
	MsgUnicast MessageType = 3
)

// EventMessage is the wire envelope for events transported over NATS.
//
// See spec §9.1.3.
type EventMessage struct {
	// Type is the routing type (1=broadcast, 2=multicast, 3=unicast).
	Type MessageType `json:"type"`

	// Target is the destination peer ID for unicast (base-10 muid.ID string).
	Target string `json:"target,omitempty"`

	// Targets is the list of destination peer IDs for multicast.
	Targets []string `json:"targets,omitempty"`

	// Event is the wrapped event.
	Event WireEvent `json:"event"`
}

// WireEvent carries an event name and its opaque JSON data.
type WireEvent struct {
	// Name is the event type identifier (e.g., "process.spawned").
	Name string `json:"name"`

	// Data is the event payload as raw JSON.
	Data json.RawMessage `json:"data"`
}

// NewBroadcastEvent creates a broadcast EventMessage.
func NewBroadcastEvent(name string, data any) (*EventMessage, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return &EventMessage{
		Type: MsgBroadcast,
		Event: WireEvent{
			Name: name,
			Data: json.RawMessage(raw),
		},
	}, nil
}

// NewUnicastEvent creates a unicast EventMessage.
func NewUnicastEvent(name string, target string, data any) (*EventMessage, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return &EventMessage{
		Type:   MsgUnicast,
		Target: target,
		Event: WireEvent{
			Name: name,
			Data: json.RawMessage(raw),
		},
	}, nil
}

// --- Required event payload schemas per spec §9.1.3.2 ---

// ConnectionEstablishedEvent is the payload for "connection.established" events.
type ConnectionEstablishedEvent struct {
	PeerID    string `json:"peer_id"`
	Timestamp string `json:"timestamp"`
}

// ConnectionLostEvent is the payload for "connection.lost" events.
type ConnectionLostEvent struct {
	PeerID    string `json:"peer_id"`
	Timestamp string `json:"timestamp"`
	Reason    string `json:"reason"`
}

// ConnectionRecoveredEvent is the payload for "connection.recovered" events.
type ConnectionRecoveredEvent struct {
	PeerID    string `json:"peer_id"`
	Timestamp string `json:"timestamp"`
}

// ProcessSpawnedEvent is the payload for "process.spawned" events.
type ProcessSpawnedEvent struct {
	PID       int    `json:"pid"`
	AgentID   string `json:"agent_id"`
	ProcessID string `json:"process_id"`
	Command   string `json:"command"`
	Args      []string `json:"args"`
	WorkDir   string `json:"work_dir"`
	ParentPID int    `json:"parent_pid"`
	StartedAt string `json:"started_at"`
}

// ProcessCompletedEvent is the payload for "process.completed", "process.failed",
// and "process.killed" events.
type ProcessCompletedEvent struct {
	PID       int    `json:"pid"`
	AgentID   string `json:"agent_id"`
	ProcessID string `json:"process_id"`
	Command   string `json:"command"`
	ExitCode  int    `json:"exit_code"`
	Signal    *int   `json:"signal"`
	StartedAt string `json:"started_at"`
	EndedAt   string `json:"ended_at"`
	Duration  string `json:"duration"`
}

// ProcessIOEvent is the payload for "process.stdout", "process.stderr",
// and "process.stdin" events.
type ProcessIOEvent struct {
	PID       int    `json:"pid"`
	AgentID   string `json:"agent_id"`
	ProcessID string `json:"process_id"`
	Command   string `json:"command"`
	Stream    string `json:"stream"`
	DataB64   string `json:"data_b64"`
	Timestamp string `json:"timestamp"`
}
