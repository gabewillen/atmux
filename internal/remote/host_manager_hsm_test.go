package remote

import (
	"context"
	"testing"

	"github.com/stateforward/hsm-go"
)

func TestHostManagerLifecycleTransitions(t *testing.T) {
	lifecycle := newHostManagerLifecycle(&HostManager{})
	started := hsm.Started(context.Background(), lifecycle, &hostManagerLifecycleModel)
	if started.State() != "/host_manager.lifecycle/pending" {
		t.Fatalf("unexpected initial state: %s", started.State())
	}
	<-hsm.Dispatch(started.Context(), started, hsm.Event{Name: hostManagerEventStart})
	if started.State() != "/host_manager.lifecycle/starting" {
		t.Fatalf("unexpected start state: %s", started.State())
	}
	<-hsm.Dispatch(started.Context(), started, hsm.Event{Name: hostManagerEventReady})
	if started.State() != "/host_manager.lifecycle/running" {
		t.Fatalf("unexpected ready state: %s", started.State())
	}
	<-hsm.Dispatch(started.Context(), started, hsm.Event{Name: hostManagerEventStop})
	if started.State() != "/host_manager.lifecycle/terminated" {
		t.Fatalf("unexpected stop state: %s", started.State())
	}
}

func TestHostManagerLifecycleError(t *testing.T) {
	lifecycle := newHostManagerLifecycle(&HostManager{})
	started := hsm.Started(context.Background(), lifecycle, &hostManagerLifecycleModel)
	<-hsm.Dispatch(started.Context(), started, hsm.Event{Name: hostManagerEventError})
	if started.State() != "/host_manager.lifecycle/errored" {
		t.Fatalf("unexpected error state: %s", started.State())
	}
}
