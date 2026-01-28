package config

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// ParseTOML parses a minimal TOML subset into a nested map.
func ParseTOML(data []byte) (map[string]any, error) {
	root := make(map[string]any)
	current := root
	lines := strings.Split(string(data), "\n")
	for i, raw := range lines {
		line := stripTOMLComment(raw)
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[[") {
			path, err := parseTablePath(line, 2)
			if err != nil {
				return nil, fmt.Errorf("parse toml: line %d: %w", i+1, err)
			}
			arr, err := getOrCreateArrayTable(root, path)
			if err != nil {
				return nil, fmt.Errorf("parse toml: line %d: %w", i+1, err)
			}
			entry := make(map[string]any)
			arr = append(arr, entry)
			if err := setPath(root, path, arr); err != nil {
				return nil, fmt.Errorf("parse toml: line %d: %w", i+1, err)
			}
			current = entry
			continue
		}
		if strings.HasPrefix(line, "[") {
			path, err := parseTablePath(line, 1)
			if err != nil {
				return nil, fmt.Errorf("parse toml: line %d: %w", i+1, err)
			}
			current, err = getOrCreateTable(root, path)
			if err != nil {
				return nil, fmt.Errorf("parse toml: line %d: %w", i+1, err)
			}
			continue
		}
		key, valueRaw, err := splitKeyValue(line)
		if err != nil {
			return nil, fmt.Errorf("parse toml: line %d: %w", i+1, err)
		}
		value, err := parseValue(valueRaw)
		if err != nil {
			return nil, fmt.Errorf("parse toml: line %d: %w", i+1, err)
		}
		if err := setKey(current, key, value); err != nil {
			return nil, fmt.Errorf("parse toml: line %d: %w", i+1, err)
		}
	}
	return root, nil
}

func stripTOMLComment(line string) string {
	inString := false
	var b strings.Builder
	for i := 0; i < len(line); i++ {
		ch := line[i]
		if ch == '"' {
			if i == 0 || line[i-1] != '\\' {
				inString = !inString
			}
		}
		if ch == '#' && !inString {
			break
		}
		b.WriteByte(ch)
	}
	return b.String()
}

func parseTablePath(line string, brackets int) ([]string, error) {
	end := strings.LastIndex(line, "]")
	if end == -1 {
		return nil, fmt.Errorf("table missing closing bracket")
	}
	content := strings.TrimSpace(line[brackets : end-(brackets-1)])
	if content == "" {
		return nil, fmt.Errorf("empty table name")
	}
	parts := strings.Split(content, ".")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
		if parts[i] == "" {
			return nil, fmt.Errorf("invalid table path")
		}
	}
	return parts, nil
}

func splitKeyValue(line string) (string, string, error) {
	idx := strings.Index(line, "=")
	if idx == -1 {
		return "", "", fmt.Errorf("missing '='")
	}
	key := strings.TrimSpace(line[:idx])
	value := strings.TrimSpace(line[idx+1:])
	if key == "" {
		return "", "", fmt.Errorf("empty key")
	}
	if value == "" {
		return "", "", fmt.Errorf("empty value")
	}
	return key, value, nil
}

func parseValue(raw string) (any, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("empty value")
	}
	if strings.HasPrefix(raw, "\"") && strings.HasSuffix(raw, "\"") {
		value, err := strconv.Unquote(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid string: %w", err)
		}
		return value, nil
	}
	if strings.HasPrefix(raw, "'") && strings.HasSuffix(raw, "'") {
		value := strings.Trim(raw, "'")
		return value, nil
	}
	if raw == "true" || raw == "false" {
		return raw == "true", nil
	}
	if strings.HasPrefix(raw, "[") && strings.HasSuffix(raw, "]") {
		items, err := parseArray(raw)
		if err != nil {
			return nil, err
		}
		return items, nil
	}
	if strings.HasPrefix(raw, "{") {
		return nil, fmt.Errorf("inline tables not supported")
	}
	if i, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return i, nil
	}
	if f, err := strconv.ParseFloat(raw, 64); err == nil {
		return f, nil
	}
	return raw, nil
}

func parseArray(raw string) ([]any, error) {
	trimmed := strings.TrimSpace(raw[1 : len(raw)-1])
	if trimmed == "" {
		return []any{}, nil
	}
	var (
		items []any
		buf   bytes.Buffer
		inStr bool
	)
	for i := 0; i < len(trimmed); i++ {
		ch := trimmed[i]
		if ch == '"' {
			if i == 0 || trimmed[i-1] != '\\' {
				inStr = !inStr
			}
		}
		if ch == ',' && !inStr {
			item := strings.TrimSpace(buf.String())
			buf.Reset()
			if item == "" {
				return nil, fmt.Errorf("empty array item")
			}
			parsed, err := parseValue(item)
			if err != nil {
				return nil, err
			}
			items = append(items, parsed)
			continue
		}
		buf.WriteByte(ch)
	}
	last := strings.TrimSpace(buf.String())
	if last != "" {
		parsed, err := parseValue(last)
		if err != nil {
			return nil, err
		}
		items = append(items, parsed)
	}
	return items, nil
}

func getOrCreateTable(root map[string]any, path []string) (map[string]any, error) {
	current := root
	for _, part := range path {
		nextRaw, ok := current[part]
		if !ok {
			next := make(map[string]any)
			current[part] = next
			current = next
			continue
		}
		next, ok := nextRaw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("table conflict at %s", part)
		}
		current = next
	}
	return current, nil
}

func getOrCreateArrayTable(root map[string]any, path []string) ([]any, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("empty array table path")
	}
	parentPath := path[:len(path)-1]
	key := path[len(path)-1]
	parent, err := getOrCreateTable(root, parentPath)
	if err != nil {
		return nil, err
	}
	if existing, ok := parent[key]; ok {
		arr, ok := existing.([]any)
		if !ok {
			return nil, fmt.Errorf("array table conflict at %s", key)
		}
		return arr, nil
	}
	arr := []any{}
	parent[key] = arr
	return arr, nil
}

func setKey(current map[string]any, key string, value any) error {
	parts := strings.Split(key, ".")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	if len(parts) == 1 {
		current[parts[0]] = value
		return nil
	}
	nested, err := getOrCreateTable(current, parts[:len(parts)-1])
	if err != nil {
		return err
	}
	nested[parts[len(parts)-1]] = value
	return nil
}

func setPath(root map[string]any, path []string, value any) error {
	if len(path) == 0 {
		return fmt.Errorf("empty path")
	}
	current := root
	for _, part := range path[:len(path)-1] {
		nextRaw, ok := current[part]
		if !ok {
			next := make(map[string]any)
			current[part] = next
			current = next
			continue
		}
		next, ok := nextRaw.(map[string]any)
		if !ok {
			return fmt.Errorf("path conflict at %s", part)
		}
		current = next
	}
	current[path[len(path)-1]] = value
	return nil
}
