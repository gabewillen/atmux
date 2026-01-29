package adapter

import (
	"errors"
	"fmt"
	"os"

	"github.com/agentflare-ai/amux/internal/paths"
)

// FindWasmPath locates the WASM module for the named adapter.
func FindWasmPath(resolver *paths.Resolver, name string) (string, error) {
	if resolver == nil || name == "" {
		return "", fmt.Errorf("find adapter wasm: %w", ErrAdapterInvalid)
	}
	pathsToCheck := []string{
		resolver.ProjectAdapterWasmPath(name),
		resolver.UserAdapterWasmPath(name),
	}
	for _, path := range pathsToCheck {
		if path == "" {
			continue
		}
		if _, err := os.Stat(path); err == nil {
			return path, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("find adapter wasm: %w", err)
		}
	}
	return "", ErrAdapterNotFound
}
