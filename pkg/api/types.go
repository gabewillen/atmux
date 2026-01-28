package api

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/stateforward/hsm-go/muid"
)

// AdapterRef is the string name of an adapter loaded from the WASM registry.
type AdapterRef string

// ID is a JSON-safe wrapper around muid.MUID that encodes as base-10 strings.
type ID struct {
	value muid.MUID
}

var defaultIDGenerator = muid.NewGenerator(muid.DefaultConfig(), 0, 0)

// NewID creates a new non-zero ID suitable for runtime use.
func NewID() ID {
	for {
		id := defaultIDGenerator.ID()
		if id != 0 {
			return ID{value: id}
		}
	}
}

// ParseID parses a base-10 encoded ID string.
func ParseID(raw string) (ID, error) {
	if raw == "" {
		return ID{}, fmt.Errorf("parse id: empty string")
	}
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return ID{}, fmt.Errorf("parse id: %w", err)
	}
	if value == 0 {
		return ID{}, fmt.Errorf("parse id: zero is reserved")
	}
	return ID{value: muid.MUID(value)}, nil
}

// MustParseID parses a base-10 encoded ID string and panics on failure.
func MustParseID(raw string) ID {
	id, err := ParseID(raw)
	if err != nil {
		panic(err)
	}
	return id
}

// Value returns the underlying muid.MUID.
func (id ID) Value() muid.MUID {
	return id.value
}

// String returns the base-10 string form of the ID.
func (id ID) String() string {
	return fmt.Sprintf("%d", uint64(id.value))
}

// MarshalJSON encodes the ID as a JSON string containing a base-10 integer.
func (id ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

// UnmarshalJSON decodes a JSON string containing a base-10 integer ID.
func (id *ID) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("unmarshal id: %w", err)
	}
	parsed, err := ParseID(raw)
	if err != nil {
		return fmt.Errorf("unmarshal id: %w", err)
	}
	*id = parsed
	return nil
}
