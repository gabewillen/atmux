// Package ids implements identifier utilities and normalization functions for the amux project
package ids

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/stateforward/hsm-go/muid"
)

// New generates a new unique ID using muid
func New() muid.MUID {
	id := muid.Make()
	// Ensure the ID is not zero (reserved sentinel value)
	for uint64(id) == 0 {
		id = muid.Make()
	}
	return id
}

// EncodeID encodes an muid.MUID as a base-10 string
func EncodeID(id muid.MUID) string {
	return fmt.Sprintf("%d", uint64(id))
}

// DecodeID decodes a base-10 string to an muid.MUID
func DecodeID(s string) (muid.MUID, error) {
	var id uint64
	_, err := fmt.Sscanf(s, "%d", &id)
	if err != nil {
		return 0, fmt.Errorf("invalid ID format: %w", err)
	}
	return muid.MUID(id), nil
}

// NormalizeAgentSlug normalizes an agent name to a filesystem-safe slug
// according to the spec rules:
// - lowercase
// - non-[a-z0-9-] → '-'
// - collapse multiple consecutive '-' into single '-'
// - trim leading/trailing '-'
// - max 63 chars
func NormalizeAgentSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

	// Replace non-alphanumeric characters with hyphens
	reg := regexp.MustCompile("[^a-z0-9-]+")
	slug = reg.ReplaceAllString(slug, "-")

	// Collapse multiple consecutive hyphens into single hyphen
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")

	// Trim leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	// Limit to 63 characters
	if len(slug) > 63 {
		slug = slug[:63]
		// Ensure we don't end up with trailing hyphen after truncation
		slug = strings.Trim(slug, "-")
	}

	return slug
}

// CanonicalizeRepoRoot canonicalizes a repository root path according to spec rules:
// - expand ~/ to target host's home directory
// - convert to absolute path
// - clean . and .. segments
// - resolve symbolic links to their target path where possible
func CanonicalizeRepoRoot(path string) (string, error) {
	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		path = filepath.Join(homeDir, path[2:])
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to convert to absolute path: %w", err)
	}

	// Clean path (resolve . and .. segments)
	cleanPath := filepath.Clean(absPath)

	// Resolve symbolic links where possible
	// Note: filepath.EvalSymlinks may fail in some environments, so we handle the error
	resolvedPath, err := filepath.EvalSymlinks(cleanPath)
	if err != nil {
		// If EvalSymlinks fails, fall back to the clean path
		// This handles cases where symlinks point to non-existent paths, etc.
		resolvedPath = cleanPath
	}

	return resolvedPath, nil
}