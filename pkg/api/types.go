// Package api contains public API types (Agent.Adapter is a string)
package api

import (
	"github.com/stateforward/hsm-go/muid"
)

// Agent represents an agent instance
type Agent struct {
	ID       muid.MUID `json:"id"`
	Name     string    `json:"name"`
	Adapter  string    `json:"adapter"`
	Location string    `json:"location,omitempty"`
}

// Session represents an amux session
type Session struct {
	ID     muid.MUID `json:"id"`
	Agents []Agent   `json:"agents"`
}