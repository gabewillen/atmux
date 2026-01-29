// Package agent provides agent orchestration: lifecycle, presence, and messaging.
// local.go implements local agent lifecycle operations: spawn, stop, restart (spec §5.4, §5.6).
package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/pty"
	"github.com/agentflare-ai/amux/internal/worktree"
	"github.com/agentflare-ai/amux/pkg/api"
)

// LocalSession holds a local agent's actor, PTY session, and worktree path (spec §5.4, §7).
// Lifecycle HSM transitions align to Spawn/Stop; PTY is started in the agent workdir.
type LocalSession struct {
	AgentConfig config.AgentConfig
	RepoRoot    string

	actor *Actor
	sess  *pty.Session
	agent *api.Agent
	disp  protocol.Dispatcher
	mu    sync.Mutex
}

// NewLocalSession creates a session holder for a local agent. Call Spawn to start.
func NewLocalSession(ac config.AgentConfig, repoRoot string, disp protocol.Dispatcher) *LocalSession {
	return &LocalSession{
		AgentConfig: ac,
		RepoRoot:    repoRoot,
		disp:        disp,
	}
}

// Spawn ensures the worktree exists, starts the PTY in the workdir, and runs the lifecycle to Running (spec §5.4, §5.6).
// command is the argv to run in the PTY (e.g. []string{"bash"} or adapter CLI). env is the environment; TERM etc. may be set by caller.
func (s *LocalSession) Spawn(ctx context.Context, command []string, env []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.actor != nil {
		return fmt.Errorf("session already spawned")
	}
	if len(command) == 0 {
		return fmt.Errorf("command is required")
	}

	slug := s.AgentConfig.Slug
	if slug == "" {
		slug = api.NormalizeAgentSlug(s.AgentConfig.Name)
	}
	wtPath, err := worktree.Create(s.RepoRoot, slug)
	if err != nil {
		return fmt.Errorf("worktree: %w", err)
	}

	agentID := api.NextRuntimeID()
	loc := api.Location{Type: api.LocationLocal}
	if s.AgentConfig.Location.Type == "ssh" {
		loc.Type = api.LocationSSH
		loc.Host = s.AgentConfig.Location.Host
		loc.User = s.AgentConfig.Location.User
		loc.Port = s.AgentConfig.Location.Port
		loc.RepoPath = s.AgentConfig.Location.RepoPath
	}
	ag := &api.Agent{
		ID:       agentID,
		Name:     s.AgentConfig.Name,
		About:    s.AgentConfig.About,
		Adapter:  s.AgentConfig.Adapter,
		RepoRoot: s.RepoRoot,
		Worktree: wtPath,
		Location: loc,
	}
	actor, err := NewActor(ag, s.disp)
	if err != nil {
		return fmt.Errorf("new actor: %w", err)
	}
	actor.Start(ctx)
	s.actor = actor
	s.agent = ag

	actor.DispatchLifecycle(ctx, EventLifecycleStart, nil)

	env = append(env, "TERM=xterm-256color")
	sess, err := pty.NewSession(wtPath, command[0], command[1:], env)
	if err != nil {
		actor.DispatchLifecycle(ctx, EventLifecycleError, err)
		return fmt.Errorf("pty: %w", err)
	}
	s.sess = sess
	actor.DispatchLifecycle(ctx, EventLifecycleReady, nil)
	return nil
}

// Stop drains the lifecycle to Terminated and closes the PTY (spec §5.6).
func (s *LocalSession) Stop(ctx context.Context) error {
	s.mu.Lock()
	actor := s.actor
	sess := s.sess
	s.actor = nil
	s.sess = nil
	s.agent = nil
	s.mu.Unlock()

	if actor != nil {
		actor.DispatchLifecycle(ctx, EventLifecycleStop, nil)
	}
	if sess != nil {
		_ = sess.Close()
	}
	return nil
}

// Restart stops then spawns again with the same command. Caller must pass the same command and env as Spawn.
func (s *LocalSession) Restart(ctx context.Context, command []string, env []string) error {
	if err := s.Stop(ctx); err != nil {
		return err
	}
	return s.Spawn(ctx, command, env)
}

// Actor returns the agent actor, or nil if not spawned.
func (s *LocalSession) Actor() *Actor {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.actor
}

// PTY returns the PTY session, or nil if not spawned.
func (s *LocalSession) PTY() *pty.Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sess
}

// Agent returns the api.Agent for this session, or nil if not spawned.
func (s *LocalSession) Agent() *api.Agent {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.agent
}
