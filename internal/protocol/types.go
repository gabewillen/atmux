package protocol

import "time"

// Event represents a system event.
type Event struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Source    string      `json:"source"`
	Payload   interface{} `json:"payload"`
}

// Dispatcher handles event routing.
type Dispatcher interface {
	Dispatch(event Event) error
	Subscribe(pattern string) (<-chan Event, func())
}
