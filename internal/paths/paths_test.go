// Package paths implements tests for the path resolver
package paths

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNew tests creating a new path resolver
func TestNew(t *testing.T) {
	config := Config{
		BaseDir:   "/base",
		HomeDir:   "/home/user",
		RepoRoot:  "/repo/root",
		CacheDir:  "/cache",
		ConfigDir: "/config",
	}
	
	resolver := New(config)
	
	if resolver == nil {
		t.Fatal("Expected resolver to be created")
	}
	
	if resolver.config.BaseDir != "/base" {
		t.Errorf("Expected BaseDir '/base', got '%s'", resolver.config.BaseDir)
	}
	
	if resolver.config.HomeDir != "/home/user" {
		t.Errorf("Expected HomeDir '/home/user', got '%s'", resolver.config.HomeDir)
	}
	
	if resolver.config.RepoRoot != "/repo/root" {
		t.Errorf("Expected RepoRoot '/repo/root', got '%s'", resolver.config.RepoRoot)
	}
}

// TestResolve tests path resolution
func TestResolve(t *testing.T) {
	config := Config{
		BaseDir: "/base/dir",
		HomeDir: "/home/user",
	}
	
	resolver := New(config)
	
	// Test absolute path (should remain unchanged)
	absPath := "/absolute/path"
	resolved := resolver.Resolve(absPath)
	if resolved != "/absolute/path" {
		t.Errorf("Expected absolute path to remain unchanged, got '%s'", resolved)
	}
	
	// Test relative path (should be resolved relative to base dir)
	relPath := "relative/path"
	resolved = resolver.Resolve(relPath)
	expected := filepath.Join("/base/dir", "relative/path")
	if resolved != expected {
		t.Errorf("Expected '%s', got '%s'", expected, resolved)
	}
	
	// Test home directory expansion
	homePath := "~/my/path"
	resolved = resolver.Resolve(homePath)
	expected = filepath.Join("/home/user", "my/path")
	if resolved != expected {
		t.Errorf("Expected '%s', got '%s'", expected, resolved)
	}
	
	// Test path cleaning
	dirtyPath := "path/../clean/path"
	resolved = resolver.Resolve(dirtyPath)
	expected = filepath.Join("/base/dir", "clean/path")
	if resolved != expected {
		t.Errorf("Expected cleaned path '%s', got '%s'", expected, resolved)
	}
}

// TestAmuxDir tests getting the .amux directory path
func TestAmuxDir(t *testing.T) {
	config := Config{
		BaseDir:  "/base",
		RepoRoot: "/repo/root",
	}
	
	resolver := New(config)
	
	amuxDir := resolver.AmuxDir()
	expected := filepath.Join("/repo/root", ".amux")
	if amuxDir != expected {
		t.Errorf("Expected '%s', got '%s'", expected, amuxDir)
	}
}

// TestWorktreesDir tests getting the worktrees directory path
func TestWorktreesDir(t *testing.T) {
	config := Config{
		BaseDir:  "/base",
		RepoRoot: "/repo/root",
	}
	
	resolver := New(config)
	
	worktreesDir := resolver.WorktreesDir()
	expected := filepath.Join("/repo/root", ".amux", "worktrees")
	if worktreesDir != expected {
		t.Errorf("Expected '%s', got '%s'", expected, worktreesDir)
	}
}

// TestWorktreeDir tests getting a specific agent's worktree directory
func TestWorktreeDir(t *testing.T) {
	config := Config{
		BaseDir:  "/base",
		RepoRoot: "/repo/root",
	}
	
	resolver := New(config)
	
	worktreeDir := resolver.WorktreeDir("test-agent")
	expected := filepath.Join("/repo/root", ".amux", "worktrees", "test-agent")
	if worktreeDir != expected {
		t.Errorf("Expected '%s', got '%s'", expected, worktreeDir)
	}
}

// TestAgentBranch tests getting the git branch name for an agent
func TestAgentBranch(t *testing.T) {
	config := Config{
		BaseDir:  "/base",
		RepoRoot: "/repo/root",
	}
	
	resolver := New(config)
	
	branch := resolver.AgentBranch("test-agent")
	expected := "amux/test-agent"
	if branch != expected {
		t.Errorf("Expected '%s', got '%s'", expected, branch)
	}
}

// TestSocketPath tests getting the daemon socket path
func TestSocketPath(t *testing.T) {
	config := Config{
		BaseDir:  "/base",
		RepoRoot: "/repo/root",
	}
	
	resolver := New(config)
	
	socketPath := resolver.SocketPath()
	expected := filepath.Join("/repo/root", ".amux", "amuxd.sock")
	if socketPath != expected {
		t.Errorf("Expected '%s', got '%s'", expected, socketPath)
	}
}

// TestCachePath tests getting a path within the cache directory
func TestCachePath(t *testing.T) {
	config := Config{
		BaseDir:   "/base",
		RepoRoot:  "/repo/root",
		CacheDir:  "/cache",
	}
	
	resolver := New(config)
	
	cachePath := resolver.CachePath("subdir", "file.txt")
	expected := filepath.Join("/cache", "subdir", "file.txt")
	if cachePath != expected {
		t.Errorf("Expected '%s', got '%s'", expected, cachePath)
	}
}

// TestConfigPath tests getting a path within the config directory
func TestConfigPath(t *testing.T) {
	config := Config{
		BaseDir:   "/base",
		RepoRoot:  "/repo/root",
		ConfigDir: "/config",
	}
	
	resolver := New(config)
	
	configPath := resolver.ConfigPath("subdir", "config.toml")
	expected := filepath.Join("/config", "subdir", "config.toml")
	if configPath != expected {
		t.Errorf("Expected '%s', got '%s'", expected, configPath)
	}
}

// TestRepoRoot tests getting the resolved repository root
func TestRepoRoot(t *testing.T) {
	config := Config{
		BaseDir:  "/base",
		RepoRoot: "/repo/root",
	}
	
	resolver := New(config)
	
	repoRoot := resolver.RepoRoot()
	expected := "/repo/root"
	if repoRoot != expected {
		t.Errorf("Expected '%s', got '%s'", expected, repoRoot)
	}
}

// TestExpandHome tests expanding the home directory
func TestExpandHome(t *testing.T) {
	// Note: This test depends on the actual user's home directory
	// In a real test, we might want to mock this, but for now we'll work with the actual home
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("Could not get user home directory: %v", err)
	}
	
	expanded := ExpandHome("~/test/path")
	expected := filepath.Join(homeDir, "test/path")
	if expanded != expected {
		t.Errorf("Expected '%s', got '%s'", expected, expanded)
	}
	
	// Test with non-home path
	nonHomePath := "/absolute/path"
	expanded = ExpandHome(nonHomePath)
	if expanded != nonHomePath {
		t.Errorf("Expected non-home path to remain unchanged, got '%s'", expanded)
	}
}

// TestEnsureDir tests ensuring a directory exists
func TestEnsureDir(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "test", "subdir")
	
	err := EnsureDir(testDir)
	if err != nil {
		t.Fatalf("Unexpected error ensuring directory: %v", err)
	}
	
	// Check that directory exists
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Error("Expected directory to exist after EnsureDir")
	}
}

// TestEnsureParentDir tests ensuring a parent directory exists
func TestEnsureParentDir(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "parent", "subdir", "file.txt")
	
	err := EnsureParentDir(testFile)
	if err != nil {
		t.Fatalf("Unexpected error ensuring parent directory: %v", err)
	}
	
	// Check that parent directory exists
	parentDir := filepath.Dir(testFile)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		t.Error("Expected parent directory to exist after EnsureParentDir")
	}
}