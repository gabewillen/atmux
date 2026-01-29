package remote

import "fmt"

// Credential holds the per-host NATS credential file bytes.
type Credential struct {
	data []byte
}

// ParseCredential validates and wraps credential bytes.
func ParseCredential(data []byte) (Credential, error) {
	if len(data) == 0 {
		return Credential{}, fmt.Errorf("parse credential: %w", ErrInvalidMessage)
	}
	return Credential{data: append([]byte(nil), data...)}, nil
}

// Marshal returns the credential bytes.
func (c Credential) Marshal() ([]byte, error) {
	if len(c.data) == 0 {
		return nil, fmt.Errorf("marshal credential: %w", ErrInvalidMessage)
	}
	return append([]byte(nil), c.data...), nil
}
