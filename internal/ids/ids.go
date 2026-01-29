// Package ids provides identifier utilities for amux.
//
// This package implements ID generation, validation, and normalization
// following the spec requirements for agent_slug (§5.3.1), repo_key (§3.24),
// and runtime IDs (§3.21, §3.22).
//
// All runtime IDs use muid.MUID from hsm-go. The value 0 is reserved
// as a sentinel (e.g., BroadcastID) and SHALL NOT be assigned as a
// runtime ID for agents, processes, sessions, or peers.
package ids

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/stateforward/hsm-go/muid"

	"github.com/agentflare-ai/amux/pkg/api"
)

// BroadcastID is the sentinel ID value (0) used for broadcast messages.
// See spec §3.22.
const BroadcastID muid.MUID = 0

// MaxAgentSlugLength is the maximum length for agent_slug values.
// See spec §5.3.1.
const MaxAgentSlugLength = 63

// DefaultAgentSlug is used when normalization produces an empty string.
const DefaultAgentSlug = "agent"

// allowedChars matches characters that don't need replacement in agent_slug.
var allowedChars = regexp.MustCompile(`[^a-z0-9-]`)

// consecutiveDashes matches runs of consecutive dashes.
var consecutiveDashes = regexp.MustCompile(`-+`)

// NewID generates a new globally unique runtime ID.
// It will never return 0 (the reserved sentinel value).
func NewID() muid.MUID {
	for {
		id := muid.Make()
		if id != 0 {
			return id
		}
		// Extremely unlikely to hit this path, but if muid.Make()
		// produces 0, we regenerate per spec §3.22.
	}
}

// IsValidRuntimeID returns true if the ID is valid for runtime use.
// Runtime IDs must be non-zero per spec §3.22.
func IsValidRuntimeID(id muid.MUID) bool {
	return id != 0
}

// NormalizeAgentSlug normalizes an agent name into a filesystem-safe slug.
// The normalization follows spec §5.3.1:
//
//   - Convert to lowercase
//   - Replace any character not in [a-z0-9-] with "-"
//   - Collapse consecutive "-" characters to a single "-"
//   - Trim leading and trailing "-"
//   - Truncate to at most 63 characters
//   - If the result is empty, use "agent"
func NormalizeAgentSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

	// Replace any character not in [a-z0-9-] with "-"
	slug = allowedChars.ReplaceAllString(slug, "-")

	// Collapse consecutive "-" characters to a single "-"
	slug = consecutiveDashes.ReplaceAllString(slug, "-")

	// Trim leading and trailing "-"
	slug = strings.Trim(slug, "-")

	// Truncate to at most 63 characters
	if len(slug) > MaxAgentSlugLength {
		slug = slug[:MaxAgentSlugLength]
		// Re-trim in case truncation created a trailing dash
		slug = strings.TrimRight(slug, "-")
	}

	// If the result is empty, use "agent"
	if slug == "" {
		slug = DefaultAgentSlug
	}

	return slug
}

// UniqueAgentSlug returns a unique agent_slug by appending numeric suffixes
// if needed. The exists function should return true if a slug is already in use.
// See spec §5.3.1 for collision handling.
func UniqueAgentSlug(name string, exists func(slug string) bool) string {
	slug := NormalizeAgentSlug(name)

	if !exists(slug) {
		return slug
	}

	// Append numeric suffixes -2, -3, ... until unique
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d", slug, i)
		// Ensure the suffix doesn't exceed max length
		if len(candidate) > MaxAgentSlugLength {
			// Truncate the base slug to make room for suffix
			baseLen := MaxAgentSlugLength - len(fmt.Sprintf("-%d", i))
			if baseLen <= 0 {
				// Extremely unlikely: suffix alone exceeds max length
				candidate = fmt.Sprintf("%d", i)[:MaxAgentSlugLength]
			} else {
				candidate = fmt.Sprintf("%s-%d", slug[:baseLen], i)
			}
		}
		if !exists(candidate) {
			return candidate
		}
	}
}

// RepoKey computes the stable repository key from location and repo_root.
// See spec §3.24: repo_key is derived from (location.type, location.host, repo_root).
//
// For local agents, the key is "local:<repo_root>".
// For SSH agents, the key is "ssh:<host>:<repo_root>".
//
// The repo_root should already be canonicalized per spec §3.23.
func RepoKey(location api.Location, repoRoot string) string {
	switch location.Type {
	case api.LocationLocal:
		return fmt.Sprintf("local:%s", repoRoot)
	case api.LocationSSH:
		return fmt.Sprintf("ssh:%s:%s", location.Host, repoRoot)
	default:
		// Treat unknown types as local for safety
		return fmt.Sprintf("unknown:%s", repoRoot)
	}
}

// RepoKeyHash returns a truncated SHA-256 hash of the repo_key.
// This can be used when a shorter identifier is needed.
func RepoKeyHash(location api.Location, repoRoot string) string {
	key := RepoKey(location, repoRoot)
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])[:16]
}

// EncodeID encodes a muid.MUID as a base-10 string for JSON wire format.
// See spec §9.1.3.1.
func EncodeID(id muid.MUID) string {
	return fmt.Sprintf("%d", uint64(id))
}

// DecodeID decodes a base-10 string to a muid.MUID.
// Returns an error if the string is not a valid base-10 unsigned integer.
func DecodeID(s string) (muid.MUID, error) {
	if s == "" {
		return 0, fmt.Errorf("invalid ID format: empty string")
	}

	// Validate that the string contains only digits
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid ID format: contains non-digit character")
		}
	}

	var id uint64
	_, err := fmt.Sscanf(s, "%d", &id)
	if err != nil {
		return 0, fmt.Errorf("invalid ID format: %w", err)
	}
	return muid.MUID(id), nil
}

// EncodeIDs encodes a slice of muid.MUID values as base-10 strings.
func EncodeIDs(ids []muid.MUID) []string {
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = EncodeID(id)
	}
	return result
}

// DecodeIDs decodes a slice of base-10 strings to muid.MUID values.
// Returns an error if any string is invalid.
func DecodeIDs(strs []string) ([]muid.MUID, error) {
	result := make([]muid.MUID, len(strs))
	for i, s := range strs {
		id, err := DecodeID(s)
		if err != nil {
			return nil, fmt.Errorf("invalid ID at index %d: %w", i, err)
		}
		result[i] = id
	}
	return result, nil
}
