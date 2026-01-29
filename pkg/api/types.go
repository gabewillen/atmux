package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/stateforward/hsm-go/muid"
)

// AdapterRef is the string name of an adapter loaded from the WASM registry.
type AdapterRef string

var (
	// ErrEmptyID is returned when parsing an empty ID string.
	ErrEmptyID = errors.New("empty id")
	// ErrZeroID is returned when an ID is zero or reserved.
	ErrZeroID = errors.New("zero id")
)

// RuntimeID is a JSON-safe wrapper around muid.MUID that encodes as base-10 strings.
type RuntimeID struct {
	value muid.MUID
}

var defaultIDGenerator = muid.NewGenerator(muid.DefaultConfig(), 0, 0)

// NewRuntimeID creates a new non-zero ID suitable for runtime use.
func NewRuntimeID() RuntimeID {
	for {
		id := defaultIDGenerator.ID()
		if id != 0 {
			return RuntimeID{value: id}
		}
	}
}

// ParseRuntimeID parses a base-10 encoded ID string.
func ParseRuntimeID(raw string) (RuntimeID, error) {
	if raw == "" {
		return RuntimeID{}, fmt.Errorf("parse id: %w", ErrEmptyID)
	}
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return RuntimeID{}, fmt.Errorf("parse id: %w", err)
	}
	if value == 0 {
		return RuntimeID{}, fmt.Errorf("parse id: %w", ErrZeroID)
	}
	return RuntimeID{value: muid.MUID(value)}, nil
}

// MustParseRuntimeID parses a base-10 encoded ID string and panics on failure.
func MustParseRuntimeID(raw string) RuntimeID {
	id, err := ParseRuntimeID(raw)
	if err != nil {
		panic(err)
	}
	return id
}

// Value returns the underlying muid.MUID.
func (id RuntimeID) Value() muid.MUID {
	return id.value
}

// IsZero reports whether the ID is the reserved zero value.
func (id RuntimeID) IsZero() bool {
	return id.value == 0
}

// String returns the base-10 string form of the ID.
func (id RuntimeID) String() string {
	return fmt.Sprintf("%d", uint64(id.value))
}

// MarshalJSON encodes the ID as a JSON string containing a base-10 integer.
func (id RuntimeID) MarshalJSON() ([]byte, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("marshal id: %w", ErrZeroID)
	}
	return json.Marshal(id.String())
}

// UnmarshalJSON decodes a JSON string containing a base-10 integer ID.
func (id *RuntimeID) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("unmarshal id: %w", err)
	}
	parsed, err := ParseRuntimeID(raw)
	if err != nil {
		return fmt.Errorf("unmarshal id: %w", err)
	}
	*id = parsed
	return nil
}

// AgentID is the runtime identifier for an agent.
type AgentID struct {
	RuntimeID
}

// NewAgentID creates a new agent runtime ID.
func NewAgentID() AgentID {
	return AgentID{RuntimeID: NewRuntimeID()}
}

// ParseAgentID parses a base-10 encoded agent ID string.
func ParseAgentID(raw string) (AgentID, error) {
	id, err := ParseRuntimeID(raw)
	if err != nil {
		return AgentID{}, fmt.Errorf("parse agent id: %w", err)
	}
	return AgentID{RuntimeID: id}, nil
}

// MustParseAgentID parses a base-10 encoded agent ID string and panics on failure.
func MustParseAgentID(raw string) AgentID {
	id, err := ParseAgentID(raw)
	if err != nil {
		panic(err)
	}
	return id
}

// SessionID is the runtime identifier for a session.
type SessionID struct {
	RuntimeID
}

// NewSessionID creates a new session runtime ID.
func NewSessionID() SessionID {
	return SessionID{RuntimeID: NewRuntimeID()}
}

// ParseSessionID parses a base-10 encoded session ID string.
func ParseSessionID(raw string) (SessionID, error) {
	id, err := ParseRuntimeID(raw)
	if err != nil {
		return SessionID{}, fmt.Errorf("parse session id: %w", err)
	}
	return SessionID{RuntimeID: id}, nil
}

// MustParseSessionID parses a base-10 encoded session ID string and panics on failure.
func MustParseSessionID(raw string) SessionID {
	id, err := ParseSessionID(raw)
	if err != nil {
		panic(err)
	}
	return id
}

// PeerID is the runtime identifier for a peer.
type PeerID struct {
	RuntimeID
}

// NewPeerID creates a new peer runtime ID.
func NewPeerID() PeerID {
	return PeerID{RuntimeID: NewRuntimeID()}
}

// ParsePeerID parses a base-10 encoded peer ID string.
func ParsePeerID(raw string) (PeerID, error) {
	id, err := ParseRuntimeID(raw)
	if err != nil {
		return PeerID{}, fmt.Errorf("parse peer id: %w", err)
	}
	return PeerID{RuntimeID: id}, nil
}

// MustParsePeerID parses a base-10 encoded peer ID string and panics on failure.
func MustParsePeerID(raw string) PeerID {
	id, err := ParsePeerID(raw)
	if err != nil {
		panic(err)
	}
	return id
}

// HostID is the identifier for a host manager.
type HostID string

// ParseHostID validates a host ID.
func ParseHostID(raw string) (HostID, error) {
	if raw == "" {
		return "", fmt.Errorf("parse host id: %w", ErrEmptyID)
	}
	return HostID(raw), nil
}

// MustParseHostID parses a host ID and panics on failure.
func MustParseHostID(raw string) HostID {
	id, err := ParseHostID(raw)
	if err != nil {
		panic(err)
	}
	return id
}

// String returns the host ID as a string.
func (id HostID) String() string {
	return string(id)
}
