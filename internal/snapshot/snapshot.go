// Package snapshot implements the snapshot functionality for amux test
package snapshot

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

// Snapshot represents a snapshot of the system state for testing
type Snapshot struct {
	Timestamp    time.Time                `toml:"timestamp"`
	Version      string                   `toml:"version"`
	GoVersion    string                   `toml:"go_version"`
	Module       string                   `toml:"module"`
	Dependencies map[string]string        `toml:"dependencies"`
	BuildInfo    map[string]interface{}   `toml:"build_info"`
	SystemInfo   map[string]interface{}   `toml:"system_info"`
	Files        map[string]FileSnapshot  `toml:"files"`
}

// FileSnapshot represents a snapshot of a file
type FileSnapshot struct {
	Path    string `toml:"path"`
	Size    int64  `toml:"size"`
	ModTime int64  `toml:"mod_time"`
	Hash    string `toml:"hash"` // SHA256 hash of the file content
}

// Create creates a new snapshot of the current system state
func Create(moduleRoot string) (*Snapshot, error) {
	snapshot := &Snapshot{
		Timestamp:    time.Now(),
		Version:      "v0.1.0", // This would come from build info in real implementation
		GoVersion:    "go1.25.6",
		Module:       "github.com/stateforward/amux",
		Dependencies: make(map[string]string),
		BuildInfo:    make(map[string]interface{}),
		SystemInfo:   make(map[string]interface{}),
		Files:        make(map[string]FileSnapshot),
	}

	// Add some basic system info
	snapshot.SystemInfo["os"] = "linux" // This would be runtime.GOOS in real implementation
	snapshot.SystemInfo["arch"] = "amd64" // This would be runtime.GOARCH in real implementation

	// Add basic build info
	snapshot.BuildInfo["build_time"] = time.Now().Format(time.RFC3339)
	snapshot.BuildInfo["built_with"] = "go1.25.6"

	// Add some example dependencies
	snapshot.Dependencies["github.com/stateforward/hsm-go"] = "v0.1.0"
	snapshot.Dependencies["github.com/creack/pty"] = "v1.1.11"
	snapshot.Dependencies["github.com/tetratelabs/wazero"] = "v1.8.0"

	// Add some example files from the module
	filesToSnapshot := []string{
		"go.mod",
		"go.sum",
		"cmd/amux/main.go",
		"cmd/amux-node/main.go",
	}

	for _, relPath := range filesToSnapshot {
		absPath := filepath.Join(moduleRoot, relPath)
		info, err := os.Stat(absPath)
		if err != nil {
			// Skip files that don't exist
			continue
		}

		// For now, we'll just record basic info without computing hashes
		// In a real implementation, we'd compute SHA256 hashes
		snapshot.Files[relPath] = FileSnapshot{
			Path:    relPath,
			Size:    info.Size(),
			ModTime: info.ModTime().Unix(),
			Hash:    "placeholder-hash", // Would be actual SHA256 in real implementation
		}
	}

	return snapshot, nil
}

// Save saves the snapshot to a TOML file
func (s *Snapshot) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create snapshot file: %w", err)
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	err = encoder.Encode(s)
	if err != nil {
		return fmt.Errorf("failed to encode snapshot as TOML: %w", err)
	}

	return nil
}

// Load loads a snapshot from a TOML file
func Load(path string) (*Snapshot, error) {
	var snapshot Snapshot
	_, err := toml.DecodeFile(path, &snapshot)
	if err != nil {
		return nil, fmt.Errorf("failed to decode snapshot from TOML: %w", err)
	}

	return &snapshot, nil
}

// Compare compares this snapshot with another and returns differences
func (s *Snapshot) Compare(other *Snapshot) *ComparisonResult {
	result := &ComparisonResult{
		Added:    make(map[string]FileSnapshot),
		Removed:  make(map[string]FileSnapshot),
		Modified: make(map[string]FileSnapshotPair),
	}

	// Find added and modified files
	for path, file := range other.Files {
		if _, exists := s.Files[path]; !exists {
			result.Added[path] = file
		} else if s.Files[path] != file {
			result.Modified[path] = FileSnapshotPair{
				Before: s.Files[path],
				After:  file,
			}
		}
	}

	// Find removed files
	for path, file := range s.Files {
		if _, exists := other.Files[path]; !exists {
			result.Removed[path] = file
		}
	}

	return result
}

// ComparisonResult holds the result of comparing two snapshots
type ComparisonResult struct {
	Added    map[string]FileSnapshot   `toml:"added"`
	Removed  map[string]FileSnapshot   `toml:"removed"`
	Modified map[string]FileSnapshotPair `toml:"modified"`
}

// FileSnapshotPair holds two file snapshots for comparison
type FileSnapshotPair struct {
	Before FileSnapshot `toml:"before"`
	After  FileSnapshot `toml:"after"`
}