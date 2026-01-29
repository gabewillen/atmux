package api

import (
	"encoding/json"
	"testing"

	"github.com/stateforward/hsm-go/muid"
)

func TestAgentID_JSON(t *testing.T) {
	// Create a random muid
	m := muid.Make()
	id := AgentID(m)

	// Marshal
	data, err := json.Marshal(id)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Expect string, not number
	s := string(data)
	if s[0] != '"' || s[len(s)-1] != '"' {
		t.Errorf("Expected JSON string, got %s", s)
	}

	// Unmarshal
	var id2 AgentID
	if err := json.Unmarshal(data, &id2); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if id != id2 {
		t.Errorf("Expected %v, got %v", id, id2)
	}
}

func TestNormalizeAgentSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Simple", "simple"},
		{"With Spaces", "with-spaces"},
		{"With_Underscore", "with-underscore"},
		{"Multi---Dash", "multi-dash"},
		{"-Trim-", "trim"},
		{"Invalid!@#Char", "invalid-char"},
		{"VeryLongNameThatExceedsTheLimitOfSixtyThreeCharactersAndShouldBeTruncated", "verylongnamethatexceedsthelimitofsixtythreecharactersandshouldb"},
	}

	for _, tc := range tests {
		got := NormalizeAgentSlug(tc.input)
		if string(got) != tc.expected {
			t.Errorf("NormalizeAgentSlug(%q) = %q, want %q", tc.input, got, tc.expected)
		}
		if err := got.Validate(); err != nil {
			t.Errorf("Validate(%q) failed: %v", got, err)
		}
	}
}

func TestAgentSlug_Validate_Error(t *testing.T) {
	badSlugs := []string{
		"",
		"-start",
		"end-",
		"Upper",
		"space ",
		"inv@lid",
	}

	for _, s := range badSlugs {
		slug := AgentSlug(s)
		if err := slug.Validate(); err == nil {
			t.Errorf("Validate(%q) should have failed", s)
		}
	}
}
