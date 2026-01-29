package remote

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/pkg/api"
)

// BootstrapRequest describes a remote bootstrap request.
type BootstrapRequest struct {
	HostID   api.HostID
	Location api.Location
	// LeafURL is the hub leaf listen URL for manager leaf connections.
	LeafURL string
	// HubClientURL is the hub client URL for direct JetStream access.
	HubClientURL  string
	CredsPath     string
	SubjectPrefix string
	KVBucket      string
	ManagerModel  string
	Adapters      []AdapterBundle
}

// SSHRunner executes SSH commands.
type SSHRunner interface {
	Run(ctx context.Context, target string, options []string, command string, stdin []byte) error
	RunOutput(ctx context.Context, target string, options []string, command string, stdin []byte) ([]byte, error)
}

// ExecSSHRunner executes SSH commands using the system ssh binary.
type ExecSSHRunner struct{}

// Run executes an SSH command.
func (ExecSSHRunner) Run(ctx context.Context, target string, options []string, command string, stdin []byte) error {
	runner := ExecSSHRunner{}
	if _, err := runner.RunOutput(ctx, target, options, command, stdin); err != nil {
		return err
	}
	return nil
}

// RunOutput executes an SSH command and returns combined output.
func (ExecSSHRunner) RunOutput(ctx context.Context, target string, options []string, command string, stdin []byte) ([]byte, error) {
	args := append([]string{}, options...)
	args = append(args, target, command)
	cmd := exec.CommandContext(ctx, "ssh", args...)
	if len(stdin) > 0 {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ssh run: %w", fmt.Errorf("%s", strings.TrimSpace(string(output))))
	}
	return output, nil
}

// AdapterBundle describes an adapter WASM module to bootstrap.
type AdapterBundle struct {
	Name string
	Wasm []byte
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
	zipBytes, err := buildBootstrapZip(ctx, req, runner)
	if err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}
	zipPath := "~/.amux/bootstrap/amux-bootstrap.zip"
	zipDir := filepath.Dir(zipPath)
	if err := runner.Run(ctx, target, sshOptions(req.Location), fmt.Sprintf("mkdir -p %s && cat > %s", shellEscape(zipDir), shellEscape(zipPath)), zipBytes); err != nil {
		return fmt.Errorf("bootstrap: %w", ErrBootstrapFailed)
	}
	if err := runner.Run(ctx, target, sshOptions(req.Location), fmt.Sprintf("unzip -o %s -d ~", shellEscape(zipPath)), nil); err != nil {
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
	statusCmd := "PATH=\"$HOME/.local/bin:$PATH\" amux-manager status"
	if output, err := runner.RunOutput(ctx, target, sshOptions(req.Location), statusCmd, nil); err == nil {
		if isHubConnected(output) {
			return nil
		}
	}
	startCmd := fmt.Sprintf("PATH=\"$HOME/.local/bin:$PATH\" amux-manager daemon --role manager --host-id %s --nats-url %s --nats-creds %s", shellEscape(req.HostID.String()), shellEscape(req.LeafURL), shellEscape(req.CredsPath))
	if err := runner.Run(ctx, target, sshOptions(req.Location), startCmd, nil); err != nil {
		return fmt.Errorf("bootstrap: %w", ErrBootstrapFailed)
	}
	if output, err := runner.RunOutput(ctx, target, sshOptions(req.Location), statusCmd, nil); err != nil {
		return fmt.Errorf("bootstrap: %w", ErrBootstrapFailed)
	} else if !isHubConnected(output) {
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
				"url":            req.LeafURL,
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
	if strings.TrimSpace(req.HubClientURL) != "" {
		data["nats"] = map[string]any{
			"hub_url": req.HubClientURL,
		}
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

func buildBootstrapZip(ctx context.Context, req BootstrapRequest, runner SSHRunner) ([]byte, error) {
	goos, goarch, err := detectRemoteArch(ctx, req.Location, runner)
	if err != nil {
		return nil, fmt.Errorf("bootstrap zip: %w", err)
	}
	binPath, err := buildManagerBinary(ctx, goos, goarch)
	if err != nil {
		return nil, fmt.Errorf("bootstrap zip: %w", err)
	}
	defer os.Remove(binPath)
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)
	if err := addZipFile(zipWriter, binPath, ".local/bin/amux-manager", 0o755); err != nil {
		_ = zipWriter.Close()
		return nil, fmt.Errorf("bootstrap zip: %w", err)
	}
	for _, adapterBundle := range req.Adapters {
		if adapterBundle.Name == "" || len(adapterBundle.Wasm) == 0 {
			continue
		}
		path := filepath.ToSlash(filepath.Join(".config", "amux", "adapters", adapterBundle.Name, adapterBundle.Name+".wasm"))
		if err := addZipBytes(zipWriter, adapterBundle.Wasm, path, 0o644); err != nil {
			_ = zipWriter.Close()
			return nil, fmt.Errorf("bootstrap zip: %w", err)
		}
	}
	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("bootstrap zip: %w", err)
	}
	return buf.Bytes(), nil
}

func detectRemoteArch(ctx context.Context, location api.Location, runner SSHRunner) (string, string, error) {
	if runner == nil {
		runner = ExecSSHRunner{}
	}
	target := sshTarget(location)
	opts := sshOptions(location)
	rawOS, err := runner.RunOutput(ctx, target, opts, "uname -s", nil)
	if err != nil {
		return "", "", fmt.Errorf("detect arch: %w", err)
	}
	rawArch, err := runner.RunOutput(ctx, target, opts, "uname -m", nil)
	if err != nil {
		return "", "", fmt.Errorf("detect arch: %w", err)
	}
	goos, err := mapGOOS(strings.TrimSpace(string(rawOS)))
	if err != nil {
		return "", "", fmt.Errorf("detect arch: %w", err)
	}
	goarch, err := mapGOARCH(strings.TrimSpace(string(rawArch)))
	if err != nil {
		return "", "", fmt.Errorf("detect arch: %w", err)
	}
	return goos, goarch, nil
}

func mapGOOS(raw string) (string, error) {
	switch strings.ToLower(raw) {
	case "linux":
		return "linux", nil
	case "darwin":
		return "darwin", nil
	default:
		return "", fmt.Errorf("unsupported os %q", raw)
	}
}

func mapGOARCH(raw string) (string, error) {
	switch strings.ToLower(raw) {
	case "x86_64", "amd64":
		return "amd64", nil
	case "arm64", "aarch64":
		return "arm64", nil
	default:
		return "", fmt.Errorf("unsupported arch %q", raw)
	}
}

func buildManagerBinary(ctx context.Context, goos, goarch string) (string, error) {
	moduleRoot, err := findModuleRoot()
	if err != nil {
		return "", fmt.Errorf("build manager: %w", err)
	}
	tmpDir := os.TempDir()
	output := filepath.Join(tmpDir, fmt.Sprintf("amux-manager-%s-%s", goos, goarch))
	cmd := exec.CommandContext(ctx, "go", "build", "-o", output, "./cmd/amux-node")
	cmd.Dir = moduleRoot
	cmd.Env = append(os.Environ(),
		"GOOS="+goos,
		"GOARCH="+goarch,
		"CGO_ENABLED=0",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("build manager: %w", fmt.Errorf("%s", strings.TrimSpace(string(out))))
	}
	return output, nil
}

func findModuleRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("module root: %w", err)
	}
	current := wd
	for {
		path := filepath.Join(current, "go.mod")
		if _, err := os.Stat(path); err == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("module root: go.mod not found")
		}
		current = parent
	}
}

func addZipFile(w *zip.Writer, srcPath, destPath string, mode os.FileMode) error {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("zip file: %w", err)
	}
	return addZipBytes(w, data, destPath, mode)
}

func addZipBytes(w *zip.Writer, data []byte, destPath string, mode os.FileMode) error {
	if w == nil {
		return fmt.Errorf("zip file: writer is nil")
	}
	header := &zip.FileHeader{
		Name:   filepath.ToSlash(destPath),
		Method: zip.Deflate,
	}
	header.SetMode(mode)
	writer, err := w.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("zip file: %w", err)
	}
	if _, err := writer.Write(data); err != nil {
		return fmt.Errorf("zip file: %w", err)
	}
	return nil
}

func isHubConnected(output []byte) bool {
	return strings.Contains(strings.ToLower(string(output)), "hub_connected=true")
}
