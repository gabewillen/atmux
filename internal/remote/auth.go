package remote

import (
	"fmt"
	"time"

	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

// GenerateAccountKey generates a new NATS Account Key Pair.
// The private key (seed) should be persisted by the Director.
// The public key is needed by the NATS Server configuration.
func GenerateAccountKey() (nkeys.KeyPair, error) {
	return nkeys.CreateAccount()
}

// GenerateHostCredentials creates a NATS User JWT and Seed for a host, signed by the provided Account Key.
// It enforces the subject permissions specified in the plan.
func GenerateHostCredentials(accountKP nkeys.KeyPair, hostID api.HostID, prefix string) (string, string, error) {
	// 1. Create User Key
	userKP, err := nkeys.CreateUser()
	if err != nil {
		return "", "", fmt.Errorf("failed to create user key: %w", err)
	}
	userPub, _ := userKP.PublicKey()
	userSeed, _ := userKP.Seed()

	// 2. Create User Claims
	claims := jwt.NewUserClaims(userPub)
	claims.Name = string(hostID)
	claims.Expires = time.Now().Add(24 * 365 * time.Hour).Unix() // 1 year

	// 3. Define Permissions
	// Pub: handshake.<host_id>, events.<host_id>, pty.<host_id>.*.out
	// Sub: handshake.<host_id>, ctl.<host_id>, pty.<host_id>.*.in, comm.manager.<host_id>, comm.agent.<host_id>.>
	
	if prefix == "" {
		prefix = "amux"
	}
	
	pubAllow := []string{
		fmt.Sprintf("%s.handshake.%s", prefix, hostID),
		fmt.Sprintf("%s.events.%s", prefix, hostID),
		fmt.Sprintf("%s.pty.%s.*.out", prefix, hostID),
	}
	
	subAllow := []string{
		fmt.Sprintf("%s.handshake.%s", prefix, hostID),
		fmt.Sprintf("%s.ctl.%s", prefix, hostID),
		fmt.Sprintf("%s.pty.%s.*.in", prefix, hostID),
		fmt.Sprintf("%s.comm.manager.%s", prefix, hostID),
		fmt.Sprintf("%s.comm.agent.%s.>", prefix, hostID),
	}

	claims.Pub.Allow.Add(pubAllow...)
	claims.Sub.Allow.Add(subAllow...)

	// 4. Sign
	token, err := claims.Encode(accountKP)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign user jwt: %w", err)
	}

	// 5. Format as .creds (JWT + Seed)
	creds := fmt.Sprintf("-----BEGIN NATS USER JWT-----\n%s\n-----END NATS USER JWT-----\n\n-----BEGIN USER NKEY SEED-----\n%s\n-----END USER NKEY SEED-----", token, string(userSeed))

	return creds, token, nil
}

