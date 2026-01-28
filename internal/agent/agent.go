// Package agent implements agent orchestration (lifecycle, presence, messaging)
package agent

import (
	"errors"

	// Import hsm-go to ensure it's included in the dependencies
	_ "github.com/stateforward/hsm-go"
)

// ErrInvalidAgent is returned when an agent is invalid
var ErrInvalidAgent = errors.New("invalid agent")