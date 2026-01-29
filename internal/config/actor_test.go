package config_test

import (
	"context"
	"testing"
	"time"

	"github.com/stateforward/amux/internal/config"
)

// TestActorConfigUpdatedEvents verifies that the config actor emits per-key
// ConfigChange events on reload per spec §4.2.8.8.
func TestActorConfigUpdatedEvents(t *testing.T) {
	ctx := context.Background()

	// Initial config
	initial := &config.Config{
		General: config.GeneralConfig{
			LogLevel:  "info",
			LogFormat: "text",
		},
		Timeouts: config.TimeoutsConfig{
			Idle:  config.Duration(30 * time.Second),
			Stuck: config.Duration(5 * time.Minute),
		},
	}

	// Mutable loader that we can swap between calls.
	current := initial
	loader := func() (*config.Config, error) {
		return current, nil
	}

	actor, err := config.NewActor(ctx, loader)
	if err != nil {
		t.Fatalf("NewActor error: %v", err)
	}

	ch := actor.Subscribe()

	// Update a couple of fields and trigger reload.
	updated := &config.Config{
		General: config.GeneralConfig{
			LogLevel:  "debug",
			LogFormat: "json",
		},
		Timeouts: config.TimeoutsConfig{
			Idle:  config.Duration(45 * time.Second),
			Stuck: config.Duration(10 * time.Minute),
		},
	}

	current = updated
	actor.NotifyFileChanged(ctx)

	// Collect a few changes; we expect specific per-key paths rather than a
	// wildcard "*" change.
	got := map[string]config.ConfigChange{}

	timeout := time.After(2 * time.Second)

collectLoop:
	for {
		select {
		case change := <-ch:
			got[change.Path] = change
			// We expect at least these four paths; once all are observed we
			// can finish early.
			if len(got) >= 4 {
				break collectLoop
			}
		case <-timeout:
			break collectLoop
		}
	}

	// We should see per-key paths for changed values.
	for _, path := range []string{
		"general.log_level",
		"general.log_format",
		"timeouts.idle",
		"timeouts.stuck",
	} {
		if _, ok := got[path]; !ok {
			t.Errorf("expected ConfigChange for path %q", path)
		}
	}

	// Ensure we no longer emit the coarse wildcard path.
	if _, ok := got["*"]; ok {
		t.Errorf("did not expect wildcard ConfigChange path '*'")
	}
}
