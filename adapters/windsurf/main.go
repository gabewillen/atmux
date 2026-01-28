// Package main implements the Windsurf adapter for amux.
// This adapter will be compiled to WASM using TinyGo.
package main

import "fmt"

//export manifest
func manifest() *byte { return nil }

//export on_output
func on_output(ptr, len uint32) uint64 { return 0 }

//export format_input
func format_input(ptr, len uint32) uint64 { return 0 }

//export on_event
func on_event(ptr, len uint32) uint64 { return 0 }

//export amux_alloc
func amux_alloc(size uint32) uint32 { return 0 }

//export amux_free
func amux_free(ptr uint32) {}

func main() {
	fmt.Println("Windsurf adapter - compile to WASM with TinyGo")
}