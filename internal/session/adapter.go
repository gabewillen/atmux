// Adapter bridges the session.Manager to the agent.SessionSpawner interface,
// breaking the import cycle between agent and session packages.
//
// The agent package defines SessionSpawner; this adapter wraps *session.Manager
// to satisfy it. The agent package never imports session directly.
package session

import (
	"context"

	"github.com/stateforward/hsm-go/muid"

	"github.com/agentflare-ai/amux/internal/agent"
)

// Adapter wraps a *Manager to satisfy agent.SessionSpawner.
type Adapter struct {
	mgr *Manager
}

// NewAdapter creates a new session adapter for the given session manager.
func NewAdapter(mgr *Manager) *Adapter {
	return &Adapter{mgr: mgr}
}

// SpawnAgent creates and starts a new PTY session for an agent.
// Returns a SessionHandle that can be used to monitor the session.
func (a *Adapter) SpawnAgent(ctx context.Context, ag *agent.Agent, shell string, args ...string) (agent.SessionHandle, error) {
	sess, err := a.mgr.Spawn(ctx, ag, shell, args...)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

// StopAgent gracefully stops the session for an agent.
func (a *Adapter) StopAgent(ctx context.Context, agentID muid.MUID) error {
	return a.mgr.Stop(ctx, agentID)
}

// KillAgent forcefully terminates the session for an agent.
func (a *Adapter) KillAgent(ctx context.Context, agentID muid.MUID) error {
	return a.mgr.Kill(ctx, agentID)
}

// RemoveSession removes a session from the session manager.
func (a *Adapter) RemoveSession(agentID muid.MUID) {
	a.mgr.Remove(agentID)
}
