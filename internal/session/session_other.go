//go:build !unix

package session

import (
	"fmt"
	"os"
)

func sendTerminate(proc *os.Process) error {
	if proc == nil {
		return fmt.Errorf("terminate: %w", ErrSessionInvalid)
	}
	return proc.Signal(os.Interrupt)
}
