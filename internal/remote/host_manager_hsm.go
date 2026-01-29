package remote

import (
	"context"

	"github.com/stateforward/hsm-go"
)

const (
	hostManagerLifecyclePending    = "pending"
	hostManagerLifecycleStarting   = "starting"
	hostManagerLifecycleRunning    = "running"
	hostManagerLifecycleTerminated = "terminated"
	hostManagerLifecycleErrored    = "errored"

	hostManagerEventStart = "start"
	hostManagerEventReady = "ready"
	hostManagerEventStop  = "stop"
	hostManagerEventError = "error"
)

var hostManagerLifecycleModel = hsm.Define(
	"host_manager.lifecycle",
	hsm.State(hostManagerLifecyclePending),
	hsm.State(hostManagerLifecycleStarting),
	hsm.State(hostManagerLifecycleRunning),
	hsm.Final(hostManagerLifecycleTerminated),
	hsm.Final(hostManagerLifecycleErrored),

	hsm.Transition(hsm.On(hsm.Event{Name: hostManagerEventStart}), hsm.Source(hostManagerLifecyclePending), hsm.Target(hostManagerLifecycleStarting)),
	hsm.Transition(hsm.On(hsm.Event{Name: hostManagerEventReady}), hsm.Source(hostManagerLifecycleStarting), hsm.Target(hostManagerLifecycleRunning)),
	hsm.Transition(hsm.On(hsm.Event{Name: hostManagerEventStop}), hsm.Source(hostManagerLifecycleRunning), hsm.Target(hostManagerLifecycleTerminated)),
	hsm.Transition(hsm.On(hsm.Event{Name: hostManagerEventError}), hsm.Source(hostManagerLifecyclePending), hsm.Target(hostManagerLifecycleErrored)),
	hsm.Transition(hsm.On(hsm.Event{Name: hostManagerEventError}), hsm.Source(hostManagerLifecycleStarting), hsm.Target(hostManagerLifecycleErrored)),
	hsm.Transition(hsm.On(hsm.Event{Name: hostManagerEventError}), hsm.Source(hostManagerLifecycleRunning), hsm.Target(hostManagerLifecycleErrored)),

	hsm.Initial(hsm.Target(hostManagerLifecyclePending)),
)

// HostManagerLifecycle drives the host manager lifecycle state machine.
type HostManagerLifecycle struct {
	hsm.HSM
	manager *HostManager
}

func newHostManagerLifecycle(manager *HostManager) *HostManagerLifecycle {
	return &HostManagerLifecycle{manager: manager}
}

// Start starts the host manager lifecycle state machine.
func (l *HostManagerLifecycle) Start(ctx context.Context) {
	hsm.Started(ctx, l, &hostManagerLifecycleModel)
}
