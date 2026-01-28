package spec

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/agentflare-ai/amux/internal/errors"
)

// expectedVersion is the authoritative version required by this plan/implementation.
const expectedVersion = "Version: v1.22"

// Verify checks that spec-v1.22.md exists and matches the expected version.
// Spec §4.3.1
func Verify(root string) error {
	path := filepath.Join(root, "docs", "spec-v1.22.md")
	f, err := os.Open(path)
	if err != nil {
		return errors.Wrap(err, "spec-v1.22.md not found in docs/")
	}
	defer f.Close()

	// Check first few lines for version
	scanner := bufio.NewScanner(f)
	found := false
	for i := 0; i < 20 && scanner.Scan(); i++ {
		line := scanner.Text()
		if strings.Contains(line, "**Version:** v1.22") || strings.Contains(line, expectedVersion) {
			found = true
			break
		}
	}

	if !found {
		return errors.New("spec-v1.22.md found but version mismatch or missing header")
	}

	return nil
}
