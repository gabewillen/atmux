package inference

import (
	"log"
	"path/filepath"
)

// NewDefaultEngine constructs the default local inference engine.
func NewDefaultEngine(repoRoot string, logger *log.Logger) (Engine, error) {
	root := filepath.Join(repoRoot, "third_party", "liquidgen")
	return NewLiquidgenEngine(root, logger)
}
