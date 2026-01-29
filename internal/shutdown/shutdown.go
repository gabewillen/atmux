// Package shutdown provides HSM-based graceful shutdown for amux.
//
// Shutdown is modeled as an event-driven process using HSM transitions
// per spec §5.6. The shutdown sequence is:
//
//	Running → Draining → Stopped    (graceful)
//	Running → Terminating → Stopped (forced)
//	Draining → Terminating → Stopped (drain timeout or second signal)
//
// See spec §5.6.1-§5.6.4 for the shutdown HSM and signal handling.
package shutdown

import (
	"context"
	"sync"
	"time"

	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/internal/session"
)

// State represents the shutdown state.
type State string

const (
	// StateRunning is the normal operating state.
	StateRunning State = "running"

	// StateDraining indicates graceful shutdown is in progress.
	StateDraining State = "draining"

	// StateTerminating indicates forced termination is in progress.
	StateTerminating State = "terminating"

	// StateStopped indicates the system has fully stopped.
	StateStopped State = "stopped"
)

// Controller manages the shutdown process for the amux system.
type Controller struct {
	mu sync.Mutex

	state        State
	drainTimeout time.Duration
	sessions     *session.Manager
	dispatcher   event.Dispatcher

	// done is closed when the system reaches StateStopped.
	done chan struct{}

	// drainTimer fires when the drain timeout expires.
	drainTimer *time.Timer
}

// NewController creates a new shutdown controller.
func NewController(sessions *session.Manager, dispatcher event.Dispatcher, drainTimeout time.Duration) *Controller {
	if dispatcher == nil {
		dispatcher = event.NewNoopDispatcher()
	}
	if drainTimeout <= 0 {
		drainTimeout = 30 * time.Second
	}

	return &Controller{
		state:        StateRunning,
		drainTimeout: drainTimeout,
		sessions:     sessions,
		dispatcher:   dispatcher,
		done:         make(chan struct{}),
	}
}

// RequestShutdown initiates a graceful shutdown (SIGTERM/SIGINT handler).
//
// This transitions the system from Running to Draining. All agents receive
// a shutdown.initiated event and have drainTimeout to terminate gracefully.
// If already draining, this escalates to forced termination.
//
// See spec §5.6.2 for signal mapping.
func (c *Controller) RequestShutdown(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch c.state {
	case StateRunning:
		c.transitionToDraining(ctx)
	case StateDraining:
		// Second signal: escalate to forced termination
		c.transitionToTerminating(ctx)
	case StateTerminating, StateStopped:
		// Already shutting down or stopped
	}
}

// ForceShutdown forces immediate termination.
//
// This transitions directly to Terminating state, killing all sessions.
func (c *Controller) ForceShutdown(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch c.state {
	case StateRunning, StateDraining:
		c.transitionToTerminating(ctx)
	case StateTerminating, StateStopped:
		// Already shutting down or stopped
	}
}

// State returns the current shutdown state.
func (c *Controller) State() State {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}

// Done returns a channel that is closed when the system reaches StateStopped.
func (c *Controller) Done() <-chan struct{} {
	return c.done
}

// transitionToDraining moves to draining state. Caller must hold mu.
func (c *Controller) transitionToDraining(ctx context.Context) {
	c.state = StateDraining

	// Dispatch shutdown.initiated to all agents
	_ = c.dispatcher.Dispatch(ctx, event.NewEvent(event.TypeShutdownInitiated, 0, nil))

	// Stop all sessions gracefully
	go func() {
		c.sessions.StopAll()
		c.onDrainComplete(ctx)
	}()

	// Start drain timeout
	c.drainTimer = time.AfterFunc(c.drainTimeout, func() {
		c.onDrainTimeout(ctx)
	})
}

// transitionToTerminating moves to terminating state. Caller must hold mu.
func (c *Controller) transitionToTerminating(ctx context.Context) {
	c.state = StateTerminating

	// Cancel drain timer if active
	if c.drainTimer != nil {
		c.drainTimer.Stop()
		c.drainTimer = nil
	}

	// Dispatch shutdown.force to all agents
	_ = c.dispatcher.Dispatch(ctx, event.NewEvent(event.TypeShutdownForce, 0, nil))

	// Kill all sessions
	go func() {
		c.sessions.KillAll()
		c.transitionToStopped(ctx)
	}()
}

// transitionToStopped moves to stopped state.
func (c *Controller) transitionToStopped(ctx context.Context) {
	c.mu.Lock()
	if c.state == StateStopped {
		c.mu.Unlock()
		return
	}
	c.state = StateStopped

	if c.drainTimer != nil {
		c.drainTimer.Stop()
		c.drainTimer = nil
	}
	c.mu.Unlock()

	_ = c.dispatcher.Dispatch(ctx, event.NewEvent(event.TypeTerminateComplete, 0, nil))
	close(c.done)
}

// onDrainComplete is called when all sessions have stopped during drain.
func (c *Controller) onDrainComplete(ctx context.Context) {
	c.mu.Lock()
	if c.state != StateDraining {
		c.mu.Unlock()
		return
	}

	if c.drainTimer != nil {
		c.drainTimer.Stop()
		c.drainTimer = nil
	}
	c.mu.Unlock()

	_ = c.dispatcher.Dispatch(ctx, event.NewEvent(event.TypeDrainComplete, 0, nil))
	c.transitionToStopped(ctx)
}

// onDrainTimeout is called when the drain timeout expires.
func (c *Controller) onDrainTimeout(ctx context.Context) {
	c.mu.Lock()
	if c.state != StateDraining {
		c.mu.Unlock()
		return
	}
	c.mu.Unlock()

	_ = c.dispatcher.Dispatch(ctx, event.NewEvent(event.TypeDrainTimeout, 0, nil))

	c.mu.Lock()
	c.transitionToTerminating(ctx)
	c.mu.Unlock()
}
