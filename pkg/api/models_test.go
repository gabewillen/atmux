package api

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestLocationTypeJSON(t *testing.T) {
	data, err := json.Marshal(LocationSSH)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(data) != "\"ssh\"" {
		t.Fatalf("unexpected json: %s", string(data))
	}
	var decoded LocationType
	if err := json.Unmarshal([]byte("\"local\""), &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded != LocationLocal {
		t.Fatalf("expected local, got %v", decoded)
	}
}

func TestAgentValidate(t *testing.T) {
	loc := Location{Type: LocationLocal}
	agent, err := NewAgent("alpha", "core work", AdapterRef("claude"), "/repo", "/repo/.amux/worktrees/alpha", loc)
	if err != nil {
		t.Fatalf("new agent: %v", err)
	}
	if err := agent.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestAgentValidateMissingName(t *testing.T) {
	loc := Location{Type: LocationLocal}
	_, err := NewAgentWithID(NewAgentID(), "", "desc", AdapterRef("adapter"), "/repo", "/repo/.amux/worktrees/alpha", loc)
	if err == nil {
		t.Fatalf("expected error for missing name")
	}
}

func TestSessionValidate(t *testing.T) {
	loc := Location{Type: LocationLocal}
	_, err := NewSession(NewAgentID(), "/repo", "/repo/.amux/worktrees/alpha", loc)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
}

func TestLocationValidateSSHRequiresRepoPath(t *testing.T) {
	loc := Location{Type: LocationSSH, Host: "devbox"}
	if err := loc.Validate(); err == nil {
		t.Fatalf("expected error for missing repo_path")
	}
}

func TestAgentValidateRejectsRelativeRepoRoot(t *testing.T) {
	loc := Location{Type: LocationLocal}
	_, err := NewAgentWithID(NewAgentID(), "alpha", "desc", AdapterRef("adapter"), "repo", "/repo/.amux/worktrees/alpha", loc)
	if err == nil {
		t.Fatalf("expected error for relative repo root")
	}
}

func TestAgentValidateRejectsWorktreeOutsideRepo(t *testing.T) {
	repo := t.TempDir()
	other := t.TempDir()
	loc := Location{Type: LocationLocal}
	_, err := NewAgentWithID(NewAgentID(), "alpha", "desc", AdapterRef("adapter"), repo, other, loc)
	if err == nil {
		t.Fatalf("expected error for worktree outside repo")
	}
}

func TestSessionValidateRejectsRepoRootAsWorktree(t *testing.T) {
	repo := t.TempDir()
	loc := Location{Type: LocationLocal}
	_, err := NewSessionWithID(NewSessionID(), NewAgentID(), repo, repo, loc)
	if err == nil {
		t.Fatalf("expected error for worktree equal to repo root")
	}
}

func TestSessionValidateAcceptsWorktreeWithinRepo(t *testing.T) {
	repo := t.TempDir()
	worktree := filepath.Join(repo, ".amux", "worktrees", "alpha")
	loc := Location{Type: LocationLocal}
	_, err := NewSessionWithID(NewSessionID(), NewAgentID(), repo, worktree, loc)
	if err != nil {
		t.Fatalf("expected valid session: %v", err)
	}
}
