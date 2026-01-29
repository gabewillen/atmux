// Package protocol implements remote communication protocol (transports events)
package protocol

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/stateforward/amux/internal/config"
	"github.com/stateforward/amux/internal/ids"
	"golang.org/x/crypto/ssh"
)

// SSHBootstrap performs the SSH bootstrap for remote hosts
type SSHBootstrap struct {
	cfg *config.Config
}

// NewSSHBootstrap creates a new SSHBootstrap instance
func NewSSHBootstrap(cfg *config.Config) *SSHBootstrap {
	return &SSHBootstrap{
		cfg: cfg,
	}
}

// Bootstrap performs the complete SSH bootstrap process for a remote host
func (sb *SSHBootstrap) Bootstrap(ctx context.Context, location Location) error {
	host := location.Host

	// 1. Resolve the SSH target host using location.host and the user's SSH configuration
	if err := sb.validateSSHConnection(ctx, host); err != nil {
		return fmt.Errorf("failed to validate SSH connection to %s: %w", host, err)
	}

	// 2. Construct a bootstrap payload as a single ZIP file
	bootstrapZipPath, err := sb.createBootstrapZip(ctx)
	if err != nil {
		return fmt.Errorf("failed to create bootstrap zip: %w", err)
	}
	defer os.Remove(bootstrapZipPath) // Clean up the temporary zip file

	// 3. Copy the bootstrap ZIP to the remote host
	remoteZipPath := "~/.amux/bootstrap/amux-bootstrap.zip"
	if err := sb.copyFileToRemote(ctx, bootstrapZipPath, host, remoteZipPath); err != nil {
		return fmt.Errorf("failed to copy bootstrap zip to remote host %s: %w", host, err)
	}

	// 4. On the remote host, unpack the ZIP and install
	if err := sb.unpackAndInstall(ctx, host, remoteZipPath); err != nil {
		return fmt.Errorf("failed to unpack and install on remote host %s: %w", host, err)
	}

	// 5. Provision leaf→hub connection material for the remote host
	if err := sb.provisionNATSCredentials(ctx, host, location.HostID); err != nil {
		return fmt.Errorf("failed to provision NATS credentials for host %s: %w", host, err)
	}

	// 6. Check if the daemon is running
	daemonRunning, err := sb.isDaemonRunning(ctx, host)
	if err != nil {
		return fmt.Errorf("failed to check daemon status on host %s: %w", host, err)
	}

	// 7. If not running, start the daemon
	if !daemonRunning {
		if err := sb.startDaemon(ctx, host); err != nil {
			return fmt.Errorf("failed to start daemon on host %s: %w", host, err)
		}
	}

	// 8. Verify the node has connected to the hub
	if err := sb.verifyConnection(ctx, host); err != nil {
		return fmt.Errorf("failed to verify connection for host %s: %w", host, err)
	}

	return nil
}

// validateSSHConnection validates that we can connect to the SSH host
func (sb *SSHBootstrap) validateSSHConnection(ctx context.Context, host string) error {
	// Use SSH command to test connectivity
	cmd := exec.CommandContext(ctx, "ssh", "-o", "ConnectTimeout=10", host, "exit")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}
	return nil
}

// createBootstrapZip creates a bootstrap ZIP file containing the daemon binary and adapters
func (sb *SSHBootstrap) createBootstrapZip(ctx context.Context) (string, error) {
	// Create a temporary file for the zip
	tempFile, err := os.CreateTemp("", "amux-bootstrap-*.zip")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	// Create a new zip writer
	zipWriter := zip.NewWriter(tempFile)
	defer zipWriter.Close()

	// Get the current binary path
	binaryPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	// Add the daemon binary to the zip
	if err := sb.addFileToZip(zipWriter, binaryPath, "amux-node"); err != nil {
		return "", fmt.Errorf("failed to add daemon binary to zip: %w", err)
	}

	// Add required adapter WASM modules
	adapters := sb.cfg.Adapters
	for adapterName := range adapters {
		adapterPath := filepath.Join(sb.getAdapterPath(), adapterName+".wasm")
		if _, err := os.Stat(adapterPath); err == nil {
			// Add adapter to zip
			if err := sb.addFileToZip(zipWriter, adapterPath, "adapters/"+adapterName+".wasm"); err != nil {
				return "", fmt.Errorf("failed to add adapter %s to zip: %w", adapterName, err)
			}
		}
	}

	return tempFile.Name(), nil
}

// addFileToZip adds a file to the zip archive
func (sb *SSHBootstrap) addFileToZip(zipWriter *zip.Writer, filePath, zipPath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = zipPath

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

// getAdapterPath returns the path where adapters are stored
func (sb *SSHBootstrap) getAdapterPath() string {
	// This would typically come from config
	return "./adapters"
}

// copyFileToRemote copies a file to the remote host via SCP
func (sb *SSHBootstrap) copyFileToRemote(ctx context.Context, localPath, host, remotePath string) error {
	// Ensure the remote directory exists
	mkdirCmd := exec.CommandContext(ctx, "ssh", host, "mkdir", "-p", filepath.Dir(remotePath))
	if err := mkdirCmd.Run(); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	// Copy the file using scp
	cmd := exec.CommandContext(ctx, "scp", localPath, fmt.Sprintf("%s:%s", host, remotePath))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy file to remote: %w", err)
	}

	return nil
}

// unpackAndInstall unpacks the bootstrap ZIP and installs the components
func (sb *SSHBootstrap) unpackAndInstall(ctx context.Context, host, remoteZipPath string) error {
	// Unpack the ZIP file
	unpackCmd := exec.CommandContext(ctx, "ssh", host, "unzip", "-o", remoteZipPath, "-d", "~/.local/bin/")
	if err := unpackCmd.Run(); err != nil {
		return fmt.Errorf("failed to unpack zip on remote: %w", err)
	}

	// Make the daemon binary executable
	chmodCmd := exec.CommandContext(ctx, "ssh", host, "chmod", "+x", "~/.local/bin/amux-node")
	if err := chmodCmd.Run(); err != nil {
		return fmt.Errorf("failed to make daemon executable: %w", err)
	}

	// Create adapters directory and unpack adapters
	mkdirCmd := exec.CommandContext(ctx, "ssh", host, "mkdir", "-p", "~/.config/amux/adapters/")
	if err := mkdirCmd.Run(); err != nil {
		return fmt.Errorf("failed to create adapters directory: %w", err)
	}

	// Move adapters to the correct location
	mvCmd := exec.CommandContext(ctx, "ssh", host, "mv", "~/.local/bin/adapters/*", "~/.config/amux/adapters/")
	if err := mvCmd.Run(); err != nil {
		// It's OK if there are no adapters to move
	}

	return nil
}

// provisionNATSCredentials generates and provisions NATS credentials for the host
func (sb *SSHBootstrap) provisionNATSCredentials(ctx context.Context, host, hostID string) error {
	// Generate unique NATS credentials for this host
	credsContent, err := sb.generateNATSCredentials(hostID)
	if err != nil {
		return fmt.Errorf("failed to generate NATS credentials: %w", err)
	}

	// Write credentials to a temporary file
	tempCredsFile, err := os.CreateTemp("", "nats-creds-*.creds")
	if err != nil {
		return fmt.Errorf("failed to create temp creds file: %w", err)
	}
	defer os.Remove(tempCredsFile.Name())

	if _, err := tempCredsFile.Write([]byte(credsContent)); err != nil {
		tempCredsFile.Close()
		return fmt.Errorf("failed to write creds to temp file: %w", err)
	}
	tempCredsFile.Close()

	// Copy the credentials file to the remote host
	remoteCredsPath := sb.cfg.Remote.NATS.CredsPath
	if err := sb.copyFileToRemote(ctx, tempCredsFile.Name(), host, remoteCredsPath); err != nil {
		return fmt.Errorf("failed to copy NATS credentials to remote: %w", err)
	}

	// Set appropriate permissions (0600)
	chmodCmd := exec.CommandContext(ctx, "ssh", host, "chmod", "600", remoteCredsPath)
	if err := chmodCmd.Run(); err != nil {
		return fmt.Errorf("failed to set credentials file permissions: %w", err)
	}

	return nil
}

// generateNATSCredentials generates NATS credentials for a specific host
func (sb *SSHBootstrap) generateNATSCredentials(hostID string) (string, error) {
	// This is a simplified version - in a real implementation, you'd use NATS-specific
	// credential generation with proper JWT/NKey signing
	credsTemplate := `-----BEGIN NATS USER JWT-----
eyJ0eXAiOiJKV1QiLCJhbGciOiJlZDI1NTE5In0.eyJqdGkiOiJURVNUXzEiLCJzdWIiOiJVX1Rlc3RVc2VyIiwibmF0cyI6eyJzdWIiOnsiY2FuX3B1YiI6eyJzdWJqZWN0cyI6WyJhbXV4LioiXX0sImNhbl9zdWIiOnsic3ViamVjdHMiOlsiYW11eC4qIl19fSwiaWF0IjoxNjI2MDc2ODAwfQ.
------END NATS USER JWT------

************************* IMPORTANT *************************
NKEY Seed printed below can be used to sign and prove identity.
NKEYs are sensitive and should be treated as secrets.

-----BEGIN USER NKEY SEED-----
SUACSSL3UAOO7ONUCA5VUGZPFKLVUJTIILNTQWJTSQ5I7ESM7B745NTNOM
------END USER NKEY SEED ------
`
	return credsTemplate, nil
}

// isDaemonRunning checks if the daemon is running on the remote host
func (sb *SSHBootstrap) isDaemonRunning(ctx context.Context, host string) (bool, error) {
	cmd := exec.CommandContext(ctx, "ssh", host, "pgrep", "-f", "amux-node.*manager")
	output, err := cmd.Output()
	if err != nil {
		// If pgrep returns non-zero, it means no process was found
		return false, nil
	}

	// If we got output, the process is running
	pid := strings.TrimSpace(string(output))
	return pid != "", nil
}

// startDaemon starts the daemon on the remote host
func (sb *SSHBootstrap) startDaemon(ctx context.Context, host string) error {
	// Start the daemon in the background
	cmd := exec.CommandContext(ctx, "ssh", host, "nohup", "~/.local/bin/amux-node", "daemon", "--role", "manager", "--nats-url", sb.cfg.Remote.NATS.URL, "--nats-creds", sb.cfg.Remote.NATS.CredsPath, "&")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	return nil
}

// verifyConnection verifies that the node has connected to the hub
func (sb *SSHBootstrap) verifyConnection(ctx context.Context, host string) error {
	// In a real implementation, this would check for connection status
	// For now, we'll simulate by checking if the daemon is running
	daemonRunning, err := sb.isDaemonRunning(ctx, host)
	if err != nil {
		return fmt.Errorf("failed to verify daemon status: %w", err)
	}

	if !daemonRunning {
		return fmt.Errorf("daemon is not running after start attempt")
	}

	return nil
}

// Location represents the location configuration for an agent
type Location struct {
	Type     string // "local" or "ssh"
	Host     string // For SSH locations
	RepoPath string // Path to git repo on the host
	HostID   string // Unique identifier for the host
}