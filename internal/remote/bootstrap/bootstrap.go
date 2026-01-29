// Package bootstrap implements SSH-based bootstrapping of remote hosts.
//
// The bootstrap process provisions a remote host to run the amux manager daemon:
//  1. Resolve the SSH target host
//  2. Construct a bootstrap ZIP payload (binary + adapter WASMs)
//  3. Copy the ZIP to the remote host
//  4. Unpack and install on the remote host
//  5. Provision NATS leaf→hub credentials
//  6. Start the daemon if not already running
//  7. Verify connectivity
//
// See spec §5.5.2 for bootstrap requirements.
package bootstrap

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/agentflare-ai/amux/internal/remote/auth"
)

// HostTarget describes the SSH target for bootstrapping.
type HostTarget struct {
	// HostID is the unique host identifier.
	HostID string

	// Host is the SSH host or alias from ~/.ssh/config.
	Host string

	// User is the SSH user (optional if defined in ssh config).
	User string

	// Port is the SSH port (optional, 0 means default).
	Port int
}

// sshTarget returns the SSH target string (user@host or just host).
func (t *HostTarget) sshTarget() string {
	if t.User != "" {
		return t.User + "@" + t.Host
	}
	return t.Host
}

// sshArgs returns the base SSH command arguments.
func (t *HostTarget) sshArgs() []string {
	var args []string
	if t.Port != 0 {
		args = append(args, "-p", fmt.Sprintf("%d", t.Port))
	}
	args = append(args, t.sshTarget())
	return args
}

// BootstrapPayload describes what to include in the bootstrap ZIP.
type BootstrapPayload struct {
	// BinaryPath is the path to the amux-manager binary for the remote OS/arch.
	BinaryPath string

	// AdapterPaths is the list of adapter WASM module paths to include.
	AdapterPaths []string
}

// BootstrapConfig holds configuration for the bootstrap process.
type BootstrapConfig struct {
	// Target is the SSH target host.
	Target *HostTarget

	// Payload describes the files to include.
	Payload *BootstrapPayload

	// HubNATSURL is the director's hub NATS URL for the manager to connect to.
	HubNATSURL string

	// Credential is the per-host NATS credential material.
	Credential *auth.HostCredential

	// RemoteInstallDir is where to install the binary on the remote host.
	// Default: ~/.local/bin
	RemoteInstallDir string

	// RemoteCredsDir is where to place the credential file on the remote host.
	// Default: ~/.amux/creds
	RemoteCredsDir string

	// RemoteBootstrapDir is the temp directory for the bootstrap ZIP.
	// Default: ~/.amux/bootstrap
	RemoteBootstrapDir string
}

// Bootstrap performs the full bootstrap sequence for a remote host.
//
// Per spec §5.5.2: the director MUST NOT require a persistent SSH connection
// after bootstrap completes successfully.
func Bootstrap(ctx context.Context, cfg *BootstrapConfig) error {
	if cfg.RemoteInstallDir == "" {
		cfg.RemoteInstallDir = "~/.local/bin"
	}
	if cfg.RemoteCredsDir == "" {
		cfg.RemoteCredsDir = "~/.amux/creds"
	}
	if cfg.RemoteBootstrapDir == "" {
		cfg.RemoteBootstrapDir = "~/.amux/bootstrap"
	}

	// Step 1: Build the bootstrap ZIP
	zipData, err := buildBootstrapZIP(cfg.Payload)
	if err != nil {
		return fmt.Errorf("bootstrap: build zip: %w", err)
	}

	// Step 2: Write credential file locally (temporary)
	credsDir, err := os.MkdirTemp("", "amux-creds-*")
	if err != nil {
		return fmt.Errorf("bootstrap: create temp creds dir: %w", err)
	}
	defer os.RemoveAll(credsDir)

	localCredsFile, err := auth.WriteCredsFile(cfg.Credential, credsDir)
	if err != nil {
		return fmt.Errorf("bootstrap: write local creds: %w", err)
	}

	// Step 3: Create remote directories
	remoteCmds := fmt.Sprintf(
		"mkdir -p %s %s %s",
		cfg.RemoteBootstrapDir,
		cfg.RemoteInstallDir,
		cfg.RemoteCredsDir,
	)
	if err := sshRun(ctx, cfg.Target, remoteCmds); err != nil {
		return fmt.Errorf("bootstrap: create remote dirs: %w", err)
	}

	// Step 4: Copy bootstrap ZIP to remote host
	remoteZipPath := cfg.RemoteBootstrapDir + "/amux-bootstrap.zip"
	if err := scpTo(ctx, cfg.Target, zipData, remoteZipPath); err != nil {
		return fmt.Errorf("bootstrap: copy zip: %w", err)
	}

	// Step 5: Unpack ZIP on remote host
	unpackCmds := fmt.Sprintf(
		"cd %s && unzip -o amux-bootstrap.zip && chmod +x amux-manager && mv amux-manager %s/amux-manager",
		cfg.RemoteBootstrapDir, cfg.RemoteInstallDir,
	)
	if err := sshRun(ctx, cfg.Target, unpackCmds); err != nil {
		return fmt.Errorf("bootstrap: unpack zip: %w", err)
	}

	// Step 6: Copy NATS credentials to remote host
	remoteCredsPath := cfg.RemoteCredsDir + "/" + cfg.Target.HostID + ".creds"
	if err := scpFile(ctx, cfg.Target, localCredsFile, remoteCredsPath); err != nil {
		return fmt.Errorf("bootstrap: copy creds: %w", err)
	}
	// Ensure permissions are <= 0600
	if err := sshRun(ctx, cfg.Target, "chmod 600 "+remoteCredsPath); err != nil {
		return fmt.Errorf("bootstrap: chmod creds: %w", err)
	}

	// Step 7: Check if daemon is running
	statusErr := sshRun(ctx, cfg.Target, cfg.RemoteInstallDir+"/amux-manager status")

	// Step 8: Start daemon if not running
	if statusErr != nil {
		startCmd := fmt.Sprintf(
			"nohup %s/amux-manager daemon --role manager --host-id %s --nats-url %s --nats-creds %s > /dev/null 2>&1 &",
			cfg.RemoteInstallDir, cfg.Target.HostID, cfg.HubNATSURL, remoteCredsPath,
		)
		if err := sshRun(ctx, cfg.Target, startCmd); err != nil {
			return fmt.Errorf("bootstrap: start daemon: %w", err)
		}
	}

	// Step 9: Verify connectivity (poll status)
	if err := waitForDaemonReady(ctx, cfg); err != nil {
		return fmt.Errorf("bootstrap: verify connectivity: %w", err)
	}

	return nil
}

// buildBootstrapZIP creates an in-memory ZIP containing the binary and adapters.
func buildBootstrapZIP(payload *BootstrapPayload) ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	// Add binary
	if payload.BinaryPath != "" {
		if err := addFileToZip(w, payload.BinaryPath, "amux-manager"); err != nil {
			return nil, fmt.Errorf("add binary to zip: %w", err)
		}
	}

	// Add adapter WASMs
	for _, adapterPath := range payload.AdapterPaths {
		name := filepath.Base(adapterPath)
		if err := addFileToZip(w, adapterPath, "adapters/"+name); err != nil {
			return nil, fmt.Errorf("add adapter %q to zip: %w", name, err)
		}
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("close zip: %w", err)
	}

	return buf.Bytes(), nil
}

// addFileToZip adds a file to a ZIP archive.
func addFileToZip(w *zip.Writer, srcPath, zipName string) error {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	f, err := w.Create(zipName)
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	return err
}

// sshRun executes a command on the remote host via SSH.
func sshRun(ctx context.Context, target *HostTarget, command string) error {
	args := append(target.sshArgs(), command)
	cmd := exec.CommandContext(ctx, "ssh", args...)
	cmd.Stdout = os.Stderr // Forward stdout to stderr for logging
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// scpTo copies data to a remote path via SSH (using stdin piping).
func scpTo(ctx context.Context, target *HostTarget, data []byte, remotePath string) error {
	// Use ssh with cat to write data to remote file
	writeCmd := fmt.Sprintf("cat > %s", remotePath)
	args := append(target.sshArgs(), writeCmd)
	cmd := exec.CommandContext(ctx, "ssh", args...)
	cmd.Stdin = bytes.NewReader(data)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// scpFile copies a local file to a remote path via scp.
func scpFile(ctx context.Context, target *HostTarget, localPath, remotePath string) error {
	var args []string
	if target.Port != 0 {
		args = append(args, "-P", fmt.Sprintf("%d", target.Port))
	}
	args = append(args, localPath, target.sshTarget()+":"+remotePath)

	cmd := exec.CommandContext(ctx, "scp", args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// waitForDaemonReady polls the remote daemon status until it reports connected.
func waitForDaemonReady(ctx context.Context, cfg *BootstrapConfig) error {
	deadline := time.After(30 * time.Second)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline:
			return fmt.Errorf("timeout waiting for daemon to become ready")
		case <-ticker.C:
			statusCmd := cfg.RemoteInstallDir + "/amux-manager status"
			args := append(cfg.Target.sshArgs(), statusCmd)
			cmd := exec.CommandContext(ctx, "ssh", args...)
			var out bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = io.Discard
			if err := cmd.Run(); err == nil {
				// Check if status indicates hub_connected=true
				if strings.Contains(out.String(), "hub_connected=true") {
					return nil
				}
			}
		}
	}
}
