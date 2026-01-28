package main

import (
	"net"
	"os"
	"syscall"
)

// SendFD sends a file descriptor over a Unix domain socket.
func SendFD(conn *net.UnixConn, fd int) error {
	rights := syscall.UnixRights(fd)
	// Send a dummy byte because some systems expect at least 1 byte of data with OOB.
	dummy := []byte{0}
	_, _, err := conn.WriteMsgUnix(dummy, rights, nil)
	return err
}

// RecvFD receives a file descriptor from a Unix domain socket.
func RecvFD(conn *net.UnixConn) (int, error) {
	// We expect 1 dummy byte and the OOB data.
	buf := make([]byte, 1)
	oob := make([]byte, syscall.CmsgSpace(4)) // 4 bytes for one FD

	_, oobn, _, _, err := conn.ReadMsgUnix(buf, oob)
	if err != nil {
		return -1, err
	}

	msgs, err := syscall.ParseSocketControlMessage(oob[:oobn])
	if err != nil {
		return -1, err
	}

	if len(msgs) > 0 {
		fds, err := syscall.ParseUnixRights(&msgs[0])
		if err != nil {
			return -1, err
		}
		if len(fds) > 0 {
			return fds[0], nil
		}
	}

	return -1, os.ErrInvalid
}
