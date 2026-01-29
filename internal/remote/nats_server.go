package remote

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/paths"
)

// ConfigureEmbeddedHub configures the embedded NATS server for the director.
func ConfigureEmbeddedHub(cfg *config.Config) error {
	if cfg.NATS.Mode != "embedded" {
		return nil
	}

	// 1. Ensure JetStream directory exists
	expandedJSDir := paths.ExpandHome(cfg.NATS.JetStreamDir)
	if err := os.MkdirAll(expandedJSDir, 0700); err != nil {
		return fmt.Errorf("failed to create JetStream directory %s: %w", expandedJSDir, err)
	}

	// 2. Generate NATS Config File (simulated/actual)
	// We'll write a config file that the embedded server *could* use, or just return options.
	// Since we are not linking nats-server yet (it's heavy), let's write the config to disk
	// as if we were going to pass it to `nats-server -c`.
	
	natsConfigContent := fmt.Sprintf(`
server_name: amux-hub
listen: %s
jetstream {
    store_dir: %s
}
leafnodes {
    port: 7422
}
`, cfg.NATS.Listen, expandedJSDir)

	configDir, _ := paths.DefaultConfigDir()
	natsConfigPath := filepath.Join(configDir, "nats-hub.conf")
	if err := os.MkdirAll(filepath.Dir(natsConfigPath), 0755); err != nil {
		return err
	}
	
	if err := os.WriteFile(natsConfigPath, []byte(natsConfigContent), 0600); err != nil {
		return fmt.Errorf("failed to write nats hub config: %w", err)
	}

	return nil
}

// ConfigureLeaf configures the NATS leaf connection for the manager.
// This is used by the daemon to connect to the hub.
func ConfigureLeaf(cfg *config.Config, credsPath string) error {
	// Logic to verify creds exist or prepare connection options.
	if _, err := os.Stat(credsPath); os.IsNotExist(err) {
		return fmt.Errorf("nats credentials missing at %s", credsPath)
	}
	return nil
}