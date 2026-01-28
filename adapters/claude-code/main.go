// Package main implements the Claude Code adapter for amux.
// This adapter will be compiled to WASM using TinyGo.
//
// Required exports per spec §8.2:
// - amux_alloc
// - amux_free  
// - manifest
// - on_output
// - format_input
// - on_event
package main

import "fmt"

// manifest returns adapter metadata as JSON.
//export manifest
func manifest() *byte {
	// Implementation deferred to Phase 8
	return nil
}

// on_output processes PTY output and returns events/actions.
//export on_output
func on_output(ptr, len uint32) uint64 {
	// Implementation deferred to Phase 8
	return 0
}

// format_input formats input for the agent.
//export format_input
func format_input(ptr, len uint32) uint64 {
	// Implementation deferred to Phase 8
	return 0
}

// on_event handles system events.
//export on_event
func on_event(ptr, len uint32) uint64 {
	// Implementation deferred to Phase 8
	return 0
}

// amux_alloc allocates memory for host-to-WASM communication.
//export amux_alloc
func amux_alloc(size uint32) uint32 {
	// Implementation deferred to Phase 8
	return 0
}

// amux_free frees allocated memory.
//export amux_free
func amux_free(ptr uint32) {
	// Implementation deferred to Phase 8
}

func main() {
	fmt.Println("Claude Code adapter - compile to WASM with TinyGo")
}