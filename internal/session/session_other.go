//go:build !unix

package session

import (
	"fmt"
	"os"
)

func dupFile(file *os.File) (*os.File, error) {
	if file == nil {
		return nil, fmt.Errorf("dup file: %w", ErrSessionInvalid)
	}
	return os.NewFile(file.Fd(), file.Name()), nil
}

func sendTerminate(proc *os.Process) error {
	if proc == nil {
		return fmt.Errorf("terminate: %w", ErrSessionInvalid)
	}
	return proc.Signal(os.Interrupt)
}
