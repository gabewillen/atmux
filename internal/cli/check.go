package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

func CheckSpec(repoRoot string) error {
	specPath := filepath.Join(repoRoot, "docs", "spec-v1.22.md")
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		return fmt.Errorf("spec-v1.22.md missing at %s", specPath)
	}
	return nil
}
