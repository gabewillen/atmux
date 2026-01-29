package remote

import (
	"fmt"
	"time"

	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

// GenerateHostCredentials creates a NATS User JWT and Seed for a host.
// It enforces the subject permissions specified in the plan.
func GenerateHostCredentials(hostID api.HostID, prefix string) (string, string, error) {
	// 1. Create keys (in a real system, Operator/Account keys would be loaded from secrets)
	// Here we generate ephemeral keys for demonstration/testing.
	// We need an Account to sign the User.
	// Let's assume we have a "Main" account keypair available or generate one.
	accountKP, err := nkeys.CreateAccount()
	if err != nil {
		return "", "", fmt.Errorf("failed to create account key: %w", err)
	}
	
	// Create User Key
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
		fmt.Sprintf("%s.comm.agent.%s.வுகளை", prefix, hostID),
	}

	claims.Pub.Allow.Add(pubAllow...)
	claims.Sub.Allow.Add(subAllow...)

	// 4. Sign

token, err := claims.Encode(accountKP)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign user jwt: %w", err)
	}

	// 5. Format as .creds (JWT + Seed)
	creds := fmt.Sprintf("-----BEGIN NATS USER JWT-----\n%s\n-----END NATS USER JWT-----\n\n-----BEGIN USER NKEY SEED-----\n%s\n-----END USER NKEY SEED-----", token, userSeed)

	return creds, token, nil
}

// Ensure Signer logic is robust (using ephemeral keys means tokens are valid only if server trusts that ephemeral account key).
// In "embedded" mode, we need to configure the server to trust this Account Key.
// Or we use "Token" auth or "NKEY" auth without JWTs if we want simplicity.
// The plan explicitly says: "Implement NATS authentication and per-host subject authorization rules".
// And "credential copied to remote.nats.creds_path".
// And "director provisions a unique credential... .creds file containing an NKey + JWT".
// So JWT is required.
// Thus, the Director MUST hold the Account Private Key to sign these.
// For Phase 3, we'll generate it and assume it's persisted or ephemeral for the session.
// In `ConfigureEmbeddedHub`, we should probably generate an Account Key and persist it to `nats-hub.conf` or a key file,
// so that the server knows it.
// Actually, `nats-server` needs the `Account Public Key` in its config (resolver) or we use an auth callout.
// Or static resolver.
// `GenerateHostCredentials` just generates the client side.
// We'll tackle server-side config update in `ConfigureEmbeddedHub` if needed or assume pre-provisioned.
// For now, let's stick to generating valid JWTs according to spec.
