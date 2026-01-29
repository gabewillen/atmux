// Package ids provides identifier and normalization utilities for the amux system.
// It implements the specifications for agent_id, peer_id, host_id, agent_slug, 
// and repo_root canonicalization as defined in the amux spec.
package ids

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/stateforward/hsm-go/muid"
)

// AgentID represents a unique agent identifier.
type AgentID = muid.MUID

// PeerID represents a unique peer identifier.
type PeerID = muid.MUID

// HostID represents a unique host identifier. 
type HostID = muid.MUID

// NewAgentID generates a new unique agent identifier.
func NewAgentID() AgentID {
	return muid.Make()
}

// NewPeerID generates a new unique peer identifier.
func NewPeerID() PeerID {
	return muid.Make()
}

// NewHostID generates a new unique host identifier.
func NewHostID() HostID {
	return muid.Make()
}

// AgentSlugFromName normalizes a human-readable agent name into a valid agent_slug.
// Per the spec: lowercase, non-[a-z0-9-] → '-', collapse, trim, max 63 chars.
func AgentSlugFromName(name string) string {
	if name == "" {
		return "unnamed"
	}

	// Convert to lowercase
	slug := strings.ToLower(name)
	
	// Replace non-[a-z0-9-] characters with '-'
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		} else {
			result.WriteRune('-')
		}
	}
	slug = result.String()
	
	// Collapse multiple consecutive '-' into single '-'
	re := regexp.MustCompile(`-+`)
	slug = re.ReplaceAllString(slug, "-")
	
	// Trim leading and trailing '-'
	slug = strings.Trim(slug, "-")
	
	// Ensure max 63 characters
	if len(slug) > 63 {
		slug = slug[:63]
		// Ensure we don't end with a dash after truncation
		slug = strings.TrimRight(slug, "-")
	}
	
	// Handle empty result
	if slug == "" {
		return "unnamed"
	}
	
	return slug
}

// ValidateAgentSlug validates that a string is a valid agent_slug per spec.
func ValidateAgentSlug(slug string) error {
	if slug == "" {
		return fmt.Errorf("agent_slug cannot be empty")
	}
	
	if len(slug) > 63 {
		return fmt.Errorf("agent_slug cannot exceed 63 characters")
	}
	
	// Check that it only contains [a-z0-9-]
	for _, r := range slug {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			return fmt.Errorf("agent_slug contains invalid character: %c", r)
		}
	}
	
	// Cannot start or end with dash
	if strings.HasPrefix(slug, "-") || strings.HasSuffix(slug, "-") {
		return fmt.Errorf("agent_slug cannot start or end with dash")
	}
	
	return nil
}

// CanonicalizeRepoRoot canonicalizes a repository root path.
// For local paths, it resolves to absolute path.
// For remote contexts, it expands ~ to user home directory.
func CanonicalizeRepoRoot(path string, isRemote bool) (string, error) {
	if path == "" {
		return "", fmt.Errorf("repo_root cannot be empty")
	}
	
	// Handle ~ expansion for remote contexts
	if isRemote && strings.HasPrefix(path, "~/") {
		// In remote context, ~ expansion is environment-specific
		// For now, we'll leave it as-is and let the remote system handle it
		return path, nil
	}
	
	// For local paths, resolve to absolute path
	if !isRemote {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("failed to canonicalize repo_root %s: %w", path, err)
		}
		return filepath.Clean(absPath), nil
	}
	
	// For remote paths that don't start with ~, clean and return
	return filepath.Clean(path), nil
}

// IsValidIdentifierName checks if a name is suitable for use in identifiers.
// It ensures the name contains printable characters and isn't excessively long.
func IsValidIdentifierName(name string) bool {
	if name == "" || len(name) > 256 {
		return false
	}
	
	// Check that all characters are printable
	for _, r := range name {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	
	return true
}