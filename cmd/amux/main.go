// Package main provides the amux CLI client.
// The CLI communicates with the amux daemon (amuxd) over JSON-RPC.
package main

import (
	"fmt"
	"log"
	"os"
)

const version = "v1.22.0-phase0"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("amux %s\n", version)
		return
	}

	log.Printf("amux CLI client %s starting...", version)
	fmt.Println("amux: phase 0 skeleton - CLI not yet implemented")
	os.Exit(1)
}