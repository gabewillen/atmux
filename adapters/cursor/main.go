// Package main implements the Cursor adapter for amux.
// This adapter implements the WASM ABI per spec §10.4 for Cursor agent integration.
package main

import (
	"encoding/json"
	"unsafe"
)

// Memory management for the WASM ABI
var (
	allocatedBlocks = make(map[uint32][]byte)
	nextPtr         = uint32(1000)
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

//export manifest
func manifest() uint64 {
	manifest := AdapterManifest{
		Name:        "cursor",
		Version:     "1.0.0",
		Description: "Cursor IDE agent adapter",
		CLI: CLIRequirement{
			Binary:     "cursor",
			VersionCmd: "cursor --version",
			VersionRe:  `(\d+\.\d+\.\d+)`,
			Constraint: ">=0.40.0",
		},
		Patterns: AdapterPatterns{
			Ready:    "Cursor>",
			Error:    "Error:",
			Complete: "Done",
		},
		Commands: AdapterCommands{
			Start:       []string{"cursor", "--cli"},
			SendMessage: "{{message}}",
		},
	}

	jsonBytes, err := json.Marshal(manifest)
	if err != nil {
		return 0
	}
	return packPtr(jsonBytes)
}

//export on_output
func on_output(ptr, len uint32) uint64 {
	data := readInput(ptr, len)
	output := string(data)
	
	var response map[string]interface{}
	if containsPattern(output, "Cursor>") {
		response = map[string]interface{}{
			"event": "agent_ready",
			"data":  map[string]interface{}{"status": "ready"},
		}
	} else if containsPattern(output, "Error:") {
		response = map[string]interface{}{
			"event": "agent_error",
			"data":  map[string]interface{}{"error": output},
		}
	} else if containsPattern(output, "Done") {
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

//export format_input
func format_input(ptr, len uint32) uint64 {
	data := readInput(ptr, len)
	formatted := string(data) + "\n"
	return packPtr([]byte(formatted))
}

//export on_event
func on_event(ptr, len uint32) uint64 {
	data := readInput(ptr, len)
	
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

//export amux_alloc
func amux_alloc(size uint32) uint32 {
	if size == 0 {
		return 0
	}
	
	block := make([]byte, size)
	ptr := nextPtr
	nextPtr += size + 64
	
	allocatedBlocks[ptr] = block
	return ptr
}

//export amux_free
func amux_free(ptr, size uint32) {
	delete(allocatedBlocks, ptr)
}

// Helper functions
func packPtr(data []byte) uint64 {
	if len(data) == 0 {
		return uint64(nextPtr) << 32
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

func readInput(ptr, len uint32) []byte {
	if len == 0 {
		return []byte{}
	}
	// Access WASM memory directly - this is safe in WASM context
	// where ptr is guaranteed to be valid by the runtime
	// Access WASM memory safely
	//lint:ignore SA1029 This is intentional for WASM memory access
	data := (*[1 << 30]byte)(unsafe.Pointer(uintptr(ptr)))[:len:len]
	return append([]byte(nil), data...)
}

func containsPattern(output, pattern string) bool {
	return len(output) >= len(pattern) && findSubstring(output, pattern) >= 0
}

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

func main() {}