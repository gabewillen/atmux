// Package adapter implements WASM adapter runtime (loads any adapter)
package adapter

import (
	"errors"

	// Import wazero to ensure it's included in the dependencies
	_ "github.com/tetratelabs/wazero"
)

// ErrAdapterNotFound is returned when an adapter is not found
var ErrAdapterNotFound = errors.New("adapter not found")