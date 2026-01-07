//go:build !linux && !darwin

package mtu

import (
	"golang.org/x/net/ipv4"
)

// setDontFragmentIPv4 is a no-op on unsupported platforms
func setDontFragmentIPv4(p *ipv4.PacketConn) error {
	_ = p // silence unused warning
	return nil
}
