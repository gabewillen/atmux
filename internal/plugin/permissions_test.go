package plugin

import "testing"

func TestHasPermission(t *testing.T) {
	gater := &DefaultGater{}
	
	p := &Plugin{
		Manifest: Manifest{
			Permissions: []string{
				"agent.list",
				"system.*",
			},
		},
	}
	
	tests := []struct {
		perm     string
		expected bool
	}{
		{"agent.list", true},
		{"agent.add", false},
		{"system.update", true},
		{"system.info", true},
		{"other.perm", false},
	}
	
	for _, tt := range tests {
		if got := gater.HasPermission(p, tt.perm); got != tt.expected {
			t.Errorf("HasPermission(%q) = %v, want %v", tt.perm, got, tt.expected)
		}
	}
}
