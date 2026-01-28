package process

import (
	"context"
	"fmt"

	"github.com/stateforward/hsm-go/muid"
)

// Tracker observes process lifecycle events.
type Tracker struct{}

// Start begins tracking processes for the given agent.
func (t *Tracker) Start(ctx context.Context, agentID muid.MUID) error {
	if agentID == 0 {
		return fmt.Errorf("tracker start: agent id is zero")
	}
	return nil
}
