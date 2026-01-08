//go:build windows

package mtu

import (
	"syscall"

	"golang.org/x/net/ipv4"
)

const (
	// IP_DONTFRAGMENT is the Windows socket option for Don't Fragment
	IP_DONTFRAGMENT = 14
)

// setDontFragmentIPv4 sets the Don't Fragment bit on Windows
func setDontFragmentIPv4(p *ipv4.PacketConn) error {
	conn := p.PacketConn

	// Get syscall.RawConn from the underlying connection
	type rawConner interface {
		SyscallConn() (syscall.RawConn, error)
	}

	rc, ok := conn.(rawConner)
	if !ok {
		return nil // Can't get raw conn, silently skip
	}

	rawConn, err := rc.SyscallConn()
	if err != nil {
		return err
	}

	var setsockoptErr error
	controlErr := rawConn.Control(func(fd uintptr) {
		setsockoptErr = syscall.SetsockoptInt(syscall.Handle(fd), syscall.IPPROTO_IP, IP_DONTFRAGMENT, 1)
	})
	if controlErr != nil {
		return controlErr
	}
	return setsockoptErr
}
