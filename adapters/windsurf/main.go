// Package main implements the Windsurf adapter for amux
package main

import (
	"unsafe"
)

//export amux_alloc
func amux_alloc(size uint32) *byte {
	// Allocate memory and return pointer
	ptr := make([]byte, size)
	return &ptr[0]
}

//export amux_free
func amux_free(ptr unsafe.Pointer, size uint32) {
	// Free memory (handled by GC in TinyGo)
}

//export manifest
func manifest() uint64 {
	// Return pointer and length of manifest JSON
	manifestStr := `{
		"name": "windsurf",
		"version": "v1.0.0",
		"description": "Adapter for Windsurf agent",
		"patterns": [
			"waiting for input",
			"rate limit",
			"error occurred",
			"task completed",
			"session started"
		],
		"actions": [
			"send_command",
			"wait_for_response",
			"handle_rate_limit",
			"start_session"
		]
	}`
	
	ptr := []byte(manifestStr)
	ptrPtr := &ptr[0]
	length := len(ptr)
	
	return (uint64(uintptr(unsafe.Pointer(ptrPtr))) << 32) | uint64(length)
}

//export on_output
func on_output(output_ptr uintptr, output_len uint32) uint64 {
	// Process output from Windsurf agent
	output := string((*[1 << 32]byte)(unsafe.Pointer(output_ptr))[:output_len:output_len])
	
	// Basic pattern matching for Windsurf
	var action string
	if contains(output, "waiting for input") {
		action = `{"type": "send_command", "data": {"command": "continue"}}`
	} else if contains(output, "rate limit") {
		action = `{"type": "handle_rate_limit", "data": {}}`
	} else if contains(output, "error occurred") {
		action = `{"type": "handle_error", "data": {"error": "` + output + `"}}`
	} else if contains(output, "task completed") {
		action = `{"type": "task_completed", "data": {}}`
	} else if contains(output, "session started") {
		action = `{"type": "session_started", "data": {}}`
	} else {
		// No specific action needed
		action = "{}"
	}
	
	ptr := []byte(action)
	ptrPtr := &ptr[0]
	length := len(ptr)
	
	return (uint64(uintptr(unsafe.Pointer(ptrPtr))) << 32) | uint64(length)
}

//export format_input
func format_input(input_ptr uintptr, input_len uint32) uint64 {
	// Format input for Windsurf agent
	input := string((*[1 << 32]byte)(unsafe.Pointer(input_ptr))[:input_len:input_len])
	
	// For Windsurf, we might format the input differently
	formattedInput := "WINDSURF_INPUT_START\n" + input + "\nWINDSURF_INPUT_END"
	
	ptr := []byte(formattedInput)
	ptrPtr := &ptr[0]
	length := len(ptr)
	
	return (uint64(uintptr(unsafe.Pointer(ptrPtr))) << 32) | uint64(length)
}

//export on_event
func on_event(event_ptr uintptr, event_len uint32) uint64 {
	// Handle events specific to Windsurf
	_ = string((*[1 << 32]byte)(unsafe.Pointer(event_ptr))[:event_len:event_len])

	// For now, just return an empty action
	action := "{}"

	ptr := []byte(action)
	ptrPtr := &ptr[0]
	length := len(ptr)

	return (uint64(uintptr(unsafe.Pointer(ptrPtr))) << 32) | uint64(length)
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && find(s, substr)
}

// Helper function to find a substring
func find(s, substr string) bool {
	sLen := len(s)
	substrLen := len(substr)
	
	if substrLen == 0 {
		return true
	}
	
	for i := 0; i <= sLen-substrLen; i++ {
		match := true
		for j := 0; j < substrLen; j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	
	return false
}

func main() {
	// No-op main function
}