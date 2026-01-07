//go:build linux

package mtu

import (
	"golang.org/x/net/ipv4"
)

// setDontFragmentIPv4 sets the Don't Fragment bit using ipv4 package
func setDontFragmentIPv4(p *ipv4.PacketConn) error {
	// On Linux, we use IP_PMTUDISC_DO via SetControlMessage
	// This enables path MTU discovery which effectively sets DF
	return p.SetControlMessage(ipv4.FlagDst|ipv4.FlagInterface, true)
}
