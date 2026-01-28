// Package api contains public API types (Agent.Adapter is a string)
package api

import (
	"encoding/json"
)

// ID represents a unique identifier for entities
type ID uint64

// MarshalJSON implements json.Marshaler for ID
func (id ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint64(id))
}

// UnmarshalJSON implements json.Unmarshaler for ID
func (id *ID) UnmarshalJSON(data []byte) error {
	var i uint64
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}
	*id = ID(i)
	return nil
}

// Agent represents an agent instance
type Agent struct {
	ID       ID     `json:"id"`
	Name     string `json:"name"`
	Adapter  string `json:"adapter"`
	Location string `json:"location,omitempty"`
}

// Session represents an amux session
type Session struct {
	ID     ID     `json:"id"`
	Agents []Agent `json:"agents"`
}