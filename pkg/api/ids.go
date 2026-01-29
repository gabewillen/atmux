// Package api provides public API types for amux.
// ids.go implements identifiers and normalization rules per spec §3, §4.2.3, §5.3.1.
package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/stateforward/hsm-go/muid"
)

// ID is the runtime identifier type for agents, sessions, peers, and hosts.
// It is a 64-bit value (muid-compatible); encoded as base-10 string in JSON per spec §4.2.3.
type ID uint64

// BroadcastID is the reserved sentinel for broadcast addressing (spec §3.22, §6.4).
// Implementations MUST NOT assign 0 as a runtime ID for any entity.
const BroadcastID ID = 0

// MarshalJSON encodes id as a base-10 string per spec §4.2.3.
func (id ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatUint(uint64(id), 10))
}

// UnmarshalJSON decodes id from a base-10 string.
func (id *ID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	dec, err := DecodeID(s)
	if err != nil {
		return err
	}
	*id = dec
	return nil
}

// MaxAgentSlugLen is the maximum length of a normalized agent_slug (spec §5.3.1).
const MaxAgentSlugLen = 63

// DefaultAgentSlug is the slug used when normalization yields an empty string (spec §5.3.1).
const DefaultAgentSlug = "agent"

// ValidRuntimeID returns true if id is non-zero and may be used as a runtime ID.
// The value 0 is reserved and MUST NOT be emitted as an entity ID (spec §3.22).
func ValidRuntimeID(id ID) bool {
	return id != BroadcastID
}

// NextRuntimeID returns a new runtime ID, retrying until non-zero.
// Callers MUST use this (or equivalent) so that 0 is never assigned (spec §3.22).
func NextRuntimeID() ID {
	for {
		id := muid.Make()
		if id != muid.MUID(BroadcastID) {
			return ID(id)
		}
	}
}

// EncodeID returns the base-10 string encoding of id for JSON/wire (spec §4.2.3).
func EncodeID(id ID) string {
	return strconv.FormatUint(uint64(id), 10)
}

// DecodeID parses a base-10 ID string. Returns an error for invalid or empty input.
func DecodeID(s string) (ID, error) {
	if s == "" {
		return 0, fmt.Errorf("empty ID string")
	}
	u, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid ID %q: %w", s, err)
	}
	return ID(u), nil
}

// NormalizeAgentSlug derives a stable, filesystem-safe agent_slug from the agent name (spec §5.3.1).
// Rules: lowercase; replace non-[a-z0-9-] with '-'; collapse consecutive '-'; trim; truncate to 63 chars.
// If the result is empty, returns DefaultAgentSlug ("agent").
func NormalizeAgentSlug(name string) string {
	var b strings.Builder
	lastDash := true
	for _, r := range name {
		if r == '-' || (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') {
			if r == '-' {
				if lastDash {
					continue
				}
				lastDash = true
			} else {
				lastDash = false
			}
			b.WriteRune(r)
			continue
		}
		if r >= 'A' && r <= 'Z' {
			b.WriteRune(unicode.ToLower(r))
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteRune('-')
			lastDash = true
		}
	}
	s := strings.Trim(b.String(), "-")
	if len(s) > MaxAgentSlugLen {
		s = s[:MaxAgentSlugLen]
		s = strings.TrimRight(s, "-")
	}
	if s == "" {
		return DefaultAgentSlug
	}
	return s
}

// UniquifyAgentSlug returns a slug that does not collide with existing.
// If baseSlug is not in existing, returns baseSlug. Otherwise returns baseSlug-2, baseSlug-3, etc.
func UniquifyAgentSlug(baseSlug string, existing map[string]struct{}) string {
	if existing == nil {
		return baseSlug
	}
	if _, ok := existing[baseSlug]; !ok {
		return baseSlug
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d", baseSlug, i)
		if _, ok := existing[candidate]; !ok {
			return candidate
		}
	}
}
