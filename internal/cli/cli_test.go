package cli

import (
	"testing"
)

func TestParseFlagsEmpty(t *testing.T) {
	flags, positional := ParseFlags([]string{})

	if len(flags) != 0 {
		t.Errorf("flags len = %d, want 0", len(flags))
	}
	if len(positional) != 0 {
		t.Errorf("positional len = %d, want 0", len(positional))
	}
}

func TestParseFlagsPositionalOnly(t *testing.T) {
	flags, positional := ParseFlags([]string{"add", "my-agent"})

	if len(flags) != 0 {
		t.Errorf("flags len = %d, want 0", len(flags))
	}
	if len(positional) != 2 {
		t.Fatalf("positional len = %d, want 2", len(positional))
	}
	if positional[0] != "add" {
		t.Errorf("positional[0] = %q, want %q", positional[0], "add")
	}
	if positional[1] != "my-agent" {
		t.Errorf("positional[1] = %q, want %q", positional[1], "my-agent")
	}
}

func TestParseFlagsLongWithValue(t *testing.T) {
	flags, positional := ParseFlags([]string{"--adapter", "claude-code", "--repo", "/path/to/repo"})

	if len(positional) != 0 {
		t.Errorf("positional len = %d, want 0", len(positional))
	}
	if flags["adapter"] != "claude-code" {
		t.Errorf("flags[adapter] = %q, want %q", flags["adapter"], "claude-code")
	}
	if flags["repo"] != "/path/to/repo" {
		t.Errorf("flags[repo] = %q, want %q", flags["repo"], "/path/to/repo")
	}
}

func TestParseFlagsLongWithEquals(t *testing.T) {
	flags, _ := ParseFlags([]string{"--adapter=claude-code", "--repo=/path/to/repo"})

	if flags["adapter"] != "claude-code" {
		t.Errorf("flags[adapter] = %q, want %q", flags["adapter"], "claude-code")
	}
	if flags["repo"] != "/path/to/repo" {
		t.Errorf("flags[repo] = %q, want %q", flags["repo"], "/path/to/repo")
	}
}

func TestParseFlagsBooleanLong(t *testing.T) {
	flags, _ := ParseFlags([]string{"--verbose", "--dry-run"})

	if flags["verbose"] != "true" {
		t.Errorf("flags[verbose] = %q, want %q", flags["verbose"], "true")
	}
	if flags["dry-run"] != "true" {
		t.Errorf("flags[dry-run] = %q, want %q", flags["dry-run"], "true")
	}
}

func TestParseFlagsShortWithValue(t *testing.T) {
	flags, _ := ParseFlags([]string{"-a", "claude-code", "-r", "/path"})

	if flags["a"] != "claude-code" {
		t.Errorf("flags[a] = %q, want %q", flags["a"], "claude-code")
	}
	if flags["r"] != "/path" {
		t.Errorf("flags[r] = %q, want %q", flags["r"], "/path")
	}
}

func TestParseFlagsBooleanShort(t *testing.T) {
	flags, _ := ParseFlags([]string{"-v"})

	if flags["v"] != "true" {
		t.Errorf("flags[v] = %q, want %q", flags["v"], "true")
	}
}

func TestParseFlagsMixed(t *testing.T) {
	// Note: ParseFlags treats a non-dash argument following a long flag
	// as that flag's value. So --verbose followed by extra-arg means
	// verbose="extra-arg", not verbose="true" + positional "extra-arg".
	flags, positional := ParseFlags([]string{
		"my-agent",
		"--adapter", "claude-code",
		"-r", "/path/to/repo",
		"--verbose",
		"extra-arg",
	})

	if len(positional) != 1 {
		t.Fatalf("positional len = %d, want 1", len(positional))
	}
	if positional[0] != "my-agent" {
		t.Errorf("positional[0] = %q, want %q", positional[0], "my-agent")
	}

	if flags["adapter"] != "claude-code" {
		t.Errorf("flags[adapter] = %q, want %q", flags["adapter"], "claude-code")
	}
	if flags["r"] != "/path/to/repo" {
		t.Errorf("flags[r] = %q, want %q", flags["r"], "/path/to/repo")
	}
	// "extra-arg" is consumed as the value for --verbose
	if flags["verbose"] != "extra-arg" {
		t.Errorf("flags[verbose] = %q, want %q", flags["verbose"], "extra-arg")
	}
}

func TestParseFlagsLongBooleanAtEnd(t *testing.T) {
	// A long flag at the end of args without a next arg should be treated as boolean
	flags, _ := ParseFlags([]string{"--force"})

	if flags["force"] != "true" {
		t.Errorf("flags[force] = %q, want %q", flags["force"], "true")
	}
}

func TestParseFlagsShortBooleanAtEnd(t *testing.T) {
	// A short flag at the end of args without a next arg should be treated as boolean
	flags, _ := ParseFlags([]string{"-f"})

	if flags["f"] != "true" {
		t.Errorf("flags[f] = %q, want %q", flags["f"], "true")
	}
}

func TestParseFlagsLongFollowedByFlag(t *testing.T) {
	// When a long flag is followed by another flag, it should be boolean
	flags, _ := ParseFlags([]string{"--verbose", "--debug"})

	if flags["verbose"] != "true" {
		t.Errorf("flags[verbose] = %q, want %q", flags["verbose"], "true")
	}
	if flags["debug"] != "true" {
		t.Errorf("flags[debug] = %q, want %q", flags["debug"], "true")
	}
}

func TestParseFlagsShortFollowedByFlag(t *testing.T) {
	// When a short flag is followed by another flag, it should be boolean
	flags, _ := ParseFlags([]string{"-v", "-d"})

	if flags["v"] != "true" {
		t.Errorf("flags[v] = %q, want %q", flags["v"], "true")
	}
	if flags["d"] != "true" {
		t.Errorf("flags[d] = %q, want %q", flags["d"], "true")
	}
}

func TestParseFlagsEqualsWithEmptyValue(t *testing.T) {
	flags, _ := ParseFlags([]string{"--key="})

	if flags["key"] != "" {
		t.Errorf("flags[key] = %q, want empty string", flags["key"])
	}
}

func TestGetFlagFound(t *testing.T) {
	flags := map[string]string{
		"adapter": "claude-code",
		"a":       "cursor",
	}

	// Should return the first matching name
	result := GetFlag(flags, []string{"adapter", "a"}, "default")
	if result != "claude-code" {
		t.Errorf("GetFlag = %q, want %q", result, "claude-code")
	}
}

func TestGetFlagShortName(t *testing.T) {
	flags := map[string]string{
		"a": "claude-code",
	}

	result := GetFlag(flags, []string{"adapter", "a"}, "default")
	if result != "claude-code" {
		t.Errorf("GetFlag = %q, want %q", result, "claude-code")
	}
}

func TestGetFlagDefault(t *testing.T) {
	flags := map[string]string{}

	result := GetFlag(flags, []string{"adapter", "a"}, "default-adapter")
	if result != "default-adapter" {
		t.Errorf("GetFlag = %q, want %q", result, "default-adapter")
	}
}

func TestGetFlagEmptyNames(t *testing.T) {
	flags := map[string]string{
		"key": "value",
	}

	result := GetFlag(flags, []string{}, "fallback")
	if result != "fallback" {
		t.Errorf("GetFlag = %q, want %q", result, "fallback")
	}
}

func TestGetFlagEmptyValue(t *testing.T) {
	flags := map[string]string{
		"key": "",
	}

	result := GetFlag(flags, []string{"key"}, "default")
	if result != "" {
		t.Errorf("GetFlag = %q, want empty string", result)
	}
}

func TestHasFlagPresent(t *testing.T) {
	flags := map[string]string{
		"verbose": "true",
		"v":       "true",
	}

	if !HasFlag(flags, "verbose") {
		t.Error("HasFlag(verbose) = false, want true")
	}
	if !HasFlag(flags, "v") {
		t.Error("HasFlag(v) = false, want true")
	}
	if !HasFlag(flags, "verbose", "v") {
		t.Error("HasFlag(verbose, v) = false, want true")
	}
}

func TestHasFlagAbsent(t *testing.T) {
	flags := map[string]string{
		"verbose": "true",
	}

	if HasFlag(flags, "debug") {
		t.Error("HasFlag(debug) = true, want false")
	}
	if HasFlag(flags, "d", "debug") {
		t.Error("HasFlag(d, debug) = true, want false")
	}
}

func TestHasFlagEmptyFlags(t *testing.T) {
	flags := map[string]string{}

	if HasFlag(flags, "anything") {
		t.Error("HasFlag on empty map = true, want false")
	}
}

func TestHasFlagNoNames(t *testing.T) {
	flags := map[string]string{
		"key": "value",
	}

	if HasFlag(flags) {
		t.Error("HasFlag with no names = true, want false")
	}
}

func TestHasFlagWithEmptyStringValue(t *testing.T) {
	flags := map[string]string{
		"key": "",
	}

	// The key exists even though the value is empty
	if !HasFlag(flags, "key") {
		t.Error("HasFlag(key) = false, want true (key exists with empty value)")
	}
}

func TestVersionConstant(t *testing.T) {
	if Version == "" {
		t.Error("Version constant is empty")
	}
}

func TestParseFlagsEqualsInValue(t *testing.T) {
	// Value itself contains an equals sign
	flags, _ := ParseFlags([]string{"--constraint=>=1.0.0"})

	if flags["constraint"] != ">=1.0.0" {
		t.Errorf("flags[constraint] = %q, want %q", flags["constraint"], ">=1.0.0")
	}
}

func TestParseFlagsMultiplePositionalWithFlags(t *testing.T) {
	flags, positional := ParseFlags([]string{
		"pos1", "--flag1", "val1", "pos2", "-f", "val2", "pos3",
	})

	if len(positional) != 3 {
		t.Fatalf("positional len = %d, want 3", len(positional))
	}
	if positional[0] != "pos1" {
		t.Errorf("positional[0] = %q, want pos1", positional[0])
	}
	if positional[1] != "pos2" {
		t.Errorf("positional[1] = %q, want pos2", positional[1])
	}
	if positional[2] != "pos3" {
		t.Errorf("positional[2] = %q, want pos3", positional[2])
	}

	if flags["flag1"] != "val1" {
		t.Errorf("flags[flag1] = %q, want val1", flags["flag1"])
	}
	if flags["f"] != "val2" {
		t.Errorf("flags[f] = %q, want val2", flags["f"])
	}
}
