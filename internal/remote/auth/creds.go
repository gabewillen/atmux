// Package auth provides per-host NATS credential generation and subject
// authorization for amux remote orchestration.
//
// Each remote host receives a unique NKey credential pair. The director
// generates the credential during SSH bootstrap and provisions it to the
// remote host. Subject authorization rules restrict each host to its
// own control, events, and PTY subjects.
//
// See spec §5.5.6.4 for authentication and authorization requirements.
package auth

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nats-io/nkeys"
)

// HostCredential holds the NATS authentication material for a single host.
type HostCredential struct {
	// HostID is the host identifier this credential is bound to.
	HostID string

	// Seed is the NKey private seed (starts with "S").
	Seed []byte

	// PublicKey is the NKey public key (starts with "U" for user keys).
	PublicKey string
}

// GenerateHostCredential creates a unique NKey credential for the given host.
//
// Per spec §5.5.6.4: "For each host_id, the director MUST create a unique
// NATS credential [...] and MUST associate it with exactly one host_id."
func GenerateHostCredential(hostID string) (*HostCredential, error) {
	kp, err := nkeys.CreateUser()
	if err != nil {
		return nil, fmt.Errorf("generate host credential for %q: %w", hostID, err)
	}

	seed, err := kp.Seed()
	if err != nil {
		return nil, fmt.Errorf("extract seed for %q: %w", hostID, err)
	}

	pub, err := kp.PublicKey()
	if err != nil {
		return nil, fmt.Errorf("extract public key for %q: %w", hostID, err)
	}

	return &HostCredential{
		HostID:    hostID,
		Seed:      seed,
		PublicKey: pub,
	}, nil
}

// WriteCredsFile writes a NATS credentials file for a host.
// The file contains the NKey seed and is written with mode 0600
// per spec §5.5.6.4 ("file permissions no more permissive than 0600").
func WriteCredsFile(cred *HostCredential, dir string) (string, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("create creds dir: %w", err)
	}

	filename := filepath.Join(dir, cred.HostID+".creds")

	// Write the NKey seed as the credential file content.
	// The NATS client can use nkey-based authentication with the seed.
	content := fmt.Sprintf("-----BEGIN NATS USER NKEY SEED-----\n%s\n-----END NATS USER NKEY SEED-----\n", string(cred.Seed))

	if err := os.WriteFile(filename, []byte(content), 0600); err != nil {
		return "", fmt.Errorf("write creds file: %w", err)
	}

	return filename, nil
}

// HostSubjectPermissions returns the publish and subscribe subject permissions
// for a given host_id and subject prefix per spec §5.5.6.4.
//
// These rules MUST be enforced by the NATS server for traffic attributable
// to the given host_id.
func HostSubjectPermissions(prefix, hostID string) (publish, subscribe []string) {
	publish = []string{
		prefix + ".handshake." + hostID,
		prefix + ".events." + hostID,
		prefix + ".pty." + hostID + ".*.out",
		prefix + ".comm.director",
		prefix + ".comm.manager.*",
		prefix + ".comm.agent.*.>",
		prefix + ".comm.broadcast",
	}

	subscribe = []string{
		prefix + ".ctl." + hostID,
		prefix + ".pty." + hostID + ".*.in",
		prefix + ".comm.manager." + hostID,
		prefix + ".comm.agent." + hostID + ".>",
		prefix + ".comm.broadcast",
		"_INBOX.>",
	}

	return publish, subscribe
}
