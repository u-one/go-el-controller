// +build !windows

package transport

import (
	"net"

	"golang.org/x/sys/unix"
)

func resolveSocketOption(conn *net.UDPConn) error {
	f, err := conn.File()
	if err != nil {
		return err
	}
	defer f.Close()

	fd := f.Fd()
	err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
	if err != nil {
		return err
	}

	err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
	if err != nil {
		return err
	}

	return nil
}
