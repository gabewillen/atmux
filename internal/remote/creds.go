package remote

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
)

// Credential holds the per-host auth token.
type Credential struct {
	Token string `json:"token"`
}

// NewCredential generates a new credential.
func NewCredential() (Credential, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return Credential{}, fmt.Errorf("new credential: %w", err)
	}
	return Credential{Token: base64.StdEncoding.EncodeToString(buf)}, nil
}

// Marshal serializes the credential to JSON.
func (c Credential) Marshal() ([]byte, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("marshal credential: %w", err)
	}
	return data, nil
}

// ParseCredential decodes a credential from JSON.
func ParseCredential(data []byte) (Credential, error) {
	var cred Credential
	if err := json.Unmarshal(data, &cred); err != nil {
		return Credential{}, fmt.Errorf("parse credential: %w", err)
	}
	if cred.Token == "" {
		return Credential{}, fmt.Errorf("parse credential: %w", ErrInvalidMessage)
	}
	return cred, nil
}

// LoadCredential reads a credential from disk.
func LoadCredential(path string) (Credential, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Credential{}, fmt.Errorf("load credential: %w", err)
	}
	cred, err := ParseCredential(data)
	if err != nil {
		return Credential{}, fmt.Errorf("load credential: %w", err)
	}
	return cred, nil
}
