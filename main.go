package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/yeya/wire-seek/mtu"
)

func main() {
	var (
		target  string
		ipv6    bool
		verbose bool
	)

	flag.StringVar(&target, "target", "", "Target host or IP address (required)")
	flag.BoolVar(&ipv6, "6", false, "Use IPv6 instead of IPv4")
	flag.BoolVar(&verbose, "v", false, "Verbose output")
	flag.Parse()

	if target == "" {
		// Check for positional argument
		if flag.NArg() > 0 {
			target = flag.Arg(0)
		} else {
			fmt.Fprintf(os.Stderr, "Usage: wire-seek [-6] [-v] -target <host>\n")
			fmt.Fprintf(os.Stderr, "       wire-seek [-6] [-v] <host>\n\n")
			fmt.Fprintf(os.Stderr, "Options:\n")
			flag.PrintDefaults()
			os.Exit(1)
		}
	}

	// Resolve target to IP address
	ip, err := resolveTarget(target, ipv6)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving target: %v\n", err)
		os.Exit(1)
	}

	isIPv6 := ip.To4() == nil

	fmt.Printf("Wire-Seek: WireGuard MTU Optimizer\n")
	fmt.Printf("Target: %s (%s)\n", target, ip.String())
	if isIPv6 {
		fmt.Printf("Protocol: IPv6\n")
	} else {
		fmt.Printf("Protocol: IPv4\n")
	}
	fmt.Println()

	// Perform MTU discovery
	discoverer := mtu.NewDiscoverer(ip, verbose)
	pathMTU, err := discoverer.FindPathMTU()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error discovering MTU: %v\n", err)
		os.Exit(1)
	}

	// Calculate WireGuard MTU
	wgMTU := mtu.CalculateWireGuardMTU(pathMTU, isIPv6)

	fmt.Println()
	fmt.Printf("Results:\n")
	fmt.Printf("  Path MTU:      %d bytes\n", pathMTU)
	fmt.Printf("  WireGuard MTU: %d bytes\n", wgMTU)
	fmt.Println()
	fmt.Printf("Add to your WireGuard config:\n")
	fmt.Printf("  MTU = %d\n", wgMTU)
}

// resolveTarget resolves a hostname or IP string to a net.IP
func resolveTarget(target string, preferIPv6 bool) (net.IP, error) {
	// First try parsing as IP directly
	if ip := net.ParseIP(target); ip != nil {
		return ip, nil
	}

	// Resolve hostname
	ips, err := net.LookupIP(target)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve %s: %w", target, err)
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no IP addresses found for %s", target)
	}

	// Find preferred IP version
	for _, ip := range ips {
		isV6 := ip.To4() == nil
		if isV6 == preferIPv6 {
			return ip, nil
		}
	}

	// Return first available if preferred version not found
	return ips[0], nil
}
