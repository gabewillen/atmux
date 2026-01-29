package merge

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	
	if cfg.Strategy != StrategyMergeCommit {
		t.Errorf("Expected default strategy %q, got %q", StrategyMergeCommit, cfg.Strategy)
	}
	
	if cfg.BaseBranch != "main" {
		t.Errorf("Expected default base branch 'main', got %q", cfg.BaseBranch)
	}
	
	if cfg.TargetBranch != "main" {
		t.Errorf("Expected default target branch 'main', got %q", cfg.TargetBranch)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Strategy:     StrategyMergeCommit,
				BaseBranch:   "main",
				TargetBranch: "main",
			},
			wantErr: false,
		},
		{
			name: "invalid strategy",
			config: Config{
				Strategy:     "invalid-strategy",
				BaseBranch:   "main",
				TargetBranch: "main",
			},
			wantErr: true,
		},
		{
			name: "empty base branch",
			config: Config{
				Strategy:     StrategyMergeCommit,
				BaseBranch:   "",
				TargetBranch: "main",
			},
			wantErr: true,
		},
		{
			name: "empty target branch",
			config: Config{
				Strategy:     StrategyMergeCommit,
				BaseBranch:   "main",
				TargetBranch: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStrategyConstants(t *testing.T) {
	strategies := []Strategy{
		StrategyMergeCommit,
		StrategySquash,
		StrategyRebase,
		StrategyFastForwardOnly,
	}

	for _, strategy := range strategies {
		if string(strategy) == "" {
			t.Errorf("Strategy constant should not be empty: %v", strategy)
		}
	}
}

func TestDryRun(t *testing.T) {
	// Test with non-existent directory
	tempDir := t.TempDir()
	
	config := Config{
		Strategy:     StrategyMergeCommit,
		BaseBranch:   "main",
		TargetBranch: "main",
	}

	// This should fail because it's not a git repository
	err := DryRun(tempDir, "feature-branch", config)
	if err == nil {
		t.Error("DryRun() should fail for non-git directory")
	}
}

func TestDryRunWithGitRepo(t *testing.T) {
	if !isGitAvailable() {
		t.Skip("Git not available, skipping test")
	}

	// Create temporary git repository
	tempDir := t.TempDir()
	if err := initTestGitRepo(tempDir); err != nil {
		t.Fatalf("Failed to init test repo: %v", err)
	}

	config := Config{
		Strategy:     StrategyMergeCommit,
		BaseBranch:   "main",
		TargetBranch: "main",
	}

	// This should succeed for existing branch
	err := DryRun(tempDir, "main", config)
	if err != nil {
		t.Errorf("DryRun() failed for valid git repo: %v", err)
	}
	
	// This should fail for non-existent branch
	err = DryRun(tempDir, "non-existent-branch", config)
	if err == nil {
		t.Error("DryRun() should fail for non-existent branch")
	}
}

func TestExecuteMergeWithoutGit(t *testing.T) {
	// Test with non-git directory
	tempDir := t.TempDir()
	
	config := Config{
		Strategy:     StrategyMergeCommit,
		BaseBranch:   "main",
		TargetBranch: "main",
	}

	// All merge strategies should fail without git
	err := ExecuteMerge(tempDir, "feature", config)
	if err == nil {
		t.Error("ExecuteMerge() should fail for non-git directory")
	}
}

// Helper functions

func isGitAvailable() bool {
	// Simple check to see if git command exists
	_, err := os.Stat("/usr/bin/git")
	if err == nil {
		return true
	}
	
	_, err = os.Stat("/bin/git")
	return err == nil
}

func initTestGitRepo(dir string) error {
	// Create .git directory to simulate a git repo
	gitDir := filepath.Join(dir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		return err
	}
	
	// Create a minimal git config
	configDir := filepath.Join(gitDir, "refs", "heads")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	
	// Create main branch reference
	mainRef := filepath.Join(configDir, "main")
	return os.WriteFile(mainRef, []byte("fake-commit-hash\n"), 0644)
}