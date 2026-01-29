// Package agent implements agent orchestration (lifecycle, presence, messaging)
package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/stateforward/hsm-go/muid"
)

// AgentMessage represents a message between agents
type AgentMessage struct {
	ID        muid.MUID   `json:"id"`
	From      muid.MUID   `json:"from"`      // Sender runtime ID (set by publishing component)
	To        muid.MUID   `json:"to"`        // Recipient runtime ID (set by publishing component, or BroadcastID)
	ToSlug    string      `json:"to_slug"`   // Recipient token captured from text (typically agent_slug); case-insensitive
	Content   string      `json:"content"`
	Timestamp time.Time   `json:"timestamp"`
}

// BroadcastID is a special ID for broadcast to all participants
const BroadcastID muid.MUID = 0

// MessageRouter handles routing of messages between agents
type MessageRouter struct {
	agents map[muid.MUID]*AgentActor
	roster *Roster
	// In a real implementation, this would connect to NATS for distributed messaging
	natsEnabled bool
}

// NewMessageRouter creates a new message router
func NewMessageRouter(roster *Roster) *MessageRouter {
	return &MessageRouter{
		agents: make(map[muid.MUID]*AgentActor),
		roster: roster,
		natsEnabled: false, // Initially disabled until NATS integration is implemented in Phase 7
	}
}

// RegisterAgent registers an agent with the message router
func (mr *MessageRouter) RegisterAgent(agent *AgentActor) {
	mr.agents[agent.ID] = agent
}

// UnregisterAgent removes an agent from the message router
func (mr *MessageRouter) UnregisterAgent(agentID muid.MUID) {
	delete(mr.agents, agentID)
}

// SendMessage sends a message to a specific agent or broadcasts to all
// Implements the inter-agent messaging routes as specified in the spec
func (mr *MessageRouter) SendMessage(ctx context.Context, msg *AgentMessage) error {
	// Validate message
	if msg.From == 0 {
		return fmt.Errorf("message must have a sender")
	}

	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	// If NATS is enabled, publish to NATS subjects for distributed messaging
	// Otherwise, route locally for single-host operation
	if mr.natsEnabled {
		return mr.publishToNATS(ctx, msg)
	}

	// Local routing for single-host operation
	if msg.To == BroadcastID {
		return mr.broadcastMessage(ctx, msg)
	}

	if msg.ToSlug != "" {
		return mr.sendToSlug(ctx, msg)
	}

	return mr.sendToAgent(ctx, msg)
}

// publishToNATS publishes the message to NATS subjects for distributed messaging
// This is a placeholder that will be fully implemented in Phase 7
func (mr *MessageRouter) publishToNATS(ctx context.Context, msg *AgentMessage) error {
	// In Phase 7, this will publish to NATS subjects like:
	// P.comm.<host_id>.<agent_id>.in for direct messages
	// P.comm.broadcast for broadcast messages
	// For now, we'll route locally as a fallback
	if msg.To == BroadcastID {
		return mr.broadcastMessage(ctx, msg)
	}

	if msg.ToSlug != "" {
		return mr.sendToSlug(ctx, msg)
	}

	return mr.sendToAgent(ctx, msg)
}

// EnableNATS enables NATS-based messaging
func (mr *MessageRouter) EnableNATS() {
	mr.natsEnabled = true
}

// DisableNATS disables NATS-based messaging (fallback to local routing)
func (mr *MessageRouter) DisableNATS() {
	mr.natsEnabled = false
}

// broadcastMessage sends a message to all registered agents
func (mr *MessageRouter) broadcastMessage(ctx context.Context, msg *AgentMessage) error {
	for agentID, agent := range mr.agents {
		if agentID == msg.From {
			continue // Skip sending to self
		}

		// In a real implementation, this would send the message to the agent
		// For now, we'll just simulate handling the message
		go mr.handleReceivedMessage(ctx, agent, msg)
	}

	return nil
}

// sendToSlug finds an agent by slug and sends the message
func (mr *MessageRouter) sendToSlug(ctx context.Context, msg *AgentMessage) error {
	// Find agent by slug in the roster
	for _, entry := range mr.roster.GetAllAgents() {
		// Compare slugs (case-insensitive)
		if strings.EqualFold(entry.Name, msg.ToSlug) {
			recipientAgent, exists := mr.agents[entry.ID]
			if !exists {
				continue
			}

			return mr.sendToAgent(ctx, msg)
		}
	}

	return fmt.Errorf("agent with slug '%s' not found", msg.ToSlug)
}

// sendToAgent sends a message to a specific agent
func (mr *MessageRouter) sendToAgent(ctx context.Context, msg *AgentMessage) error {
	recipient, exists := mr.agents[msg.To]
	if !exists {
		return fmt.Errorf("recipient agent not found")
	}

	// In a real implementation, this would send the message to the agent
	// For now, we'll just simulate handling the message
	go mr.handleReceivedMessage(ctx, recipient, msg)

	return nil
}

// handleReceivedMessage simulates handling of a received message by an agent
func (mr *MessageRouter) handleReceivedMessage(ctx context.Context, agent *AgentActor, msg *AgentMessage) {
	// In a real implementation, this would deliver the message to the agent
	// and potentially trigger state changes based on the message content

	// For example, if the message contains a task assignment, we might update presence
	if strings.Contains(strings.ToLower(msg.Content), "assign") || strings.Contains(strings.ToLower(msg.Content), "task") {
		// Attempt to set the agent busy if it's currently online
		if agent.CurrentPresenceState() == PresenceOnline {
			agent.SetBusy(ctx)
		}
	}
}