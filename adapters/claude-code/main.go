// Package main implements the Claude Code adapter for amux.
// This adapter implements the WASM ABI per spec §10.4 for Claude Code agent integration.
//
// Required exports per spec §10.4.2:
// - amux_alloc
// - amux_free  
// - manifest
// - on_output
// - format_input
// - on_event
package main

import (
	"encoding/json"
	"unsafe"
)

// Memory management for the WASM ABI
var (
	allocatedBlocks = make(map[uint32][]byte)
	nextPtr         = uint32(1000) // Start at a safe offset
)

// AdapterManifest describes this adapter's capabilities per spec §10.2
type AdapterManifest struct {
	Name        string          `json:"name"`
	Version     string          `json:"version"`
	Description string          `json:"description,omitempty"`
	CLI         CLIRequirement  `json:"cli"`
	Patterns    AdapterPatterns `json:"patterns"`
	Commands    AdapterCommands `json:"commands"`
}

type CLIRequirement struct {
	Binary     string `json:"binary"`
	VersionCmd string `json:"version_cmd"`
	VersionRe  string `json:"version_re"`
	Constraint string `json:"constraint"`
}

type AdapterPatterns struct {
	Ready    string `json:"ready"`
	Error    string `json:"error"`
	Complete string `json:"complete"`
}

type AdapterCommands struct {
	Start       []string `json:"start"`
	SendMessage string   `json:"send_message"`
}

// manifest returns adapter metadata as JSON per spec §10.2
//export manifest
func manifest() uint64 {
	manifest := AdapterManifest{
		Name:        "claude-code",
		Version:     "1.0.0",
		Description: "Claude Code agent adapter",
		CLI: CLIRequirement{
			Binary:     "claude",
			VersionCmd: "claude --version",
			VersionRe:  `v(\d+\.\d+\.\d+)`,
			Constraint: ">=1.0.20 <2.0.0",
		},
		Patterns: AdapterPatterns{
			Ready:    "Ready for your request",
			Error:    "Error:",
			Complete: "Task completed",
		},
		Commands: AdapterCommands{
			Start:       []string{"claude", "chat"},
			SendMessage: "{{message}}",
		},
	}

	jsonBytes, err := json.Marshal(manifest)
	if err != nil {
		return 0
	}

	return packPtr(jsonBytes)
}

// on_output processes PTY output and returns events/actions per spec §10.4.3
//export on_output
func on_output(ptr, len uint32) uint64 {
	data := readInput(ptr, len)
	
	// Simple pattern matching for claude-code
	output := string(data)
	
	var response map[string]interface{}
	if containsPattern(output, "Ready for your request") {
		response = map[string]interface{}{
			"event": "agent_ready",
			"data":  map[string]interface{}{"status": "ready"},
		}
	} else if containsPattern(output, "Error:") {
		response = map[string]interface{}{
			"event": "agent_error",
			"data":  map[string]interface{}{"error": output},
		}
	} else if containsPattern(output, "Task completed") {
		response = map[string]interface{}{
			"event": "task_complete",
			"data":  map[string]interface{}{"status": "complete"},
		}
	} else {
		response = map[string]interface{}{
			"event": "output",
			"data":  map[string]interface{}{"text": output},
		}
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return 0
	}

	return packPtr(jsonBytes)
}

// format_input formats input for the Claude Code agent
//export format_input
func format_input(ptr, len uint32) uint64 {
	data := readInput(ptr, len)
	
	// Simple input formatting for claude-code
	formatted := string(data) + "\n"
	
	return packPtr([]byte(formatted))
}

// on_event handles system events
//export on_event
func on_event(ptr, len uint32) uint64 {
	data := readInput(ptr, len)
	
	// Parse event and respond accordingly
	var event map[string]interface{}
	if err := json.Unmarshal(data, &event); err != nil {
		return 0
	}

	response := map[string]interface{}{
		"event": "event_handled",
		"data":  event,
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return 0
	}

	return packPtr(jsonBytes)
}

// amux_alloc allocates memory for host-to-WASM communication per spec §10.4.1
//export amux_alloc
func amux_alloc(size uint32) uint32 {
	if size == 0 {
		return 0
	}
	
	block := make([]byte, size)
	ptr := nextPtr
	nextPtr += size + 64 // Add padding to avoid collisions
	
	allocatedBlocks[ptr] = block
	return ptr
}

// amux_free frees allocated memory per spec §10.4.1
//export amux_free
func amux_free(ptr, size uint32) {
	delete(allocatedBlocks, ptr)
}

// Helper functions

// packPtr packs data into memory and returns packed (ptr << 32 | len) per spec §10.4.1
func packPtr(data []byte) uint64 {
	if len(data) == 0 {
		return uint64(nextPtr) << 32 // Non-zero ptr with len=0
	}
	
	size := uint32(len(data))
	ptr := amux_alloc(size)
	if ptr == 0 {
		return 0
	}
	
	block := allocatedBlocks[ptr]
	copy(block, data)
	
	return uint64(ptr)<<32 | uint64(size)
}

// readInput reads input data from WASM memory
func readInput(ptr, len uint32) []byte {
	if len == 0 {
		return []byte{}
	}
	
	// Access WASM memory safely
	//lint:ignore SA1029 This is intentional for WASM memory access
	//nolint:govet // This unsafe.Pointer usage is necessary for WASM memory interface
	data := (*[1 << 30]byte)(unsafe.Pointer(uintptr(ptr)))[:len:len]
	return append([]byte(nil), data...)
}

// containsPattern checks if output contains a pattern (simple substring match)
func containsPattern(output, pattern string) bool {
	return len(output) >= len(pattern) && 
		   findSubstring(output, pattern) >= 0
}

// findSubstring finds substring without using strings package (WASM compatibility)
func findSubstring(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// main is required for Go WASM modules but not called in the WASM context
func main() {
	// This function is not called when loaded as a WASM module
}