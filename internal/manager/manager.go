package manager

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/git"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/remote"
	"github.com/agentflare-ai/amux/internal/session"
	"github.com/agentflare-ai/amux/pkg/api"
)

var (
	// ErrAgentNotFound is returned when an agent cannot be found.
	ErrAgentNotFound = errors.New("agent not found")
	// ErrAgentAmbiguous is returned when a name matches multiple agents.
	ErrAgentAmbiguous = errors.New("agent name is ambiguous")
	// ErrAgentInvalid is returned when an agent request is invalid.
	ErrAgentInvalid = errors.New("agent invalid")
	// ErrRepoPathRequired is returned when repo_path is required by the spec.
	ErrRepoPathRequired = errors.New("repo path required")
	// ErrMessageTargetUnknown is returned when a message recipient cannot be resolved.
	ErrMessageTargetUnknown = errors.New("message target unknown")
)

// AddRequest describes an agent add request.
type AddRequest struct {
	Name           string
	About          string
	Adapter        string
	Location       api.Location
	Cwd            string
	ListenChannels []string
}

// RemoveRequest describes an agent removal request.
type RemoveRequest struct {
	AgentID api.AgentID
	Name    string
}

type agentState struct {
	runtime          *agent.Agent
	slug             string
	repoRoot         string
	worktree         string
	session          *session.LocalSession
	adapter          adapter.Adapter
	formatter        adapter.ActionFormatter
	remoteHost       api.HostID
	remoteSession    api.SessionID
	remote           bool
	config           config.AgentConfig
	explicitRepoPath bool
	presence         string
	task             string
	listenSubjects   []string
}

// Manager manages local and remote agents and sessions.
type Manager struct {
	resolver        *paths.Resolver
	dispatcher      protocol.Dispatcher
	cfg             config.Config
	git             *git.Runner
	remoteDirector  *remote.Director
	logger          *log.Logger
	managerID       api.PeerID
	mu              sync.Mutex
	agents          map[api.AgentID]*agentState
	nameIndex       map[string][]api.AgentID
	bases           map[string]string
	registries      map[string]adapter.Registry
	registryFactory func(*paths.Resolver) (adapter.Registry, error)
	subs            []protocol.Subscription
	listenSubs      map[string]*listenSubscription
	listenTargets   map[string]map[api.AgentID]struct{}
	shutdownMu      sync.Mutex
	shutdown        *shutdownController
}

// NewManager constructs a Manager.
func NewManager(ctx context.Context, resolver *paths.Resolver, cfg config.Config, dispatcher protocol.Dispatcher, version string) (*Manager, error) {
	if resolver == nil {
		return nil, fmt.Errorf("new manager: %w", ErrAgentInvalid)
	}
	if dispatcher == nil {
		return nil, fmt.Errorf("new manager: %w", ErrAgentInvalid)
	}
	logger := log.New(os.Stderr, "amux-manager ", log.LstdFlags)
	peerDir := cfg.NATS.JetStreamDir
	if peerDir == "" {
		peerDir = filepath.Join(resolver.HomeDir(), ".amux")
	}
	managerDir := filepath.Join(peerDir, "manager")
	managerID, err := remote.LoadOrCreatePeerID(managerDir)
	if err != nil {
		return nil, fmt.Errorf("new manager: %w", err)
	}
	mgr := &Manager{
		resolver:      resolver,
		dispatcher:    dispatcher,
		cfg:           cfg,
		git:           git.NewRunner(),
		logger:        logger,
		managerID:     managerID,
		agents:        make(map[api.AgentID]*agentState),
		nameIndex:     make(map[string][]api.AgentID),
		bases:         make(map[string]string),
		registries:    make(map[string]adapter.Registry),
		listenSubs:    make(map[string]*listenSubscription),
		listenTargets: make(map[string]map[api.AgentID]struct{}),
		registryFactory: func(resolver *paths.Resolver) (adapter.Registry, error) {
			return adapter.NewWazeroRegistry(context.Background(), resolver)
		},
	}
	remoteDirector, err := remote.NewDirector(cfg, dispatcher, remote.DirectorOptions{Version: version})
	if err != nil {
		return nil, fmt.Errorf("new manager: %w", err)
	}
	if err := remoteDirector.Start(ctx); err != nil {
		return nil, fmt.Errorf("new manager: %w", err)
	}
	mgr.remoteDirector = remoteDirector
	if err := mgr.loadFromConfig(ctx); err != nil {
		return nil, fmt.Errorf("new manager: %w", err)
	}
	if err := mgr.startPresenceRouting(ctx); err != nil {
		return nil, fmt.Errorf("new manager: %w", err)
	}
	if err := mgr.startMessageRouting(ctx); err != nil {
		return nil, fmt.Errorf("new manager: %w", err)
	}
	return mgr, nil
}

// AddAgent adds and starts a local agent.
func (m *Manager) AddAgent(ctx context.Context, req AddRequest) (api.RosterEntry, error) {
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Adapter) == "" {
		return api.RosterEntry{}, fmt.Errorf("add agent: %w", ErrAgentInvalid)
	}
	explicitRepoPath := strings.TrimSpace(req.Location.RepoPath) != ""
	location, repoRoot, err := m.resolveLocation(req)
	if err != nil {
		return api.RosterEntry{}, fmt.Errorf("add agent: %w", err)
	}
	if location.Type == api.LocationSSH {
		return m.addRemoteAgent(ctx, req, location, repoRoot, explicitRepoPath)
	}
	if err := ensureGitRepo(repoRoot); err != nil {
		return api.RosterEntry{}, fmt.Errorf("add agent: %w", err)
	}
	if err := m.validateMultiRepo(repoRoot, explicitRepoPath); err != nil {
		return api.RosterEntry{}, fmt.Errorf("add agent: %w", err)
	}
	m.mu.Lock()
	used := make(map[string]struct{})
	for _, state := range m.agents {
		used[state.slug] = struct{}{}
	}
	slug := paths.UniqueAgentSlug(req.Name, used)
	m.mu.Unlock()
	worktree, err := m.git.EnsureWorktree(ctx, repoRoot, slug)
	if err != nil {
		return api.RosterEntry{}, fmt.Errorf("add agent: %w", err)
	}
	agentMeta, err := api.NewAgent(req.Name, req.About, api.AdapterRef(req.Adapter), repoRoot, worktree.Path, location)
	if err != nil {
		return api.RosterEntry{}, fmt.Errorf("add agent: %w", err)
	}
	runtime, err := agent.NewAgent(agentMeta, m.dispatcher)
	if err != nil {
		return api.RosterEntry{}, fmt.Errorf("add agent: %w", err)
	}
	cfgEntry := config.AgentConfig{
		Name:           req.Name,
		About:          req.About,
		Adapter:        req.Adapter,
		ListenChannels: append([]string(nil), req.ListenChannels...),
		Location: config.AgentLocationConfig{
			Type:     location.Type.String(),
			Host:     location.Host,
			RepoPath: repoRoot,
		},
	}
	if err := m.appendAgentConfig(cfgEntry); err != nil {
		_ = m.git.RemoveWorktree(ctx, repoRoot, slug, m.cfg.Shutdown.CleanupWorktrees)
		return api.RosterEntry{}, fmt.Errorf("add agent: %w", err)
	}
	state := &agentState{
		runtime:          runtime,
		slug:             slug,
		repoRoot:         repoRoot,
		worktree:         worktree.Path,
		config:           cfgEntry,
		explicitRepoPath: explicitRepoPath,
		presence:         agent.PresenceOnline,
	}
	m.mu.Lock()
	m.agents[agentMeta.ID] = state
	m.nameIndex[req.Name] = append(m.nameIndex[req.Name], agentMeta.ID)
	m.cfg.Agents = append(m.cfg.Agents, cfgEntry)
	m.mu.Unlock()
	if _, err := m.baseBranch(ctx, repoRoot); err != nil {
		_ = m.removeAgentConfig(cfgEntry)
		_ = m.git.RemoveWorktree(ctx, repoRoot, slug, m.cfg.Shutdown.CleanupWorktrees)
		m.mu.Lock()
		delete(m.agents, agentMeta.ID)
		m.removeNameIndexLocked(req.Name, agentMeta.ID)
		m.mu.Unlock()
		return api.RosterEntry{}, fmt.Errorf("add agent: %w", err)
	}
	if _, err := m.startSession(ctx, agentMeta.ID); err != nil {
		_ = m.removeAgentConfig(cfgEntry)
		_ = m.git.RemoveWorktree(ctx, repoRoot, slug, m.cfg.Shutdown.CleanupWorktrees)
		m.mu.Lock()
		delete(m.agents, agentMeta.ID)
		m.removeNameIndexLocked(req.Name, agentMeta.ID)
		m.mu.Unlock()
		return api.RosterEntry{}, fmt.Errorf("add agent: %w", err)
	}
	record := m.rosterEntry(agentMeta.ID, state)
	m.emitAgentEvent(ctx, "agent.added", record)
	m.emitRosterUpdated(ctx)
	return record, nil
}

func (m *Manager) addRemoteAgent(ctx context.Context, req AddRequest, location api.Location, repoRoot string, explicitRepoPath bool) (api.RosterEntry, error) {
	if m.remoteDirector == nil {
		return api.RosterEntry{}, fmt.Errorf("add agent: %w", ErrAgentInvalid)
	}
	if strings.TrimSpace(location.Host) == "" {
		return api.RosterEntry{}, fmt.Errorf("add agent: %w", ErrAgentInvalid)
	}
	if strings.TrimSpace(repoRoot) == "" {
		return api.RosterEntry{}, fmt.Errorf("add agent: %w", ErrRepoPathRequired)
	}
	m.mu.Lock()
	used := make(map[string]struct{})
	for _, state := range m.agents {
		used[state.slug] = struct{}{}
	}
	slug := paths.UniqueAgentSlug(req.Name, used)
	m.mu.Unlock()
	worktree := paths.WorktreePathForRepo(repoRoot, slug)
	adapterBundle, err := buildAdapterBundle(m.resolver, req.Adapter)
	if err != nil {
		return api.RosterEntry{}, fmt.Errorf("add agent: %w", err)
	}
	hostID, _, err := m.remoteDirector.EnsureHost(ctx, location, []remote.AdapterBundle{adapterBundle})
	if err != nil {
		return api.RosterEntry{}, fmt.Errorf("add agent: %w", err)
	}
	registry, err := m.registry(m.resolver)
	if err != nil {
		return api.RosterEntry{}, fmt.Errorf("add agent: %w", err)
	}
	adapterInstance, err := registry.Load(ctx, req.Adapter)
	if err != nil {
		return api.RosterEntry{}, fmt.Errorf("add agent: %w", err)
	}
	manifest := adapterInstance.Manifest()
	if len(manifest.Commands.Start) == 0 {
		return api.RosterEntry{}, fmt.Errorf("add agent: %w", ErrAgentInvalid)
	}
	agentID := api.NewAgentID()
	cfgEntry := config.AgentConfig{
		Name:           req.Name,
		About:          req.About,
		Adapter:        req.Adapter,
		ListenChannels: append([]string(nil), req.ListenChannels...),
		Location: config.AgentLocationConfig{
			Type:     location.Type.String(),
			Host:     location.Host,
			RepoPath: repoRoot,
		},
	}
	if err := m.appendAgentConfig(cfgEntry); err != nil {
		return api.RosterEntry{}, fmt.Errorf("add agent: %w", err)
	}
	spawnReq := remote.SpawnRequest{
		Name:           req.Name,
		About:          req.About,
		AgentID:        agentID.String(),
		AgentSlug:      slug,
		RepoPath:       repoRoot,
		Adapter:        req.Adapter,
		Command:        manifest.Commands.Start,
		ListenChannels: append([]string(nil), req.ListenChannels...),
	}
	resp, err := m.spawnRemote(ctx, hostID, spawnReq)
	if err != nil {
		_ = m.removeAgentConfig(cfgEntry)
		return api.RosterEntry{}, fmt.Errorf("add agent: %w", err)
	}
	sessionID, err := api.ParseSessionID(resp.SessionID)
	if err != nil {
		_ = m.removeAgentConfig(cfgEntry)
		return api.RosterEntry{}, fmt.Errorf("add agent: %w", err)
	}
	state := &agentState{
		slug:             slug,
		repoRoot:         repoRoot,
		worktree:         worktree,
		config:           cfgEntry,
		explicitRepoPath: explicitRepoPath,
		remote:           true,
		remoteHost:       hostID,
		remoteSession:    sessionID,
		presence:         agent.PresenceOnline,
	}
	m.mu.Lock()
	m.agents[agentID] = state
	m.nameIndex[req.Name] = append(m.nameIndex[req.Name], agentID)
	m.cfg.Agents = append(m.cfg.Agents, cfgEntry)
	m.mu.Unlock()
	record := m.rosterEntry(agentID, state)
	m.emitAgentEvent(ctx, "agent.added", record)
	m.emitRosterUpdated(ctx)
	return record, nil
}

// RemoveAgent removes an agent and its worktree.
func (m *Manager) RemoveAgent(ctx context.Context, req RemoveRequest) error {
	state, id, err := m.findAgent(req)
	if err != nil {
		return fmt.Errorf("remove agent: %w", err)
	}
	if err := m.stopSession(ctx, id); err != nil {
		return fmt.Errorf("remove agent: %w", err)
	}
	if err := m.removeAgentConfig(state.config); err != nil {
		return fmt.Errorf("remove agent: %w", err)
	}
	if !state.remote {
		if err := m.git.RemoveWorktree(ctx, state.repoRoot, state.slug, m.cfg.Shutdown.CleanupWorktrees); err != nil {
			return fmt.Errorf("remove agent: %w", err)
		}
	}
	m.mu.Lock()
	delete(m.agents, id)
	m.removeNameIndexLocked(state.config.Name, id)
	m.removeConfigEntryLocked(state.config)
	m.mu.Unlock()
	entry := m.rosterEntry(id, state)
	m.emitAgentEvent(ctx, "agent.removed", entry)
	m.emitRosterUpdated(ctx)
	return nil
}

// ListAgents returns the current roster entries.
func (m *Manager) ListAgents() ([]api.RosterEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	records := make([]api.RosterEntry, 0, len(m.agents)+2)
	records = append(records, m.systemRosterLocked()...)
	for id, state := range m.agents {
		records = append(records, m.rosterEntry(id, state))
	}
	sortRoster(records)
	return records, nil
}

// StartAgent starts an existing agent session.
func (m *Manager) StartAgent(ctx context.Context, id api.AgentID) error {
	if _, err := m.startSession(ctx, id); err != nil {
		return fmt.Errorf("start agent: %w", err)
	}
	return nil
}

// StopAgent stops a running agent session.
func (m *Manager) StopAgent(ctx context.Context, id api.AgentID) error {
	if err := m.stopSession(ctx, id); err != nil {
		return fmt.Errorf("stop agent: %w", err)
	}
	return nil
}

// Shutdown drains all running sessions and optionally forces termination.
func (m *Manager) Shutdown(ctx context.Context, force bool) error {
	if m == nil {
		return nil
	}
	if ctx.Err() != nil {
		return fmt.Errorf("shutdown: %w", ctx.Err())
	}
	controller := m.ensureShutdownController()
	if controller == nil {
		return fmt.Errorf("shutdown: %w", ErrAgentInvalid)
	}
	defer m.releaseShutdownController(controller)
	state := lastStateSegment(controller.State())
	if force {
		if state == shutdownStateRunning || state == shutdownStateDraining {
			controller.signal(ctx, shutdownEventForce, map[string]any{})
		}
	} else if state == shutdownStateRunning {
		controller.signal(ctx, shutdownEventRequest, map[string]any{
			"drain_timeout": m.cfg.Shutdown.DrainTimeout.String(),
		})
	}
	waitErr := controller.wait(ctx)
	cleanupErr := m.cleanupWorktrees(ctx, m.shutdownTargets())
	if cleanupErr != nil {
		controller.recordError(cleanupErr)
	}
	errOut := errors.Join(waitErr, controller.error())
	if errOut != nil {
		return fmt.Errorf("shutdown: %w", errOut)
	}
	return nil
}

// KillAgent forces a running agent session to stop.
func (m *Manager) KillAgent(ctx context.Context, id api.AgentID) error {
	m.mu.Lock()
	state := m.agents[id]
	if state == nil {
		m.mu.Unlock()
		return fmt.Errorf("kill agent: %w", ErrAgentNotFound)
	}
	if state.remote {
		m.mu.Unlock()
		if err := m.stopSession(ctx, id); err != nil {
			return fmt.Errorf("kill agent: %w", err)
		}
		return nil
	}
	sess := state.session
	m.mu.Unlock()
	if sess == nil {
		return nil
	}
	if err := sess.Kill(ctx); err != nil {
		return fmt.Errorf("kill agent: %w", err)
	}
	return nil
}

// RestartAgent restarts a running agent session.
func (m *Manager) RestartAgent(ctx context.Context, id api.AgentID) error {
	m.mu.Lock()
	state := m.agents[id]
	sess := (*session.LocalSession)(nil)
	if state != nil && !state.remote {
		sess = state.session
	}
	m.mu.Unlock()
	if state != nil && state.remote {
		if err := m.stopSession(ctx, id); err != nil {
			return fmt.Errorf("restart agent: %w", err)
		}
		if _, err := m.startSession(ctx, id); err != nil {
			return fmt.Errorf("restart agent: %w", err)
		}
		return nil
	}
	if sess == nil {
		if _, err := m.startSession(ctx, id); err != nil {
			return fmt.Errorf("restart agent: %w", err)
		}
		return nil
	}
	if err := sess.Restart(ctx); err != nil {
		return fmt.Errorf("restart agent: %w", err)
	}
	return nil
}

// AttachAgent attaches to a running agent PTY.
func (m *Manager) AttachAgent(id api.AgentID) (net.Conn, error) {
	m.mu.Lock()
	state := m.agents[id]
	sess := (*session.LocalSession)(nil)
	if state != nil {
		sess = state.session
	}
	m.mu.Unlock()
	if state != nil && state.remote {
		if state.remoteSession.IsZero() {
			return nil, fmt.Errorf("attach agent: %w", ErrAgentNotFound)
		}
		conn, err := m.remoteDirector.AttachPTY(context.Background(), state.remoteHost, state.remoteSession)
		if err != nil {
			return nil, fmt.Errorf("attach agent: %w", err)
		}
		return conn, nil
	}
	if sess == nil {
		return nil, fmt.Errorf("attach agent: %w", ErrAgentNotFound)
	}
	conn, err := sess.Attach()
	if err != nil {
		return nil, fmt.Errorf("attach agent: %w", err)
	}
	return conn, nil
}

// MergeAgent integrates an agent branch into a target branch.
func (m *Manager) MergeAgent(ctx context.Context, id api.AgentID, strategy git.MergeStrategy, targetBranch string) (git.MergeResult, error) {
	m.mu.Lock()
	state := m.agents[id]
	m.mu.Unlock()
	if state == nil {
		return git.MergeResult{}, fmt.Errorf("merge agent: %w", ErrAgentNotFound)
	}
	if state.remote {
		return git.MergeResult{}, fmt.Errorf("merge agent: %w", ErrAgentInvalid)
	}
	if strategy == "" {
		strategy = git.MergeStrategy(m.cfg.Git.Merge.Strategy)
	}
	baseBranch, err := m.baseBranch(ctx, state.repoRoot)
	if err != nil {
		return git.MergeResult{}, fmt.Errorf("merge agent: %w", err)
	}
	m.emitAgentEvent(ctx, "git.merge.requested", map[string]any{
		"repo_root":     state.repoRoot,
		"agent_slug":    state.slug,
		"strategy":      string(strategy),
		"target_branch": targetBranch,
	})
	result, err := m.git.Merge(ctx, git.MergeOptions{
		RepoRoot:     state.repoRoot,
		WorktreePath: state.worktree,
		AgentSlug:    state.slug,
		Strategy:     strategy,
		TargetBranch: targetBranch,
		BaseBranch:   baseBranch,
		AllowDirty:   m.cfg.Git.Merge.AllowDirty,
	})
	if err != nil {
		name := "git.merge.failed"
		if errors.Is(err, git.ErrMergeConflict) {
			name = "git.merge.conflict"
		}
		m.emitAgentEvent(ctx, name, map[string]any{
			"repo_root":     state.repoRoot,
			"agent_slug":    state.slug,
			"strategy":      string(strategy),
			"target_branch": targetBranch,
			"error":         err.Error(),
		})
		return git.MergeResult{}, fmt.Errorf("merge agent: %w", err)
	}
	m.emitAgentEvent(ctx, "git.merge.completed", map[string]any{
		"repo_root":     state.repoRoot,
		"agent_slug":    state.slug,
		"strategy":      string(result.Strategy),
		"target_branch": result.TargetBranch,
	})
	return result, nil
}

func (m *Manager) startSession(ctx context.Context, id api.AgentID) (*session.LocalSession, error) {
	m.mu.Lock()
	state := m.agents[id]
	m.mu.Unlock()
	if state == nil {
		return nil, fmt.Errorf("start session: %w", ErrAgentNotFound)
	}
	if state.remote {
		if err := m.startRemoteSession(ctx, id, state); err != nil {
			return nil, fmt.Errorf("start session: %w", err)
		}
		return nil, nil
	}
	if state.session != nil {
		return state.session, nil
	}
	repoResolver, err := paths.NewResolver(state.repoRoot)
	if err != nil {
		return nil, fmt.Errorf("start session: %w", err)
	}
	registry, err := m.registry(repoResolver)
	if err != nil {
		return nil, fmt.Errorf("start session: %w", err)
	}
	adapterInstance, err := registry.Load(ctx, state.config.Adapter)
	if err != nil {
		return nil, fmt.Errorf("start session: %w", err)
	}
	manifest := adapterInstance.Manifest()
	if len(manifest.Commands.Start) == 0 {
		return nil, fmt.Errorf("start session: %w", ErrAgentInvalid)
	}
	formatter := adapterInstance.Formatter()
	sessionMeta, err := api.NewSession(id, state.repoRoot, state.worktree, state.runtime.Location)
	if err != nil {
		return nil, fmt.Errorf("start session: %w", err)
	}
	sess, err := session.NewLocalSession(sessionMeta, state.runtime, session.Command{Argv: manifest.Commands.Start}, state.worktree, adapterInstance.Matcher(), m.dispatcher, session.Config{DrainTimeout: m.cfg.Shutdown.DrainTimeout})
	if err != nil {
		return nil, fmt.Errorf("start session: %w", err)
	}
	if err := sess.Start(ctx); err != nil {
		return nil, fmt.Errorf("start session: %w", err)
	}
	m.mu.Lock()
	state.session = sess
	state.adapter = adapterInstance
	state.formatter = formatter
	if strings.TrimSpace(state.presence) == "" {
		state.presence = agent.PresenceOnline
	}
	m.mu.Unlock()
	m.configureListen(ctx, id, state)
	return sess, nil
}

func (m *Manager) stopSession(ctx context.Context, id api.AgentID) error {
	m.mu.Lock()
	state := m.agents[id]
	if state == nil {
		m.mu.Unlock()
		return fmt.Errorf("stop session: %w", ErrAgentNotFound)
	}
	if state.remote {
		hostID := state.remoteHost
		sessionID := state.remoteSession
		state.remoteSession = api.SessionID{}
		m.mu.Unlock()
		if sessionID.IsZero() {
			return nil
		}
		if _, err := m.remoteDirector.Kill(ctx, hostID, remote.KillRequest{SessionID: sessionID.String()}); err != nil {
			return fmt.Errorf("stop session: %w", err)
		}
		return nil
	}
	sess := state.session
	state.session = nil
	state.formatter = nil
	state.adapter = nil
	m.mu.Unlock()
	m.clearListen(id)
	if sess == nil {
		return nil
	}
	if err := sess.Stop(ctx); err != nil {
		return fmt.Errorf("stop session: %w", err)
	}
	return nil
}

func (m *Manager) startRemoteSession(ctx context.Context, id api.AgentID, state *agentState) error {
	if state == nil {
		return fmt.Errorf("start remote session: %w", ErrAgentNotFound)
	}
	if !state.remoteSession.IsZero() {
		return nil
	}
	registry, err := m.registry(m.resolver)
	if err != nil {
		return fmt.Errorf("start remote session: %w", err)
	}
	adapterInstance, err := registry.Load(ctx, state.config.Adapter)
	if err != nil {
		return fmt.Errorf("start remote session: %w", err)
	}
	manifest := adapterInstance.Manifest()
	if len(manifest.Commands.Start) == 0 {
		return fmt.Errorf("start remote session: %w", ErrAgentInvalid)
	}
	req := remote.SpawnRequest{
		Name:           state.config.Name,
		About:          state.config.About,
		AgentID:        id.String(),
		AgentSlug:      state.slug,
		RepoPath:       state.repoRoot,
		Adapter:        state.config.Adapter,
		Command:        manifest.Commands.Start,
		ListenChannels: append([]string(nil), state.config.ListenChannels...),
	}
	resp, err := m.spawnRemote(ctx, state.remoteHost, req)
	if err != nil {
		return fmt.Errorf("start remote session: %w", err)
	}
	sessionID, err := api.ParseSessionID(resp.SessionID)
	if err != nil {
		return fmt.Errorf("start remote session: %w", err)
	}
	m.mu.Lock()
	state.remoteSession = sessionID
	m.mu.Unlock()
	return nil
}

func (m *Manager) spawnRemote(ctx context.Context, hostID api.HostID, req remote.SpawnRequest) (remote.SpawnResponse, error) {
	timeout := m.cfg.Remote.RequestTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	deadline := time.Now().Add(timeout)
	for {
		resp, err := m.remoteDirector.Spawn(ctx, hostID, req)
		if err == nil {
			return resp, nil
		}
		if !errors.Is(err, remote.ErrNotReady) {
			return remote.SpawnResponse{}, err
		}
		if ctx.Err() != nil {
			return remote.SpawnResponse{}, fmt.Errorf("spawn remote: %w", ctx.Err())
		}
		if time.Now().After(deadline) {
			return remote.SpawnResponse{}, err
		}
		timer := time.NewTimer(200 * time.Millisecond)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return remote.SpawnResponse{}, fmt.Errorf("spawn remote: %w", ctx.Err())
		}
	}
}

func (m *Manager) baseBranch(ctx context.Context, repoRoot string) (string, error) {
	m.mu.Lock()
	base := m.bases[repoRoot]
	m.mu.Unlock()
	if base != "" {
		return base, nil
	}
	branch, err := m.git.DetectBaseBranch(ctx, repoRoot, m.cfg.Git.Merge.TargetBranch)
	if err != nil {
		if errors.Is(err, git.ErrDetachedHead) && strings.TrimSpace(m.cfg.Git.Merge.TargetBranch) == "" {
			return "", fmt.Errorf("base branch: %w (set git.merge.target_branch)", err)
		}
		return "", fmt.Errorf("base branch: %w", err)
	}
	m.mu.Lock()
	m.bases[repoRoot] = branch
	m.mu.Unlock()
	return branch, nil
}

func (m *Manager) registry(resolver *paths.Resolver) (adapter.Registry, error) {
	key := resolver.RepoRoot()
	m.mu.Lock()
	reg := m.registries[key]
	m.mu.Unlock()
	if reg != nil {
		return reg, nil
	}
	factory := m.registryFactory
	if factory == nil {
		factory = func(resolver *paths.Resolver) (adapter.Registry, error) {
			return adapter.NewWazeroRegistry(context.Background(), resolver)
		}
	}
	runtime, err := factory(resolver)
	if err != nil {
		return nil, fmt.Errorf("adapter registry: %w", err)
	}
	m.mu.Lock()
	m.registries[key] = runtime
	m.mu.Unlock()
	return runtime, nil
}

// SetRegistryFactory overrides the adapter registry factory.
func (m *Manager) SetRegistryFactory(factory func(*paths.Resolver) (adapter.Registry, error)) {
	if m == nil || factory == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.registryFactory = factory
}

func (m *Manager) resolveLocation(req AddRequest) (api.Location, string, error) {
	location := req.Location
	if location.Type == api.LocationLocal {
		repoRoot := location.RepoPath
		if strings.TrimSpace(repoRoot) == "" {
			cwd := req.Cwd
			if cwd == "" {
				repoRoot = m.resolver.RepoRoot()
			} else {
				root, err := paths.FindRepoRoot(cwd)
				if err != nil {
					return api.Location{}, "", fmt.Errorf("repo root: %w", err)
				}
				repoRoot = root
			}
		} else {
			root, err := paths.FindRepoRoot(repoRoot)
			if err != nil {
				return api.Location{}, "", fmt.Errorf("repo root: %w", err)
			}
			repoRoot = root
		}
		canonical, err := paths.CanonicalizeRepoRoot(repoRoot, m.resolver.HomeDir())
		if err != nil {
			return api.Location{}, "", fmt.Errorf("repo root: %w", err)
		}
		location.RepoPath = canonical
		return location, canonical, nil
	}
	if location.Type == api.LocationSSH {
		if strings.TrimSpace(location.Host) == "" {
			return api.Location{}, "", fmt.Errorf("location: %w", ErrAgentInvalid)
		}
		if strings.TrimSpace(location.RepoPath) == "" {
			return api.Location{}, "", fmt.Errorf("location: %w", ErrRepoPathRequired)
		}
		return location, location.RepoPath, nil
	}
	return api.Location{}, "", fmt.Errorf("location: %w", ErrAgentInvalid)
}

func buildAdapterBundle(resolver *paths.Resolver, name string) (remote.AdapterBundle, error) {
	if resolver == nil {
		return remote.AdapterBundle{}, fmt.Errorf("adapter bundle: %w", ErrAgentInvalid)
	}
	if strings.TrimSpace(name) == "" {
		return remote.AdapterBundle{}, fmt.Errorf("adapter bundle: %w", ErrAgentInvalid)
	}
	wasmPath, err := adapter.FindWasmPath(resolver, name)
	if err != nil {
		return remote.AdapterBundle{}, fmt.Errorf("adapter bundle: %w", err)
	}
	data, err := os.ReadFile(wasmPath)
	if err != nil {
		return remote.AdapterBundle{}, fmt.Errorf("adapter bundle: %w", err)
	}
	return remote.AdapterBundle{Name: name, Wasm: data}, nil
}

func (m *Manager) validateMultiRepo(repoRoot string, explicitRepoPath bool) error {
	if m == nil {
		return nil
	}
	directorRoot := m.resolver.RepoRoot()
	if strings.TrimSpace(repoRoot) == "" {
		return fmt.Errorf("repo root: %w", ErrAgentInvalid)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	distinct := false
	for _, state := range m.agents {
		if state == nil {
			continue
		}
		if state.repoRoot != repoRoot {
			distinct = true
		}
	}
	if !distinct {
		return nil
	}
	if repoRoot != directorRoot && !explicitRepoPath {
		return fmt.Errorf("repo root: %w", ErrRepoPathRequired)
	}
	for _, state := range m.agents {
		if state == nil {
			continue
		}
		if state.repoRoot != directorRoot && !state.explicitRepoPath {
			return fmt.Errorf("repo root: %w", ErrRepoPathRequired)
		}
	}
	return nil
}

func (m *Manager) findAgent(req RemoveRequest) (*agentState, api.AgentID, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !req.AgentID.IsZero() {
		state := m.agents[req.AgentID]
		if state == nil {
			return nil, api.AgentID{}, ErrAgentNotFound
		}
		return state, req.AgentID, nil
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, api.AgentID{}, ErrAgentNotFound
	}
	ids := m.nameIndex[name]
	if len(ids) == 0 {
		return nil, api.AgentID{}, ErrAgentNotFound
	}
	if len(ids) > 1 {
		return nil, api.AgentID{}, ErrAgentAmbiguous
	}
	id := ids[0]
	state := m.agents[id]
	if state == nil {
		return nil, api.AgentID{}, ErrAgentNotFound
	}
	return state, id, nil
}

func (m *Manager) loadFromConfig(ctx context.Context) error {
	used := make(map[string]struct{})
	for _, entry := range m.cfg.Agents {
		explicitRepoPath := strings.TrimSpace(entry.Location.RepoPath) != ""
		locTypeRaw := entry.Location.Type
		if strings.TrimSpace(locTypeRaw) == "" {
			locTypeRaw = "local"
		}
		locType, err := api.ParseLocationType(locTypeRaw)
		if err != nil {
			return fmt.Errorf("load agent: %w", err)
		}
		location := api.Location{Type: locType, Host: entry.Location.Host}
		repoRoot := entry.Location.RepoPath
		if repoRoot == "" {
			repoRoot = m.resolver.RepoRoot()
		}
		if locType == api.LocationLocal {
			if err := ensureGitRepo(repoRoot); err != nil {
				return fmt.Errorf("load agent: %w", err)
			}
			canonical, err := paths.CanonicalizeRepoRoot(repoRoot, m.resolver.HomeDir())
			if err != nil {
				return fmt.Errorf("load agent: %w", err)
			}
			if err := m.validateMultiRepo(canonical, explicitRepoPath); err != nil {
				return fmt.Errorf("load agent: %w", err)
			}
			location.RepoPath = canonical
			slug := paths.UniqueAgentSlug(entry.Name, used)
			used[slug] = struct{}{}
			worktree := paths.WorktreePathForRepo(canonical, slug)
			agentMeta, err := api.NewAgent(entry.Name, entry.About, api.AdapterRef(entry.Adapter), canonical, worktree, location)
			if err != nil {
				return fmt.Errorf("load agent: %w", err)
			}
			runtime, err := agent.NewAgent(agentMeta, m.dispatcher)
			if err != nil {
				return fmt.Errorf("load agent: %w", err)
			}
			state := &agentState{
				runtime:          runtime,
				slug:             slug,
				repoRoot:         canonical,
				worktree:         worktree,
				config:           entry,
				explicitRepoPath: explicitRepoPath,
				presence:         agent.PresenceOffline,
			}
			m.agents[agentMeta.ID] = state
			m.nameIndex[entry.Name] = append(m.nameIndex[entry.Name], agentMeta.ID)
			if _, err := m.baseBranch(ctx, canonical); err != nil {
				return fmt.Errorf("load agent: %w", err)
			}
			continue
		}
		if locType == api.LocationSSH {
			if strings.TrimSpace(location.Host) == "" {
				return fmt.Errorf("load agent: %w", ErrAgentInvalid)
			}
			if strings.TrimSpace(repoRoot) == "" {
				return fmt.Errorf("load agent: %w", ErrRepoPathRequired)
			}
			location.RepoPath = repoRoot
			slug := paths.UniqueAgentSlug(entry.Name, used)
			used[slug] = struct{}{}
			worktree := paths.WorktreePathForRepo(repoRoot, slug)
			agentID := api.NewAgentID()
			hostID, err := api.ParseHostID(location.Host)
			if err != nil {
				return fmt.Errorf("load agent: %w", err)
			}
			state := &agentState{
				slug:             slug,
				repoRoot:         repoRoot,
				worktree:         worktree,
				config:           entry,
				explicitRepoPath: explicitRepoPath,
				remote:           true,
				remoteHost:       hostID,
				presence:         agent.PresenceOffline,
			}
			m.agents[agentID] = state
			m.nameIndex[entry.Name] = append(m.nameIndex[entry.Name], agentID)
			continue
		}
		return fmt.Errorf("load agent: %w", ErrAgentInvalid)
	}
	return nil
}

func (m *Manager) appendAgentConfig(entry config.AgentConfig) error {
	path := m.resolver.ProjectConfigPath()
	raw, err := config.LoadConfigFile(path)
	if err != nil {
		return fmt.Errorf("append config: %w", err)
	}
	agents := extractAgents(raw)
	agents = append(agents, entry)
	if len(agents) == 0 {
		delete(raw, "agents")
	} else {
		raw["agents"] = encodeAgents(agents)
	}
	if err := config.WriteConfigFile(path, raw); err != nil {
		return fmt.Errorf("append config: %w", err)
	}
	return nil
}

func (m *Manager) removeAgentConfig(entry config.AgentConfig) error {
	path := m.resolver.ProjectConfigPath()
	raw, err := config.LoadConfigFile(path)
	if err != nil {
		return fmt.Errorf("remove config: %w", err)
	}
	agents := extractAgents(raw)
	filtered := make([]config.AgentConfig, 0, len(agents))
	removed := false
	for _, candidate := range agents {
		if !removed && sameAgent(candidate, entry) {
			removed = true
			continue
		}
		filtered = append(filtered, candidate)
	}
	if len(filtered) == 0 {
		delete(raw, "agents")
	} else {
		raw["agents"] = encodeAgents(filtered)
	}
	if err := config.WriteConfigFile(path, raw); err != nil {
		return fmt.Errorf("remove config: %w", err)
	}
	return nil
}

func extractAgents(raw map[string]any) []config.AgentConfig {
	listRaw, ok := raw["agents"].([]any)
	if !ok {
		return nil
	}
	var agents []config.AgentConfig
	for _, item := range listRaw {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		cfg := config.AgentConfig{}
		if value, ok := entry["name"].(string); ok {
			cfg.Name = value
		}
		if value, ok := entry["about"].(string); ok {
			cfg.About = value
		}
		if value, ok := entry["adapter"].(string); ok {
			cfg.Adapter = value
		}
		if rawList, ok := entry["listen_channels"].([]any); ok {
			for _, item := range rawList {
				if value, ok := item.(string); ok {
					cfg.ListenChannels = append(cfg.ListenChannels, value)
				}
			}
		}
		if locRaw, ok := entry["location"].(map[string]any); ok {
			if value, ok := locRaw["type"].(string); ok {
				cfg.Location.Type = value
			}
			if value, ok := locRaw["host"].(string); ok {
				cfg.Location.Host = value
			}
			if value, ok := locRaw["repo_path"].(string); ok {
				cfg.Location.RepoPath = value
			}
		}
		agents = append(agents, cfg)
	}
	return agents
}

func encodeAgents(agents []config.AgentConfig) []any {
	entries := make([]any, 0, len(agents))
	for _, agentCfg := range agents {
		entry := map[string]any{
			"name":    agentCfg.Name,
			"about":   agentCfg.About,
			"adapter": agentCfg.Adapter,
			"location": map[string]any{
				"type":      agentCfg.Location.Type,
				"host":      agentCfg.Location.Host,
				"repo_path": agentCfg.Location.RepoPath,
			},
		}
		if len(agentCfg.ListenChannels) > 0 {
			channels := make([]any, 0, len(agentCfg.ListenChannels))
			for _, channel := range agentCfg.ListenChannels {
				channels = append(channels, channel)
			}
			entry["listen_channels"] = channels
		}
		entries = append(entries, entry)
	}
	return entries
}

func sameAgent(a, b config.AgentConfig) bool {
	if a.Name != b.Name || a.Adapter != b.Adapter {
		return false
	}
	if !sameStringList(a.ListenChannels, b.ListenChannels) {
		return false
	}
	if a.Location.Type != b.Location.Type || a.Location.Host != b.Location.Host {
		return false
	}
	if filepath.Clean(a.Location.RepoPath) != filepath.Clean(b.Location.RepoPath) {
		return false
	}
	return true
}

func sameStringList(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func ensureGitRepo(repoRoot string) error {
	if strings.TrimSpace(repoRoot) == "" {
		return fmt.Errorf("repo root: %w", ErrAgentInvalid)
	}
	gitDir := filepath.Join(repoRoot, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return fmt.Errorf("repo root: %w", err)
	}
	if info.IsDir() {
		return nil
	}
	if info.Mode().IsRegular() {
		data, readErr := os.ReadFile(gitDir)
		if readErr != nil {
			return fmt.Errorf("repo root: %w", readErr)
		}
		if strings.Contains(string(data), "gitdir:") {
			return nil
		}
	}
	return fmt.Errorf("repo root: %w", ErrAgentInvalid)
}

func statePresence(state *agentState) string {
	if state == nil {
		return agent.PresenceOffline
	}
	if strings.TrimSpace(state.presence) != "" {
		return strings.ToLower(strings.TrimSpace(state.presence))
	}
	if state.runtime == nil || state.runtime.Presence == nil {
		return agent.PresenceOffline
	}
	current := state.runtime.Presence.State()
	if current == "" {
		return agent.PresenceOffline
	}
	return lastStateSegment(current)
}

func lastStateSegment(state string) string {
	idx := strings.LastIndex(state, "/")
	if idx == -1 {
		return state
	}
	return state[idx+1:]
}

func (m *Manager) removeNameIndexLocked(name string, id api.AgentID) {
	ids := m.nameIndex[name]
	if len(ids) == 0 {
		return
	}
	updated := ids[:0]
	for _, candidate := range ids {
		if candidate == id {
			continue
		}
		updated = append(updated, candidate)
	}
	if len(updated) == 0 {
		delete(m.nameIndex, name)
		return
	}
	m.nameIndex[name] = updated
}

func (m *Manager) removeConfigEntryLocked(entry config.AgentConfig) {
	filtered := m.cfg.Agents[:0]
	removed := false
	for _, candidate := range m.cfg.Agents {
		if !removed && sameAgent(candidate, entry) {
			removed = true
			continue
		}
		filtered = append(filtered, candidate)
	}
	m.cfg.Agents = filtered
}
