package config

import (
	"strings"
	"testing"
	"time"
)

func TestParseKeyPathQuotedSegments(t *testing.T) {
	parts, err := parseKeyPath(`root."a.b".'c.d'`)
	if err != nil {
		t.Fatalf("parse key path: %v", err)
	}
	if len(parts) != 3 || parts[1] != "a.b" || parts[2] != "c.d" {
		t.Fatalf("unexpected key parts: %#v", parts)
	}
}

func TestSplitKeyValueIgnoresQuotedEquals(t *testing.T) {
	key, value, err := splitKeyValue(`alpha = "a=b"`)
	if err != nil {
		t.Fatalf("split key/value: %v", err)
	}
	if key != "alpha" || value != `"a=b"` {
		t.Fatalf("unexpected split: %q %q", key, value)
	}
}

func TestParseValueCoverage(t *testing.T) {
	if value, err := parseValue("true"); err != nil || value != true {
		t.Fatalf("parse bool: %v %v", value, err)
	}
	if value, err := parseValue("inf"); err != nil || value.(float64) <= 0 {
		t.Fatalf("parse inf: %v %v", value, err)
	}
	if value, err := parseValue("nan"); err != nil {
		t.Fatalf("parse nan: %v", err)
	} else if _, ok := value.(float64); !ok {
		t.Fatalf("expected float nan")
	}
	if value, err := parseValue("2024-01-02T03:04:05Z"); err != nil {
		t.Fatalf("parse datetime: %v", err)
	} else if _, ok := value.(time.Time); !ok {
		t.Fatalf("expected time value")
	}
	if value, err := parseValue("15:04:05.999"); err != nil {
		t.Fatalf("parse time: %v", err)
	} else if _, ok := value.(time.Time); !ok {
		t.Fatalf("expected time value")
	}
	if value, err := parseValue("1_024.5"); err != nil || value.(float64) != 1024.5 {
		t.Fatalf("parse float: %v %v", value, err)
	}
	if value, err := parseValue("0x10"); err != nil || value.(int64) != 16 {
		t.Fatalf("parse hex: %v %v", value, err)
	}
	if value, err := parseValue("[1, 2, 3]"); err != nil || len(value.([]any)) != 3 {
		t.Fatalf("parse array: %v %v", value, err)
	}
	if value, err := parseValue(`{ foo.bar = "baz" }`); err != nil {
		t.Fatalf("parse inline table: %v", err)
	} else {
		table := value.(map[string]any)
		nested := table["foo"].(map[string]any)
		if nested["bar"] != "baz" {
			t.Fatalf("unexpected inline table value: %#v", table)
		}
	}
}

func TestParseValueErrors(t *testing.T) {
	if _, err := parseValue(`"bad\x"`); err == nil {
		t.Fatalf("expected invalid escape error")
	}
	if _, err := parseValue(`"\u12"`); err == nil {
		t.Fatalf("expected unicode error")
	}
	if _, err := parseValue("[1 2]"); err == nil {
		t.Fatalf("expected array separator error")
	}
	if _, err := parseValue(`{ a = 1 b = 2 }`); err == nil {
		t.Fatalf("expected inline table separator error")
	}
	if _, err := parseValue(`"unterminated`); err == nil {
		t.Fatalf("expected unterminated string error")
	}
}

func TestParseTablePathErrors(t *testing.T) {
	if _, err := parseTablePath("[invalid", 1); err == nil {
		t.Fatalf("expected missing bracket error")
	}
	if _, err := parseTablePath("[]", 1); err == nil {
		t.Fatalf("expected empty table name error")
	}
}

func TestParseStringValueErrors(t *testing.T) {
	if _, err := parseStringValue(`"ok" trailing`); err == nil {
		t.Fatalf("expected trailing data error")
	}
	if _, err := parseStringValue(`123`); err == nil {
		t.Fatalf("expected invalid string error")
	}
}

func TestSetInlineKeyDuplicates(t *testing.T) {
	root := map[string]any{"a": 1}
	if err := setInlineKey(root, "a", 2); err == nil {
		t.Fatalf("expected duplicate key error")
	}
}

func TestGetOrCreateConflicts(t *testing.T) {
	root := map[string]any{"a": 1}
	if _, err := getOrCreateTable(root, []string{"a"}); err == nil {
		t.Fatalf("expected table conflict error")
	}
	root = map[string]any{"a": map[string]any{"b": 1}}
	if _, err := getOrCreateArrayTable(root, []string{"a", "b"}); err == nil {
		t.Fatalf("expected array table conflict error")
	}
}

func TestSetPathErrors(t *testing.T) {
	if err := setPath(map[string]any{}, nil, 1); err == nil {
		t.Fatalf("expected empty path error")
	}
	root := map[string]any{"a": 1}
	if err := setPath(root, []string{"a", "b"}, 2); err == nil {
		t.Fatalf("expected path conflict error")
	}
}

func TestParseTOMLComplexStructures(t *testing.T) {
	doc := `
title = "example" # comment
[servers."alpha.beta"]
ip = "127.0.0.1"
ports = [8000, 8001]
[[clients]]
name = 'alpha'
[[clients]]
name = "beta"
data = """line1
line2"""
[inline]
kv = { a = 1, "b.c" = 2 }
`
	parsed, err := ParseTOML([]byte(doc))
	if err != nil {
		t.Fatalf("parse toml: %v", err)
	}
	if parsed["title"] != "example" {
		t.Fatalf("unexpected title")
	}
	servers := parsed["servers"].(map[string]any)
	ab := servers["alpha.beta"].(map[string]any)
	if ab["ip"] != "127.0.0.1" {
		t.Fatalf("unexpected ip")
	}
	clients := parsed["clients"].([]any)
	if len(clients) != 2 {
		t.Fatalf("unexpected clients")
	}
	inline := parsed["inline"].(map[string]any)
	kv := inline["kv"].(map[string]any)
	if kv["a"].(int64) != 1 {
		t.Fatalf("unexpected inline a")
	}
	b := kv["b"].(map[string]any)
	if b["c"].(int64) != 2 {
		t.Fatalf("unexpected inline b.c")
	}
}

func TestParseValueNumbersAndTimes(t *testing.T) {
	cases := []string{
		"0b1010",
		"0o77",
		"0xFF",
		"1.25e2",
		"2006-01-02",
		"15:04:05",
	}
	for _, raw := range cases {
		if _, err := parseValue(raw); err != nil {
			t.Fatalf("parse value %q: %v", raw, err)
		}
	}
}

func TestSplitOnEqualsComplex(t *testing.T) {
	cases := []struct {
		line string
		want string
	}{
		{`key = "value=with=equals"`, `key`},
		{`key = 'value=with=equals'`, `key`},
		{`key = [1, 2, "a=b"]`, `key`},
		{`key = { a = 1, b = "c=d" }`, `key`},
		{"key = \"\"\"multi=basic\"\"\"", "key"},
		{"key = '''multi=literal'''", "key"},
	}
	for _, tc := range cases {
		idx := splitOnEquals(tc.line)
		if idx == -1 {
			t.Fatalf("expected '=' for %q", tc.line)
		}
		if got := strings.TrimSpace(tc.line[:idx]); got != tc.want {
			t.Fatalf("unexpected key for %q: %q", tc.line, got)
		}
	}
}

func TestParseValueMultiLineStrings(t *testing.T) {
	value, err := parseValue("\"\"\"line1\\\nline2\"\"\"")
	if err != nil {
		t.Fatalf("parse multiline string: %v", err)
	}
	if value.(string) != "line1line2" {
		t.Fatalf("unexpected multiline result: %q", value.(string))
	}
	value, err = parseValue("'''line1\nline2'''")
	if err != nil {
		t.Fatalf("parse multiline literal: %v", err)
	}
	if value.(string) != "line1\nline2" {
		t.Fatalf("unexpected multiline literal result: %q", value.(string))
	}
}
