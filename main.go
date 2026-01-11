package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/yeya/wire-seek/mtu"
	"github.com/yeya/wire-seek/output"
)

func main() {
	var (
		target  string
		ipv6    bool
		verbose bool
		quiet   bool
		minMTU  int
		maxMTU  int
	)

	flag.StringVar(&target, "target", "", "Target host or IP address (required)")
	flag.BoolVar(&ipv6, "6", false, "Use IPv6 instead of IPv4")
	flag.BoolVar(&verbose, "v", false, "Verbose output (debug diagnostics)")
	flag.BoolVar(&quiet, "q", false, "Quiet output (only print MTU value, for scripting)")
	flag.IntVar(&minMTU, "min-mtu", 0, "Minimum MTU to test (default: 576 for IPv4, 1280 for IPv6)")
	flag.IntVar(&maxMTU, "max-mtu", 0, "Maximum MTU to test (default: 1500)")
	flag.Parse()

	if target == "" {
		// Check for positional argument
		if flag.NArg() > 0 {
			target = flag.Arg(0)
		} else {
			fmt.Fprintf(os.Stderr, "Usage: wire-seek [-6] [-v] [-q] [-min-mtu N] [-max-mtu N] -target <host>\n")
			fmt.Fprintf(os.Stderr, "       wire-seek [-6] [-v] [-q] [-min-mtu N] [-max-mtu N] <host>\n\n")
			fmt.Fprintf(os.Stderr, "Options:\n")
			flag.PrintDefaults()
			os.Exit(1)
		}
	}

	// Determine output level (quiet takes precedence if both are specified)
	level := output.LevelNormal
	if quiet {
		level = output.LevelQuiet
	} else if verbose {
		level = output.LevelVerbose
	}
	log := output.New(level)

	// Resolve target to IP address
	ip, err := resolveTarget(target, ipv6)
	if err != nil {
		log.Error("Error resolving target: %v\n", err)
		os.Exit(1)
	}

	isIPv6 := ip.To4() == nil

	log.Info("Wire-Seek: WireGuard MTU Optimizer\n")
	log.Info("Target: %s (%s)\n", target, ip.String())
	if isIPv6 {
		log.Info("Protocol: IPv6\n")
	} else {
		log.Info("Protocol: IPv4\n")
	}
	log.Info("\n")

	// Validate and set MTU range
	protocolMin := mtu.MinMTU_IPv4
	if isIPv6 {
		protocolMin = mtu.MinMTU_IPv6
	}

	if minMTU == 0 {
		minMTU = protocolMin
	}
	if maxMTU == 0 {
		maxMTU = mtu.MaxMTU
	}

	// Validate min <= max
	if minMTU > maxMTU {
		log.Error("Error: min-mtu (%d) cannot be greater than max-mtu (%d)\n", minMTU, maxMTU)
		os.Exit(1)
	}

	// Warn if min is below protocol minimum
	if minMTU < protocolMin {
		log.Info("Warning: min-mtu %d is below the protocol minimum (%d for %s)\n", minMTU, protocolMin, map[bool]string{true: "IPv6", false: "IPv4"}[isIPv6])
	}

	// Perform MTU discovery
	discoverer := mtu.NewDiscoverer(ip, log, minMTU, maxMTU)
	pathMTU, err := discoverer.FindPathMTU()
	if err != nil {
		log.Error("Error discovering MTU: %v\n", err)
		os.Exit(1)
	}

	// Calculate WireGuard MTU
	wgMTU := mtu.CalculateWireGuardMTU(pathMTU, isIPv6)

	log.Info("\n")
	log.Info("Results:\n")
	log.Info("  Path MTU:      %d bytes\n", pathMTU)
	log.Info("  WireGuard MTU: %d bytes\n", wgMTU)
	log.Info("\n")
	log.Info("Add to your WireGuard config:\n")
	log.Info("  MTU = %d\n", wgMTU)

	// In quiet mode, output only the MTU value
	if quiet {
		log.Result("%d\n", wgMTU)
	}
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
