package config

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// EncodeTOML encodes a nested map into TOML.
func EncodeTOML(data map[string]any) ([]byte, error) {
	var b strings.Builder
	if err := writeTable(&b, nil, data); err != nil {
		return nil, err
	}
	return []byte(b.String()), nil
}

func writeTable(b *strings.Builder, path []string, table map[string]any) error {
	if table == nil {
		return nil
	}
	if len(path) > 0 {
		b.WriteString("[")
		b.WriteString(strings.Join(path, "."))
		b.WriteString("]\n")
	}
	keys := make([]string, 0, len(table))
	for key := range table {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var nestedTables []string
	var arrayTables []string
	for _, key := range keys {
		value := table[key]
		if isTable(value) {
			nestedTables = append(nestedTables, key)
			continue
		}
		if isArrayTable(value) {
			arrayTables = append(arrayTables, key)
			continue
		}
		formatted, err := formatValue(value)
		if err != nil {
			return fmt.Errorf("encode toml: %s: %w", key, err)
		}
		b.WriteString(key)
		b.WriteString(" = ")
		b.WriteString(formatted)
		b.WriteString("\n")
	}
	for _, key := range nestedTables {
		child, _ := table[key].(map[string]any)
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		if err := writeTable(b, append(path, key), child); err != nil {
			return err
		}
	}
	for _, key := range arrayTables {
		child, _ := table[key].([]any)
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		if err := writeArrayTable(b, append(path, key), child); err != nil {
			return err
		}
	}
	return nil
}

func writeArrayTable(b *strings.Builder, path []string, entries []any) error {
	for i, entry := range entries {
		item, ok := entry.(map[string]any)
		if !ok {
			return fmt.Errorf("array table entry is not a table")
		}
		if i > 0 || b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString("[[")
		b.WriteString(strings.Join(path, "."))
		b.WriteString("]]\n")
		flat := make(map[string]any)
		if err := flattenEntry("", item, flat); err != nil {
			return err
		}
		keys := make([]string, 0, len(flat))
		for key := range flat {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			formatted, err := formatValue(flat[key])
			if err != nil {
				return fmt.Errorf("encode toml: %s: %w", key, err)
			}
			b.WriteString(key)
			b.WriteString(" = ")
			b.WriteString(formatted)
			b.WriteString("\n")
		}
	}
	return nil
}

func flattenEntry(prefix string, value any, out map[string]any) error {
	switch v := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			child := v[key]
			joined := key
			if prefix != "" {
				joined = prefix + "." + key
			}
			if err := flattenEntry(joined, child, out); err != nil {
				return err
			}
		}
	case []any:
		if isArrayTable(v) {
			return fmt.Errorf("nested array tables not supported")
		}
		out[prefix] = v
	default:
		out[prefix] = v
	}
	return nil
}

func formatValue(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return strconv.Quote(v), nil
	case bool:
		if v {
			return "true", nil
		}
		return "false", nil
	case int:
		return strconv.Itoa(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case []any:
		return formatArray(v)
	default:
		return "", fmt.Errorf("unsupported type %T", value)
	}
}

func formatArray(values []any) (string, error) {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		if isTable(value) || isArrayTable(value) {
			return "", fmt.Errorf("array tables not supported in arrays")
		}
		formatted, err := formatValue(value)
		if err != nil {
			return "", err
		}
		parts = append(parts, formatted)
	}
	return "[" + strings.Join(parts, ", ") + "]", nil
}

func isTable(value any) bool {
	_, ok := value.(map[string]any)
	return ok
}

func isArrayTable(value any) bool {
	arr, ok := value.([]any)
	if !ok || len(arr) == 0 {
		return false
	}
	_, ok = arr[0].(map[string]any)
	return ok
}
