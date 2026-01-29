// Package protocol implements remote communication protocol (transports events)
package protocol

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// ControlOperations handles request-reply control operations (spawn/kill/replay)
type ControlOperations struct {
	nc           *nats.Conn
	subjectBuilder *SubjectBuilder
	requestTimeout time.Duration
}

// NewControlOperations creates a new ControlOperations instance
func NewControlOperations(nc *nats.Conn, subjectBuilder *SubjectBuilder, requestTimeout time.Duration) *ControlOperations {
	return &ControlOperations{
		nc:             nc,
		subjectBuilder: subjectBuilder,
		requestTimeout: requestTimeout,
	}
}

// Spawn sends a spawn request to the daemon
func (co *ControlOperations) Spawn(ctx context.Context, hostID string, payload SpawnPayload) (*SpawnResponsePayload, error) {
	subject := co.subjectBuilder.ControlSubject(hostID)
	
	resp, err := co.subjectBuilder.RequestControlMessage(
		co.nc,
		subject,
		SpawnType,
		payload,
		co.requestTimeout,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send spawn request: %w", err)
	}

	var controlResp ControlMessage
	if err := json.Unmarshal(resp.Data, &controlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spawn response: %w", err)
	}

	if controlResp.Type == ErrorType {
		var errorPayload ErrorPayload
		if err := json.Unmarshal(controlResp.Payload, &errorPayload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal error response: %w", err)
		}
		return nil, fmt.Errorf("spawn failed: %s (%s)", errorPayload.Message, errorPayload.Code)
	}

	if controlResp.Type != SpawnType {
		return nil, fmt.Errorf("unexpected response type: %s", controlResp.Type)
	}

	var spawnResp SpawnResponsePayload
	if err := json.Unmarshal(controlResp.Payload, &spawnResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spawn response payload: %w", err)
	}

	return &spawnResp, nil
}

// Kill sends a kill request to the daemon
func (co *ControlOperations) Kill(ctx context.Context, hostID string, payload KillPayload) (*KillResponsePayload, error) {
	subject := co.subjectBuilder.ControlSubject(hostID)
	
	resp, err := co.subjectBuilder.RequestControlMessage(
		co.nc,
		subject,
		KillType,
		payload,
		co.requestTimeout,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send kill request: %w", err)
	}

	var controlResp ControlMessage
	if err := json.Unmarshal(resp.Data, &controlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal kill response: %w", err)
	}

	if controlResp.Type == ErrorType {
		var errorPayload ErrorPayload
		if err := json.Unmarshal(controlResp.Payload, &errorPayload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal error response: %w", err)
		}
		return nil, fmt.Errorf("kill failed: %s (%s)", errorPayload.Message, errorPayload.Code)
	}

	if controlResp.Type != KillType {
		return nil, fmt.Errorf("unexpected response type: %s", controlResp.Type)
	}

	var killResp KillResponsePayload
	if err := json.Unmarshal(controlResp.Payload, &killResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal kill response payload: %w", err)
	}

	return &killResp, nil
}

// Replay sends a replay request to the daemon
func (co *ControlOperations) Replay(ctx context.Context, hostID string, payload ReplayPayload) (*ReplayResponsePayload, error) {
	subject := co.subjectBuilder.ControlSubject(hostID)
	
	resp, err := co.subjectBuilder.RequestControlMessage(
		co.nc,
		subject,
		ReplayType,
		payload,
		co.requestTimeout,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send replay request: %w", err)
	}

	var controlResp ControlMessage
	if err := json.Unmarshal(resp.Data, &controlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal replay response: %w", err)
	}

	if controlResp.Type == ErrorType {
		var errorPayload ErrorPayload
		if err := json.Unmarshal(controlResp.Payload, &errorPayload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal error response: %w", err)
		}
		return nil, fmt.Errorf("replay failed: %s (%s)", errorPayload.Message, errorPayload.Code)
	}

	if controlResp.Type != ReplayType {
		return nil, fmt.Errorf("unexpected response type: %s", controlResp.Type)
	}

	var replayResp ReplayResponsePayload
	if err := json.Unmarshal(controlResp.Payload, &replayResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal replay response payload: %w", err)
	}

	return &replayResp, nil
}

// FailFastCheck checks if the host is considered disconnected before attempting operations
func (co *ControlOperations) FailFastCheck(hostID string, agentLifecycleState string) error {
	// If the director considers the target host disconnected (e.g., agent lifecycle is `Away`),
	// it MUST fail fast by rejecting remote control operations (`spawn`, `kill`, `replay`) without issuing a NATS request.
	if agentLifecycleState == "Away" {
		return fmt.Errorf("host %s is disconnected (lifecycle state: %s), failing fast", hostID, agentLifecycleState)
	}
	
	// Additional checks could go here
	return nil
}

// IsNotReadyError checks if the daemon replied with an error whose code is "not_ready"
func (co *ControlOperations) IsNotReadyError(err error) bool {
	if err == nil {
		return false
	}
	
	// Check if the error message contains "not_ready"
	return fmt.Sprintf("%v", err) == "spawn failed: host not ready (not_ready)" ||
		   fmt.Sprintf("%v", err) == "kill failed: host not ready (not_ready)" ||
		   fmt.Sprintf("%v", err) == "replay failed: host not ready (not_ready)"
}