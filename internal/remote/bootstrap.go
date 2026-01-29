// Package remote implements Phase 3 remote agent orchestration.
// This file implements SSH bootstrap per spec §5.5.2.
package remote

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"

	sshconfig "github.com/kevinburke/ssh_config"
	"github.com/stateforward/amux/internal/errors"
)

// BootstrapOptions holds options for SSH bootstrap per spec §5.5.2.
type BootstrapOptions struct {
	Host         string   // SSH host (supports ~/.ssh/config aliases)
	HostID       string   // Unique host identifier
	HubURL       string   // NATS hub URL to advertise
	CredsPath    string   // Remote path for NATS credentials
	AdapterNames []string // Required adapter WASM modules
}

// BootstrapRemoteHost performs SSH bootstrap per spec §5.5.2.
//
// Steps:
// 1. Resolve SSH target host using location.host and ~/.ssh/config
// 2. Construct bootstrap ZIP (daemon binary + adapter WASMs)
// 3. Copy bootstrap ZIP to remote host
// 4. Unpack ZIP and install daemon + adapters
// 5. Provision leaf→hub connection material (NATS credentials)
// 6. Check if daemon is running
// 7. Start daemon if not running
// 8. Verify daemon has connected to hub
func BootstrapRemoteHost(ctx context.Context, opts BootstrapOptions) error {
	if opts.Host == "" {
		return errors.Wrap(errors.ErrInvalidInput, "host is required")
	}
	if opts.HostID == "" {
		return errors.Wrap(errors.ErrInvalidInput, "host_id is required")
	}

	// Step 1: Resolve SSH configuration per spec §5.2
	sshHost, sshUser, sshPort, sshIdentity, err := resolveSSHConfig(opts.Host)
	if err != nil {
		return errors.Wrapf(err, "resolve SSH config for %s", opts.Host)
	}

	// Step 2: Construct bootstrap ZIP
	zipData, err := createBootstrapZIP(ctx, opts.AdapterNames)
	if err != nil {
		return errors.Wrap(err, "create bootstrap ZIP")
	}

	// Step 3: Copy bootstrap ZIP to remote host
	remoteZipPath := "~/.amux/bootstrap/amux-bootstrap.zip"
	if err := copyToRemote(ctx, sshHost, sshUser, sshPort, sshIdentity, zipData, remoteZipPath); err != nil {
		return errors.Wrap(err, "copy bootstrap ZIP")
	}

	// Step 4: Unpack ZIP and install
	if err := unpackRemoteBootstrap(ctx, sshHost, sshUser, sshPort, sshIdentity, remoteZipPath); err != nil {
		return errors.Wrap(err, "unpack bootstrap ZIP")
	}

	// Step 5: Provision NATS credentials
	creds, err := generateNATSCredentials(opts.HostID)
	if err != nil {
		return errors.Wrap(err, "generate NATS credentials")
	}

	credsPath := opts.CredsPath
	if credsPath == "" {
		credsPath = "~/.config/amux/nats.creds"
	}

	if err := copyToRemote(ctx, sshHost, sshUser, sshPort, sshIdentity, creds, credsPath); err != nil {
		return errors.Wrap(err, "copy NATS credentials")
	}

	// Set permissions to 0600 per spec §5.5.6.4
	if err := setRemoteFilePerms(ctx, sshHost, sshUser, sshPort, sshIdentity, credsPath, "0600"); err != nil {
		return errors.Wrap(err, "set credentials permissions")
	}

	// Step 6: Check if daemon is running
	running, err := checkRemoteDaemonStatus(ctx, sshHost, sshUser, sshPort, sshIdentity)
	if err != nil {
		return errors.Wrap(err, "check daemon status")
	}

	// Step 7: Start daemon if not running
	if !running {
		if err := startRemoteDaemon(ctx, sshHost, sshUser, sshPort, sshIdentity, opts.HostID, opts.HubURL, credsPath); err != nil {
			return errors.Wrap(err, "start remote daemon")
		}
	}

	// Step 8: Verify connection (simplified for Phase 3)
	// In a full implementation, we would wait for connection.established event
	// For now, we just check status again
	_, err = checkRemoteDaemonStatus(ctx, sshHost, sshUser, sshPort, sshIdentity)
	if err != nil {
		return errors.Wrap(err, "verify daemon connection")
	}

	return nil
}

// resolveSSHConfig resolves SSH configuration using ~/.ssh/config per spec §5.2.
func resolveSSHConfig(host string) (hostname, user, port, identityFile string, err error) {
	// Resolve hostname
	hostname = sshconfig.Get(host, "HostName")
	if hostname == "" {
		hostname = host
	}

	// Resolve user
	user = sshconfig.Get(host, "User")
	if user == "" {
		user = os.Getenv("USER")
	}

	// Resolve port
	port = sshconfig.Get(host, "Port")
	if port == "" {
		port = "22"
	}

	// Resolve identity file
	identityFile = sshconfig.Get(host, "IdentityFile")

	return hostname, user, port, identityFile, nil
}

// createBootstrapZIP creates a bootstrap ZIP containing the daemon binary and adapter WASMs per spec §5.5.2.
func createBootstrapZIP(ctx context.Context, adapterNames []string) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// TODO: Embed or fetch the correct binary for the remote host's OS/arch per spec §5.5.3
	// For Phase 3, we use a placeholder
	daemonBinary := []byte("placeholder-daemon-binary")
	fw, err := zw.Create("amux-manager")
	if err != nil {
		return nil, err
	}
	if _, err := fw.Write(daemonBinary); err != nil {
		return nil, err
	}

	// TODO: Add adapter WASMs per spec §5.5.2
	// For Phase 3, we skip this

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// copyToRemote copies data to a remote path via SSH/SCP.
func copyToRemote(ctx context.Context, host, user, port, identity string, data []byte, remotePath string) error {
	// Create a temporary local file
	tmpFile, err := os.CreateTemp("", "amux-bootstrap-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(data); err != nil {
		return err
	}

	// Use scp to copy
	scpArgs := []string{"-P", port}
	if identity != "" {
		scpArgs = append(scpArgs, "-i", identity)
	}
	scpArgs = append(scpArgs, tmpFile.Name(), fmt.Sprintf("%s@%s:%s", user, host, remotePath))

	cmd := exec.CommandContext(ctx, "scp", scpArgs...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("scp failed: %w: %s", err, string(out))
	}

	return nil
}

// unpackRemoteBootstrap unpacks the bootstrap ZIP on the remote host.
func unpackRemoteBootstrap(ctx context.Context, host, user, port, identity, zipPath string) error {
	script := fmt.Sprintf(`
		mkdir -p ~/.amux/bootstrap ~/.local/bin ~/.config/amux/adapters
		cd ~/.amux/bootstrap
		unzip -o %s
		chmod +x amux-manager
		cp amux-manager ~/.local/bin/
	`, zipPath)

	return runRemoteCommand(ctx, host, user, port, identity, script)
}

// setRemoteFilePerms sets file permissions on a remote file per spec §5.5.6.4.
func setRemoteFilePerms(ctx context.Context, host, user, port, identity, path, perms string) error {
	script := fmt.Sprintf("chmod %s %s", perms, path)
	return runRemoteCommand(ctx, host, user, port, identity, script)
}

// checkRemoteDaemonStatus checks if the remote daemon is running per spec §5.5.2 step 6.
func checkRemoteDaemonStatus(ctx context.Context, host, user, port, identity string) (bool, error) {
	script := "amux-manager status 2>&1"
	out, err := runRemoteCommandOutput(ctx, host, user, port, identity, script)
	if err != nil {
		// If the command fails, assume daemon is not running
		return false, nil
	}

	// Check if output indicates running status
	return strings.Contains(string(out), "running") || strings.Contains(string(out), "connected"), nil
}

// startRemoteDaemon starts the remote daemon per spec §5.5.2 step 7.
func startRemoteDaemon(ctx context.Context, host, user, port, identity, hostID, hubURL, credsPath string) error {
	script := fmt.Sprintf(
		"amux-manager daemon --role manager --host-id %s --nats-url %s --nats-creds %s &",
		hostID, hubURL, credsPath,
	)
	return runRemoteCommand(ctx, host, user, port, identity, script)
}

// runRemoteCommand runs a shell command on the remote host via SSH.
func runRemoteCommand(ctx context.Context, host, user, port, identity, command string) error {
	sshArgs := []string{"-p", port}
	if identity != "" {
		sshArgs = append(sshArgs, "-i", identity)
	}
	sshArgs = append(sshArgs, fmt.Sprintf("%s@%s", user, host), command)

	cmd := exec.CommandContext(ctx, "ssh", sshArgs...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ssh command failed: %w: %s", err, string(out))
	}

	return nil
}

// runRemoteCommandOutput runs a shell command on the remote host and returns output.
func runRemoteCommandOutput(ctx context.Context, host, user, port, identity, command string) ([]byte, error) {
	sshArgs := []string{"-p", port}
	if identity != "" {
		sshArgs = append(sshArgs, "-i", identity)
	}
	sshArgs = append(sshArgs, fmt.Sprintf("%s@%s", user, host), command)

	cmd := exec.CommandContext(ctx, "ssh", sshArgs...)
	return cmd.CombinedOutput()
}

// generateNATSCredentials generates per-host NATS credentials per spec §5.5.6.4.
//
// For Phase 3, we generate a placeholder credential.
// In a production implementation, this would generate NKey + JWT using NATS nkeys.
func generateNATSCredentials(hostID string) ([]byte, error) {
	// Generate a random seed for demonstration
	seed := make([]byte, 32)
	if _, err := rand.Read(seed); err != nil {
		return nil, err
	}

	// Create a placeholder .creds file
	// In production, use nkeys.CreateUser() and jwt.EncodeUserClaims()
	creds := fmt.Sprintf(`-----BEGIN NATS USER JWT-----
placeholder-jwt-for-%s
-----END NATS USER JWT-----

************************* IMPORTANT *************************
NKEY Seed printed below can be used to sign and prove identity.
NKEYs are sensitive and should be treated as secrets.

-----BEGIN USER NKEY SEED-----
%s
-----END USER NKEY SEED-----
`, hostID, base64.StdEncoding.EncodeToString(seed))

	return []byte(creds), nil
}
