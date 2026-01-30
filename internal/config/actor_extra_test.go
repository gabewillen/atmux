package config

import "testing"

func TestConfigActorSubscribeUnsubscribe(t *testing.T) {
	t.Parallel()
	actor := &ConfigActor{subscribers: make(map[uint64]func(ConfigChange))}
	id := actor.Subscribe(func(ConfigChange) {})
	if id == 0 {
		t.Fatalf("expected non-zero subscription id")
	}
	if len(actor.subscribers) != 1 {
		t.Fatalf("expected subscriber")
	}
	actor.Unsubscribe(id)
	if len(actor.subscribers) != 0 {
		t.Fatalf("expected unsubscribe to remove subscriber")
	}
}

