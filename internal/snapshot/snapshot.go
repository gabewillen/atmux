package snapshot

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/pelletier/go-toml/v2"
)

type Snapshot struct {
	Meta   MetaSnapshot   `toml:"meta"`
	System SystemSnapshot `toml:"system"`
	Config config.Config  `toml:"config"`
}

type MetaSnapshot struct {
	Timestamp string `toml:"timestamp"`
	GoVersion string `toml:"go_version"`
}

type SystemSnapshot struct {
	OS   string `toml:"os"`
	Arch string `toml:"arch"`
}

func Capture() (*Snapshot, error) {
	// Load config
	cwd, _ := os.Getwd()
	cfg, err := config.Load(cwd)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &Snapshot{
		Meta: MetaSnapshot{
			Timestamp: time.Now().Format(time.RFC3339),
			GoVersion: runtime.Version(),
		},
		System: SystemSnapshot{
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
		},
		Config: *cfg,
	}, nil
}

func Save(path string, snap *Snapshot) error {
	data, err := toml.Marshal(snap)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func Load(path string) (*Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var snap Snapshot
	if err := toml.Unmarshal(data, &snap); err != nil {
		return nil, err
	}
	
	// Normalize maps
	if snap.Config.Adapters == nil {
		snap.Config.Adapters = make(map[string]map[string]any)
	}
	
	return &snap, nil
}

func Compare(old, new *Snapshot) error {
	// Normalize timestamps
	old.Meta.Timestamp = ""
	new.Meta.Timestamp = ""
	
	oldData, err := toml.Marshal(old)
	if err != nil {
		return err
	}
	newData, err := toml.Marshal(new)
	if err != nil {
		return err
	}
	
	if string(oldData) != string(newData) {
		return fmt.Errorf("snapshot mismatch: \nOLD:\n%s\nNEW:\n%s", string(oldData), string(newData))
	}
	return nil
}
