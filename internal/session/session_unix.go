//go:build unix

package session

import (
	"fmt"
	"os"
	"syscall"
)

func dupFile(file *os.File) (*os.File, error) {
	fd, err := syscall.Dup(int(file.Fd()))
	if err != nil {
		return nil, fmt.Errorf("dup file: %w", err)
	}
	return os.NewFile(uintptr(fd), file.Name()), nil
}

func sendTerminate(proc *os.Process) error {
	return proc.Signal(syscall.SIGTERM)
}
