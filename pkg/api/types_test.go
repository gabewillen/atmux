package api

import (
	"testing"

	"github.com/stateforward/hsm-go/muid"
)

func TestAgentValidate(t *testing.T) {
	validAgent := Agent{
		ID:       muid.MUID(123),
		Name:     "frontend-dev",
		Slug:     "frontend-dev",
		Adapter:  "claude-code",
		RepoRoot: "/home/user/project",
	}

	tests := []struct {
		name    string
		agent   Agent
		wantErr string
	}{
		{"valid agent", validAgent, ""},
		{
			"zero ID",
			func() Agent { a := validAgent; a.ID = 0; return a }(),
			"ID",
		},
		{
			"empty name",
			func() Agent { a := validAgent; a.Name = ""; return a }(),
			"Name",
		},
		{
			"empty slug",
			func() Agent { a := validAgent; a.Slug = ""; return a }(),
			"Slug",
		},
		{
			"empty adapter",
			func() Agent { a := validAgent; a.Adapter = ""; return a }(),
			"Adapter",
		},
		{
			"empty repo root",
			func() Agent { a := validAgent; a.RepoRoot = ""; return a }(),
			"RepoRoot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.agent.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Validate() expected error containing %q, got nil", tt.wantErr)
				} else {
					valErr, ok := err.(*AgentValidationError)
					if !ok {
						t.Errorf("Validate() expected AgentValidationError, got %T", err)
					} else if valErr.Field != tt.wantErr {
						t.Errorf("Validate() error field = %q, want %q", valErr.Field, tt.wantErr)
					}
				}
			}
		})
	}
}

func TestSessionValidate(t *testing.T) {
	tests := []struct {
		name    string
		session Session
		wantErr string
	}{
		{
			"valid session",
			Session{ID: muid.MUID(1), Agents: []muid.MUID{2, 3}},
			"",
		},
		{
			"valid session no agents",
			Session{ID: muid.MUID(1), Agents: nil},
			"",
		},
		{
			"zero ID",
			Session{ID: 0, Agents: nil},
			"ID",
		},
		{
			"zero agent ID",
			Session{ID: muid.MUID(1), Agents: []muid.MUID{2, 0, 3}},
			"Agents",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.session.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Validate() expected error containing %q, got nil", tt.wantErr)
				} else {
					valErr, ok := err.(*SessionValidationError)
					if !ok {
						t.Errorf("Validate() expected SessionValidationError, got %T", err)
					} else if valErr.Field != tt.wantErr {
						t.Errorf("Validate() error field = %q, want %q", valErr.Field, tt.wantErr)
					}
				}
			}
		})
	}
}

func TestSessionHasAgent(t *testing.T) {
	s := Session{ID: 1, Agents: []muid.MUID{2, 3, 4}}

	if !s.HasAgent(2) {
		t.Error("HasAgent(2) = false, want true")
	}
	if s.HasAgent(5) {
		t.Error("HasAgent(5) = true, want false")
	}
}

func TestSessionAddAgent(t *testing.T) {
	s := Session{ID: 1, Agents: []muid.MUID{2, 3}}

	// Add new agent
	if !s.AddAgent(4) {
		t.Error("AddAgent(4) = false, want true")
	}
	if !s.HasAgent(4) {
		t.Error("After AddAgent(4), HasAgent(4) = false")
	}

	// Add existing agent
	if s.AddAgent(2) {
		t.Error("AddAgent(2) = true, want false (already exists)")
	}
}

func TestSessionRemoveAgent(t *testing.T) {
	s := Session{ID: 1, Agents: []muid.MUID{2, 3, 4}}

	// Remove existing agent
	if !s.RemoveAgent(3) {
		t.Error("RemoveAgent(3) = false, want true")
	}
	if s.HasAgent(3) {
		t.Error("After RemoveAgent(3), HasAgent(3) = true")
	}

	// Remove non-existing agent
	if s.RemoveAgent(5) {
		t.Error("RemoveAgent(5) = true, want false (not present)")
	}
}

func TestLifecycleStateIsFinal(t *testing.T) {
	tests := []struct {
		state   LifecycleState
		isFinal bool
	}{
		{LifecyclePending, false},
		{LifecycleStarting, false},
		{LifecycleRunning, false},
		{LifecycleTerminated, true},
		{LifecycleErrored, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			if got := tt.state.IsFinal(); got != tt.isFinal {
				t.Errorf("%s.IsFinal() = %v, want %v", tt.state, got, tt.isFinal)
			}
		})
	}
}

func TestLifecycleStateIsValid(t *testing.T) {
	tests := []struct {
		state   LifecycleState
		isValid bool
	}{
		{LifecyclePending, true},
		{LifecycleStarting, true},
		{LifecycleRunning, true},
		{LifecycleTerminated, true},
		{LifecycleErrored, true},
		{LifecycleState("invalid"), false},
		{LifecycleState(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			if got := tt.state.IsValid(); got != tt.isValid {
				t.Errorf("%q.IsValid() = %v, want %v", tt.state, got, tt.isValid)
			}
		})
	}
}

func TestPresenceStateCanAcceptTasks(t *testing.T) {
	tests := []struct {
		state      PresenceState
		canAccept  bool
	}{
		{PresenceOnline, true},
		{PresenceBusy, false},
		{PresenceOffline, false},
		{PresenceAway, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			if got := tt.state.CanAcceptTasks(); got != tt.canAccept {
				t.Errorf("%s.CanAcceptTasks() = %v, want %v", tt.state, got, tt.canAccept)
			}
		})
	}
}

func TestPresenceStateIsValid(t *testing.T) {
	tests := []struct {
		state   PresenceState
		isValid bool
	}{
		{PresenceOnline, true},
		{PresenceBusy, true},
		{PresenceOffline, true},
		{PresenceAway, true},
		{PresenceState("invalid"), false},
		{PresenceState(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			if got := tt.state.IsValid(); got != tt.isValid {
				t.Errorf("%q.IsValid() = %v, want %v", tt.state, got, tt.isValid)
			}
		})
	}
}

func TestLocationTypeString(t *testing.T) {
	tests := []struct {
		lt       LocationType
		expected string
	}{
		{LocationLocal, "local"},
		{LocationSSH, "ssh"},
		{LocationType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.lt.String(); got != tt.expected {
				t.Errorf("LocationType(%d).String() = %q, want %q", tt.lt, got, tt.expected)
			}
		})
	}
}

func TestParseLocationType(t *testing.T) {
	tests := []struct {
		input    string
		expected LocationType
		wantErr  bool
	}{
		{"local", LocationLocal, false},
		{"LOCAL", LocationLocal, false},
		{"Local", LocationLocal, false},
		{"lOcAl", LocationLocal, false},
		{"lOCAL", LocationLocal, false},
		{"ssh", LocationSSH, false},
		{"SSH", LocationSSH, false},
		{"Ssh", LocationSSH, false},
		{"sSH", LocationSSH, false},
		{"sSh", LocationSSH, false},
		{"invalid", LocationLocal, true},
		{"", LocationLocal, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseLocationType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLocationType(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if err == nil && got != tt.expected {
				t.Errorf("ParseLocationType(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
