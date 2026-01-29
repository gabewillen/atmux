// Package messaging implements inter-agent messaging routes per spec §6.4.
// This package handles message routing, addressing, and delivery over NATS.
package messaging

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/stateforward/hsm-go/muid"

	"github.com/copilot-claude-sonnet-4/amux/internal/roster"
)

// BroadcastID is the special ID for broadcast messages per spec §6.4.
const BroadcastID muid.MUID = 0

// AgentMessage represents a message sent between agents per spec §6.4.
type AgentMessage struct {
	// ID is the unique identifier for this message.
	ID muid.MUID `json:"id"`

	// From is the sender runtime ID (set by publishing component).
	From muid.MUID `json:"from"`

	// To is the recipient runtime ID (set by publishing component, or BroadcastID).
	To muid.MUID `json:"to"`

	// ToSlug is the recipient token captured from text (typically agent_slug).
	ToSlug string `json:"to_slug"`

	// Content is the message content.
	Content string `json:"content"`

	// Timestamp is when this message was sent.
	Timestamp time.Time `json:"timestamp"`
}

// Router handles message routing and addressing per spec §6.4.1.
type Router struct {
	// roster is the roster store for participant lookup.
	roster *roster.Store

	// hostID is this host's identifier.
	hostID string

	// directorID is the director's runtime ID.
	directorID muid.MUID

	// localManagerID is the local host manager's runtime ID.
	localManagerID muid.MUID

	// mu protects concurrent access to router state.
	mu sync.RWMutex

	// ctx is the router context.
	ctx context.Context

	// cancel cancels the router context.
	cancel context.CancelFunc
}

// MessageEvent represents message routing events per spec §6.4.2.
type MessageEvent struct {
	// Type is the event type: "message.outbound", "message.inbound", or "message.broadcast".
	Type string `json:"type"`

	// Message is the agent message.
	Message *AgentMessage `json:"message"`

	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`
}

// NewRouter creates a new message router.
func NewRouter(roster *roster.Store, hostID string, directorID, localManagerID muid.MUID) *Router {
	ctx, cancel := context.WithCancel(context.Background())
	return &Router{
		roster:         roster,
		hostID:         hostID,
		directorID:     directorID,
		localManagerID: localManagerID,
		ctx:            ctx,
		cancel:         cancel,
	}
}

// ProcessOutboundMessage processes an outbound message from an agent per spec §6.4.1.
func (r *Router) ProcessOutboundMessage(fromID muid.MUID, toSlug, content string) (*AgentMessage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Create the message with enriched metadata
	msg := &AgentMessage{
		ID:        muid.Make(),
		From:      fromID,
		ToSlug:    toSlug,
		Content:   content,
		Timestamp: time.Now().UTC(),
	}

	// Resolve ToSlug per spec §6.4.1.3
	resolvedID, err := r.resolveToSlug(strings.ToLower(toSlug))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve recipient '%s': %w", toSlug, err)
	}

	msg.To = resolvedID
	return msg, nil
}

// resolveToSlug resolves ToSlug to a runtime ID per spec §6.4.1.3.
func (r *Router) resolveToSlug(toSlug string) (muid.MUID, error) {
	// Case-insensitive resolution per spec
	toSlug = strings.ToLower(toSlug)

	// Handle broadcast aliases
	if toSlug == "all" || toSlug == "broadcast" || toSlug == "*" {
		return BroadcastID, nil
	}

	// Handle director
	if toSlug == "director" {
		return r.directorID, nil
	}

	// Handle local manager alias
	if toSlug == "manager" {
		return r.localManagerID, nil
	}

	// Handle manager@<host_id> format
	if strings.HasPrefix(toSlug, "manager@") {
		hostID := strings.TrimPrefix(toSlug, "manager@")
		managerSlug := fmt.Sprintf("manager@%s", hostID)
		entry, err := r.roster.GetBySlug(managerSlug)
		if err != nil {
			return 0, fmt.Errorf("manager for host '%s' not found", hostID)
		}
		return entry.ID, nil
	}

	// Look up in roster by slug
	entry, err := r.roster.GetBySlug(toSlug)
	if err != nil {
		return 0, fmt.Errorf("recipient with slug '%s' not found in roster", toSlug)
	}

	return entry.ID, nil
}

// GetParticipantChannels returns NATS subject names for a participant per spec §5.5.7.1.
func (r *Router) GetParticipantChannels(participantID muid.MUID) ([]string, error) {
	entry, err := r.roster.GetByID(participantID)
	if err != nil {
		return nil, fmt.Errorf("participant %s not found", participantID)
	}

	var channels []string

	switch entry.Type {
	case "director":
		// Director subscribes to P.comm.> for observation
		channels = []string{"P.comm.>"}
	case "manager":
		// Host manager subscribes to manager.<host_id>, agent.<host_id>.>, and broadcast
		hostID := entry.HostID
		channels = []string{
			fmt.Sprintf("P.comm.manager.%s", hostID),
			fmt.Sprintf("P.comm.agent.%s.>", hostID),
			"P.comm.broadcast",
		}
	case "agent":
		// Agent can listen to specific channels (mechanism may be configuration-driven)
		hostID := entry.HostID
		agentSlug := entry.Slug
		channels = []string{
			fmt.Sprintf("P.comm.agent.%s.%s", hostID, agentSlug),
		}
	}

	return channels, nil
}

// GetDeliveryChannels returns channels where a message should be published.
func (r *Router) GetDeliveryChannels(msg *AgentMessage) ([]string, error) {
	if msg.To == BroadcastID {
		// Broadcast message
		return []string{"P.comm.broadcast"}, nil
	}

	// Unicast message - get recipient channels
	recipientEntry, err := r.roster.GetByID(msg.To)
	if err != nil {
		return nil, fmt.Errorf("recipient %s not found", msg.To)
	}

	senderEntry, err := r.roster.GetByID(msg.From)
	if err != nil {
		return nil, fmt.Errorf("sender %s not found", msg.From)
	}

	var channels []string

	// For unicast, publish to both sender and recipient channels per spec
	switch recipientEntry.Type {
	case "director":
		channels = append(channels, "P.comm.director")
	case "manager":
		hostID := recipientEntry.HostID
		channels = append(channels, fmt.Sprintf("P.comm.manager.%s", hostID))
	case "agent":
		hostID := recipientEntry.HostID
		agentSlug := recipientEntry.Slug
		channels = append(channels, fmt.Sprintf("P.comm.agent.%s.%s", hostID, agentSlug))
	}

	// Also add sender's channel for unicast delivery confirmation
	switch senderEntry.Type {
	case "director":
		channels = append(channels, "P.comm.director")
	case "manager":
		hostID := senderEntry.HostID
		channels = append(channels, fmt.Sprintf("P.comm.manager.%s", hostID))
	case "agent":
		hostID := senderEntry.HostID
		agentSlug := senderEntry.Slug
		channels = append(channels, fmt.Sprintf("P.comm.agent.%s.%s", hostID, agentSlug))
	}

	return channels, nil
}

// AdapterPatterns represents adapter-specific message detection patterns per spec §6.4.3.
type AdapterPatterns struct {
	// Prompt is the pattern for detecting prompt readiness.
	Prompt string `json:"prompt"`

	// RateLimit is the pattern for detecting rate limiting.
	RateLimit string `json:"rate_limit"`

	// Error is the pattern for detecting errors.
	Error string `json:"error"`

	// Completion is the pattern for detecting task completion.
	Completion string `json:"completion"`

	// Message is the pattern for detecting outbound messages (optional).
	Message string `json:"message,omitempty"`
}

// MessageDetector detects outbound messages from PTY output using adapter patterns.
type MessageDetector struct {
	// patterns holds compiled regex patterns for message detection.
	patterns map[string]*regexp.Regexp

	// mu protects concurrent access to patterns.
	mu sync.RWMutex
}

// NewMessageDetector creates a new message detector with adapter patterns.
func NewMessageDetector() *MessageDetector {
	return &MessageDetector{
		patterns: make(map[string]*regexp.Regexp),
	}
}

// LoadPatterns loads adapter patterns for message detection.
func (d *MessageDetector) LoadPatterns(adapterName string, patterns *AdapterPatterns) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if patterns.Message == "" {
		// No message pattern defined for this adapter
		return nil
	}

	// Compile the message pattern
	regex, err := regexp.Compile(patterns.Message)
	if err != nil {
		return fmt.Errorf("invalid message pattern for adapter %s: %w", adapterName, err)
	}

	d.patterns[adapterName] = regex
	return nil
}

// DetectMessage attempts to detect an outbound message from PTY output.
func (d *MessageDetector) DetectMessage(adapterName, output string) (toSlug, content string, detected bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	pattern, exists := d.patterns[adapterName]
	if !exists {
		// No pattern defined for this adapter
		return "", "", false
	}

	matches := pattern.FindStringSubmatch(output)
	if len(matches) < 3 {
		// Pattern didn't match or insufficient capture groups
		return "", "", false
	}

	// First capture group is recipient (ToSlug), second is content
	toSlug = strings.TrimSpace(matches[1])
	content = strings.TrimSpace(matches[2])

	return toSlug, content, true
}

// Close gracefully shuts down the router.
func (r *Router) Close() error {
	r.cancel()
	return nil
}