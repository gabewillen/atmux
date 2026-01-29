package ids

import (
	"strings"
	"testing"
)

func TestAgentSlugFromName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name",
			input:    "test-agent",
			expected: "test-agent",
		},
		{
			name:     "uppercase conversion",
			input:    "Test-Agent",
			expected: "test-agent",
		},
		{
			name:     "special characters",
			input:    "my@agent#1",
			expected: "my-agent-1",
		},
		{
			name:     "multiple consecutive dashes",
			input:    "my---agent",
			expected: "my-agent",
		},
		{
			name:     "leading and trailing dashes",
			input:    "-my-agent-",
			expected: "my-agent",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "unnamed",
		},
		{
			name:     "only special characters",
			input:    "@#$",
			expected: "unnamed",
		},
		{
			name:     "long name truncation",
			input:    strings.Repeat("a", 70),
			expected: strings.Repeat("a", 63),
		},
		{
			name:     "unicode characters",
			input:    "my-agent-κόσμος",
			expected: "my-agent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AgentSlugFromName(tt.input)
			if result != tt.expected {
				t.Errorf("AgentSlugFromName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateAgentSlug(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid slug",
			input:   "test-agent",
			wantErr: false,
		},
		{
			name:    "valid single character",
			input:   "a",
			wantErr: false,
		},
		{
			name:    "valid with numbers",
			input:   "agent-123",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "too long",
			input:   strings.Repeat("a", 64),
			wantErr: true,
		},
		{
			name:    "starts with dash",
			input:   "-agent",
			wantErr: true,
		},
		{
			name:    "ends with dash",
			input:   "agent-",
			wantErr: true,
		},
		{
			name:    "contains uppercase",
			input:   "Agent",
			wantErr: true,
		},
		{
			name:    "contains special character",
			input:   "my@agent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAgentSlug(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAgentSlug(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestCanonicalizeRepoRoot(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		isRemote bool
		wantErr  bool
	}{
		{
			name:     "local absolute path",
			path:     "/home/user/repo",
			isRemote: false,
			wantErr:  false,
		},
		{
			name:     "local relative path",
			path:     "./repo",
			isRemote: false,
			wantErr:  false,
		},
		{
			name:     "remote home path",
			path:     "~/repo",
			isRemote: true,
			wantErr:  false,
		},
		{
			name:     "empty path",
			path:     "",
			isRemote: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CanonicalizeRepoRoot(tt.path, tt.isRemote)
			if (err != nil) != tt.wantErr {
				t.Errorf("CanonicalizeRepoRoot(%q, %v) error = %v, wantErr %v", tt.path, tt.isRemote, err, tt.wantErr)
			}
			if !tt.wantErr && result == "" {
				t.Errorf("CanonicalizeRepoRoot(%q, %v) returned empty result", tt.path, tt.isRemote)
			}
		})
	}
}

func TestIsValidIdentifierName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid simple name",
			input:    "agent-name",
			expected: true,
		},
		{
			name:     "valid with spaces",
			input:    "My Agent Name",
			expected: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "too long",
			input:    strings.Repeat("a", 257),
			expected: false,
		},
		{
			name:     "contains non-printable",
			input:    "agent\x00name",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidIdentifierName(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidIdentifierName(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNewIDs(t *testing.T) {
	// Test that ID generation functions return non-zero values
	agentID := NewAgentID()
	peerID := NewPeerID() 
	hostID := NewHostID()
	
	if agentID == 0 {
		t.Error("NewAgentID() returned zero value")
	}
	
	if peerID == 0 {
		t.Error("NewPeerID() returned zero value")
	}
	
	if hostID == 0 {
		t.Error("NewHostID() returned zero value")
	}
	
	// Test that multiple calls return different values
	agentID2 := NewAgentID()
	if agentID == agentID2 {
		t.Error("NewAgentID() returned duplicate values")
	}
}