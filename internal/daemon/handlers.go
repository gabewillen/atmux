package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/agentflare-ai/amux/internal/git"
	"github.com/agentflare-ai/amux/internal/manager"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/internal/rpc"
	"github.com/agentflare-ai/amux/pkg/api"
)

type agentAddParams struct {
	Name           string        `json:"name"`
	About          string        `json:"about"`
	Adapter        string        `json:"adapter"`
	Location       locationParam `json:"location"`
	Cwd            string        `json:"cwd"`
	ListenChannels []string      `json:"listen_channels"`
}

type daemonStopParams struct {
	Force bool `json:"force"`
}

type daemonStatusResult struct {
	Role         string `json:"role"`
	HubConnected bool   `json:"hub_connected"`
	Ready        bool   `json:"ready"`
	HostID       string `json:"host_id,omitempty"`
}

type locationParam struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	RepoPath string `json:"repo_path"`
}

type agentRefParams struct {
	AgentID string `json:"agent_id"`
	Name    string `json:"name"`
}

type agentAddResult struct {
	AgentID api.AgentID `json:"agent_id"`
}

type agentListResult struct {
	Roster []api.RosterEntry `json:"roster"`
}

type attachResult struct {
	SocketPath string `json:"socket_path"`
}

type mergeParams struct {
	AgentID      string `json:"agent_id"`
	Name         string `json:"name"`
	Strategy     string `json:"strategy"`
	TargetBranch string `json:"target_branch"`
}

func (d *Daemon) registerHandlers() {
	d.server.Register("daemon.ping", d.handlePing)
	d.server.Register("daemon.version", d.handleVersion)
	d.server.Register("daemon.status", d.handleStatus)
	d.server.Register("daemon.stop", d.handleStop)
	d.server.Register("agent.add", d.handleAgentAdd)
	d.server.Register("agent.list", d.handleAgentList)
	d.server.Register("agent.remove", d.handleAgentRemove)
	d.server.Register("agent.start", d.handleAgentStart)
	d.server.Register("agent.stop", d.handleAgentStop)
	d.server.Register("agent.kill", d.handleAgentKill)
	d.server.Register("agent.restart", d.handleAgentRestart)
	d.server.Register("agent.attach", d.handleAgentAttach)
	d.server.Register("git.merge", d.handleGitMerge)
}

func (d *Daemon) handlePing(ctx context.Context, raw json.RawMessage) (any, *rpc.Error) {
	_ = ctx
	_ = raw
	return map[string]any{"ok": true}, nil
}

func (d *Daemon) handleVersion(ctx context.Context, raw json.RawMessage) (any, *rpc.Error) {
	_ = ctx
	_ = raw
	return map[string]any{"amux_version": AmuxVersion, "spec_version": SpecVersion}, nil
}

func (d *Daemon) handleStatus(ctx context.Context, raw json.RawMessage) (any, *rpc.Error) {
	_ = ctx
	_ = raw
	if d == nil {
		return nil, rpcInternal(fmt.Errorf("daemon status: daemon is nil"))
	}
	status := daemonStatusResult{
		Role: strings.TrimSpace(d.cfg.Node.Role),
	}
	if status.Role == "" {
		status.Role = "director"
	}
	if d.hostMgr != nil {
		hostStatus := d.hostMgr.Status()
		status.HubConnected = hostStatus.Connected
		status.Ready = hostStatus.Ready
		status.HostID = hostStatus.HostID
	} else if d.dispatcher != nil {
		status.HubConnected = true
		status.Ready = true
	}
	return status, nil
}

func (d *Daemon) handleStop(ctx context.Context, raw json.RawMessage) (any, *rpc.Error) {
	var params daemonStopParams
	if len(raw) > 0 {
		if err := decodeParams(raw, &params); err != nil {
			return nil, err
		}
	}
	go func() {
		_ = d.Close(context.Background(), params.Force)
	}()
	return map[string]any{"ok": true}, nil
}

func (d *Daemon) handleAgentAdd(ctx context.Context, raw json.RawMessage) (any, *rpc.Error) {
	var params agentAddParams
	if err := decodeParams(raw, &params); err != nil {
		return nil, err
	}
	locTypeRaw := params.Location.Type
	if locTypeRaw == "" {
		locTypeRaw = "local"
	}
	locType, err := api.ParseLocationType(locTypeRaw)
	if err != nil {
		return nil, rpcInvalidParams(err)
	}
	location := api.Location{Type: locType, Host: params.Location.Host, RepoPath: params.Location.RepoPath}
	record, err := d.manager.AddAgent(ctx, manager.AddRequest{
		Name:           params.Name,
		About:          params.About,
		Adapter:        params.Adapter,
		Location:       location,
		Cwd:            params.Cwd,
		ListenChannels: params.ListenChannels,
	})
	if err != nil {
		return nil, rpcInternal(err)
	}
	if record.AgentID == nil {
		return nil, rpcInternal(fmt.Errorf("agent add: %w", manager.ErrAgentInvalid))
	}
	return agentAddResult{AgentID: *record.AgentID}, nil
}

func (d *Daemon) handleAgentList(ctx context.Context, raw json.RawMessage) (any, *rpc.Error) {
	_ = raw
	records, err := d.manager.ListAgents()
	if err != nil {
		return nil, rpcInternal(err)
	}
	return agentListResult{Roster: records}, nil
}

func (d *Daemon) handleAgentRemove(ctx context.Context, raw json.RawMessage) (any, *rpc.Error) {
	var params agentRefParams
	if err := decodeParams(raw, &params); err != nil {
		return nil, err
	}
	id, err := d.resolveAgentID(params)
	if err != nil {
		return nil, rpcInvalidParams(err)
	}
	if err := d.manager.RemoveAgent(ctx, manager.RemoveRequest{AgentID: id, Name: params.Name}); err != nil {
		return nil, rpcInternal(err)
	}
	return map[string]any{"ok": true}, nil
}

func (d *Daemon) handleAgentStart(ctx context.Context, raw json.RawMessage) (any, *rpc.Error) {
	var params agentRefParams
	if err := decodeParams(raw, &params); err != nil {
		return nil, err
	}
	id, err := d.resolveAgentID(params)
	if err != nil {
		return nil, rpcInvalidParams(err)
	}
	if err := d.manager.StartAgent(ctx, id); err != nil {
		return nil, rpcInternal(err)
	}
	return map[string]any{"ok": true}, nil
}

func (d *Daemon) handleAgentStop(ctx context.Context, raw json.RawMessage) (any, *rpc.Error) {
	var params agentRefParams
	if err := decodeParams(raw, &params); err != nil {
		return nil, err
	}
	id, err := d.resolveAgentID(params)
	if err != nil {
		return nil, rpcInvalidParams(err)
	}
	if err := d.manager.StopAgent(ctx, id); err != nil {
		return nil, rpcInternal(err)
	}
	return map[string]any{"ok": true}, nil
}

func (d *Daemon) handleAgentKill(ctx context.Context, raw json.RawMessage) (any, *rpc.Error) {
	var params agentRefParams
	if err := decodeParams(raw, &params); err != nil {
		return nil, err
	}
	id, err := d.resolveAgentID(params)
	if err != nil {
		return nil, rpcInvalidParams(err)
	}
	if err := d.manager.KillAgent(ctx, id); err != nil {
		return nil, rpcInternal(err)
	}
	return map[string]any{"ok": true}, nil
}

func (d *Daemon) handleAgentRestart(ctx context.Context, raw json.RawMessage) (any, *rpc.Error) {
	var params agentRefParams
	if err := decodeParams(raw, &params); err != nil {
		return nil, err
	}
	id, err := d.resolveAgentID(params)
	if err != nil {
		return nil, rpcInvalidParams(err)
	}
	if err := d.manager.RestartAgent(ctx, id); err != nil {
		return nil, rpcInternal(err)
	}
	return map[string]any{"ok": true}, nil
}

func (d *Daemon) handleAgentAttach(ctx context.Context, raw json.RawMessage) (any, *rpc.Error) {
	var params agentRefParams
	if err := decodeParams(raw, &params); err != nil {
		return nil, err
	}
	id, err := d.resolveAgentID(params)
	if err != nil {
		return nil, rpcInvalidParams(err)
	}
	records, err := d.manager.ListAgents()
	if err != nil {
		return nil, rpcInternal(err)
	}
	var repoRoot string
	for _, record := range records {
		if record.AgentID != nil && *record.AgentID == id {
			repoRoot = record.RepoRoot
			break
		}
	}
	if repoRoot == "" {
		return nil, rpcInvalidParams(fmt.Errorf("agent not found"))
	}
	ptyConn, err := d.manager.AttachAgent(id)
	if err != nil {
		return nil, rpcInternal(err)
	}
	socketPath, err := d.startAttachProxy(ctx, repoRoot, id, ptyConn)
	if err != nil {
		return nil, rpcInternal(err)
	}
	return attachResult{SocketPath: socketPath}, nil
}

func (d *Daemon) handleGitMerge(ctx context.Context, raw json.RawMessage) (any, *rpc.Error) {
	var params mergeParams
	if err := decodeParams(raw, &params); err != nil {
		return nil, err
	}
	id, err := d.resolveAgentID(agentRefParams{AgentID: params.AgentID, Name: params.Name})
	if err != nil {
		return nil, rpcInvalidParams(err)
	}
	strategy := git.MergeStrategy(params.Strategy)
	result, err := d.manager.MergeAgent(ctx, id, strategy, params.TargetBranch)
	if err != nil {
		return nil, rpcInternal(err)
	}
	return map[string]any{"target_branch": result.TargetBranch, "strategy": string(result.Strategy)}, nil
}

func (d *Daemon) resolveAgentID(params agentRefParams) (api.AgentID, error) {
	if params.AgentID != "" {
		id, err := api.ParseAgentID(params.AgentID)
		if err != nil {
			return api.AgentID{}, err
		}
		return id, nil
	}
	if params.Name == "" {
		return api.AgentID{}, fmt.Errorf("agent reference required")
	}
	records, err := d.manager.ListAgents()
	if err != nil {
		return api.AgentID{}, err
	}
	var match api.AgentID
	for _, record := range records {
		if record.Kind != api.RosterAgent || record.AgentID == nil {
			continue
		}
		if record.Name == params.Name {
			if !match.IsZero() {
				return api.AgentID{}, fmt.Errorf("agent name is ambiguous")
			}
			match = *record.AgentID
		}
	}
	if match.IsZero() {
		return api.AgentID{}, fmt.Errorf("agent not found")
	}
	return match, nil
}

func (d *Daemon) startAttachProxy(ctx context.Context, repoRoot string, agentID api.AgentID, stream io.ReadWriteCloser) (string, error) {
	_ = ctx
	if stream == nil {
		return "", fmt.Errorf("attach: pty file missing")
	}
	dir := paths.PTYDirForRepo(repoRoot)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("attach: %w", err)
	}
	name := fmt.Sprintf("attach-%s-%d.sock", agentID.String(), time.Now().UTC().UnixNano())
	socketPath := filepath.Join(dir, name)
	if err := os.RemoveAll(socketPath); err != nil {
		return "", fmt.Errorf("attach: %w", err)
	}
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return "", fmt.Errorf("attach: %w", err)
	}
	go func() {
		defer func() {
			_ = listener.Close()
			_ = os.Remove(socketPath)
			_ = stream.Close()
		}()
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		_ = listener.Close()
		done := make(chan struct{}, 2)
		go func() {
			_, _ = io.Copy(conn, stream)
			done <- struct{}{}
		}()
		go func() {
			_, _ = io.Copy(stream, conn)
			done <- struct{}{}
		}()
		<-done
		_ = conn.Close()
		<-done
	}()
	return socketPath, nil
}

func decodeParams(raw json.RawMessage, dest any) *rpc.Error {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		return rpcInvalidParams(err)
	}
	return nil
}

func rpcInvalidParams(err error) *rpc.Error {
	return &rpc.Error{Code: rpc.CodeInvalidParams, Message: err.Error()}
}

func rpcInternal(err error) *rpc.Error {
	return &rpc.Error{Code: rpc.CodeInternalError, Message: err.Error()}
}
