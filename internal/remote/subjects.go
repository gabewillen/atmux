// Package remote implements remote agent orchestration: NATS hub/leaf, JetStream KV,
// handshake, request-reply control (spawn/kill/replay), PTY I/O subjects, and replay buffering
// per spec §5.5, §5.5.6–§5.5.8.
//
// This file defines NATS subject namespaces per spec §5.5.7.1.
package remote

import "fmt"

// SubjectPrefix returns the subject prefix P (default "amux"). All remote subjects use P as literal prefix.
func SubjectPrefix(prefix string) string {
	if prefix == "" {
		return "amux"
	}
	return prefix
}

// SubjectHandshake returns P.handshake.<host_id> (daemon → director, request-reply).
func SubjectHandshake(prefix, hostID string) string {
	return fmt.Sprintf("%s.handshake.%s", SubjectPrefix(prefix), hostID)
}

// SubjectHandshakeAll returns P.handshake.> for subscribing to all handshakes (director).
func SubjectHandshakeAll(prefix string) string {
	return fmt.Sprintf("%s.handshake.>", SubjectPrefix(prefix))
}

// SubjectCtl returns P.ctl.<host_id> (director → daemon, request-reply).
func SubjectCtl(prefix, hostID string) string {
	return fmt.Sprintf("%s.ctl.%s", SubjectPrefix(prefix), hostID)
}

// SubjectEvents returns P.events.<host_id> (daemon → director, host events).
func SubjectEvents(prefix, hostID string) string {
	return fmt.Sprintf("%s.events.%s", SubjectPrefix(prefix), hostID)
}

// SubjectPTYOut returns P.pty.<host_id>.<session_id>.out (daemon → director).
func SubjectPTYOut(prefix, hostID, sessionID string) string {
	return fmt.Sprintf("%s.pty.%s.%s.out", SubjectPrefix(prefix), hostID, sessionID)
}

// SubjectPTYIn returns P.pty.<host_id>.<session_id>.in (director → daemon).
func SubjectPTYIn(prefix, hostID, sessionID string) string {
	return fmt.Sprintf("%s.pty.%s.%s.in", SubjectPrefix(prefix), hostID, sessionID)
}

// SubjectPTYInWildcard returns P.pty.<host_id>.*.in for subscribing to all sessions on a host.
func SubjectPTYInWildcard(prefix, hostID string) string {
	return fmt.Sprintf("%s.pty.%s.*.in", SubjectPrefix(prefix), hostID)
}
