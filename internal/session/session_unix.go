//go:build unix

package session

import (
	"fmt"
	"os"
	"syscall"
)

func sendTerminate(proc *os.Process) error {
	return proc.Signal(syscall.SIGTERM)
}
