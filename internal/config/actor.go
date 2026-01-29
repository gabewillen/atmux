package config

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/stateforward/hsm-go"
)

// ConfigActor manages live configuration reloading.
type ConfigActor struct {
	hsm.HSM
	opts        LoadOptions
	mu          sync.RWMutex
	current     Config
	watcher     *watcher
	subscribers map[uint64]func(ConfigChange)
	nextSubID   uint64
}

// ConfigModel defines the configuration actor state machine.
var ConfigModel = hsm.Define(
	"config",
	hsm.State(
		"loading",
		hsm.Entry(func(ctx context.Context, actor *ConfigActor, event hsm.Event) {
			actor.loadAll(ctx)
		}),
	),
	hsm.State(
		"ready",
		hsm.Entry(func(ctx context.Context, actor *ConfigActor, event hsm.Event) {
			actor.startWatching(ctx)
		}),
		hsm.Exit(func(ctx context.Context, actor *ConfigActor, event hsm.Event) {
			actor.stopWatching()
		}),
	),
	hsm.State(
		"reloading",
		hsm.Entry(func(ctx context.Context, actor *ConfigActor, event hsm.Event) {
			actor.reload(ctx)
		}),
	),
	hsm.Transition(hsm.On(hsm.Event{Name: ConfigLoaded}), hsm.Source("loading"), hsm.Target("ready")),
	hsm.Transition(hsm.On(hsm.Event{Name: ConfigFileChanged}), hsm.Source("ready"), hsm.Target("reloading")),
	hsm.Transition(hsm.On(hsm.Event{Name: ConfigReloaded}), hsm.Source("reloading"), hsm.Target("ready")),
	hsm.Transition(hsm.On(hsm.Event{Name: ConfigReloadFailed}), hsm.Source("reloading"), hsm.Target("ready")),
	hsm.Initial(hsm.Target("loading")),
)

// StartConfigActor constructs and starts the configuration actor.
func StartConfigActor(ctx context.Context, opts LoadOptions) (*ConfigActor, error) {
	actor := &ConfigActor{
		opts:        opts,
		subscribers: make(map[uint64]func(ConfigChange)),
	}
	started := hsm.Started(ctx, actor, &ConfigModel)
	return started, nil
}

// Current returns the current configuration snapshot.
func (a *ConfigActor) Current() Config {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.current
}

// Subscribe registers a callback for configuration updates.
func (a *ConfigActor) Subscribe(callback func(ConfigChange)) uint64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.nextSubID++
	id := a.nextSubID
	a.subscribers[id] = callback
	return id
}

// Unsubscribe removes a configuration subscription.
func (a *ConfigActor) Unsubscribe(id uint64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.subscribers, id)
}

func (a *ConfigActor) loadAll(ctx context.Context) {
	cfg, err := Load(a.opts)
	if err != nil {
		hsm.Dispatch(ctx, a, hsm.Event{Name: ConfigReloadFailed, Data: err})
		return
	}
	a.mu.Lock()
	a.current = cfg
	a.mu.Unlock()
	hsm.Dispatch(ctx, a, hsm.Event{Name: ConfigLoaded})
}

func (a *ConfigActor) reload(ctx context.Context) {
	a.mu.RLock()
	oldConfig := a.current
	a.mu.RUnlock()
	cfg, err := Load(a.opts)
	if err != nil {
		hsm.Dispatch(ctx, a, hsm.Event{Name: ConfigReloadFailed, Data: err})
		return
	}
	changes := DiffConfig(oldConfig, cfg)
	a.mu.Lock()
	a.current = cfg
	a.mu.Unlock()
	for _, change := range changes {
		a.dispatchChange(ctx, change)
	}
	hsm.Dispatch(ctx, a, hsm.Event{Name: ConfigReloaded})
}

func (a *ConfigActor) dispatchChange(ctx context.Context, change ConfigChange) {
	a.mu.RLock()
	subs := make([]func(ConfigChange), 0, len(a.subscribers))
	for _, cb := range a.subscribers {
		subs = append(subs, cb)
	}
	a.mu.RUnlock()
	for _, cb := range subs {
		cb(change)
	}
	hsm.Dispatch(ctx, a, hsm.Event{Name: ConfigUpdated, Data: change})
}

func (a *ConfigActor) startWatching(ctx context.Context) {
	paths := a.watchPaths()
	a.watcher = newWatcher(paths, func() {
		hsm.Dispatch(ctx, a, hsm.Event{Name: ConfigFileChanged})
	}, a.opts.WatchPollInterval)
	a.watcher.start(ctx)
}

func (a *ConfigActor) stopWatching() {
	if a.watcher == nil {
		return
	}
	a.watcher.stop()
	a.watcher = nil
}

func (a *ConfigActor) watchPaths() []string {
	cfg := a.Current()
	paths := []string{
		a.opts.Resolver.UserConfigPath(),
		a.opts.Resolver.ProjectConfigPath(),
	}
	adapters := AdapterNames(cfg)
	for _, name := range adapters {
		paths = append(paths, a.opts.Resolver.UserAdapterConfigPath(name))
		paths = append(paths, a.opts.Resolver.ProjectAdapterConfigPath(name))
	}
	return uniqueStrings(paths)
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{})
	var result []string
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

type watcher struct {
	paths     []string
	onChange  func()
	pollEvery time.Duration
	cancel    context.CancelFunc
	lastMod   map[string]time.Time
}

func newWatcher(paths []string, onChange func(), pollEvery time.Duration) *watcher {
	if pollEvery <= 0 {
		pollEvery = 2 * time.Second
	}
	return &watcher{
		paths:     paths,
		onChange:  onChange,
		pollEvery: pollEvery,
		lastMod:   make(map[string]time.Time),
	}
}

func (w *watcher) start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	w.cancel = cancel
	for _, path := range w.paths {
		mod, err := modTime(path)
		if err == nil {
			w.lastMod[path] = mod
		}
	}
	go func() {
		ticker := time.NewTicker(w.pollEvery)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				if w.check() {
					w.onChange()
				}
			}
		}
	}()
}

func (w *watcher) stop() {
	if w.cancel == nil {
		return
	}
	w.cancel()
	w.cancel = nil
}

func (w *watcher) check() bool {
	changed := false
	for _, path := range w.paths {
		mod, err := modTime(path)
		if err != nil {
			continue
		}
		last, ok := w.lastMod[path]
		if !ok || mod.After(last) {
			w.lastMod[path] = mod
			changed = true
		}
	}
	return changed
}

func modTime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	if info.IsDir() {
		return time.Time{}, fmt.Errorf("config path is directory")
	}
	return info.ModTime(), nil
}
