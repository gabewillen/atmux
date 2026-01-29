// Package remote implements remote host management, SSH bootstrap, NATS connectivity
// and control plane operations for distributed agent orchestration.
package remote

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

	"github.com/stateforward/hsm-go/muid"
)

// Common sentinel errors for remote operations.
var (
	// ErrBootstrapFailed indicates SSH bootstrap failed.
	ErrBootstrapFailed = error(fmt.Errorf("bootstrap failed"))
	
	// ErrSSHFailed indicates SSH command execution failed.
	ErrSSHFailed = error(fmt.Errorf("ssh failed"))
	
	// ErrCredentialsFailed indicates credential provisioning failed.
	ErrCredentialsFailed = error(fmt.Errorf("credentials failed"))
	
	// ErrHostUnreachable indicates remote host is not reachable.
	ErrHostUnreachable = error(fmt.Errorf("host unreachable"))
)

// BootstrapConfig holds configuration for SSH bootstrap operations.
type BootstrapConfig struct {
	SSHHost       string        // SSH target (e.g., "user@host")
	BinaryPath    string        // Path to amux binary for target arch
	AdapterPaths  []string      // Paths to required adapter WASM modules
	CredsPath     string        // Local path to NATS credentials file
	RemoteCredsPath string      // Remote path for credentials (e.g., ~/.amux/nats.creds)
	HubURL        string        // NATS hub URL to configure
	Timeout       time.Duration // SSH operation timeout
}

// Bootstrap performs SSH bootstrap for a remote host.
// Implements §5.5.2 daemon bootstrap requirements.
func Bootstrap(ctx context.Context, hostID string, config BootstrapConfig) error {
	if hostID == "" {
		return fmt.Errorf("hostID required: %w", ErrBootstrapFailed)
	}
	
	// Create bootstrap payload ZIP
	zipData, err := createBootstrapZip(config)
	if err != nil {
		return fmt.Errorf("failed to create bootstrap ZIP: %w", err)
	}
	
	// Copy bootstrap ZIP to remote host
	remotePath := "~/.amux/bootstrap/amux-bootstrap.zip"
	if err := copyToRemote(ctx, config.SSHHost, zipData, remotePath, config.Timeout); err != nil {
		return fmt.Errorf("failed to copy bootstrap to %s: %w", config.SSHHost, err)
	}
	
	// Extract bootstrap ZIP on remote host
	if err := extractBootstrapOnRemote(ctx, config.SSHHost, remotePath, config.Timeout); err != nil {
		return fmt.Errorf("failed to extract bootstrap on %s: %w", config.SSHHost, err)
	}
	
	// Provision NATS credentials
	if err := provisionCredentials(ctx, config.SSHHost, config.CredsPath, config.RemoteCredsPath, config.Timeout); err != nil {
		return fmt.Errorf("failed to provision credentials to %s: %w", config.SSHHost, err)
	}
	
	// Start daemon if not running
	if err := ensureDaemonRunning(ctx, hostID, config); err != nil {
		return fmt.Errorf("failed to start daemon on %s: %w", config.SSHHost, err)
	}
	
	return nil
}

// createBootstrapZip creates a ZIP file containing the binary and adapter modules.
func createBootstrapZip(config BootstrapConfig) ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	
	// Add binary
	if err := addFileToZip(w, config.BinaryPath, "amux-manager"); err != nil {
		w.Close()
		return nil, fmt.Errorf("failed to add binary to ZIP: %w", err)
	}
	
	// Add adapter modules
	for _, adapterPath := range config.AdapterPaths {
		filename := filepath.Base(adapterPath)
		zipPath := "adapters/" + filename
		if err := addFileToZip(w, adapterPath, zipPath); err != nil {
			w.Close()
			return nil, fmt.Errorf("failed to add adapter %s to ZIP: %w", adapterPath, err)
		}
	}
	
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close ZIP writer: %w", err)
	}
	
	return buf.Bytes(), nil
}

// addFileToZip adds a file to a ZIP archive.
func addFileToZip(w *zip.Writer, srcPath, zipPath string) error {
	file, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", srcPath, err)
	}
	defer file.Close()
	
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat %s: %w", srcPath, err)
	}
	
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return fmt.Errorf("failed to create ZIP header for %s: %w", srcPath, err)
	}
	
	header.Name = zipPath
	header.Method = zip.Deflate
	
	// Preserve execute permission for binaries
	if strings.Contains(zipPath, "amux") {
		header.SetMode(0755)
	}
	
	writer, err := w.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("failed to create ZIP entry for %s: %w", zipPath, err)
	}
	
	if _, err := io.Copy(writer, file); err != nil {
		return fmt.Errorf("failed to write %s to ZIP: %w", srcPath, err)
	}
	
	return nil
}

// copyToRemote copies data to a remote path via SSH.
func copyToRemote(ctx context.Context, sshHost string, data []byte, remotePath string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	// Create remote directory
	mkdirCmd := exec.CommandContext(ctx, "ssh", sshHost, fmt.Sprintf("mkdir -p %s", filepath.Dir(remotePath)))
	if err := mkdirCmd.Run(); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", ErrSSHFailed)
	}
	
	// Copy data via SSH
	cmd := exec.CommandContext(ctx, "ssh", sshHost, fmt.Sprintf("cat > %s", remotePath))
	cmd.Stdin = bytes.NewReader(data)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy data to remote: %w", ErrSSHFailed)
	}
	
	return nil
}

// extractBootstrapOnRemote extracts the bootstrap ZIP on the remote host.
func extractBootstrapOnRemote(ctx context.Context, sshHost, remotePath string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	commands := []string{
		"mkdir -p ~/.local/bin ~/.config/amux/adapters",
		fmt.Sprintf("cd ~/.amux/bootstrap && unzip -o %s", filepath.Base(remotePath)),
		"cp ~/.amux/bootstrap/amux-manager ~/.local/bin/amux-manager",
		"chmod +x ~/.local/bin/amux-manager",
		"cp ~/.amux/bootstrap/adapters/* ~/.config/amux/adapters/ 2>/dev/null || true",
	}
	
	for _, cmdStr := range commands {
		cmd := exec.CommandContext(ctx, "ssh", sshHost, cmdStr)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to execute '%s': %w", cmdStr, ErrSSHFailed)
		}
	}
	
	return nil
}

// provisionCredentials copies NATS credentials to the remote host with proper permissions.
func provisionCredentials(ctx context.Context, sshHost, localPath, remotePath string, timeout time.Duration) error {
	if localPath == "" {
		// No credentials to provision
		return nil
	}
	
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	// Read local credentials
	credsData, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read credentials from %s: %w", localPath, ErrCredentialsFailed)
	}
	
	// Create remote directory with proper permissions
	mkdirCmd := exec.CommandContext(ctx, "ssh", sshHost, fmt.Sprintf("mkdir -p %s && chmod 700 %s", filepath.Dir(remotePath), filepath.Dir(remotePath)))
	if err := mkdirCmd.Run(); err != nil {
		return fmt.Errorf("failed to create remote credentials directory: %w", ErrSSHFailed)
	}
	
	// Copy credentials with restricted permissions
	copyCmd := exec.CommandContext(ctx, "ssh", sshHost, fmt.Sprintf("cat > %s && chmod 600 %s", remotePath, remotePath))
	copyCmd.Stdin = bytes.NewReader(credsData)
	if err := copyCmd.Run(); err != nil {
		return fmt.Errorf("failed to copy credentials to remote: %w", ErrCredentialsFailed)
	}
	
	return nil
}

// ensureDaemonRunning starts the daemon on the remote host if not already running.
func ensureDaemonRunning(ctx context.Context, hostID string, config BootstrapConfig) error {
	ctx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()
	
	// Check if daemon is already running
	statusCmd := exec.CommandContext(ctx, "ssh", config.SSHHost, "amux-manager status 2>/dev/null")
	if err := statusCmd.Run(); err == nil {
		// Daemon is running
		return nil
	}
	
	// Start daemon
	daemonArgs := []string{
		"daemon",
		"--role", "manager",
		"--host-id", hostID,
	}
	
	if config.HubURL != "" {
		daemonArgs = append(daemonArgs, "--nats-url", config.HubURL)
	}
	
	if config.RemoteCredsPath != "" {
		daemonArgs = append(daemonArgs, "--nats-creds", config.RemoteCredsPath)
	}
	
	cmdStr := fmt.Sprintf("nohup amux-manager %s > ~/.amux/daemon.log 2>&1 &", strings.Join(daemonArgs, " "))
	startCmd := exec.CommandContext(ctx, "ssh", config.SSHHost, cmdStr)
	if err := startCmd.Run(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", ErrSSHFailed)
	}
	
	// Brief wait for daemon to start
	time.Sleep(2 * time.Second)
	
	// Verify daemon started
	verifyCmd := exec.CommandContext(ctx, "ssh", config.SSHHost, "amux-manager status")
	if err := verifyCmd.Run(); err != nil {
		return fmt.Errorf("daemon failed to start: %w", ErrSSHFailed)
	}
	
	return nil
}

// GenerateHostID creates a new unique host identifier.
func GenerateHostID() string {
	return muid.Make().String()
}