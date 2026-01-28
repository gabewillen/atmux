// Package hook implements the exec hook library with CGO
package hook

/*
#include <sys/socket.h>
#include <sys/un.h>
*/
import "C"

import (
	"errors"
)

// ErrHook is returned when hook operations fail
var ErrHook = errors.New("hook operation failed")