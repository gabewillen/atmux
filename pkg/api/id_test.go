package api

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stateforward/hsm-go/muid"
)

func TestEncodeParseID(t *testing.T) {
	tests := []struct {
		desc    string
		inputID muid.MUID
	}{
		{"valid ID", 1234567890},
		{"max uint64 ID", muid.MUID(^uint64(0))},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			str := EncodeID(tc.inputID)
			parsed, err := ParseID(str)
			if err != nil {
				t.Fatalf("ParseID failed: %v", err)
			}
			if parsed != tc.inputID {
				t.Errorf("got %d, want %d", parsed, tc.inputID)
			}
		})
	}
}

func TestParseIDErrors(t *testing.T) {
	invalid := []string{
		"", "abc", "-1", "12.34", "18446744073709551616", // overflow
	}
	for _, s := range invalid {
		t.Run("invalid_"+s, func(t *testing.T) {
			_, err := ParseID(s)
			if err == nil {
				t.Errorf("expected error for invalid input %q, got nil", s)
			}
		})
	}
}

func TestAgentSlug(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Simple", "simple"},
		{"With Spaces", "with-spaces"},
		{"SpecialChars!@#", "specialchars"},
		{"Multiple---Dashes", "multiple-dashes"},
		{"Trailing-Dash-", "trailing-dash"},
		{"-Leading-Dash", "leading-dash"},
		{"", "agent"},                                       // Fallback
		{strings.Repeat("a", 100), strings.Repeat("a", 63)}, // Truncation
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%q", tc.input), func(t *testing.T) {
			got := NewAgentSlug(tc.input)
			if got.String() != tc.want {
				t.Errorf("NewAgentSlug(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestReservedID(t *testing.T) {
	if ReservedID != 0 {
		t.Errorf("ReservedID must be 0, got %d", ReservedID)
	}
}
