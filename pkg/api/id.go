package api

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/stateforward/hsm-go/muid"
)

// AgentID is a globally unique identifier for an agent instance.
type AgentID = muid.MUID

// SessionID is a globally unique identifier for an amux session.
type SessionID = muid.MUID

// HostID is a globally unique identifier for a host running amux.
type HostID = muid.MUID

// AgentSlug is a normalized, filesystem-safe string derived from an agent's name.
type AgentSlug string

var (
	nonAlphanumeric = regexp.MustCompile(`[^a-z0-9-]+`)
	multipleDashes  = regexp.MustCompile(`-+`)
)

// NewAgentSlug creates a normalized AgentSlug from a raw name.
// Rules: lowercase, replace non-alphanumeric with dash, collapse dashes, trim dashes, max 63 chars.
func NewAgentSlug(name string) AgentSlug {
	s := strings.ToLower(name)
	s = nonAlphanumeric.ReplaceAllString(s, "-")
	s = multipleDashes.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 63 {
		s = s[:63]
		s = strings.TrimSuffix(s, "-") // Ensure we don't end with a dash after truncation
	}
	if s == "" {
		// Fallback for empty or all-invalid strings to avoid empty slugs
		return "agent"
	}
	return AgentSlug(s)
}

func (s AgentSlug) String() string {
	return string(s)
}

// EncodeID returns the base-10 string representation of an ID.
// This matches the spec requirement for JSON encoding.
func EncodeID(id muid.MUID) string {
	return strconv.FormatUint(uint64(id), 10)
}

// ParseID parses a base-10 string into an ID.
func ParseID(s string) (muid.MUID, error) {
	u, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return muid.MUID(u), nil
}

const (
	// ReservedID is the sentinel value 0, which MUST NOT be used as a runtime ID.
	ReservedID muid.MUID = 0
)
