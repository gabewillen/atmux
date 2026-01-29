// Control provides the control plane that wires lifecycle HSM transitions
// to actual session spawn/stop/kill operations.
//
// The SessionSpawner interface breaks the import cycle between agent and
// session packages: agent defines the interface, session/adapter.go adapts
// *session.Manager to satisfy it.
//
// See spec §5.4 for lifecycle state machine and §5.6 for shutdown behavior.
package agent

import (
	"context"
	"fmt"
	"sync"

	hsm "github.com/stateforward/hsm-go"
	"github.com/stateforward/hsm-go/muid"

	amuxerrors "github.com/agentflare-ai/amux/internal/errors"
	"github.com/agentflare-ai/amux/pkg/api"
)

// SessionHandle represents a running session. The agent package uses this
// to monitor session lifetime without importing the session package.
type SessionHandle interface {
	// Done returns a channel that is closed when the session exits.
	Done() <-chan struct{}

	// ExitErr returns the process exit error, or nil if exited cleanly.
	ExitErr() error
}

// SessionSpawner is the interface that the agent control plane uses to
// spawn, stop, and kill sessions. It is satisfied by session.Adapter.
type SessionSpawner interface {
	// SpawnAgent creates and starts a new PTY session for an agent.
	SpawnAgent(ctx context.Context, ag *Agent, shell string, args ...string) (SessionHandle, error)

	// StopAgent gracefully stops the session for an agent.
	StopAgent(ctx context.Context, agentID muid.MUID) error

	// KillAgent forcefully terminates the session for an agent.
	KillAgent(ctx context.Context, agentID muid.MUID) error

	// RemoveSession removes a session from the session manager.
	RemoveSession(agentID muid.MUID)
}

// agentHSMs holds the per-agent lifecycle and presence HSMs plus control state.
type agentHSMs struct {
	lifecycle *LifecycleHSM
	presence  *PresenceHSM

	// lifecycleInstance is the started HSM instance for lifecycle.
	lifecycleInstance hsm.Instance

	// presenceInstance is the started HSM instance for presence.
	presenceInstance hsm.Instance

	// stopping is set to true when Stop or Kill is called intentionally.
	// watchSession uses this to distinguish intentional shutdown from crash.
	stopping bool

	// mu protects the stopping flag.
	mu sync.Mutex
}

// setStopping atomically sets the stopping flag.
func (a *agentHSMs) setStopping(v bool) {
	a.mu.Lock()
	a.stopping = v
	a.mu.Unlock()
}

// isStopping atomically reads the stopping flag.
func (a *agentHSMs) isStopping() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.stopping
}

// SetSessionSpawner sets the session spawner used by control plane methods.
func (m *Manager) SetSessionSpawner(s SessionSpawner) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions = s
}

// Start transitions an agent from Pending to Running by spawning a session.
//
// The lifecycle HSM is driven through: Pending → Starting → Running.
// If the spawn fails, the lifecycle transitions to Errored.
// A watchSession goroutine is launched to monitor the session.
func (m *Manager) Start(ctx context.Context, agentID muid.MUID, shell string, args ...string) error {
	m.mu.RLock()
	agent, ok := m.agents[agentID]
	hsms, hsmOK := m.hsms[agentID]
	spawner := m.sessions
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("agent start: %w", amuxerrors.ErrAgentNotFound)
	}
	if !hsmOK || hsms == nil {
		return fmt.Errorf("agent start: HSMs not initialized for agent %d", agentID)
	}
	if spawner == nil {
		return fmt.Errorf("agent start: session spawner not configured")
	}

	// Verify agent is in Pending state (the only valid starting point)
	if hsms.lifecycle.LifecycleState() != api.LifecyclePending {
		return fmt.Errorf("agent start: agent is in state %q, must be %q",
			hsms.lifecycle.LifecycleState(), api.LifecyclePending)
	}

	// Pending → Starting
	<-DispatchStart(ctx, hsms.lifecycleInstance)

	// Spawn the session
	handle, err := spawner.SpawnAgent(ctx, agent, shell, args...)
	if err != nil {
		// Starting → Errored
		<-DispatchError(ctx, hsms.lifecycleInstance, fmt.Errorf("spawn failed: %w", err))
		return fmt.Errorf("agent start: %w", err)
	}

	// Starting → Running
	<-DispatchReady(ctx, hsms.lifecycleInstance)

	// Launch watcher goroutine with its own context. The watcher outlives
	// the Start call, so it must not use the caller's context.
	go m.watchSession(agentID, handle)

	return nil
}

// Stop gracefully stops an agent's session and waits for it to exit.
//
// The stopping flag is set so watchSession knows this was intentional
// and transitions to Terminated (not Errored).
func (m *Manager) Stop(ctx context.Context, agentID muid.MUID) error {
	m.mu.RLock()
	_, ok := m.agents[agentID]
	hsms, hsmOK := m.hsms[agentID]
	spawner := m.sessions
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("agent stop: %w", amuxerrors.ErrAgentNotFound)
	}
	if !hsmOK || hsms == nil {
		return fmt.Errorf("agent stop: HSMs not initialized for agent %d", agentID)
	}
	if spawner == nil {
		return fmt.Errorf("agent stop: session spawner not configured")
	}

	// Must be Running
	if hsms.lifecycle.LifecycleState() != api.LifecycleRunning {
		return fmt.Errorf("agent stop: %w", amuxerrors.ErrAgentNotRunning)
	}

	// Mark as intentional stop
	hsms.setStopping(true)

	// Stop the session (watchSession handles lifecycle transition)
	if err := spawner.StopAgent(ctx, agentID); err != nil {
		hsms.setStopping(false)
		return fmt.Errorf("agent stop: %w", err)
	}

	return nil
}

// Kill forcefully terminates an agent's session.
//
// Like Stop, the stopping flag is set so watchSession transitions to
// Terminated rather than Errored.
func (m *Manager) Kill(ctx context.Context, agentID muid.MUID) error {
	m.mu.RLock()
	_, ok := m.agents[agentID]
	hsms, hsmOK := m.hsms[agentID]
	spawner := m.sessions
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("agent kill: %w", amuxerrors.ErrAgentNotFound)
	}
	if !hsmOK || hsms == nil {
		return fmt.Errorf("agent kill: HSMs not initialized for agent %d", agentID)
	}
	if spawner == nil {
		return fmt.Errorf("agent kill: session spawner not configured")
	}

	// Must be Running
	if hsms.lifecycle.LifecycleState() != api.LifecycleRunning {
		return fmt.Errorf("agent kill: %w", amuxerrors.ErrAgentNotRunning)
	}

	// Mark as intentional kill
	hsms.setStopping(true)

	// Kill the session (watchSession handles lifecycle transition)
	if err := spawner.KillAgent(ctx, agentID); err != nil {
		hsms.setStopping(false)
		return fmt.Errorf("agent kill: %w", err)
	}

	return nil
}

// watchSession monitors a session and drives the lifecycle HSM when it exits.
//
// Uses context.Background() because this goroutine outlives the Start call
// that launched it. The caller's context may be canceled independently.
//
// If stopping is true (intentional Stop/Kill), lifecycle → Terminated.
// If stopping is false (unexpected crash), lifecycle → Errored.
func (m *Manager) watchSession(agentID muid.MUID, handle SessionHandle) {
	// Block until session exits
	<-handle.Done()

	ctx := context.Background()

	m.mu.RLock()
	hsms, ok := m.hsms[agentID]
	spawner := m.sessions
	m.mu.RUnlock()

	if !ok {
		// Agent was removed while session was running; nothing to do
		return
	}

	// Clean up session from session manager
	if spawner != nil {
		spawner.RemoveSession(agentID)
	}

	if hsms.isStopping() {
		// Intentional stop/kill → Terminated
		<-DispatchStop(ctx, hsms.lifecycleInstance)
	} else {
		// Unexpected exit → Errored
		exitErr := handle.ExitErr()
		if exitErr == nil {
			exitErr = fmt.Errorf("session exited unexpectedly")
		}
		<-DispatchError(ctx, hsms.lifecycleInstance, exitErr)
	}
}

// LifecycleHSMFor returns the lifecycle HSM for an agent, or nil if not found.
func (m *Manager) LifecycleHSMFor(agentID muid.MUID) *LifecycleHSM {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if h, ok := m.hsms[agentID]; ok {
		return h.lifecycle
	}
	return nil
}

// PresenceHSMFor returns the presence HSM for an agent, or nil if not found.
func (m *Manager) PresenceHSMFor(agentID muid.MUID) *PresenceHSM {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if h, ok := m.hsms[agentID]; ok {
		return h.presence
	}
	return nil
}
