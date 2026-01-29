// Package remote implements control operations for remote agent management.
package remote

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

// ControlOperations provides remote control operations for director role.
type ControlOperations struct {
	nm    *NATSManager
	subs  map[string]*nats.Subscription
	mutex sync.RWMutex
}

// NewControlOperations creates a new control operations manager.
func NewControlOperations(nm *NATSManager) *ControlOperations {
	return &ControlOperations{
		nm:   nm,
		subs: make(map[string]*nats.Subscription),
	}
}

// StartControlSubscriptions starts control request subscriptions for manager role.
func (co *ControlOperations) StartControlSubscriptions() error {
	if co.nm.role != "manager" {
		return fmt.Errorf("control subscriptions only for manager role")
	}
	
	// Subscribe to control requests for this host
	subject := co.nm.Subject("ctl", co.nm.hostID)
	sub, err := co.nm.conn.Subscribe(subject, co.handleControlRequest)
	if err != nil {
		return fmt.Errorf("failed to subscribe to control requests: %w", err)
	}
	
	co.mutex.Lock()
	co.subs["control"] = sub
	co.mutex.Unlock()
	
	return nil
}

// handleControlRequest handles incoming control requests (manager role).
func (co *ControlOperations) handleControlRequest(msg *nats.Msg) {
	if !co.nm.IsReady() {
		co.nm.sendErrorResponse(msg.Reply, "unknown", "not_ready", "handshake not completed")
		return
	}
	
	var reqMsg ControlMessage
	if err := json.Unmarshal(msg.Data, &reqMsg); err != nil {
		co.nm.sendErrorResponse(msg.Reply, "unknown", "invalid_payload", "failed to parse control request")
		return
	}
	
	switch reqMsg.Type {
	case "spawn":
		co.handleSpawnRequest(msg.Reply, reqMsg.Payload)
	case "kill":
		co.handleKillRequest(msg.Reply, reqMsg.Payload)
	case "replay":
		co.handleReplayRequest(msg.Reply, reqMsg.Payload)
	case "ping":
		co.handlePingRequest(msg.Reply, reqMsg.Payload)
	default:
		co.nm.sendErrorResponse(msg.Reply, reqMsg.Type, "unknown_type", "unsupported control message type")
	}
}

// handleSpawnRequest handles agent spawn requests.
func (co *ControlOperations) handleSpawnRequest(replyTo string, payload json.RawMessage) {
	var spawnReq SpawnPayload
	if err := json.Unmarshal(payload, &spawnReq); err != nil {
		co.nm.sendErrorResponse(replyTo, "spawn", "invalid_payload", "failed to parse spawn payload")
		return
	}
	
	if spawnReq.AgentID == "" {
		co.nm.sendErrorResponse(replyTo, "spawn", "missing_agent_id", "agent_id is required")
		return
	}
	
	if spawnReq.AgentSlug == "" {
		co.nm.sendErrorResponse(replyTo, "spawn", "missing_agent_slug", "agent_slug is required")
		return
	}
	
	// Check if session already exists (idempotency)
	if sessionID := co.getExistingSession(spawnReq.AgentID); sessionID != "" {
		response := SpawnPayload{
			AgentID:   spawnReq.AgentID,
			SessionID: sessionID,
		}
		co.nm.sendControlResponse(replyTo, "spawn", response)
		return
	}
	
	// TODO: Implement actual agent spawning
	// For now, return a mock response
	sessionID := fmt.Sprintf("%d", time.Now().UnixNano())
	
	response := SpawnPayload{
		AgentID:   spawnReq.AgentID,
		SessionID: sessionID,
	}
	
	// TODO: Store session metadata in KV
	
	co.nm.sendControlResponse(replyTo, "spawn", response)
}

// handleKillRequest handles agent kill requests.
func (co *ControlOperations) handleKillRequest(replyTo string, payload json.RawMessage) {
	// TODO: Implement kill logic
	co.nm.sendErrorResponse(replyTo, "kill", "not_implemented", "kill operation not yet implemented")
}

// handleReplayRequest handles PTY replay requests.
func (co *ControlOperations) handleReplayRequest(replyTo string, payload json.RawMessage) {
	// TODO: Implement replay logic
	co.nm.sendErrorResponse(replyTo, "replay", "not_implemented", "replay operation not yet implemented")
}

// handlePingRequest handles ping requests.
func (co *ControlOperations) handlePingRequest(replyTo string, payload json.RawMessage) {
	co.nm.sendControlResponse(replyTo, "pong", map[string]interface{}{"timestamp": time.Now().UTC().Format(time.RFC3339)})
}

// getExistingSession checks if a session already exists for the given agent ID.
func (co *ControlOperations) getExistingSession(agentID string) string {
	// TODO: Check actual session state
	return ""
}

// DirectorOperations provides control operations for director role.
type DirectorOperations struct {
	nm         *NATSManager
	connStates map[string]bool // hostID -> connected
	mutex      sync.RWMutex
}

// NewDirectorOperations creates a new director operations manager.
func NewDirectorOperations(nm *NATSManager) *DirectorOperations {
	return &DirectorOperations{
		nm:         nm,
		connStates: make(map[string]bool),
	}
}

// SpawnAgent spawns an agent on the specified remote host.
func (do *DirectorOperations) SpawnAgent(ctx context.Context, hostID string, req SpawnPayload) (*SpawnPayload, error) {
	if !do.isHostConnected(hostID) {
		return nil, fmt.Errorf("host %s is not connected", hostID)
	}
	
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal spawn request: %w", err)
	}
	
	msg := ControlMessage{
		Type:    "spawn",
		Payload: reqData,
	}
	
	msgData, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal control message: %w", err)
	}
	
	// Send request
	subject := do.nm.Subject("ctl", hostID)
	resp, err := do.nm.conn.RequestWithContext(ctx, subject, msgData)
	if err != nil {
		return nil, fmt.Errorf("spawn request failed: %w", err)
	}
	
	// Parse response
	var respMsg ControlMessage
	if err := json.Unmarshal(resp.Data, &respMsg); err != nil {
		return nil, fmt.Errorf("failed to parse spawn response: %w", err)
	}
	
	if respMsg.Type == "error" {
		var errPayload ErrorPayload
		if err := json.Unmarshal(respMsg.Payload, &errPayload); err == nil {
			return nil, fmt.Errorf("spawn failed: %s - %s", errPayload.Code, errPayload.Message)
		}
		return nil, fmt.Errorf("spawn failed with unknown error")
	}
	
	if respMsg.Type != "spawn" {
		return nil, fmt.Errorf("unexpected response type: %s", respMsg.Type)
	}
	
	var response SpawnPayload
	if err := json.Unmarshal(respMsg.Payload, &response); err != nil {
		return nil, fmt.Errorf("failed to parse spawn response payload: %w", err)
	}
	
	return &response, nil
}

// KillAgent kills an agent on the specified remote host.
func (do *DirectorOperations) KillAgent(ctx context.Context, hostID, agentID string) error {
	if !do.isHostConnected(hostID) {
		return fmt.Errorf("host %s is not connected", hostID)
	}
	
	// TODO: Implement kill request
	return fmt.Errorf("kill operation not yet implemented")
}

// ReplayPTY requests PTY replay for a session.
func (do *DirectorOperations) ReplayPTY(ctx context.Context, hostID, sessionID string) error {
	if !do.isHostConnected(hostID) {
		return fmt.Errorf("host %s is not connected", hostID)
	}
	
	// TODO: Implement replay request
	return fmt.Errorf("replay operation not yet implemented")
}

// PingHost sends a ping to verify host connectivity.
func (do *DirectorOperations) PingHost(ctx context.Context, hostID string) error {
	if !do.isHostConnected(hostID) {
		return fmt.Errorf("host %s is not connected", hostID)
	}
	
	msg := ControlMessage{
		Type:    "ping",
		Payload: json.RawMessage("{}"),
	}
	
	msgData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal ping: %w", err)
	}
	
	subject := do.nm.Subject("ctl", hostID)
	resp, err := do.nm.conn.RequestWithContext(ctx, subject, msgData)
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	
	var respMsg ControlMessage
	if err := json.Unmarshal(resp.Data, &respMsg); err != nil {
		return fmt.Errorf("failed to parse ping response: %w", err)
	}
	
	if respMsg.Type != "pong" {
		return fmt.Errorf("unexpected ping response type: %s", respMsg.Type)
	}
	
	return nil
}

// MarkHostConnected marks a host as connected.
func (do *DirectorOperations) MarkHostConnected(hostID string) {
	do.mutex.Lock()
	defer do.mutex.Unlock()
	do.connStates[hostID] = true
}

// MarkHostDisconnected marks a host as disconnected.
func (do *DirectorOperations) MarkHostDisconnected(hostID string) {
	do.mutex.Lock()
	defer do.mutex.Unlock()
	do.connStates[hostID] = false
}

// isHostConnected checks if a host is currently connected.
func (do *DirectorOperations) isHostConnected(hostID string) bool {
	do.mutex.RLock()
	defer do.mutex.RUnlock()
	return do.connStates[hostID]
}

// Close cleans up control operations.
func (co *ControlOperations) Close() {
	co.mutex.Lock()
	defer co.mutex.Unlock()
	
	for _, sub := range co.subs {
		if sub != nil {
			sub.Unsubscribe()
		}
	}
	
	co.subs = make(map[string]*nats.Subscription)
}