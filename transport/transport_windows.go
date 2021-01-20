// +build windows

package transport

import (
	"net"
)

func resolveSocketOption(conn *net.UDPConn) error {
	// Do nothing
	return nil
}
