package remote

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/pkg/api"
)

// BootstrapRequest describes a remote bootstrap request.
type BootstrapRequest struct {
	HostID        api.HostID
	Location      api.Location
	HubURL        string
	CredsPath     string
	SubjectPrefix string
	KVBucket      string
	ManagerModel  string
}

// SSHRunner executes SSH commands.
type SSHRunner interface {
	Run(ctx context.Context, target string, options []string, command string, stdin []byte) error
}

// ExecSSHRunner executes SSH commands using the system ssh binary.
type ExecSSHRunner struct{}

// Run executes an SSH command.
func (ExecSSHRunner) Run(ctx context.Context, target string, options []string, command string, stdin []byte) error {
	args := append([]string{}, options...)
	args = append(args, target, command)
	cmd := exec.CommandContext(ctx, "ssh", args...)
	if len(stdin) > 0 {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ssh run: %w", fmt.Errorf("%s", strings.TrimSpace(string(output))))
	}
	return nil
}

// Bootstrapper provisions remote credentials and configuration.
type Bootstrapper struct {
	Runner SSHRunner
}

// Bootstrap performs SSH bootstrap for a remote host.
func (b *Bootstrapper) Bootstrap(ctx context.Context, req BootstrapRequest, cred Credential) error {
	if req.HostID == "" {
		return fmt.Errorf("bootstrap: %w", ErrInvalidMessage)
	}
	if req.Location.Host == "" {
		return fmt.Errorf("bootstrap: %w", ErrInvalidMessage)
	}
	runner := b.Runner
	if runner == nil {
		runner = ExecSSHRunner{}
	}
	target := sshTarget(req.Location)
	credsDir := filepath.Dir(req.CredsPath)
	credBytes, err := cred.Marshal()
	if err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}
	if err := runner.Run(ctx, target, sshOptions(req.Location), fmt.Sprintf("mkdir -p %s && umask 077 && cat > %s && chmod 600 %s", shellEscape(credsDir), shellEscape(req.CredsPath), shellEscape(req.CredsPath)), credBytes); err != nil {
		return fmt.Errorf("bootstrap: %w", ErrBootstrapFailed)
	}
	configPath := "~/.config/amux/config.toml"
	configBytes, err := bootstrapConfig(req)
	if err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}
	configDir := filepath.Dir(configPath)
	if err := runner.Run(ctx, target, sshOptions(req.Location), fmt.Sprintf("mkdir -p %s && cat > %s", shellEscape(configDir), shellEscape(configPath)), configBytes); err != nil {
		return fmt.Errorf("bootstrap: %w", ErrBootstrapFailed)
	}
	if err := runner.Run(ctx, target, sshOptions(req.Location), "nohup amux-node >/tmp/amux-node.log 2>&1 &", nil); err != nil {
		return fmt.Errorf("bootstrap: %w", ErrBootstrapFailed)
	}
	return nil
}

func bootstrapConfig(req BootstrapRequest) ([]byte, error) {
	data := map[string]any{
		"node": map[string]any{
			"role": "manager",
		},
		"remote": map[string]any{
			"transport": "nats",
			"nats": map[string]any{
				"url":            req.HubURL,
				"creds_path":     req.CredsPath,
				"subject_prefix": req.SubjectPrefix,
				"kv_bucket":      req.KVBucket,
			},
			"manager": map[string]any{
				"enabled": true,
				"model":   req.ManagerModel,
				"host_id": req.HostID.String(),
			},
		},
	}
	encoded, err := config.EncodeTOML(data)
	if err != nil {
		return nil, fmt.Errorf("bootstrap config: %w", err)
	}
	return encoded, nil
}

func sshTarget(location api.Location) string {
	host := location.Host
	if location.User != "" {
		host = location.User + "@" + host
	}
	return host
}

func sshOptions(location api.Location) []string {
	args := []string{}
	if location.Port > 0 {
		args = append(args, "-p", fmt.Sprintf("%d", location.Port))
	}
	return args
}

func shellEscape(raw string) string {
	if raw == "" {
		return "''"
	}
	replacer := strings.NewReplacer("'", "'\\''")
	return "'" + replacer.Replace(raw) + "'"
}
