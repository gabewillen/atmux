package agent

import (
	"github.com/stateforward/hsm-go"
)

// Model defines the agent state machine.
// Phase 1 will implement the full logic.
var Model = hsm.Define("agent",
	hsm.State("pending"),
	hsm.Initial(hsm.Target("pending")),
)
