//go:build unix

package session

import (
	"os"
	"syscall"
)

func sendTerminate(proc *os.Process) error {
	return proc.Signal(syscall.SIGTERM)
}
