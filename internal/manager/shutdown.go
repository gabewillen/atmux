package manager

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/session"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go"
)

const (
	shutdownStateRunning     = "running"
	shutdownStateDraining    = "draining"
	shutdownStateTerminating = "terminating"
	shutdownStateStopped     = "stopped"

	shutdownEventRequest           = "shutdown.request"
	shutdownEventForce             = "shutdown.force"
	shutdownEventDrainComplete     = "drain.complete"
	shutdownEventDrainTimeout      = "drain.timeout"
	shutdownEventTerminateComplete = "terminate.complete"
)

var shutdownModel = hsm.Define(
	"system.shutdown",
	hsm.State(shutdownStateRunning),
	hsm.State(
		shutdownStateDraining,
		hsm.Entry(func(ctx context.Context, actor *shutdownController, event hsm.Event) {
			actor.onDraining(ctx)
		}),
	),
	hsm.State(
		shutdownStateTerminating,
		hsm.Entry(func(ctx context.Context, actor *shutdownController, event hsm.Event) {
			actor.onTerminating(ctx)
		}),
	),
	hsm.Final(shutdownStateStopped),

	hsm.Transition(hsm.On(hsm.Event{Name: shutdownEventRequest}), hsm.Source(shutdownStateRunning), hsm.Target(shutdownStateDraining)),
	hsm.Transition(hsm.On(hsm.Event{Name: shutdownEventForce}), hsm.Source(shutdownStateRunning), hsm.Target(shutdownStateTerminating)),
	hsm.Transition(hsm.On(hsm.Event{Name: shutdownEventForce}), hsm.Source(shutdownStateDraining), hsm.Target(shutdownStateTerminating)),
	hsm.Transition(
		hsm.On(hsm.Event{Name: shutdownEventDrainComplete}),
		hsm.Source(shutdownStateDraining),
		hsm.Target(shutdownStateStopped),
		hsm.Effect(func(ctx context.Context, actor *shutdownController, event hsm.Event) {
			actor.onStopped()
		}),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: shutdownEventDrainTimeout}),
		hsm.Source(shutdownStateDraining),
		hsm.Target(shutdownStateTerminating),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: shutdownEventTerminateComplete}),
		hsm.Source(shutdownStateTerminating),
		hsm.Target(shutdownStateStopped),
		hsm.Effect(func(ctx context.Context, actor *shutdownController, event hsm.Event) {
			actor.onStopped()
		}),
	),

	hsm.Initial(hsm.Target(shutdownStateRunning)),
)

type shutdownController struct {
	hsm.HSM
	manager *LocalManager
	done    chan struct{}
	errMu   sync.Mutex
	err     error
	once    sync.Once
}

type shutdownTarget struct {
	id       api.AgentID
	repoRoot string
	slug     string
	session  *session.LocalSession
	runtime  *agent.Agent
}

func newShutdownController(m *LocalManager) *shutdownController {
	return &shutdownController{
		manager: m,
		done:    make(chan struct{}),
	}
}

func (s *shutdownController) signal(ctx context.Context, name string, payload any) {
	if s == nil || s.manager == nil {
		return
	}
	if ctx == nil || ctx.Err() != nil {
		ctx = context.Background()
	}
	s.manager.emitSystemEvent(ctx, name, payload)
	<-hsm.Dispatch(s.Context(), s, hsm.Event{Name: name, Data: payload})
}

func (s *shutdownController) wait(ctx context.Context) error {
	if s == nil {
		return nil
	}
	select {
	case <-s.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *shutdownController) onDraining(ctx context.Context) {
	if s == nil || s.manager == nil {
		return
	}
	targets := s.manager.shutdownTargets()
	s.manager.dispatchAgentLifecycle(ctx, targets, agent.EventTaskCancel)
	s.manager.dispatchAgentLifecycle(ctx, targets, agent.EventShutdownInitiated)
	go func() {
		timedOut, err := s.manager.drainSessions(ctx, targets)
		if err != nil {
			s.recordError(err)
		}
		if timedOut {
			s.signal(context.Background(), shutdownEventDrainTimeout, map[string]any{})
			return
		}
		s.manager.clearSessions(targets)
		s.signal(context.Background(), shutdownEventDrainComplete, map[string]any{})
	}()
}

func (s *shutdownController) onTerminating(ctx context.Context) {
	if s == nil || s.manager == nil {
		return
	}
	targets := s.manager.shutdownTargets()
	s.manager.dispatchAgentLifecycle(ctx, targets, agent.EventShutdownForce)
	go func() {
		if err := s.manager.forceTerminate(ctx, targets); err != nil {
			s.recordError(err)
		}
		s.manager.clearSessions(targets)
		s.signal(context.Background(), shutdownEventTerminateComplete, map[string]any{})
	}()
}

func (s *shutdownController) onStopped() {
	if s == nil {
		return
	}
	s.once.Do(func() {
		close(s.done)
	})
}

func (s *shutdownController) recordError(err error) {
	if err == nil {
		return
	}
	s.errMu.Lock()
	defer s.errMu.Unlock()
	s.err = errors.Join(s.err, err)
}

func (s *shutdownController) error() error {
	s.errMu.Lock()
	defer s.errMu.Unlock()
	return s.err
}

func (m *LocalManager) shutdownTargets() []shutdownTarget {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	targets := make([]shutdownTarget, 0, len(m.agents))
	for id, state := range m.agents {
		if state == nil {
			continue
		}
		targets = append(targets, shutdownTarget{
			id:       id,
			repoRoot: state.repoRoot,
			slug:     state.slug,
			session:  state.session,
			runtime:  state.runtime,
		})
	}
	return targets
}

func (m *LocalManager) dispatchAgentLifecycle(ctx context.Context, targets []shutdownTarget, name string) {
	if m == nil {
		return
	}
	for _, target := range targets {
		if target.runtime == nil {
			continue
		}
		_ = target.runtime.EmitLifecycle(ctx, name, nil)
	}
}

func (m *LocalManager) drainSessions(ctx context.Context, targets []shutdownTarget) (bool, error) {
	if m == nil {
		return false, nil
	}
	timeout := m.cfg.Shutdown.DrainTimeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	drainCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var wg sync.WaitGroup
	errCh := make(chan error, len(targets))
	for _, target := range targets {
		if target.session == nil {
			continue
		}
		wg.Add(1)
		go func(sess *session.LocalSession) {
			defer wg.Done()
			if err := sess.Stop(drainCtx); err != nil {
				if drainCtx.Err() != nil {
					_ = sess.Kill(context.Background())
				}
				errCh <- err
			}
		}(target.session)
	}
	wg.Wait()
	close(errCh)
	var errOut error
	for err := range errCh {
		errOut = errors.Join(errOut, err)
	}
	return errors.Is(drainCtx.Err(), context.DeadlineExceeded), errOut
}

func (m *LocalManager) forceTerminate(ctx context.Context, targets []shutdownTarget) error {
	if m == nil {
		return nil
	}
	var wg sync.WaitGroup
	errCh := make(chan error, len(targets))
	for _, target := range targets {
		if target.session == nil {
			continue
		}
		wg.Add(1)
		go func(sess *session.LocalSession) {
			defer wg.Done()
			if err := sess.Kill(ctx); err != nil {
				if !errors.Is(err, session.ErrSessionNotRunning) {
					errCh <- err
				}
			}
		}(target.session)
	}
	wg.Wait()
	close(errCh)
	var errOut error
	for err := range errCh {
		errOut = errors.Join(errOut, err)
	}
	return errOut
}

func (m *LocalManager) clearSessions(targets []shutdownTarget) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, target := range targets {
		state := m.agents[target.id]
		if state == nil {
			continue
		}
		if state.session == target.session {
			state.session = nil
		}
	}
}

func (m *LocalManager) cleanupWorktrees(ctx context.Context, targets []shutdownTarget) error {
	if m == nil {
		return nil
	}
	if !m.cfg.Shutdown.CleanupWorktrees {
		return nil
	}
	var errOut error
	for _, target := range targets {
		if target.slug == "" || target.repoRoot == "" {
			continue
		}
		if err := m.git.RemoveWorktree(ctx, target.repoRoot, target.slug, false); err != nil {
			errOut = errors.Join(errOut, fmt.Errorf("cleanup worktree: %w", err))
		}
	}
	return errOut
}

func (m *LocalManager) emitAgentEvent(ctx context.Context, name string, payload any) {
	m.emitEvent(ctx, "agent", name, payload)
}

func (m *LocalManager) emitSystemEvent(ctx context.Context, name string, payload any) {
	m.emitEvent(ctx, "system", name, payload)
}

func (m *LocalManager) emitEvent(ctx context.Context, category string, name string, payload any) {
	if m == nil || m.dispatcher == nil {
		return
	}
	event := protocol.Event{Name: name, Payload: payload, OccurredAt: time.Now().UTC()}
	subject := protocol.Subject("events", category)
	_ = m.dispatcher.Publish(ctx, subject, event)
}

func (m *LocalManager) ensureShutdownController() *shutdownController {
	if m == nil {
		return nil
	}
	m.shutdownMu.Lock()
	defer m.shutdownMu.Unlock()
	if m.shutdown != nil {
		return m.shutdown
	}
	controller := newShutdownController(m)
	hsm.Started(context.Background(), controller, &shutdownModel)
	m.shutdown = controller
	return controller
}

func (m *LocalManager) releaseShutdownController(controller *shutdownController) {
	if m == nil || controller == nil {
		return
	}
	select {
	case <-controller.done:
	default:
		return
	}
	m.shutdownMu.Lock()
	defer m.shutdownMu.Unlock()
	if m.shutdown == controller {
		m.shutdown = nil
	}
}
