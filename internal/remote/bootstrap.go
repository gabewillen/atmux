// bootstrap.go implements SSH bootstrap for remote hosts per spec §5.5.2, §5.5.3, §5.5.6.4.
// The director provisions per-host NATS creds, copies bootstrap payload, and starts the daemon.
package remote

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// BootstrapConfig holds SSH target and paths for bootstrap (spec §5.5.2).
type BootstrapConfig struct {
	Host        string // SSH host (location.host)
	User        string // SSH user (optional; from location or SSH config)
	Port        int    // SSH port (optional; 0 = default)
	CredsPath   string // Remote path for NATS creds (remote.nats.creds_path)
	BootstrapDir string // Remote dir for bootstrap zip (e.g. ~/.amux/bootstrap)
	DaemonPath  string // Remote path for amux-manager binary (e.g. ~/.local/bin/amux-manager)
}

// SSHTarget returns the SSH target string (user@host or host, with optional -p port).
func (c *BootstrapConfig) SSHTarget() string {
	if c.User != "" {
		if c.Port != 0 && c.Port != 22 {
			return fmt.Sprintf("%s@%s", c.User, c.Host)
		}
		return fmt.Sprintf("%s@%s", c.User, c.Host)
	}
	return c.Host
}

// SSHArgs returns base SSH args (e.g. ["-p", "22", "user@host"]).
func (c *BootstrapConfig) SSHArgs() []string {
	args := []string{}
	if c.Port != 0 && c.Port != 22 {
		args = append(args, "-p", fmt.Sprintf("%d", c.Port))
	}
	args = append(args, c.SSHTarget())
	return args
}

// RunSSH runs a remote command via ssh. cmd is the remote command string (e.g. "amux-manager status").
func RunSSH(ctx context.Context, cfg *BootstrapConfig, cmd string) ([]byte, error) {
	args := append(cfg.SSHArgs(), cmd)
	c := exec.CommandContext(ctx, "ssh", args...)
	c.Stdin = nil
	out, err := c.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return ee.Stderr, fmt.Errorf("ssh: %w", err)
		}
		return nil, fmt.Errorf("ssh: %w", err)
	}
	return out, nil
}

// CopyFile copies a local file to the remote path via scp. Remote path is interpreted on the remote host.
func CopyFile(ctx context.Context, cfg *BootstrapConfig, localPath, remotePath string) error {
	dest := cfg.SSHTarget() + ":" + remotePath
	c := exec.CommandContext(ctx, "scp", "-q")
	if cfg.Port != 0 && cfg.Port != 22 {
		c.Args = append(c.Args, "-P", fmt.Sprintf("%d", cfg.Port))
	}
	c.Args = append(c.Args, localPath, dest)
	if err := c.Run(); err != nil {
		return fmt.Errorf("scp %s -> %s: %w", localPath, dest, err)
	}
	return nil
}

// ProvisionCreds writes credential content to the remote path with permissions 0600 (spec §5.5.6.4).
// The director MUST ensure file permissions are no more permissive than 0600.
func ProvisionCreds(ctx context.Context, cfg *BootstrapConfig, credsContent []byte) error {
	remotePath := cfg.CredsPath
	if remotePath == "" {
		remotePath = "~/.config/amux/nats.creds"
	}
	// Write to a temp file locally, then scp, then ssh chmod 0600
	tmp, err := os.CreateTemp("", "amux-creds-*.creds")
	if err != nil {
		return fmt.Errorf("create temp creds: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.Write(credsContent); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp creds: %w", err)
	}
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return fmt.Errorf("chmod temp creds: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp creds: %w", err)
	}
	if err := CopyFile(ctx, cfg, tmpPath, remotePath); err != nil {
		return err
	}
	_, err = RunSSH(ctx, cfg, fmt.Sprintf("chmod 0600 %s", quoteRemotePath(remotePath)))
	if err != nil {
		return fmt.Errorf("chmod 0600 on remote: %w", err)
	}
	return nil
}

func quoteRemotePath(p string) string {
	if strings.Contains(p, " ") || strings.Contains(p, "'") {
		return "'" + strings.ReplaceAll(p, "'", "'\"'\"'") + "'"
	}
	return p
}

// DaemonStatus runs "amux-manager status" on the remote host (spec §5.5.2 step 6).
func DaemonStatus(ctx context.Context, cfg *BootstrapConfig) ([]byte, error) {
	return RunSSH(ctx, cfg, "amux-manager status")
}

// StartDaemon starts the daemon in manager role so it survives the SSH session (spec §5.5.2 step 7).
// It runs amux-manager daemon --role manager --host-id <hostID> --nats-url <url> --nats-creds <path>.
func StartDaemon(ctx context.Context, cfg *BootstrapConfig, hostID, natsURL, credsPath string) error {
	if credsPath == "" {
		credsPath = cfg.CredsPath
	}
	cmd := fmt.Sprintf("amux-manager daemon --role manager --host-id %s --nats-url %s --nats-creds %s",
		quoteRemotePath(hostID), quoteRemotePath(natsURL), quoteRemotePath(credsPath))
	_, err := RunSSH(ctx, cfg, cmd)
	return err
}

// WaitForConnection waits until the daemon reports hub_connected or timeout.
func WaitForConnection(ctx context.Context, cfg *BootstrapConfig, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		out, err := DaemonStatus(ctx, cfg)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if strings.Contains(string(out), "hub_connected=true") {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	return fmt.Errorf("timeout waiting for hub connection")
}

// EnsureBootstrapDir creates the remote bootstrap directory (e.g. ~/.amux/bootstrap).
func EnsureBootstrapDir(ctx context.Context, cfg *BootstrapConfig) error {
	dir := cfg.BootstrapDir
	if dir == "" {
		dir = "~/.amux/bootstrap"
	}
	_, err := RunSSH(ctx, cfg, fmt.Sprintf("mkdir -p %s", quoteRemotePath(dir)))
	return err
}

// CopyBootstrapZip copies the local bootstrap zip to the remote bootstrap dir (spec §5.5.2 step 3).
func CopyBootstrapZip(ctx context.Context, cfg *BootstrapConfig, localZipPath string) (remotePath string, err error) {
	if err := EnsureBootstrapDir(ctx, cfg); err != nil {
		return "", fmt.Errorf("ensure bootstrap dir: %w", err)
	}
	if cfg.BootstrapDir != "" {
		remotePath = cfg.BootstrapDir + "/amux-bootstrap.zip"
	} else {
		remotePath = "~/.amux/bootstrap/amux-bootstrap.zip"
	}
	if err := CopyFile(ctx, cfg, localZipPath, remotePath); err != nil {
		return "", err
	}
	return remotePath, nil
}
