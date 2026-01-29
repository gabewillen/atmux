package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/stateforward/hsm-go/muid"
)

// BroadcastID is the reserved runtime ID for broadcast messages.
const BroadcastID muid.MUID = 0

// TargetID is a runtime ID that permits the broadcast sentinel (0).
type TargetID struct {
	value muid.MUID
}

// ParseTargetID parses a base-10 encoded ID string, allowing zero.
func ParseTargetID(raw string) (TargetID, error) {
	if raw == "" {
		return TargetID{}, fmt.Errorf("parse target id: %w", ErrEmptyID)
	}
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return TargetID{}, fmt.Errorf("parse target id: %w", err)
	}
	return TargetID{value: muid.MUID(value)}, nil
}

// TargetIDFromRuntime converts a runtime ID to a target ID.
func TargetIDFromRuntime(id RuntimeID) TargetID {
	return TargetID{value: id.Value()}
}

// Value returns the underlying muid.MUID.
func (id TargetID) Value() muid.MUID {
	return id.value
}

// IsBroadcast reports whether the ID is the broadcast sentinel.
func (id TargetID) IsBroadcast() bool {
	return id.value == 0
}

// String returns the base-10 string form of the ID.
func (id TargetID) String() string {
	return fmt.Sprintf("%d", uint64(id.value))
}

// MarshalJSON encodes the ID as a JSON string containing a base-10 integer.
func (id TargetID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

// UnmarshalJSON decodes a JSON string containing a base-10 integer ID.
func (id *TargetID) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("unmarshal target id: %w", err)
	}
	parsed, err := ParseTargetID(raw)
	if err != nil {
		return fmt.Errorf("unmarshal target id: %w", err)
	}
	*id = parsed
	return nil
}

// OutboundMessage describes an adapter-emitted outbound message payload.
type OutboundMessage struct {
	AgentID   *AgentID `json:"agent_id,omitempty"`
	ToSlug    string   `json:"to_slug"`
	Content   string   `json:"content"`
	ID        string   `json:"id,omitempty"`
	From      string   `json:"from,omitempty"`
	To        string   `json:"to,omitempty"`
	Timestamp string   `json:"timestamp,omitempty"`
}

// AgentMessage represents a participant communication payload.
type AgentMessage struct {
	ID        RuntimeID `json:"id"`
	From      RuntimeID `json:"from"`
	To        TargetID  `json:"to"`
	ToSlug    string    `json:"to_slug"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}
