package config

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/agentflare-ai/amux/internal/paths"
)

var byteSizePattern = regexp.MustCompile(`^(\d+)(B|KB|MB|GB)?$`)

// ParseByteSize parses a byte size string or integer.
func ParseByteSize(raw string) (ByteSize, error) {
	match := byteSizePattern.FindStringSubmatch(strings.TrimSpace(raw))
	if match == nil {
		return 0, fmt.Errorf("invalid byte size: %s", raw)
	}
	value, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid byte size: %w", err)
	}
	unit := match[2]
	switch unit {
	case "", "B":
		return ByteSize(value), nil
	case "KB":
		return ByteSize(value * 1024), nil
	case "MB":
		return ByteSize(value * 1024 * 1024), nil
	case "GB":
		return ByteSize(value * 1024 * 1024 * 1024), nil
	default:
		return 0, fmt.Errorf("invalid byte size unit: %s", unit)
	}
}

// ParseByteSizeValue parses a byte size from an interface value.
func ParseByteSizeValue(value any) (ByteSize, error) {
	switch v := value.(type) {
	case int64:
		return ByteSize(v), nil
	case int:
		return ByteSize(v), nil
	case float64:
		return ByteSize(int64(v)), nil
	case string:
		return ParseByteSize(v)
	default:
		return 0, fmt.Errorf("invalid byte size type: %T", value)
	}
}

func parseDurationValue(value any) (time.Duration, error) {
	switch v := value.(type) {
	case time.Duration:
		return v, nil
	case string:
		parsed, err := time.ParseDuration(v)
		if err != nil {
			return 0, fmt.Errorf("invalid duration: %w", err)
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("invalid duration type: %T", value)
	}
}

func parseString(value any) (string, bool) {
	str, ok := value.(string)
	return str, ok
}

func parseBool(value any) (bool, bool) {
	b, ok := value.(bool)
	return b, ok
}

func parseInt(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

func expandPath(resolver *paths.Resolver, value string) string {
	if resolver == nil {
		return value
	}
	return resolver.ExpandHome(value)
}
