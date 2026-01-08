//go:build linux

package mtu

import (
	"syscall"

	"golang.org/x/net/ipv4"
)

const (
	// IP_MTU_DISCOVER is the Linux socket option for MTU discovery
	IP_MTU_DISCOVER = 10
	// IP_PMTUDISC_DO enables PMTU discovery and sets Don't Fragment bit
	IP_PMTUDISC_DO = 2
)

// setDontFragmentIPv4 sets the Don't Fragment bit on Linux
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
		setsockoptErr = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, IP_MTU_DISCOVER, IP_PMTUDISC_DO)
	})
	if controlErr != nil {
		return controlErr
	}
	return setsockoptErr
}
