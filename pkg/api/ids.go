// Package api defines public types and interfaces for the amux system.
// This package contains the core data structures that are shared across
// the entire amux ecosystem and are stable for external consumption.
package api

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/stateforward/hsm-go/muid"
)

// ID represents a globally unique identifier using muid.MUID.
// All runtime IDs (agents, sessions, processes, peers) are of this type.
// The value 0 is reserved and must not be assigned as a runtime ID.
type ID = muid.MUID

// AgentID represents the runtime ID of an agent instance.
type AgentID = ID

// SessionID represents the runtime ID of a session.
type SessionID = ID

// ProcessID represents the runtime ID of a process.
type ProcessID = ID

// PeerID represents the network peer identifier.
type PeerID = ID

// HostID represents a stable host identifier.
type HostID string

// AgentSlug represents a filesystem-safe identifier derived from agent names.
type AgentSlug string

// RepoRoot represents the canonical absolute path to a git repository root.
type RepoRoot string

// NewID generates a new unique ID. If the generated ID is 0 (reserved),
// it generates a new one until a non-zero value is obtained.
func NewID() ID {
	id := muid.Make()
	for id == 0 {
		id = muid.Make()
	}
	return id
}

// MustParseID parses an ID from a base-10 string and panics on error.
// This should only be used for testing or when the input is guaranteed valid.
func MustParseID(s string) ID {
	id, err := ParseID(s)
	if err != nil {
		panic(fmt.Sprintf("invalid ID %q: %v", s, err))
	}
	return id
}

// ParseID parses an ID from a base-10 string representation.
func ParseID(s string) (ID, error) {
	if s == "" {
		return 0, fmt.Errorf("empty ID string")
	}

	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid ID format %q: %w", s, err)
	}

	id := ID(val)
	if !IDIsValid(id) {
		return 0, fmt.Errorf("ID value 0 is reserved")
	}

	return id, nil
}

// String returns the base-10 string representation of an ID.
func IDToString(id ID) string {
	return strconv.FormatUint(uint64(id), 10)
}

// IsValid checks if an ID is valid (non-zero).
func IDIsValid(id ID) bool {
	return id != 0
}

// NormalizeAgentSlug converts an agent name to a filesystem-safe slug.
// The rules are:
//   - Convert to lowercase
//   - Replace any character not in [a-z0-9-] with -
//   - Collapse consecutive - characters to a single -
//   - Trim leading and trailing -
//   - Truncate to at most 63 characters
//   - If the result is empty, use "agent"
func NormalizeAgentSlug(name string) AgentSlug {
	if name == "" {
		return "agent"
	}

	// Convert to lowercase
	slug := strings.ToLower(name)

	// Replace any character not in [a-z0-9-] with -
	re := regexp.MustCompile(`[^a-z0-9-]`)
	slug = re.ReplaceAllString(slug, "-")

	// Collapse consecutive - characters to a single -
	re = regexp.MustCompile(`-+`)
	slug = re.ReplaceAllString(slug, "-")

	// Trim leading and trailing -
	slug = strings.Trim(slug, "-")

	// Truncate to at most 63 characters
	if len(slug) > 63 {
		slug = slug[:63]
	}

	// If the result is empty, use "agent"
	if slug == "" {
		slug = "agent"
	}

	return AgentSlug(slug)
}

// MakeUniqueAgentSlug ensures a slug is unique by appending numeric suffixes.
// If the base slug is not in the existing set, it returns the base slug.
// Otherwise, it returns base-2, base-3, etc. until a unique slug is found.
func MakeUniqueAgentSlug(base AgentSlug, existing map[AgentSlug]bool) AgentSlug {
	if !existing[base] {
		return base
	}

	for i := 2; i < 1000; i++ {
		slug := AgentSlug(fmt.Sprintf("%s-%d", base, i))
		if !existing[slug] {
			return slug
		}
	}

	// This should never happen in practice, but handle gracefully
	suffix := fmt.Sprintf("-%d", muid.Make())
	return AgentSlug(fmt.Sprintf("%s%s", base, suffix))
}

// CanonicalizeRepoRoot canonicalizes a repository path according to spec §3.23.
// For local paths, it:
//   - Expands ~ to the user's home directory
//   - Converts to an absolute path
//   - Cleans . and .. segments
//   - Resolves symbolic links where possible
func CanonicalizeRepoRoot(path string) (RepoRoot, error) {
	if path == "" {
		return "", fmt.Errorf("empty repository path")
	}

	// Expand ~ to home directory
	expanded, err := expandHome(path)
	if err != nil {
		return "", fmt.Errorf("failed to expand home directory: %w", err)
	}

	// Convert to absolute path
	abs, err := filepath.Abs(expanded)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Clean the path
	clean := filepath.Clean(abs)

	// Resolve symbolic links where possible
	resolved, err := filepath.EvalSymlinks(clean)
	if err != nil {
		// If symlink resolution fails, use the cleaned path as canonical
		resolved = clean
	}

	return RepoRoot(resolved), nil
}

// expandHome expands ~ to the user's home directory.
func expandHome(path string) (string, error) {
	if !strings.HasPrefix(path, "~/") && path != "~" {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	if path == "~" {
		return home, nil
	}

	return filepath.Join(home, path[2:]), nil
}

// String returns the string representation of various types.
func (h HostID) String() string {
	return string(h)
}

func (s AgentSlug) String() string {
	return string(s)
}

func (r RepoRoot) String() string {
	return string(r)
}
