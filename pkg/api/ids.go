package api

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/stateforward/hsm-go/muid"
)

// AgentID is a globally unique identifier for an agent.
// It is serialized as a base-10 string in JSON.
type AgentID muid.MUID

// PeerID is a unique identifier for a peer in the hsmnet.
// It is serialized as a base-10 string in JSON.
type PeerID muid.MUID

// SessionID is a unique identifier for a session.
// It is serialized as a base-10 string in JSON.
type SessionID muid.MUID

// ProcessID is a unique identifier for a process.
// It is serialized as a base-10 string in JSON.
type ProcessID muid.MUID

// HostID is a stable identifier for a host.
type HostID string

// AgentSlug is a filesystem-safe identifier derived from an agent's name.
type AgentSlug string

// RepoRoot is a canonical absolute path to a git repository.
type RepoRoot string

// repoKey represents a stable, session-scoped identifier for a repository.
// Derived from (location.type, location.host, repo_root).
type RepoKey string

// MarshalJSON encodes AgentID as a base-10 string.
func (id AgentID) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatUint(uint64(id), 10))
}

// UnmarshalJSON decodes AgentID from a base-10 string.
func (id *AgentID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	u, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	*id = AgentID(u)
	return nil
}

func (id AgentID) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

// MarshalJSON encodes PeerID as a base-10 string.
func (id PeerID) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatUint(uint64(id), 10))
}

// UnmarshalJSON decodes PeerID from a base-10 string.
func (id *PeerID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	u, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	*id = PeerID(u)
	return nil
}

func (id PeerID) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

// MarshalJSON encodes SessionID as a base-10 string.
func (id SessionID) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatUint(uint64(id), 10))
}

// UnmarshalJSON decodes SessionID from a base-10 string.
func (id *SessionID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	u, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	*id = SessionID(u)
	return nil
}

func (id SessionID) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

// MarshalJSON encodes ProcessID as a base-10 string.
func (id ProcessID) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatUint(uint64(id), 10))
}

// UnmarshalJSON decodes ProcessID from a base-10 string.
func (id *ProcessID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	u, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	*id = ProcessID(u)
	return nil
}

func (id ProcessID) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

// NormalizeAgentSlug derives a stable, filesystem-safe identifier from a name.
// Rules: lowercase, non-[a-z0-9-] -> -, collapse -, trim -, max 63 chars.
func NormalizeAgentSlug(name string) AgentSlug {
	// Lowercase
	s := strings.ToLower(name)

	// Replace non-alphanumeric with hyphen
	reg := regexp.MustCompile(`[^a-z0-9-]+`)
	s = reg.ReplaceAllString(s, "-")

	// Collapse multiple hyphens
	regCollapse := regexp.MustCompile(`-+`)
	s = regCollapse.ReplaceAllString(s, "-")

	// Trim hyphens
	s = strings.Trim(s, "-")

	// Truncate to 63 chars
	if len(s) > 63 {
		s = s[:63]
		// Trim again in case truncation left a trailing hyphen
		s = strings.Trim(s, "-")
	}

	return AgentSlug(s)
}

// Validate checks if the slug is valid.
func (s AgentSlug) Validate() error {
	if len(s) == 0 {
		return fmt.Errorf("slug cannot be empty")
	}
	if len(s) > 63 {
		return fmt.Errorf("slug cannot exceed 63 characters")
	}
	for _, r := range s {
		if !unicode.IsLower(r) && !unicode.IsDigit(r) && r != '-' {
			return fmt.Errorf("slug contains invalid character: %c", r)
		}
	}
	if strings.HasPrefix(string(s), "-") || strings.HasSuffix(string(s), "-") {
		return fmt.Errorf("slug cannot start or end with hyphen")
	}
	return nil
}

// String returns the string representation.
func (s AgentSlug) String() string {
	return string(s)
}

// ParseRepoRoot validates that the path looks like a repo root (simple check).
// Full canonicalization requires filesystem access (internal/paths).
func ParseRepoRoot(path string) (RepoRoot, error) {
	if path == "" {
		return "", fmt.Errorf("repo root cannot be empty")
	}
	// Further validation requires FS
	return RepoRoot(path), nil
}

func (r RepoRoot) String() string {
	return string(r)
}
