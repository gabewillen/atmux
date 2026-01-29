// Package protocol implements remote communication protocol (transports events)
package protocol

import (
	"context"
	"fmt"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
)

// NATSAuth provides NATS authentication and authorization functionality
type NATSAuth struct {
	nc *nats.Conn
}

// NewNATSAuth creates a new NATSAuth instance
func NewNATSAuth(nc *nats.Conn) *NATSAuth {
	return &NATSAuth{
		nc: nc,
	}
}

// GenerateHostCredentials generates unique NATS credentials for a specific host
func (na *NATSAuth) GenerateHostCredentials(hostID string) (string, error) {
	// Create a user keypair for this host
	userKP, err := nkeys.CreateUser()
	if err != nil {
		return "", fmt.Errorf("failed to create user keypair: %w", err)
	}

	// Get the public key (seed) for the user
	seed, err := userKP.Seed()
	if err != nil {
		return "", fmt.Errorf("failed to get user seed: %w", err)
	}

	// Create a NATS credential file content
	// Note: In a real implementation, this would involve generating proper JWTs
	// with appropriate permissions, but for this example we'll create a basic structure
	credsContent := fmt.Sprintf(`-----BEGIN NATS USER JWT-----
%s
------END NATS USER JWT------

************************* IMPORTANT *************************
NKEY Seed printed below can be used to sign and prove identity.
NKEYs are sensitive and should be treated as secrets.

-----BEGIN USER NKEY SEED-----
%s
------END USER NKEY SEED ------
`, generateJWTPlaceholder(hostID), string(seed))

	return credsContent, nil
}

// generateJWTPlaceholder creates a placeholder JWT for demonstration purposes
// In a real implementation, this would generate a proper signed JWT with permissions
func generateJWTPlaceholder(hostID string) string {
	// This is a placeholder - in a real implementation, you'd use a proper JWT library
	// and sign the JWT with an account/issuer key
	return "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJVbmRlZmluZWQiLCJuYXRzIjp7InN1YiI6eyJjYW5fcHViIjp7InN1YmplY3RzIjpbImFtdXguKiJdfSwiY2FuX3N1YiI6eyJzdWJqZWN0cyI6WyJhbXV4LioiXX19fQ.5mYB8j1a2f3c4d5e6f7g8h9i0j1k2l3m4n5o6p7q8r9s0t1u2v3w4x5y6z7a8b9c0d1"
}

// ValidateHostPermissions validates that a host has appropriate permissions for its subjects
func (na *NATSAuth) ValidateHostPermissions(hostID string, subject string) bool {
	// Define the allowed subject patterns for this host
	prefix := "amux" // This would come from config.Remote.NATS.SubjectPrefix
	
	allowedPatterns := []string{
		fmt.Sprintf("%s.handshake.%s", prefix, hostID), // Publish: handshake request
		fmt.Sprintf("%s.events.%s", prefix, hostID),    // Publish: host events
		fmt.Sprintf("%s.pty.%s.*.out", prefix, hostID), // Publish: PTY output from daemon to director
		fmt.Sprintf("%s.comm.director", prefix),        // Publish: messages to the director channel
		fmt.Sprintf("%s.comm.manager.%s", prefix, hostID), // Publish: messages to this host's manager channel
		fmt.Sprintf("%s.comm.agent.%s.>", prefix, hostID), // Publish: messages to agents on this host
		fmt.Sprintf("%s.comm.broadcast", prefix),       // Publish: broadcast messages
		
		fmt.Sprintf("%s.ctl.%s", prefix, hostID),      // Subscribe: control requests
		fmt.Sprintf("%s.pty.%s.*.in", prefix, hostID), // Subscribe: PTY input from director to daemon
		fmt.Sprintf("%s.comm.manager.%s", prefix, hostID), // Subscribe: this host's manager channel
		fmt.Sprintf("%s.comm.agent.%s.>", prefix, hostID), // Subscribe: channels for agents on this host
		fmt.Sprintf("%s.comm.broadcast", prefix),       // Subscribe: broadcast messages
		"_INBOX.>",                                    // Subscribe: required for NATS request-reply replies
	}

	// Check if the subject matches any of the allowed patterns
	for _, pattern := range allowedPatterns {
		if subjectMatchesPattern(subject, pattern) {
			return true
		}
	}

	return false
}

// subjectMatchesPattern checks if a subject matches a pattern that may contain wildcards
func subjectMatchesPattern(subject, pattern string) bool {
	// Replace NATS wildcards with regex equivalents for matching
	pattern = strings.Replace(pattern, ".", "\\.", -1) // Escape dots
	pattern = strings.Replace(pattern, "*", "([^\\.]+)", -1) // Replace * with [^.] group
	pattern = strings.Replace(pattern, ">", "(.*)", -1) // Replace > with .* group

	// Use regex to match
	// Note: For simplicity, we're using strings.HasPrefix/HasSuffix here
	// A full implementation would use proper regex matching
	
	if strings.Contains(pattern, "(.*)") {
		// Handle > wildcard (matches everything after)
		prefix := strings.Split(pattern, "(.*)")[0]
		return strings.HasPrefix(subject, prefix)
	} else if strings.Contains(pattern, "([^\\.]+)") {
		// Handle * wildcard (matches single segment)
		// This is a simplified check - a full implementation would use regex
		parts := strings.Split(pattern, "([^\\.]+)")
		subjectParts := strings.Split(subject, ".")
		
		if len(parts) != 2 {
			return false
		}
		
		// Check prefix matches
		if !strings.HasPrefix(subject, parts[0]) {
			return false
		}
		
		// Check suffix matches
		return strings.HasSuffix(subject, parts[1])
	}
	
	// No wildcards, direct comparison
	return subject == pattern
}

// EnforceSubjectAuthorization enforces per-host subject authorization
func (na *NATSAuth) EnforceSubjectAuthorization(hostID string, subject string, operation string) error {
	// Operation is "publish" or "subscribe"
	isAllowed := na.ValidateHostPermissions(hostID, subject)
	
	if !isAllowed {
		return fmt.Errorf("host %s is not authorized to %s on subject %s", hostID, operation, subject)
	}
	
	return nil
}