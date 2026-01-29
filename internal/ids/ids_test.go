package ids

import (
	"testing"

	"github.com/stateforward/hsm-go/muid"

	"github.com/agentflare-ai/amux/pkg/api"
)

func TestNewID(t *testing.T) {
	// Generate several IDs and verify they are non-zero
	for i := 0; i < 100; i++ {
		id := NewID()
		if id == 0 {
			t.Errorf("NewID() returned 0, which is reserved")
		}
	}
}

func TestIsValidRuntimeID(t *testing.T) {
	tests := []struct {
		name  string
		id    muid.MUID
		valid bool
	}{
		{"zero is invalid", 0, false},
		{"one is valid", 1, true},
		{"max uint64 is valid", muid.MUID(^uint64(0)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidRuntimeID(tt.id); got != tt.valid {
				t.Errorf("IsValidRuntimeID(%d) = %v, want %v", tt.id, got, tt.valid)
			}
		})
	}
}

func TestNormalizeAgentSlug(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic cases
		{"simple lowercase", "frontend-dev", "frontend-dev"},
		{"simple uppercase", "FRONTEND-DEV", "frontend-dev"},
		{"mixed case", "Frontend-Dev", "frontend-dev"},

		// Special character replacement
		{"spaces to dashes", "frontend dev", "frontend-dev"},
		{"underscores to dashes", "frontend_dev", "frontend-dev"},
		{"dots to dashes", "frontend.dev", "frontend-dev"},
		{"multiple special chars", "my agent name!", "my-agent-name"},

		// Dash collapsing
		{"multiple dashes", "frontend--dev", "frontend-dev"},
		{"many dashes", "a---b---c", "a-b-c"},

		// Dash trimming
		{"leading dash", "-frontend", "frontend"},
		{"trailing dash", "frontend-", "frontend"},
		{"both leading and trailing", "-frontend-", "frontend"},

		// Empty and default cases
		{"empty string", "", "agent"},
		{"only special chars", "!!!", "agent"},
		{"only dashes", "---", "agent"},

		// Length truncation
		{"exactly 63 chars", "abcdefghijklmnopqrstuvwxyz012345678901234567890123456789012", "abcdefghijklmnopqrstuvwxyz012345678901234567890123456789012"},
		{"over 63 chars", "abcdefghijklmnopqrstuvwxyz0123456789012345678901234567890123456789", "abcdefghijklmnopqrstuvwxyz0123456789012345678901234567890123456"},

		// Complex cases
		{"unicode to dash", "frntnd-dv", "frntnd-dv"},
		{"numbers preserved", "agent123", "agent123"},
		{"mixed complexity", "  My Agent__Name--Here!  ", "my-agent-name-here"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeAgentSlug(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizeAgentSlug(%q) = %q, want %q", tt.input, got, tt.expected)
			}

			// Verify length constraint
			if len(got) > MaxAgentSlugLength {
				t.Errorf("NormalizeAgentSlug(%q) produced slug of length %d, max is %d", tt.input, len(got), MaxAgentSlugLength)
			}
		})
	}
}

func TestUniqueAgentSlug(t *testing.T) {
	t.Run("no collision", func(t *testing.T) {
		exists := func(slug string) bool { return false }
		got := UniqueAgentSlug("frontend", exists)
		if got != "frontend" {
			t.Errorf("UniqueAgentSlug() = %q, want %q", got, "frontend")
		}
	})

	t.Run("with collision", func(t *testing.T) {
		existing := map[string]bool{
			"frontend":   true,
			"frontend-2": true,
		}
		exists := func(slug string) bool { return existing[slug] }
		got := UniqueAgentSlug("frontend", exists)
		if got != "frontend-3" {
			t.Errorf("UniqueAgentSlug() = %q, want %q", got, "frontend-3")
		}
	})

	t.Run("collision with normalized name", func(t *testing.T) {
		existing := map[string]bool{
			"frontend-dev": true,
		}
		exists := func(slug string) bool { return existing[slug] }
		got := UniqueAgentSlug("Frontend Dev", exists)
		if got != "frontend-dev-2" {
			t.Errorf("UniqueAgentSlug() = %q, want %q", got, "frontend-dev-2")
		}
	})
}

func TestRepoKey(t *testing.T) {
	tests := []struct {
		name     string
		location api.Location
		repoRoot string
		expected string
	}{
		{
			"local agent",
			api.Location{Type: api.LocationLocal},
			"/home/user/project",
			"local:/home/user/project",
		},
		{
			"ssh agent",
			api.Location{Type: api.LocationSSH, Host: "server.example.com"},
			"/home/user/project",
			"ssh:server.example.com:/home/user/project",
		},
		{
			"ssh agent with user",
			api.Location{Type: api.LocationSSH, Host: "server.example.com", User: "deploy"},
			"/home/deploy/project",
			"ssh:server.example.com:/home/deploy/project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RepoKey(tt.location, tt.repoRoot)
			if got != tt.expected {
				t.Errorf("RepoKey() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestEncodeDecodeID(t *testing.T) {
	tests := []struct {
		name string
		id   muid.MUID
	}{
		{"zero", 0},
		{"one", 1},
		{"small", 12345},
		{"large", muid.MUID(1 << 48)},
		{"max", muid.MUID(^uint64(0))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := EncodeID(tt.id)
			decoded, err := DecodeID(encoded)
			if err != nil {
				t.Fatalf("DecodeID(%q) failed: %v", encoded, err)
			}
			if decoded != tt.id {
				t.Errorf("round-trip failed: got %d, want %d", decoded, tt.id)
			}
		})
	}
}

func TestDecodeIDError(t *testing.T) {
	tests := []string{
		"",
		"abc",
		"12.34",
		"-1",
	}

	for _, s := range tests {
		t.Run(s, func(t *testing.T) {
			_, err := DecodeID(s)
			if err == nil {
				t.Errorf("DecodeID(%q) should have failed", s)
			}
		})
	}
}

func TestEncodeDecodeIDs(t *testing.T) {
	ids := []muid.MUID{1, 2, 3, 12345, muid.MUID(1 << 48)}
	encoded := EncodeIDs(ids)

	if len(encoded) != len(ids) {
		t.Fatalf("EncodeIDs length mismatch: got %d, want %d", len(encoded), len(ids))
	}

	decoded, err := DecodeIDs(encoded)
	if err != nil {
		t.Fatalf("DecodeIDs failed: %v", err)
	}

	if len(decoded) != len(ids) {
		t.Fatalf("DecodeIDs length mismatch: got %d, want %d", len(decoded), len(ids))
	}

	for i, id := range ids {
		if decoded[i] != id {
			t.Errorf("decoded[%d] = %d, want %d", i, decoded[i], id)
		}
	}
}

func TestBroadcastID(t *testing.T) {
	if BroadcastID != 0 {
		t.Errorf("BroadcastID = %d, want 0", BroadcastID)
	}

	// BroadcastID should be invalid as a runtime ID
	if IsValidRuntimeID(BroadcastID) {
		t.Error("BroadcastID should not be a valid runtime ID")
	}
}
