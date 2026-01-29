package remote

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/git"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/session"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go"
)

type remoteSession struct {
	agentID        api.AgentID
	sessionID      api.SessionID
	slug           string
	adapter        string
	repoPath       string
	worktree       string
	agentRuntime   *agent.Agent
	runtime        *session.LocalSession
	buffer         *ReplayBuffer
	matcher        adapter.PatternMatcher
	formatter      adapter.ActionFormatter
	adapterRef     adapter.Adapter
	replayGate     bool
	replaying      bool
	pending        [][]byte
	presence       string
	listenSubjects []string
	mu             sync.Mutex
}

// HostManager runs sessions and responds to remote control requests.
type HostManager struct {
	cfg           config.Config
	resolver      *paths.Resolver
	dispatcher    protocol.Dispatcher
	subjectPrefix string
	hostID        api.HostID
	peerID        api.PeerID
	directorPeer  api.PeerID
	version       string
	bufferSize    int
	outbox        *Outbox
	kv            *KVStore
	leaf          *protocol.NATSServer
	registry      adapter.Registry
	registryClose func(context.Context) error
	logger        *log.Logger
	lifecycle     *HostManagerLifecycle
	mu            sync.Mutex
	sessions      map[api.SessionID]*remoteSession
	agentIndex    map[api.AgentID]*remoteSession
	listenSubs    map[string]*listenSubscription
	listenTargets map[string]map[api.AgentID]struct{}
	subscribed    bool
	ready         bool
	connected     bool
	everConnected bool
}

// HostManagerStatus reports manager connection state.
type HostManagerStatus struct {
	Connected bool
	Ready     bool
	HostID    string
}

// SetRegistry overrides the adapter registry used by the host manager.
func (m *HostManager) SetRegistry(reg adapter.Registry, closer func(context.Context) error) {
	if m == nil {
		return
	}
	m.registry = reg
	m.registryClose = closer
}

// NewHostManager constructs a host manager.
func NewHostManager(cfg config.Config, resolver *paths.Resolver, version string) (*HostManager, error) {
	hostID := strings.TrimSpace(cfg.Remote.Manager.HostID)
	if hostID == "" {
		name, err := os.Hostname()
		if err != nil || name == "" {
			hostID = "manager"
		} else {
			hostID = strings.ToLower(name)
		}
	}
	parsedHostID, err := api.ParseHostID(hostID)
	if err != nil {
		return nil, fmt.Errorf("host manager: %w", err)
	}
	bufferSize := int(cfg.Remote.BufferSize)
	if bufferSize < 0 {
		bufferSize = 0
	}
	peerDir := cfg.NATS.JetStreamDir
	if peerDir == "" && resolver != nil {
		peerDir = filepath.Join(resolver.HomeDir(), ".amux")
	}
	peerID, err := LoadOrCreatePeerID(peerDir)
	if err != nil {
		return nil, fmt.Errorf("host manager: %w", err)
	}
	logger := log.New(os.Stderr, "amux-hostmgr ", log.LstdFlags)
	manager := &HostManager{
		cfg:           cfg,
		resolver:      resolver,
		subjectPrefix: SubjectPrefix(cfg.Remote.NATS.SubjectPrefix),
		hostID:        parsedHostID,
		peerID:        peerID,
		version:       version,
		bufferSize:    bufferSize,
		outbox:        NewOutbox(bufferSize),
		logger:        logger,
		sessions:      make(map[api.SessionID]*remoteSession),
		agentIndex:    make(map[api.AgentID]*remoteSession),
		listenSubs:    make(map[string]*listenSubscription),
		listenTargets: make(map[string]map[api.AgentID]struct{}),
	}
	manager.lifecycle = newHostManagerLifecycle(manager)
	return manager, nil
}

// Status returns the current connection state for the host manager.
func (m *HostManager) Status() HostManagerStatus {
	if m == nil {
		return HostManagerStatus{}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return HostManagerStatus{
		Connected: m.connected,
		Ready:     m.ready,
		HostID:    m.hostID.String(),
	}
}

// Start connects to NATS and begins serving control requests.
func (m *HostManager) Start(ctx context.Context) error {
	if m.lifecycle != nil {
		m.lifecycle.Start(ctx)
		hsm.Dispatch(ctx, m.lifecycle, hsm.Event{Name: hostManagerEventStart})
	}
	if err := m.startInternal(ctx); err != nil {
		if m.lifecycle != nil {
			hsm.Dispatch(ctx, m.lifecycle, hsm.Event{Name: hostManagerEventError, Data: err})
		}
		return err
	}
	if m.lifecycle != nil {
		hsm.Dispatch(ctx, m.lifecycle, hsm.Event{Name: hostManagerEventStop})
	}
	return nil
}

func (m *HostManager) startInternal(ctx context.Context) error {
	if m.registry == nil {
		registry, err := adapter.NewWazeroRegistry(ctx, m.resolver)
		if err != nil {
			return fmt.Errorf("host manager: %w", err)
		}
		m.registry = registry
		m.registryClose = registry.Close
		defer func() {
			if m.registryClose != nil {
				_ = m.registryClose(context.Background())
			}
		}()
	}
	if err := m.ensureLeafServer(ctx); err != nil {
		return fmt.Errorf("host manager: %w", err)
	}
	if err := m.connect(ctx); err != nil {
		return err
	}
	if m.lifecycle != nil {
		hsm.Dispatch(ctx, m.lifecycle, hsm.Event{Name: hostManagerEventReady})
	}
	go m.monitorLeaf(ctx)
	go m.heartbeatLoop(ctx)
	select {
	case <-ctx.Done():
		return nil
	case <-m.dispatcher.Closed():
		m.markDisconnected("io_error")
		return fmt.Errorf("host manager: dispatcher closed")
	}
}

func (m *HostManager) connect(ctx context.Context) error {
	credsPath := strings.TrimSpace(m.cfg.Remote.NATS.CredsPath)
	if credsPath == "" {
		return fmt.Errorf("host manager: %w", ErrInvalidMessage)
	}
	if _, err := os.Stat(credsPath); err != nil {
		return fmt.Errorf("host manager: %w", err)
	}
	if m.leaf == nil {
		return fmt.Errorf("host manager: leaf server unavailable")
	}
	leafDispatcher, err := protocol.NewNATSDispatcher(ctx, m.leaf.URL(), protocol.NATSOptions{
		Name:             "amux-manager-leaf",
		AllowNoJetStream: true,
	})
	if err != nil {
		return fmt.Errorf("host manager: %w", err)
	}
	hubClientURL, err := hubClientURL(m.cfg)
	if err != nil {
		_ = leafDispatcher.Close(ctx)
		return fmt.Errorf("host manager: %w", err)
	}
	hubDispatcher, err := protocol.NewNATSDispatcher(ctx, hubClientURL, protocol.NATSOptions{
		Name:      "amux-manager-hub",
		CredsPath: credsPath,
	})
	if err != nil {
		_ = leafDispatcher.Close(ctx)
		return fmt.Errorf("host manager: %w", err)
	}
	kv, err := NewKVStore(hubDispatcher.JetStream(), m.cfg.Remote.NATS.KVBucket)
	if err != nil {
		_ = hubDispatcher.Close(ctx)
		_ = leafDispatcher.Close(ctx)
		return fmt.Errorf("host manager: %w", err)
	}
	m.mu.Lock()
	m.dispatcher = leafDispatcher
	m.kv = kv
	m.mu.Unlock()
	return nil
}

func (m *HostManager) performHandshake(ctx context.Context, recovered bool) error {
	payload := HandshakePayload{
		Protocol: 1,
		PeerID:   m.peerID.String(),
		Role:     "daemon",
		HostID:   m.hostID.String(),
	}
	msg, err := EncodePayload("handshake", payload)
	if err != nil {
		return fmt.Errorf("host manager: %w", err)
	}
	data, err := EncodeControlMessage(msg)
	if err != nil {
		return fmt.Errorf("host manager: %w", err)
	}
	subject := HandshakeSubject(m.subjectPrefix, m.hostID)
	timeout := m.cfg.Remote.RequestTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	if timeout > time.Second {
		timeout = time.Second
	}
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	reply, err := m.dispatcher.Request(reqCtx, subject, data, timeout)
	if err != nil {
		return fmt.Errorf("host manager: %w", err)
	}
	resp, err := DecodeControlMessage(reply.Data)
	if err != nil {
		return fmt.Errorf("host manager: %w", err)
	}
	if resp.Type == "error" {
		var payload ErrorPayload
		if err := DecodePayload(resp, &payload); err != nil {
			return fmt.Errorf("host manager: %w", err)
		}
		return fmt.Errorf("host manager: %w", ErrNotReady)
	}
	var respPayload HandshakePayload
	if err := DecodePayload(resp, &respPayload); err != nil {
		return fmt.Errorf("host manager: %w", err)
	}
	if respPayload.Role == "director" {
		if peerID, err := api.ParsePeerID(respPayload.PeerID); err == nil {
			m.mu.Lock()
			m.directorPeer = peerID
			m.mu.Unlock()
		}
	}
	m.mu.Lock()
	m.ready = true
	m.connected = true
	m.everConnected = true
	needSubscribe := !m.subscribed
	m.subscribed = true
	m.mu.Unlock()
	if needSubscribe {
		if err := m.subscribeControl(ctx); err != nil {
			return fmt.Errorf("host manager: %w", err)
		}
	}
	if err := m.writeHostKV(context.Background()); err != nil {
		return fmt.Errorf("host manager: %w", err)
	}
	if recovered {
		m.publishConnectionEvent(ctx, "connection.recovered", ConnectionRecoveredPayload{
			PeerID:    m.peerID.String(),
			Timestamp: NowRFC3339(),
		})
	} else {
		m.publishConnectionEvent(ctx, "connection.established", ConnectionEstablishedPayload{
			PeerID:    m.peerID.String(),
			Timestamp: NowRFC3339(),
		})
	}
	return nil
}

func (m *HostManager) ensureLeafServer(ctx context.Context) error {
	m.mu.Lock()
	if m.leaf != nil {
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock()
	listen := strings.TrimSpace(m.cfg.NATS.Listen)
	if listen == "" {
		listen = "127.0.0.1:-1"
	}
	leaf, err := protocol.StartLeafServer(ctx, protocol.LeafServerConfig{
		Listen:    listen,
		HubURL:    hubURL(m.cfg),
		CredsPath: m.cfg.Remote.NATS.CredsPath,
	})
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.leaf = leaf
	m.mu.Unlock()
	return nil
}

func (m *HostManager) monitorLeaf(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	attempts := 0
	nextAttempt := time.Time{}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		if time.Now().After(nextAttempt) && !m.isReady() {
			recovered := m.wasConnected()
			if err := m.performHandshake(ctx, recovered); err != nil {
				if recovered {
					m.markDisconnected("handshake_failed")
				}
				attempts++
				nextAttempt = time.Now().Add(reconnectDelay(m.cfg, attempts))
			} else {
				attempts = 0
				nextAttempt = time.Time{}
				m.flushOutbox()
			}
		}
	}
}

func (m *HostManager) heartbeatLoop(ctx context.Context) {
	interval := m.cfg.Remote.NATS.HeartbeatInterval
	if interval <= 0 {
		interval = 5 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		if !m.isReady() {
			continue
		}
		if err := m.writeHeartbeat(ctx); err != nil {
			m.markDisconnected("heartbeat_failed")
			continue
		}
	}
}

func (m *HostManager) writeHeartbeat(ctx context.Context) error {
	m.mu.Lock()
	kv := m.kv
	hostID := m.hostID
	m.mu.Unlock()
	if kv == nil {
		return fmt.Errorf("heartbeat: kv unavailable")
	}
	payload := map[string]any{"timestamp": NowRFC3339()}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("heartbeat: %w", err)
	}
	if err := kv.Put(ctx, fmt.Sprintf("hosts/%s/heartbeat", hostID.String()), data); err != nil {
		return fmt.Errorf("heartbeat: %w", err)
	}
	return nil
}

func (m *HostManager) writeHostKV(ctx context.Context) error {
	m.mu.Lock()
	kv := m.kv
	hostID := m.hostID
	peerID := m.peerID
	version := m.version
	m.mu.Unlock()
	if kv == nil {
		return fmt.Errorf("host kv: kv unavailable")
	}
	payload := map[string]any{
		"version":    version,
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"peer_id":    peerID.String(),
		"started_at": NowRFC3339(),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("host kv: %w", err)
	}
	if err := kv.Put(ctx, fmt.Sprintf("hosts/%s/info", hostID.String()), data); err != nil {
		return fmt.Errorf("host kv: %w", err)
	}
	return nil
}

func (m *HostManager) writeSessionKV(ctx context.Context, session *remoteSession, state string, sessionErr error) error {
	if session == nil {
		return fmt.Errorf("session kv: missing session")
	}
	m.mu.Lock()
	kv := m.kv
	hostID := m.hostID
	m.mu.Unlock()
	if kv == nil {
		return fmt.Errorf("session kv: kv unavailable")
	}
	payload := map[string]any{
		"agent_id":   session.agentID.String(),
		"agent_slug": session.slug,
		"repo_path":  session.repoPath,
		"state":      state,
	}
	if sessionErr != nil {
		payload["error"] = sessionErr.Error()
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("session kv: %w", err)
	}
	key := fmt.Sprintf("sessions/%s/%s", hostID.String(), session.sessionID.String())
	if err := kv.Put(ctx, key, data); err != nil {
		return fmt.Errorf("session kv: %w", err)
	}
	return nil
}

func (m *HostManager) wasConnected() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.everConnected
}

func (m *HostManager) subscribeControl(ctx context.Context) error {
	ctlSubject := ControlSubject(m.subjectPrefix, m.hostID)
	_, err := m.dispatcher.SubscribeRaw(ctx, ctlSubject, m.handleControl)
	if err != nil {
		return fmt.Errorf("host manager: %w", err)
	}
	ptySubject := protocol.Subject(m.subjectPrefix, "pty", m.hostID.String(), "*", "in")
	_, err = m.dispatcher.SubscribeRaw(ctx, ptySubject, m.handlePTYInput)
	if err != nil {
		return fmt.Errorf("host manager: %w", err)
	}
	if err := m.subscribeComm(ctx); err != nil {
		return err
	}
	if err := m.subscribePresence(ctx); err != nil {
		return err
	}
	return nil
}

func (m *HostManager) subscribeComm(ctx context.Context) error {
	managerSubject := ManagerCommSubject(m.subjectPrefix, m.hostID)
	_, err := m.dispatcher.SubscribeRaw(ctx, managerSubject, m.handleCommMessage)
	if err != nil {
		return fmt.Errorf("host manager: %w", err)
	}
	agentSubject := protocol.Subject(m.subjectPrefix, "comm", "agent", m.hostID.String(), ">")
	_, err = m.dispatcher.SubscribeRaw(ctx, agentSubject, m.handleCommMessage)
	if err != nil {
		return fmt.Errorf("host manager: %w", err)
	}
	broadcastSubject := BroadcastCommSubject(m.subjectPrefix)
	_, err = m.dispatcher.SubscribeRaw(ctx, broadcastSubject, m.handleCommMessage)
	if err != nil {
		return fmt.Errorf("host manager: %w", err)
	}
	return nil
}

func (m *HostManager) handleControl(msg protocol.Message) {
	if msg.Reply == "" {
		return
	}
	control, err := DecodeControlMessage(msg.Data)
	requestType := "unknown"
	if err == nil && control.Type != "" {
		requestType = control.Type
	}
	if !m.isReady() {
		_ = m.replyError(msg.Reply, requestType, "not_ready", "handshake incomplete")
		return
	}
	if err != nil {
		_ = m.replyError(msg.Reply, "unknown", "invalid_request", "invalid control message")
		return
	}
	switch control.Type {
	case "spawn":
		m.handleSpawn(msg.Reply, control)
	case "kill":
		m.handleKill(msg.Reply, control)
	case "replay":
		m.handleReplay(msg.Reply, control)
	case "ping":
		m.handlePing(msg.Reply, control)
	default:
		_ = m.replyError(msg.Reply, "unknown", "unknown", "unknown request")
	}
}

func (m *HostManager) handleSpawn(reply string, control ControlMessage) {
	var req SpawnRequest
	if err := DecodePayload(control, &req); err != nil {
		_ = m.replyError(reply, "spawn", "invalid_request", "invalid spawn")
		return
	}
	if len(req.Command) == 0 {
		_ = m.replyError(reply, "spawn", "invalid_request", "missing command")
		return
	}
	agentID, err := api.ParseAgentID(req.AgentID)
	if err != nil {
		_ = m.replyError(reply, "spawn", "invalid_request", "invalid agent id")
		return
	}
	if strings.TrimSpace(req.AgentSlug) == "" {
		_ = m.replyError(reply, "spawn", "invalid_request", "missing agent slug")
		return
	}
	if strings.TrimSpace(req.Adapter) == "" {
		_ = m.replyError(reply, "spawn", "invalid_request", "missing adapter")
		return
	}
	repoRoot := m.expandPath(req.RepoPath)
	if repoRoot == "" {
		_ = m.replyError(reply, "spawn", "invalid_repo", "repo path required")
		return
	}
	if err := ensureRepo(repoRoot); err != nil {
		_ = m.replyError(reply, "spawn", "invalid_repo", err.Error())
		return
	}
	m.mu.Lock()
	if existing, ok := m.agentIndex[agentID]; ok {
		if existing.slug != req.AgentSlug || existing.repoPath != repoRoot || existing.adapter != req.Adapter {
			m.mu.Unlock()
			_ = m.replyError(reply, "spawn", "session_conflict", "session conflict")
			return
		}
		sessionID := existing.sessionID
		m.mu.Unlock()
		m.replySpawn(reply, agentID, sessionID)
		return
	}
	m.mu.Unlock()
	worktree := paths.WorktreePathForRepo(repoRoot, req.AgentSlug)
	runner := git.NewRunner()
	if _, err := runner.EnsureWorktree(context.Background(), repoRoot, req.AgentSlug); err != nil {
		_ = m.replyError(reply, "spawn", "invalid_repo", "failed to create worktree")
		return
	}
	location := api.Location{Type: api.LocationSSH, Host: m.hostID.String(), RepoPath: repoRoot}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = req.AgentSlug
	}
	agentMeta, err := api.NewAgentWithID(agentID, name, req.About, api.AdapterRef(req.Adapter), repoRoot, worktree, location)
	if err != nil {
		_ = m.replyError(reply, "spawn", "invalid_request", "invalid agent metadata")
		return
	}
	runtime, err := agent.NewAgent(agentMeta, m.dispatcher)
	if err != nil {
		_ = m.replyError(reply, "spawn", "invalid_request", "failed to create agent runtime")
		return
	}
	sessionMeta, err := api.NewSession(agentID, repoRoot, worktree, location)
	if err != nil {
		_ = m.replyError(reply, "spawn", "invalid_request", "invalid session")
		return
	}
	cmd := session.Command{Argv: req.Command}
	if len(req.Env) > 0 {
		env := make([]string, 0, len(req.Env))
		for key, value := range req.Env {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = env
	}
	registry := m.registry
	if registry == nil {
		_ = m.replyError(reply, "spawn", "internal", "adapter registry unavailable")
		return
	}
	adapterInstance, err := registry.Load(context.Background(), req.Adapter)
	if err != nil {
		_ = m.replyError(reply, "spawn", "invalid_request", "failed to load adapter")
		return
	}
	matcher := adapterInstance.Matcher()
	formatter := adapterInstance.Formatter()
	sess, err := session.NewLocalSession(sessionMeta, runtime, cmd, worktree, matcher, m.dispatcher, session.Config{DrainTimeout: m.cfg.Shutdown.DrainTimeout})
	if err != nil {
		_ = m.replyError(reply, "spawn", "invalid_request", "failed to start session")
		return
	}
	if err := sess.Start(context.Background()); err != nil {
		_ = m.replyError(reply, "spawn", "invalid_request", "failed to start session")
		return
	}
	remoteSess := &remoteSession{
		agentID:      agentID,
		sessionID:    sessionMeta.ID,
		slug:         req.AgentSlug,
		adapter:      req.Adapter,
		repoPath:     repoRoot,
		worktree:     worktree,
		agentRuntime: runtime,
		runtime:      sess,
		buffer:       NewReplayBuffer(m.bufferSize),
		matcher:      matcher,
		formatter:    formatter,
		adapterRef:   adapterInstance,
		presence:     agent.PresenceOnline,
	}
	if err := m.writeSessionKV(context.Background(), remoteSess, "running", nil); err != nil {
		_ = sess.Kill(context.Background())
		_ = m.replyError(reply, "spawn", "internal", "failed to persist session")
		return
	}
	m.mu.Lock()
	m.sessions[sessionMeta.ID] = remoteSess
	m.agentIndex[agentID] = remoteSess
	m.mu.Unlock()
	m.configureListen(context.Background(), remoteSess, req.ListenChannels)
	sess.AddOutputObserver(func(chunk []byte) {
		m.handleOutput(remoteSess, chunk)
	})
	go m.observeSession(remoteSess)
	m.replySpawn(reply, agentID, sessionMeta.ID)
}

func (m *HostManager) replySpawn(reply string, agentID api.AgentID, sessionID api.SessionID) {
	payload := SpawnResponse{AgentID: agentID.String(), SessionID: sessionID.String()}
	msg, err := EncodePayload("spawn", payload)
	if err != nil {
		_ = m.replyError(reply, "spawn", "internal", "failed to encode response")
		return
	}
	data, err := EncodeControlMessage(msg)
	if err != nil {
		_ = m.replyError(reply, "spawn", "internal", "failed to encode response")
		return
	}
	_ = m.dispatcher.PublishRaw(context.Background(), reply, data, "")
}

func (m *HostManager) handleKill(reply string, control ControlMessage) {
	var req KillRequest
	if err := DecodePayload(control, &req); err != nil {
		_ = m.replyError(reply, "kill", "invalid_request", "invalid kill")
		return
	}
	sessionID, err := api.ParseSessionID(req.SessionID)
	if err != nil {
		_ = m.replyError(reply, "kill", "invalid_request", "invalid session id")
		return
	}
	m.mu.Lock()
	session := m.sessions[sessionID]
	m.mu.Unlock()
	killed := false
	if session != nil {
		if err := session.runtime.Kill(context.Background()); err == nil {
			killed = true
			m.clearListen(session)
			m.mu.Lock()
			delete(m.sessions, sessionID)
			delete(m.agentIndex, session.agentID)
			m.mu.Unlock()
			_ = m.writeSessionKV(context.Background(), session, "terminated", nil)
		}
	}
	payload := KillResponse{SessionID: sessionID.String(), Killed: killed}
	msg, err := EncodePayload("kill", payload)
	if err != nil {
		_ = m.replyError(reply, "kill", "internal", "failed to encode response")
		return
	}
	data, err := EncodeControlMessage(msg)
	if err != nil {
		_ = m.replyError(reply, "kill", "internal", "failed to encode response")
		return
	}
	_ = m.dispatcher.PublishRaw(context.Background(), reply, data, "")
}

func (m *HostManager) handleReplay(reply string, control ControlMessage) {
	var req ReplayRequest
	if err := DecodePayload(control, &req); err != nil {
		_ = m.replyError(reply, "replay", "invalid_request", "invalid replay")
		return
	}
	sessionID, err := api.ParseSessionID(req.SessionID)
	if err != nil {
		_ = m.replyError(reply, "replay", "invalid_request", "invalid session id")
		return
	}
	m.mu.Lock()
	session := m.sessions[sessionID]
	m.mu.Unlock()
	accepted := false
	if session != nil && session.buffer.Enabled() {
		accepted = true
		m.replaySession(session)
	} else if session != nil {
		session.mu.Lock()
		session.replayGate = false
		session.mu.Unlock()
	}
	payload := ReplayResponse{SessionID: sessionID.String(), Accepted: accepted}
	msg, err := EncodePayload("replay", payload)
	if err != nil {
		_ = m.replyError(reply, "replay", "internal", "failed to encode response")
		return
	}
	data, err := EncodeControlMessage(msg)
	if err != nil {
		_ = m.replyError(reply, "replay", "internal", "failed to encode response")
		return
	}
	_ = m.dispatcher.PublishRaw(context.Background(), reply, data, "")
}

func (m *HostManager) handlePing(reply string, control ControlMessage) {
	var req PingPayload
	if err := DecodePayload(control, &req); err != nil {
		_ = m.replyError(reply, "ping", "invalid_request", "invalid ping")
		return
	}
	payload := PingPayload{UnixMS: req.UnixMS}
	msg, err := EncodePayload("pong", payload)
	if err != nil {
		_ = m.replyError(reply, "ping", "internal", "failed to encode response")
		return
	}
	data, err := EncodeControlMessage(msg)
	if err != nil {
		_ = m.replyError(reply, "ping", "internal", "failed to encode response")
		return
	}
	_ = m.dispatcher.PublishRaw(context.Background(), reply, data, "")
}

func (m *HostManager) handlePTYInput(msg protocol.Message) {
	hostID, sessionID, direction, err := ParseSessionSubject(m.subjectPrefix, msg.Subject)
	if err != nil {
		return
	}
	if hostID != m.hostID || direction != "in" {
		return
	}
	m.mu.Lock()
	session := m.sessions[sessionID]
	m.mu.Unlock()
	if session == nil {
		return
	}
	_ = session.runtime.Send(msg.Data)
}

func (m *HostManager) handleCommMessage(msg protocol.Message) {
	if len(msg.Data) == 0 {
		return
	}
	var payload api.AgentMessage
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		return
	}
	m.mirrorListenedMessage(msg.Subject, payload)
	target := payload.To
	if target.IsBroadcast() {
		m.mu.Lock()
		sessions := make([]*remoteSession, 0, len(m.sessions))
		for _, sess := range m.sessions {
			sessions = append(sessions, sess)
		}
		m.mu.Unlock()
		for _, sess := range sessions {
			m.deliverMessage(sess, payload)
		}
		return
	}
	var targetSession *remoteSession
	m.mu.Lock()
	for _, sess := range m.sessions {
		if sess != nil && sess.agentID.Value() == target.Value() {
			targetSession = sess
			break
		}
	}
	m.mu.Unlock()
	if targetSession == nil {
		return
	}
	m.deliverMessage(targetSession, payload)
}

func (m *HostManager) deliverMessage(session *remoteSession, payload api.AgentMessage) {
	if session == nil || session.runtime == nil || session.formatter == nil {
		return
	}
	formatted, err := session.formatter.Format(context.Background(), payload.Content)
	if err != nil {
		return
	}
	if formatted == "" {
		return
	}
	_ = session.runtime.Send([]byte(formatted))
}

func (m *HostManager) handleOutput(session *remoteSession, chunk []byte) {
	if session == nil || len(chunk) == 0 {
		return
	}
	session.buffer.Add(chunk)
	m.handleOutboundMessages(session, chunk)
	m.mu.Lock()
	connected := m.connected
	m.mu.Unlock()
	session.mu.Lock()
	if session.replayGate || !connected {
		session.mu.Unlock()
		return
	}
	if session.replaying {
		session.pending = append(session.pending, append([]byte(nil), chunk...))
		session.mu.Unlock()
		return
	}
	session.mu.Unlock()
	m.publishPTY(session.sessionID, chunk)
}

func (m *HostManager) handleOutboundMessages(session *remoteSession, chunk []byte) {
	if session == nil || session.matcher == nil {
		return
	}
	matches, err := session.matcher.Match(context.Background(), chunk)
	if err != nil {
		return
	}
	for _, match := range matches {
		if strings.ToLower(strings.TrimSpace(match.Pattern)) != "message" {
			continue
		}
		var payload api.OutboundMessage
		if err := json.Unmarshal([]byte(match.Text), &payload); err != nil {
			continue
		}
		if payload.ToSlug == "" || payload.Content == "" {
			continue
		}
		msg, err := m.buildAgentMessage(session, payload)
		if err != nil {
			if errors.Is(err, ErrMessageTargetUnknown) {
				m.notifyUnknownRecipient(session, payload.ToSlug)
			} else if m.logger != nil {
				m.logger.Printf("message build failed: agent=%s error=%v", session.slug, err)
			}
			continue
		}
		m.publishAgentMessage(session, msg)
	}
}

func (m *HostManager) buildAgentMessage(session *remoteSession, payload api.OutboundMessage) (api.AgentMessage, error) {
	if session == nil {
		return api.AgentMessage{}, fmt.Errorf("message build: %w", ErrInvalidMessage)
	}
	msg := api.AgentMessage{
		ToSlug:  payload.ToSlug,
		Content: payload.Content,
	}
	msg.ID = api.NewRuntimeID()
	msg.From = session.agentID.RuntimeID
	msg.Timestamp = time.Now().UTC()
	target, ok := m.resolveToID(payload.ToSlug)
	if !ok {
		return api.AgentMessage{}, fmt.Errorf("message build: %w", ErrMessageTargetUnknown)
	}
	msg.To = target
	return msg, nil
}

func (m *HostManager) notifyUnknownRecipient(session *remoteSession, toSlug string) {
	if m.logger != nil {
		m.logger.Printf("message target unresolved: to_slug=%q sender=%s", toSlug, session.slug)
	}
	if session == nil || session.runtime == nil {
		return
	}
	text := fmt.Sprintf("amux: unknown recipient %q\n", toSlug)
	formatted := text
	if session.formatter != nil {
		if out, err := session.formatter.Format(context.Background(), text); err == nil && out != "" {
			formatted = out
		}
	}
	_ = session.runtime.Send([]byte(formatted))
}

func (m *HostManager) resolveToID(slug string) (api.TargetID, bool) {
	target := strings.ToLower(strings.TrimSpace(slug))
	switch target {
	case "all", "broadcast", "*":
		return api.TargetID{}, true
	case "director":
		m.mu.Lock()
		peer := m.directorPeer
		m.mu.Unlock()
		if peer.IsZero() {
			return api.TargetID{}, false
		}
		return api.TargetIDFromRuntime(peer.RuntimeID), true
	case "manager":
		return api.TargetIDFromRuntime(m.peerID.RuntimeID), true
	}
	if strings.HasPrefix(target, "manager@") {
		if strings.TrimPrefix(target, "manager@") == strings.ToLower(m.hostID.String()) {
			return api.TargetIDFromRuntime(m.peerID.RuntimeID), true
		}
		return api.TargetID{}, false
	}
	m.mu.Lock()
	for _, sess := range m.sessions {
		if sess != nil && strings.EqualFold(sess.slug, slug) {
			m.mu.Unlock()
			return api.TargetIDFromRuntime(sess.agentID.RuntimeID), true
		}
	}
	m.mu.Unlock()
	return api.TargetID{}, false
}

func (m *HostManager) publishAgentMessage(session *remoteSession, msg api.AgentMessage) {
	if session == nil {
		return
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	sender := AgentCommSubject(m.subjectPrefix, m.hostID, session.agentID)
	if msg.To.IsBroadcast() {
		m.publishComm(sender, data)
		m.publishComm(BroadcastCommSubject(m.subjectPrefix), data)
		return
	}
	if msg.To.Value() == session.agentID.Value() {
		m.publishComm(sender, data)
		return
	}
	m.publishComm(sender, data)
	recipient := m.commSubjectForTarget(msg.To)
	if recipient != "" {
		m.publishComm(recipient, data)
	}
}

func (m *HostManager) commSubjectForTarget(target api.TargetID) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if target.Value() == m.peerID.Value() {
		return ManagerCommSubject(m.subjectPrefix, m.hostID)
	}
	if m.directorPeer.Value() != 0 && target.Value() == m.directorPeer.Value() {
		return DirectorCommSubject(m.subjectPrefix)
	}
	for _, sess := range m.sessions {
		if sess != nil && sess.agentID.Value() == target.Value() {
			return AgentCommSubject(m.subjectPrefix, m.hostID, sess.agentID)
		}
	}
	return ""
}

func (m *HostManager) publishComm(subject string, payload []byte) {
	if subject == "" {
		return
	}
	m.mu.Lock()
	connected := m.connected
	m.mu.Unlock()
	if !connected {
		m.outbox.Enqueue(subject, payload)
		return
	}
	_ = m.dispatcher.PublishRaw(context.Background(), subject, payload, "")
}

func (m *HostManager) replaySession(session *remoteSession) {
	snapshot := session.buffer.Snapshot()
	session.mu.Lock()
	session.replaying = true
	session.replayGate = false
	session.mu.Unlock()
	if len(snapshot) > 0 {
		m.publishPTY(session.sessionID, snapshot)
	}
	session.mu.Lock()
	pending := session.pending
	session.pending = nil
	session.replaying = false
	session.mu.Unlock()
	for _, chunk := range pending {
		m.publishPTY(session.sessionID, chunk)
	}
}

func (m *HostManager) publishPTY(sessionID api.SessionID, data []byte) {
	subject := PtyOutSubject(m.subjectPrefix, m.hostID, sessionID)
	chunks := chunkBytes(m.dispatcher.MaxPayload(), data)
	for _, chunk := range chunks {
		_ = m.dispatcher.PublishRaw(context.Background(), subject, chunk, "")
	}
}

func (m *HostManager) publishConnectionEvent(ctx context.Context, name string, payload any) {
	event, err := EncodeEventMessage(name, payload)
	if err != nil {
		return
	}
	data, err := EncodeEventMessageJSON(event)
	if err != nil {
		return
	}
	m.publishEvent(ctx, data)
}

func (m *HostManager) publishHostEvent(ctx context.Context, name string, payload any) {
	m.publishConnectionEvent(ctx, name, payload)
}

func (m *HostManager) publishEvent(ctx context.Context, payload []byte) {
	subject := EventsSubject(m.subjectPrefix, m.hostID)
	m.mu.Lock()
	connected := m.connected
	m.mu.Unlock()
	if !connected {
		m.outbox.Enqueue(subject, payload)
		return
	}
	_ = m.dispatcher.PublishRaw(ctx, subject, payload, "")
}

func (m *HostManager) flushOutbox() {
	entries := m.outbox.Drain()
	for _, entry := range entries {
		_ = m.dispatcher.PublishRaw(context.Background(), entry.subject, entry.payload, "")
	}
}

func (m *HostManager) markDisconnected(reason string) {
	m.mu.Lock()
	if !m.connected && !m.ready {
		m.mu.Unlock()
		return
	}
	m.connected = false
	m.ready = false
	for _, sess := range m.sessions {
		sess.mu.Lock()
		sess.replayGate = true
		sess.mu.Unlock()
	}
	m.mu.Unlock()
	m.publishConnectionEvent(context.Background(), "connection.lost", ConnectionLostPayload{
		PeerID:    m.peerID.String(),
		Timestamp: NowRFC3339(),
		Reason:    reason,
	})
}

func (m *HostManager) observeSession(session *remoteSession) {
	if session == nil || session.runtime == nil {
		return
	}
	done := session.runtime.Done()
	if done == nil {
		return
	}
	err := <-done
	payload := map[string]any{
		"agent_id":   session.agentID.String(),
		"session_id": session.sessionID.String(),
	}
	if err != nil {
		payload["error"] = err.Error()
	}
	m.publishHostEvent(context.Background(), "session.exited", payload)
	_ = m.writeSessionKV(context.Background(), session, "exited", err)
}

func (m *HostManager) isReady() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ready
}

func (m *HostManager) replyError(reply, requestType, code, message string) error {
	msg, err := NewErrorMessage(requestType, code, message)
	if err != nil {
		return err
	}
	data, err := EncodeControlMessage(msg)
	if err != nil {
		return err
	}
	return m.dispatcher.PublishRaw(context.Background(), reply, data, "")
}

func (m *HostManager) expandPath(path string) string {
	if m.resolver == nil {
		return path
	}
	return m.resolver.ExpandHome(path)
}

func ensureRepo(repoRoot string) error {
	gitDir := filepath.Join(repoRoot, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return fmt.Errorf("repo missing: %w", err)
	}
	if info.IsDir() {
		return nil
	}
	if info.Mode().IsRegular() {
		return nil
	}
	return fmt.Errorf("repo missing: %w", errors.New("invalid git dir"))
}

func reconnectDelay(cfg config.Config, attempt int) time.Duration {
	base := cfg.Remote.ReconnectBackoffBase
	if base <= 0 {
		base = time.Second
	}
	max := cfg.Remote.ReconnectBackoffMax
	if max <= 0 {
		max = 30 * time.Second
	}
	delay := base
	for i := 1; i < attempt; i++ {
		delay *= 2
		if delay >= max {
			return max
		}
	}
	if delay > max {
		return max
	}
	return delay
}

func hubClientURL(cfg config.Config) (string, error) {
	clientURL := strings.TrimSpace(cfg.NATS.HubURL)
	if clientURL != "" {
		return clientURL, nil
	}
	leafURL := strings.TrimSpace(cfg.Remote.NATS.URL)
	if leafURL == "" {
		return "", fmt.Errorf("hub url unavailable")
	}
	derived, err := deriveHubURLFromLeaf(leafURL)
	if err != nil {
		return "", err
	}
	return derived, nil
}

func deriveHubURLFromLeaf(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("derive hub url: %w", err)
	}
	host := parsed.Hostname()
	portStr := parsed.Port()
	if host == "" || portStr == "" {
		return "", fmt.Errorf("derive hub url: missing host or port")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", fmt.Errorf("derive hub url: %w", err)
	}
	if port <= 3200 {
		return "", fmt.Errorf("derive hub url: invalid leaf port")
	}
	port -= 3200
	parsed.Host = net.JoinHostPort(host, strconv.Itoa(port))
	return parsed.String(), nil
}
