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

	hsm "github.com/stateforward/hsm-go"

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

// HSM event names for shutdown transitions per spec §5.6.1-§5.6.2.
const (
	EventShutdownRequest  = "shutdown.request"
	EventShutdownForce    = "shutdown.force"
	EventDrainComplete    = "drain.complete"
	EventDrainTimeout     = "drain.timeout"
	EventTerminateComplete = "terminate.complete"
)

// ShutdownModel defines the HSM model for system shutdown per spec §5.6.1.
var ShutdownModel = hsm.Define(
	"system.shutdown",

	// States
	hsm.State("running"),
	hsm.State("draining",
		hsm.Entry(func(ctx context.Context, c *Controller, e hsm.Event) {
			c.onEnterDraining(ctx)
		}),
	),
	hsm.State("terminating",
		hsm.Entry(func(ctx context.Context, c *Controller, e hsm.Event) {
			c.onEnterTerminating(ctx)
		}),
	),
	hsm.State("stopped",
		hsm.Entry(func(ctx context.Context, c *Controller, e hsm.Event) {
			c.onEnterStopped(ctx)
		}),
	),

	// Transitions per spec §5.6.1
	hsm.Transition(hsm.On(hsm.Event{Name: EventShutdownRequest}),
		hsm.Source("running"), hsm.Target("draining")),
	hsm.Transition(hsm.On(hsm.Event{Name: EventShutdownForce}),
		hsm.Source("running"), hsm.Target("terminating")),
	hsm.Transition(hsm.On(hsm.Event{Name: EventShutdownForce}),
		hsm.Source("draining"), hsm.Target("terminating")),
	hsm.Transition(hsm.On(hsm.Event{Name: EventDrainComplete}),
		hsm.Source("draining"), hsm.Target("stopped")),
	hsm.Transition(hsm.On(hsm.Event{Name: EventDrainTimeout}),
		hsm.Source("draining"), hsm.Target("terminating")),
	hsm.Transition(hsm.On(hsm.Event{Name: EventTerminateComplete}),
		hsm.Source("terminating"), hsm.Target("stopped")),

	hsm.Initial(hsm.Target("running")),
)

// Controller manages the shutdown process for the amux system using
// an HSM-driven state machine per spec §5.6.1.
type Controller struct {
	hsm.HSM

	mu           sync.Mutex
	state        State
	drainTimeout time.Duration
	sessions     *session.Manager
	dispatcher   event.Dispatcher
	requested    bool

	// done is closed when the system reaches StateStopped.
	done chan struct{}
	// doneOnce prevents double-close of the done channel.
	doneOnce sync.Once

	// drainTimer fires when the drain timeout expires.
	drainTimer *time.Timer
}

// NewController creates a new shutdown controller. The HSM is initialized
// in the Running state per spec §5.6.1.
func NewController(sessions *session.Manager, dispatcher event.Dispatcher, drainTimeout time.Duration) *Controller {
	if dispatcher == nil {
		dispatcher = event.NewNoopDispatcher()
	}
	if drainTimeout <= 0 {
		drainTimeout = 30 * time.Second
	}

	c := &Controller{
		state:        StateRunning,
		drainTimeout: drainTimeout,
		sessions:     sessions,
		dispatcher:   dispatcher,
		done:         make(chan struct{}),
	}

	hsm.Started(context.Background(), c, &ShutdownModel)
	return c
}

// RequestShutdown initiates a graceful shutdown (SIGTERM/SIGINT handler).
//
// First call dispatches shutdown.request (Running → Draining).
// Second call dispatches shutdown.force (Draining → Terminating).
//
// See spec §5.6.2 for signal mapping.
func (c *Controller) RequestShutdown(ctx context.Context) {
	c.mu.Lock()
	alreadyRequested := c.requested
	c.requested = true
	c.mu.Unlock()

	if alreadyRequested {
		// Second signal: escalate to forced termination (spec §5.6.2)
		hsm.Dispatch(ctx, c, hsm.Event{Name: EventShutdownForce})
	} else {
		hsm.Dispatch(ctx, c, hsm.Event{Name: EventShutdownRequest})
	}
}

// ForceShutdown forces immediate termination.
//
// This dispatches shutdown.force, transitioning to Terminating from
// either Running or Draining state.
func (c *Controller) ForceShutdown(ctx context.Context) {
	hsm.Dispatch(ctx, c, hsm.Event{Name: EventShutdownForce})
}

// ShutdownState returns the current shutdown state.
// Named ShutdownState (not State) to avoid shadowing hsm.HSM.State()
// which is required by the hsm.Instance interface.
func (c *Controller) ShutdownState() State {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}

// Done returns a channel that is closed when the system reaches StateStopped.
func (c *Controller) Done() <-chan struct{} {
	return c.done
}

// onEnterDraining is the entry action for the draining state.
// It dispatches shutdown.initiated, stops all sessions gracefully,
// and starts the drain timeout timer per spec §5.6.3-§5.6.4.
func (c *Controller) onEnterDraining(ctx context.Context) {
	c.mu.Lock()
	c.state = StateDraining
	c.mu.Unlock()

	// Dispatch shutdown.initiated to all agents
	_ = c.dispatcher.Dispatch(ctx, event.NewEvent(event.TypeShutdownInitiated, 0, nil))

	// Stop all sessions gracefully in background
	go func() {
		c.sessions.StopAll()
		hsm.Dispatch(ctx, c, hsm.Event{Name: EventDrainComplete})
	}()

	// Start drain timeout per spec §5.6.4
	c.mu.Lock()
	c.drainTimer = time.AfterFunc(c.drainTimeout, func() {
		hsm.Dispatch(ctx, c, hsm.Event{Name: EventDrainTimeout})
	})
	c.mu.Unlock()
}

// onEnterTerminating is the entry action for the terminating state.
// It cancels the drain timer, dispatches shutdown.force, and kills
// all sessions per spec §5.6.4.
func (c *Controller) onEnterTerminating(ctx context.Context) {
	c.mu.Lock()
	c.state = StateTerminating
	if c.drainTimer != nil {
		c.drainTimer.Stop()
		c.drainTimer = nil
	}
	c.mu.Unlock()

	// Dispatch shutdown.force to all agents
	_ = c.dispatcher.Dispatch(ctx, event.NewEvent(event.TypeShutdownForce, 0, nil))

	// Kill all sessions in background
	go func() {
		c.sessions.KillAll()
		hsm.Dispatch(ctx, c, hsm.Event{Name: EventTerminateComplete})
	}()
}

// onEnterStopped is the entry action for the stopped state.
// It closes the done channel to signal shutdown completion.
func (c *Controller) onEnterStopped(ctx context.Context) {
	c.mu.Lock()
	c.state = StateStopped
	if c.drainTimer != nil {
		c.drainTimer.Stop()
		c.drainTimer = nil
	}
	c.mu.Unlock()

	_ = c.dispatcher.Dispatch(ctx, event.NewEvent(event.TypeTerminateComplete, 0, nil))
	c.doneOnce.Do(func() {
		close(c.done)
	})
}
