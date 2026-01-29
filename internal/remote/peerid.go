package remote

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/agentflare-ai/amux/pkg/api"
)

// LoadOrCreatePeerID loads a persisted peer ID or creates a new one.
func LoadOrCreatePeerID(dir string) (api.PeerID, error) {
	if dir == "" {
		return api.PeerID{}, fmt.Errorf("peer id: directory is empty")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return api.PeerID{}, fmt.Errorf("peer id: %w", err)
	}
	path := filepath.Join(dir, "peer_id")
	if data, err := os.ReadFile(path); err == nil {
		raw := strings.TrimSpace(string(data))
		id, err := api.ParsePeerID(raw)
		if err != nil {
			return api.PeerID{}, fmt.Errorf("peer id: %w", err)
		}
		return id, nil
	}
	id := api.NewPeerID()
	if err := os.WriteFile(path, []byte(id.String()), 0o600); err != nil {
		return api.PeerID{}, fmt.Errorf("peer id: %w", err)
	}
	return id, nil
}
