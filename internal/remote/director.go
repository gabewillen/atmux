package remote

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
	"github.com/stateforward/hsm-go/muid"
)

// DirectorProtocol manages the director-side remote protocol.
type DirectorProtocol struct {
	nc            *nats.Conn
	subjectPrefix string
}

// NewDirectorProtocol creates a new DirectorProtocol.
func NewDirectorProtocol(nc *nats.Conn, prefix string) *DirectorProtocol {
	return &DirectorProtocol{
		nc:            nc,
		subjectPrefix: prefix,
	}
}

// Start starts listening for handshake requests.
func (d *DirectorProtocol) Start(ctx context.Context) error {
	// Subscribe to P.handshake.>
	subject := fmt.Sprintf("%s.handshake.>", d.subjectPrefix)
	
	sub, err := d.nc.QueueSubscribe(subject, "director", func(msg *nats.Msg) {
		d.handleHandshake(msg)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to handshake: %w", err)
	}
	
	go func() {
		<-ctx.Done()
		sub.Unsubscribe()
	}()
	
	return nil
}

func (d *DirectorProtocol) handleHandshake(msg *nats.Msg) {
	var req protocol.HandshakeRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		d.replyError(msg, "handshake", "malformed_payload", "Invalid JSON")
		return
	}
	
	// Extract host_id from subject: prefix.handshake.<host_id>
	// We assume prefix does not contain dots for simplicity or split carefully.
	// Tokens: [prefix..., "handshake", hostID]
	// Better: Trim prefix + ".handshake."
	prefixLen := len(d.subjectPrefix) + 1 + len("handshake") + 1
	if len(msg.Subject) <= prefixLen {
		d.replyError(msg, "handshake", "invalid_subject", "Subject too short")
		return
	}
	subjectHostID := msg.Subject[prefixLen:]
	
	// Spec: "The director MUST treat the <host_id> token in the request subject as canonical. 
	// If the handshake payload contains a different host_id, the director MUST reject the handshake."
	if req.HostID != api.HostID(subjectHostID) {
		d.replyError(msg, "handshake", "host_id_mismatch", fmt.Sprintf("Subject hostID %s != Payload hostID %s", subjectHostID, req.HostID))
		return
	}

	if req.Role != "daemon" {
		d.replyError(msg, "handshake", "invalid_role", "Expected role 'daemon'")
		return
	}
	
	// Reply with success
	// In a real implementation these IDs would come from config/runtime
	resp := protocol.HandshakeResponse{
		Protocol: 1,
		Role:     "director",
		HostID:   "director-host", 
		PeerID:   api.PeerID(muid.Make()),
	}
	
	data, _ := json.Marshal(resp)
	d.nc.Publish(msg.Reply, data)
}

func (d *DirectorProtocol) replyError(msg *nats.Msg, reqType, code, message string) {
	// Handshake response usually uses HandshakeResponse struct but Spec 5.5.7.3 says 
	// "reply with an error whose payload has request_type set to ... code ... message".
	// The Error envelope logic in 5.5.7.2 applies.
	// But HandshakeResponse has `Error *ControlError`.
	resp := protocol.HandshakeResponse{
		Error: &protocol.Error{
			RequestType: reqType,
			Code:        code,
			Message:     message,
		},
	}
	data, _ := json.Marshal(resp)
	d.nc.Publish(msg.Reply, data)
}