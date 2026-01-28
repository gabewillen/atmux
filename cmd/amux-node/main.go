// Package main provides the unified daemon binary.
// This binary serves as both amuxd (daemon) and amux-manager roles,
// with the active role determined by configuration and/or flags.
package main

import (
	"fmt"
	"log"
	"os"
)

const version = "v1.22.0-phase0"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("amux-node %s\n", version)
		return
	}

	log.Printf("amux-node daemon %s starting...", version)
	fmt.Println("amux-node: phase 0 skeleton - daemon not yet implemented")
	os.Exit(1)
}