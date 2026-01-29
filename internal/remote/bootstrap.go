package remote

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/pkg/api"
	"golang.org/x/crypto/ssh"
)

// SSHConfig holds SSH connection parameters.
type SSHConfig struct {
	Host     string
	User     string
	Port     int
	KeyPath  string
	Password string
}

// BootstrapRemote installs/configures the daemon on a remote host via SSH.
func BootstrapRemote(cfg config.AgentConfig, hostID api.HostID) error {
	// 1. Resolve SSH config
	sshCfg := SSHConfig{
		Host: cfg.Location.Host,
		User: cfg.Location.User,
		Port: cfg.Location.Port,
		// In a real app, we'd load keys from ~/.ssh/id_rsa or agent
	}
	if sshCfg.Port == 0 {
		sshCfg.Port = 22
	}
	if sshCfg.User == "" {
		sshCfg.User = os.Getenv("USER")
	}

	// 2. Load Amux Account Key (Director's signing key)
	accountKP, err := LoadAmuxAccountKey()
	if err != nil {
		return fmt.Errorf("failed to load amux account key (director must be configured): %w", err)
	}

	// 3. Generate NATS Creds for this host
	credsContent, _, err := GenerateHostCredentials(accountKP, hostID, "amux")
	if err != nil {
		return fmt.Errorf("failed to generate host credentials: %w", err)
	}

	// 4. Create Bootstrap ZIP
	zipBytes, err := createBootstrapZip()
	if err != nil {
		return fmt.Errorf("failed to create bootstrap zip: %w", err)
	}

	// 5. Connect SSH (simulated or real)
	if os.Getenv("AMUX_TEST_SKIP_SSH") == "1" {
		return nil
	}

	return executeSSHBootstrap(sshCfg, credsContent, zipBytes, hostID)
}

func executeSSHBootstrap(cfg SSHConfig, credsContent string, zipBytes []byte, hostID api.HostID) error {
	authMethods := []ssh.AuthMethod{}
	
	// Try loading default key
	keyPath := os.Getenv("HOME") + "/.ssh/id_rsa"
	if key, err := os.ReadFile(keyPath); err == nil {
		signer, err := ssh.ParsePrivateKey(key)
		if err == nil {
			authMethods = append(authMethods, ssh.PublicKeys(signer))
		}
	}
	// Also support agent? For now, keep it simple.
	
	clientConfig := &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: meaningful host key verification
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return fmt.Errorf("ssh dial failed: %w", err)
	}
	defer client.Close()

	// 6. Copy Bootstrap ZIP
	// We'll use a new session to cat the zip to a file
	if err := copyBytesToRemote(client, zipBytes, ".amux/bootstrap.zip"); err != nil {
		return fmt.Errorf("failed to copy bootstrap zip: %w", err)
	}

	// 7. Setup Remote Environment
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create ssh session: %w", err)
	}
	defer session.Close()

	remoteCredsPath := fmt.Sprintf(".amux/nats/%s.creds", hostID)
	// We also need a minimal config file for the manager
	remoteConfigPath := ".amux/config.toml"
	managerConfig := `
[node]
role = "manager"

[remote.manager]
enabled = true

[remote.nats]
creds_path = "~/.amux/nats/` + string(hostID) + `.creds"
`

	// Combine commands:
	// 1. Mkdir
	// 2. Write creds
	// 3. Write config
	// 4. Unzip bootstrap
	// 5. Install binary and adapters
	// 6. Start daemon if not running
	
	// We'll stream the content via stdin to a script
	script := fmt.Sprintf(`
set -e
mkdir -p .amux/nats
mkdir -p .amux/bin
mkdir -p .config/amux/adapters

cat > %s <<EOF
%s
EOF
chmod 600 %s

cat > %s <<EOF
%s
EOF

# Unzip bootstrap
unzip -o .amux/bootstrap.zip -d .amux/bootstrap_tmp

# Install binary
if [ -f .amux/bootstrap_tmp/amux-node ]; then
  mv .amux/bootstrap_tmp/amux-node .amux/bin/amux-node
  chmod +x .amux/bin/amux-node
fi

# Install adapters (simple copy for now)
if [ -d .amux/bootstrap_tmp/adapters ]; then
  cp -r .amux/bootstrap_tmp/adapters/* .config/amux/adapters/
fi

rm -rf .amux/bootstrap_tmp .amux/bootstrap.zip

# Update PATH for this session
export PATH=$HOME/.amux/bin:$PATH

if ! pgrep amux-node >/dev/null; then
  echo "Starting amux-node..."
  nohup amux-node --config %s > .amux/amux.log 2>&1 &
  # Wait a bit to ensure it started
  sleep 1
  if ! pgrep amux-node >/dev/null; then
    echo "Failed to start amux-node"
    exit 1
  fi
else
  echo "amux-node already running"
fi
`, remoteCredsPath, credsContent, remoteCredsPath, remoteConfigPath, managerConfig, remoteConfigPath)

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(script); err != nil {
		return fmt.Errorf("remote bootstrap failed: %w, stderr: %s", err, stderr.String())
	}

	return nil
}

func copyBytesToRemote(client *ssh.Client, data []byte, remotePath string) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdin = bytes.NewReader(data)
	// Ensure dir exists
	dir := filepath.Dir(remotePath)
	cmd := fmt.Sprintf("mkdir -p %s && cat > %s", dir, remotePath)
	return session.Run(cmd)
}

func createBootstrapZip() ([]byte, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	// 1. Add amux-node binary
	// For this plan, we assume we are cross-compiling or running from same arch.
	// Spec says "binary for the remote host's OS/arch".
	// We'll assume linux/amd64 for now or use the current binary if matching.
	// In a real CLI, we'd invoke the build.
	// We will attempt to use the current executable.
	
	selfPath, err := os.Executable()
	if err == nil {
		f, err := w.Create("amux-node")
		if err != nil {
			return nil, err
		}
		
		// Copy self content
		src, err := os.Open(selfPath)
		if err == nil {
			_, _ = io.Copy(f, src)
			src.Close()
		}
	}
	
	// 2. Add adapters
	// We look in ~/.config/amux/adapters or .amux/adapters
	// Just scaffolding a README for now to prevent empty zip error if no binary
	f, err := w.Create("adapters/README.txt")
	if err == nil {
		_, _ = f.Write([]byte("Adapters go here"))
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}