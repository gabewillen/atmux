package remote

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
)

// SendControlRequest sends a control request to a remote host and waits for a response.
// It enforces the timeout and error handling specified in the plan.
func SendControlRequest(ctx context.Context, nc *nats.Conn, hostID api.HostID, req protocol.ControlRequest) (*protocol.ControlResponse, error) {
	subject := protocol.SubjectForCtl("amux", hostID) // TODO: prefix from config

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Use context deadline for request timeout
	// "director uses remote.request_timeout"
	// We assume ctx has the timeout set by caller.
	
	// nats.RequestWithContext requires a context.
	msg, err := nc.RequestWithContext(ctx, subject, payload)
	if err != nil {
		return nil, fmt.Errorf("nats request failed: %w", err)
	}

	var resp protocol.ControlResponse
	if err := json.Unmarshal(msg.Data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.Error != nil {
		// "not_ready errors block retries until connection.established"
		// The caller handles the retry logic based on the error code.
		return &resp, fmt.Errorf("remote error: %s (code=%s)", resp.Error.Message, resp.Error.Code)
	}

	return &resp, nil
}

// Helper to construct spawn request
func NewSpawnRequest(agentID api.AgentID, slug api.AgentSlug, repoPath string, cmd []string, env map[string]string) protocol.ControlRequest {
	payload := protocol.SpawnPayload{
		AgentID:  agentID,
		Slug:     slug,
		RepoPath: repoPath,
		Command:  cmd,
		Env:      env,
	}
	data, _ := json.Marshal(payload)
	return protocol.ControlRequest{
		Type:    "spawn",
		Payload: data,
	}
}
