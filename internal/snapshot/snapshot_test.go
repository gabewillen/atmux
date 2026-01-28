// Package snapshot implements tests for the snapshot functionality
package snapshot

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestCreate tests creating a snapshot
func TestCreate(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create some test files
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	snapshot, err := Create(tempDir)
	if err != nil {
		t.Fatalf("Unexpected error creating snapshot: %v", err)
	}
	
	if snapshot.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
	
	if snapshot.Version == "" {
		t.Error("Expected version to be set")
	}
	
	if snapshot.Module == "" {
		t.Error("Expected module to be set")
	}
	
	if snapshot.Dependencies == nil {
		t.Error("Expected dependencies map to be initialized")
	}
	
	if snapshot.BuildInfo == nil {
		t.Error("Expected build info map to be initialized")
	}
	
	if snapshot.SystemInfo == nil {
		t.Error("Expected system info map to be initialized")
	}
	
	if snapshot.Files == nil {
		t.Error("Expected files map to be initialized")
	}
}

// TestSaveAndLoad tests saving and loading a snapshot
func TestSaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create a basic snapshot
	snapshot := &Snapshot{
		Timestamp: time.Now(),
		Version:   "v1.0.0",
		GoVersion: "go1.25.6",
		Module:    "test.module",
		Dependencies: map[string]string{
			"dep1": "v1.0.0",
		},
		BuildInfo: map[string]interface{}{
			"build_time": time.Now().Format(time.RFC3339),
		},
		SystemInfo: map[string]interface{}{
			"os":   "linux",
			"arch": "amd64",
		},
		Files: map[string]FileSnapshot{
			"test.txt": {
				Path:    "test.txt",
				Size:    12,
				ModTime: time.Now().Unix(),
				Hash:    "test-hash",
			},
		},
	}
	
	snapshotPath := filepath.Join(tempDir, "test-snapshot.toml")
	err := snapshot.Save(snapshotPath)
	if err != nil {
		t.Fatalf("Unexpected error saving snapshot: %v", err)
	}
	
	// Check that file was created
	if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
		t.Fatal("Expected snapshot file to be created")
	}
	
	// Load the snapshot
	loadedSnapshot, err := Load(snapshotPath)
	if err != nil {
		t.Fatalf("Unexpected error loading snapshot: %v", err)
	}
	
	// Verify the loaded snapshot matches the original
	if loadedSnapshot.Version != snapshot.Version {
		t.Errorf("Expected version '%s', got '%s'", snapshot.Version, loadedSnapshot.Version)
	}
	
	if loadedSnapshot.Module != snapshot.Module {
		t.Errorf("Expected module '%s', got '%s'", snapshot.Module, loadedSnapshot.Module)
	}
	
	if len(loadedSnapshot.Dependencies) != len(snapshot.Dependencies) {
		t.Errorf("Expected %d dependencies, got %d", len(snapshot.Dependencies), len(loadedSnapshot.Dependencies))
	}
	
	if len(loadedSnapshot.Files) != len(snapshot.Files) {
		t.Errorf("Expected %d files, got %d", len(snapshot.Files), len(loadedSnapshot.Files))
	}
	
	// Check a specific file
	if _, exists := loadedSnapshot.Files["test.txt"]; !exists {
		t.Error("Expected 'test.txt' file to exist in loaded snapshot")
	}
}

// TestCompare tests comparing two snapshots
func TestCompare(t *testing.T) {
	baseTime := time.Now()
	
	snapshot1 := &Snapshot{
		Files: map[string]FileSnapshot{
			"file1.txt": {
				Path:    "file1.txt",
				Size:    10,
				ModTime: baseTime.Unix(),
				Hash:    "hash1",
			},
			"file2.txt": {
				Path:    "file2.txt",
				Size:    20,
				ModTime: baseTime.Unix(),
				Hash:    "hash2",
			},
		},
	}
	
	snapshot2 := &Snapshot{
		Files: map[string]FileSnapshot{
			"file1.txt": { // Same as in snapshot1
				Path:    "file1.txt",
				Size:    10,
				ModTime: baseTime.Unix(),
				Hash:    "hash1",
			},
			"file3.txt": { // New file
				Path:    "file3.txt",
				Size:    30,
				ModTime: baseTime.Unix(),
				Hash:    "hash3",
			},
			"file2.txt": { // Modified file
				Path:    "file2.txt",
				Size:    25, // Different size
				ModTime: baseTime.Unix(),
				Hash:    "hash2-modified",
			},
		},
	}
	
	result := snapshot1.Compare(snapshot2)
	
	// Check added files
	if len(result.Added) != 1 {
		t.Errorf("Expected 1 added file, got %d", len(result.Added))
	}
	if _, exists := result.Added["file3.txt"]; !exists {
		t.Error("Expected 'file3.txt' to be in added files")
	}
	
	// The above assertions are incorrect - if file2.txt exists in both snapshots but is modified,
	// it should be in Modified, not Removed. Let me fix the test.
	// Check that file2.txt is NOT in removed files since it exists in both snapshots
	if _, exists := result.Removed["file2.txt"]; exists {
		t.Error("Did not expect 'file2.txt' to be in removed files (it was modified, not removed)")
	}
	
	// Actually, looking at the logic, file2.txt is modified, not removed
	// Let me fix the test to reflect the actual behavior:
	// In the Compare function, if a file exists in both snapshots but differs,
	// it's added to Modified, not Removed. Only files that exist in the first
	// snapshot but not in the second are in Removed.

	// Reset for correct expectations:
	snapshot1 = &Snapshot{
		Files: map[string]FileSnapshot{
			"file1.txt": {
				Path:    "file1.txt",
				Size:    10,
				ModTime: baseTime.Unix(),
				Hash:    "hash1",
			},
			"file2.txt": {
				Path:    "file2.txt",
				Size:    20,
				ModTime: baseTime.Unix(),
				Hash:    "hash2",
			},
			"file4.txt": { // This will be removed
				Path:    "file4.txt",
				Size:    40,
				ModTime: baseTime.Unix(),
				Hash:    "hash4",
			},
		},
	}

	snapshot2 = &Snapshot{
		Files: map[string]FileSnapshot{
			"file1.txt": { // Same as in snapshot1
				Path:    "file1.txt",
				Size:    10,
				ModTime: baseTime.Unix(),
				Hash:    "hash1",
			},
			"file3.txt": { // New file
				Path:    "file3.txt",
				Size:    30,
				ModTime: baseTime.Unix(),
				Hash:    "hash3",
			},
			"file2.txt": { // Modified file
				Path:    "file2.txt",
				Size:    25, // Different size
				ModTime: baseTime.Unix(),
				Hash:    "hash2-modified", // Different hash
			},
		},
	}

	result = snapshot1.Compare(snapshot2)

	// Check added files
	if len(result.Added) != 1 {
		t.Errorf("Expected 1 added file, got %d", len(result.Added))
	}
	if _, exists := result.Added["file3.txt"]; !exists {
		t.Error("Expected 'file3.txt' to be in added files")
	}

	// Check removed files
	if len(result.Removed) != 1 {
		t.Errorf("Expected 1 removed file, got %d", len(result.Removed))
	}
	if _, exists := result.Removed["file4.txt"]; !exists {
		t.Error("Expected 'file4.txt' to be in removed files")
	}

	// Check modified files
	if len(result.Modified) != 1 {
		t.Errorf("Expected 1 modified file, got %d", len(result.Modified))
	}
	if _, exists := result.Modified["file2.txt"]; !exists {
		t.Error("Expected 'file2.txt' to be in modified files")
	}

	// Check that unchanged file is not in any category
	if _, exists := result.Added["file1.txt"]; exists {
		t.Error("Did not expect 'file1.txt' to be in added files (it was unchanged)")
	}
	if _, exists := result.Removed["file1.txt"]; exists {
		t.Error("Did not expect 'file1.txt' to be in removed files (it was unchanged)")
	}
	if _, exists := result.Modified["file1.txt"]; exists {
		t.Error("Did not expect 'file1.txt' to be in modified files (it was unchanged)")
	}
}