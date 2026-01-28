package main

import (
	"fmt"
	"log"
	"os"

	"github.com/agentflare-ai/amux/internal/inference"
	"github.com/agentflare-ai/amux/internal/paths"
)

func main() {
	repoRoot, err := paths.FindRepoRoot(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	logger := log.New(os.Stderr, "amux-node ", log.LstdFlags)
	if _, err := inference.NewDefaultEngine(repoRoot, logger); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, "amux-node stub: Phase 0")
}
