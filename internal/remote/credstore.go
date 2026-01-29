package remote

import (
	"fmt"
	"os"
	"path/filepath"
)

// CredentialStore persists host credentials on disk.
type CredentialStore struct {
	baseDir string
}

// NewCredentialStore constructs a credential store rooted at baseDir.
func NewCredentialStore(baseDir string) (*CredentialStore, error) {
	if baseDir == "" {
		return nil, fmt.Errorf("credential store: base dir is empty")
	}
	if err := os.MkdirAll(filepath.Join(baseDir, "creds"), 0o755); err != nil {
		return nil, fmt.Errorf("credential store: %w", err)
	}
	return &CredentialStore{baseDir: baseDir}, nil
}

// GetOrCreate returns a credential for the host, creating one if missing.
func (c *CredentialStore) GetOrCreate(hostID string) (Credential, error) {
	if c == nil {
		return Credential{}, fmt.Errorf("credential store: nil")
	}
	path := filepath.Join(c.baseDir, "creds", hostID+".json")
	data, err := os.ReadFile(path)
	if err == nil {
		cred, err := ParseCredential(data)
		if err != nil {
			return Credential{}, fmt.Errorf("credential store: %w", err)
		}
		return cred, nil
	}
	cred, err := NewCredential()
	if err != nil {
		return Credential{}, fmt.Errorf("credential store: %w", err)
	}
	encoded, err := cred.Marshal()
	if err != nil {
		return Credential{}, fmt.Errorf("credential store: %w", err)
	}
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		return Credential{}, fmt.Errorf("credential store: %w", err)
	}
	return cred, nil
}
