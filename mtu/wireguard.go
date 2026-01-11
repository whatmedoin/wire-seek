package mtu

// WireGuard overhead constants
// WireGuard adds the following overhead to each packet:
// - IPv4/IPv6 header (20/40 bytes)
// - UDP header (8 bytes)
// - WireGuard header (32 bytes)
//
// Total overhead:
// - IPv4: 20 + 8 + 32 = 60 bytes
// - IPv6: 40 + 8 + 32 = 80 bytes
const (
	// WireGuardOverheadIPv4 is the overhead for WireGuard over IPv4
	// (20 bytes IPv4 + 8 bytes UDP + 32 bytes WireGuard)
	WireGuardOverheadIPv4 = 60

	// WireGuardOverheadIPv6 is the overhead for WireGuard over IPv6
	// (40 bytes IPv6 + 8 bytes UDP + 32 bytes WireGuard)
	WireGuardOverheadIPv6 = 80
)

// CalculateWireGuardMTU calculates the optimal WireGuard interface MTU
// given the path MTU and IP version
func CalculateWireGuardMTU(pathMTU int, isIPv6 bool) int {
	if isIPv6 {
		return pathMTU - WireGuardOverheadIPv6
	}
	return pathMTU - WireGuardOverheadIPv4
}

// CalculateMTU returns the appropriate MTU based on whether we're testing
// through an existing tunnel or to a WireGuard endpoint.
//
// When tunnelMode is true, the path MTU is returned as-is because the
// WireGuard overhead is already applied by the tunnel.
//
// When tunnelMode is false, the WireGuard overhead is subtracted to give
// the optimal MTU setting for a new tunnel.
func CalculateMTU(pathMTU int, isIPv6 bool, tunnelMode bool) int {
	if tunnelMode {
		return pathMTU
	}
	return CalculateWireGuardMTU(pathMTU, isIPv6)
}
