package agent

import (
	"context"
	"testing"

	"github.com/stateforward/hsm-go"
)

type testActor struct {
	hsm.HSM
}

func TestHSMDispatch(t *testing.T) {
	model := hsm.Define(
		"test",
		hsm.State("pending"),
		hsm.State("running"),
		hsm.Transition(hsm.On(hsm.Event{Name: "start"}), hsm.Source("pending"), hsm.Target("running")),
		hsm.Initial(hsm.Target("pending")),
	)
	actor := &testActor{}
	started := hsm.Started(context.Background(), actor, &model)
	<-hsm.Dispatch(started.Context(), started, hsm.Event{Name: "start"})
	if started.State() == "" {
		t.Fatalf("state missing")
	}
	if started.State() != "/test/running" {
		t.Fatalf("unexpected state: %s", started.State())
	}
}
