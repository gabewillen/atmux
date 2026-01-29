package remote

import (
	"fmt"
	"strings"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
)

// SubjectPrefix normalizes the configured subject prefix.
func SubjectPrefix(prefix string) string {
	trimmed := strings.Trim(prefix, ".")
	if trimmed == "" {
		return "amux"
	}
	return trimmed
}

// HandshakeSubject returns the subject for handshake requests.
func HandshakeSubject(prefix string, hostID api.HostID) string {
	return protocol.Subject(SubjectPrefix(prefix), "handshake", hostID.String())
}

// ControlSubject returns the subject for control requests.
func ControlSubject(prefix string, hostID api.HostID) string {
	return protocol.Subject(SubjectPrefix(prefix), "ctl", hostID.String())
}

// EventsSubject returns the subject for host events.
func EventsSubject(prefix string, hostID api.HostID) string {
	return protocol.Subject(SubjectPrefix(prefix), "events", hostID.String())
}

// PtyOutSubject returns the subject for PTY output.
func PtyOutSubject(prefix string, hostID api.HostID, sessionID api.SessionID) string {
	return protocol.Subject(SubjectPrefix(prefix), "pty", hostID.String(), sessionID.String(), "out")
}

// PtyInSubject returns the subject for PTY input.
func PtyInSubject(prefix string, hostID api.HostID, sessionID api.SessionID) string {
	return protocol.Subject(SubjectPrefix(prefix), "pty", hostID.String(), sessionID.String(), "in")
}

// ManagerCommSubject returns the subject for a manager channel.
func ManagerCommSubject(prefix string, hostID api.HostID) string {
	return protocol.Subject(SubjectPrefix(prefix), "comm", "manager", hostID.String())
}

// AgentCommSubject returns the subject for an agent channel.
func AgentCommSubject(prefix string, hostID api.HostID, agentID api.AgentID) string {
	return protocol.Subject(SubjectPrefix(prefix), "comm", "agent", hostID.String(), agentID.String())
}

// DirectorCommSubject returns the subject for the director channel.
func DirectorCommSubject(prefix string) string {
	return protocol.Subject(SubjectPrefix(prefix), "comm", "director")
}

// BroadcastCommSubject returns the broadcast channel subject.
func BroadcastCommSubject(prefix string) string {
	return protocol.Subject(SubjectPrefix(prefix), "comm", "broadcast")
}

// ParseHandshakeSubject extracts the host_id from a handshake subject.
func ParseHandshakeSubject(prefix string, subject string) (api.HostID, error) {
	prefixParts := strings.Split(SubjectPrefix(prefix), ".")
	parts := strings.Split(subject, ".")
	if len(parts) != len(prefixParts)+2 {
		return "", fmt.Errorf("parse handshake subject: %w", ErrInvalidSubject)
	}
	for i, part := range prefixParts {
		if parts[i] != part {
			return "", fmt.Errorf("parse handshake subject: %w", ErrInvalidSubject)
		}
	}
	if parts[len(prefixParts)] != "handshake" {
		return "", fmt.Errorf("parse handshake subject: %w", ErrInvalidSubject)
	}
	id, err := api.ParseHostID(parts[len(prefixParts)+1])
	if err != nil {
		return "", fmt.Errorf("parse handshake subject: %w", err)
	}
	return id, nil
}

// ParseSessionSubject extracts the session_id from a PTY subject.
func ParseSessionSubject(prefix string, subject string) (api.HostID, api.SessionID, string, error) {
	prefixParts := strings.Split(SubjectPrefix(prefix), ".")
	parts := strings.Split(subject, ".")
	if len(parts) < len(prefixParts)+4 {
		return "", api.SessionID{}, "", fmt.Errorf("parse pty subject: %w", ErrInvalidSubject)
	}
	for i, part := range prefixParts {
		if parts[i] != part {
			return "", api.SessionID{}, "", fmt.Errorf("parse pty subject: %w", ErrInvalidSubject)
		}
	}
	if parts[len(prefixParts)] != "pty" {
		return "", api.SessionID{}, "", fmt.Errorf("parse pty subject: %w", ErrInvalidSubject)
	}
	hostID, err := api.ParseHostID(parts[len(prefixParts)+1])
	if err != nil {
		return "", api.SessionID{}, "", fmt.Errorf("parse pty subject: %w", err)
	}
	sessionID, err := api.ParseSessionID(parts[len(prefixParts)+2])
	if err != nil {
		return "", api.SessionID{}, "", fmt.Errorf("parse pty subject: %w", err)
	}
	return hostID, sessionID, parts[len(prefixParts)+3], nil
}
