// Package api provides public types for the amux system.
package api

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/stateforward/hsm-go/muid"
)

// BroadcastID is the reserved ID value (0) for broadcast messages to all participants.
// Per spec §3.22, implementations SHALL NOT assign 0 as a runtime ID for any agent,
// process, session, peer, or message.
const BroadcastID muid.MUID = 0

// NormalizeAgentSlug derives a stable, filesystem-safe identifier from an agent name.
// Per spec §5.3.1:
//   - Convert to lowercase
//   - Replace any character not in [a-z0-9-] with -
//   - Collapse consecutive - characters to a single -
//   - Trim leading and trailing -
//   - Truncate to at most 63 characters
//   - If the result is empty, use "agent"
func NormalizeAgentSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

	// Replace any character not in [a-z0-9-] with -
	re := regexp.MustCompile(`[^a-z0-9-]`)
	slug = re.ReplaceAllString(slug, "-")

	// Collapse consecutive - characters to a single -
	collapseRe := regexp.MustCompile(`-+`)
	slug = collapseRe.ReplaceAllString(slug, "-")

	// Trim leading and trailing -
	slug = strings.Trim(slug, "-")

	// Truncate to at most 63 characters
	if len(slug) > 63 {
		slug = slug[:63]
		// Re-trim in case truncation left a trailing -
		slug = strings.TrimRight(slug, "-")
	}

	// If the result is empty, use "agent"
	if slug == "" {
		slug = "agent"
	}

	return slug
}

// CanonicalizeRepoRoot canonicalizes a repository root path.
// Per spec §3.23:
//   - Expand ~/ to the target host's home directory
//   - Convert to an absolute path
//   - Clean ./.. segments
//   - Resolve symbolic links to their target path where possible
//
// If symbolic link resolution is not possible (insufficient permissions or missing OS support),
// this function still applies (a)-(c) and treats the result as canonical.
func CanonicalizeRepoRoot(repoPath string) (string, error) {
	// Expand ~/ to home directory
	if strings.HasPrefix(repoPath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("expand home directory: %w", err)
		}
		repoPath = filepath.Join(homeDir, repoPath[2:])
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return "", fmt.Errorf("convert to absolute path: %w", err)
	}

	// Clean ./.. segments
	cleanPath := filepath.Clean(absPath)

	// Resolve symbolic links to their target path where possible
	// If resolution fails (insufficient permissions or missing OS support),
	// we still return the cleaned path as canonical per spec.
	evalPath, err := filepath.EvalSymlinks(cleanPath)
	if err != nil {
		// Per spec: if symbolic link resolution is not possible,
		// treat the cleaned path as canonical
		return cleanPath, nil
	}

	return evalPath, nil
}

// GenerateID generates a new muid.MUID and ensures it is not the reserved value 0.
// Per spec §3.22: If an ID generator produces 0, the implementation SHALL generate a new ID.
func GenerateID() muid.MUID {
	for {
		id := muid.Make()
		if id != BroadcastID {
			return id
		}
		// Retry if we got the reserved value 0
	}
}
