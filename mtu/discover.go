package mtu

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

const (
	// MinMTU_IPv4 is the minimum MTU for IPv4 networks (RFC 791)
	MinMTU_IPv4 = 576
	// MinMTU_IPv6 is the minimum MTU for IPv6 networks (RFC 8200)
	MinMTU_IPv6 = 1280
	// MaxMTU is the standard Ethernet MTU
	MaxMTU = 1500
	// ICMPHeaderSize is the size of ICMP header
	ICMPHeaderSize = 8
	// IPv4HeaderSize is the size of IPv4 header
	IPv4HeaderSize = 20
	// IPv6HeaderSize is the size of IPv6 header
	IPv6HeaderSize = 40

	// pingTimeout is the timeout for each ping attempt
	pingTimeout = 2 * time.Second
)

// Discoverer handles MTU discovery operations
type Discoverer struct {
	target  net.IP
	isIPv6  bool
	verbose bool
}

// NewDiscoverer creates a new MTU discoverer for the given target
func NewDiscoverer(target net.IP, verbose bool) *Discoverer {
	return &Discoverer{
		target:  target,
		isIPv6:  target.To4() == nil,
		verbose: verbose,
	}
}

// FindPathMTU uses binary search to find the path MTU to the target
func (d *Discoverer) FindPathMTU() (int, error) {
	low := MinMTU_IPv4
	if d.isIPv6 {
		low = MinMTU_IPv6
	}
	high := MaxMTU

	fmt.Printf("Discovering path MTU (range: %d-%d)...\n", low, high)

	// First, verify we can reach the target at all
	if !d.canSendSize(low) {
		return 0, fmt.Errorf("cannot reach target even with minimum MTU (%d)", low)
	}

	// Binary search for optimal MTU
	for low < high-1 {
		mid := (low + high + 1) / 2

		if d.verbose {
			fmt.Printf("  Testing MTU %d... ", mid)
		}

		if d.canSendSize(mid) {
			if d.verbose {
				fmt.Printf("OK\n")
			}
			low = mid
		} else {
			if d.verbose {
				fmt.Printf("Too large\n")
			}
			high = mid - 1
		}
	}

	// Final verification
	if d.canSendSize(high) {
		return high, nil
	}
	return low, nil
}

// canSendSize tests if a packet of the given size can reach the target
// without fragmentation
func (d *Discoverer) canSendSize(size int) bool {
	var headerSize int
	if d.isIPv6 {
		headerSize = IPv6HeaderSize
	} else {
		headerSize = IPv4HeaderSize
	}

	// Calculate payload size (MTU - IP header - ICMP header)
	payloadSize := size - headerSize - ICMPHeaderSize
	if payloadSize < 0 {
		return false
	}

	return d.pingWithDF(payloadSize)
}

// pingWithDF sends an ICMP echo request with the Don't Fragment bit set
func (d *Discoverer) pingWithDF(payloadSize int) bool {
	var (
		network string
		proto   int
		msgType icmp.Type
	)

	if d.isIPv6 {
		network = "ip6:ipv6-icmp"
		proto = 58 // ICMPv6
		msgType = ipv6.ICMPTypeEchoRequest
	} else {
		network = "ip4:icmp"
		proto = 1 // ICMP
		msgType = ipv4.ICMPTypeEcho
	}

	// Create ICMP connection
	conn, err := icmp.ListenPacket(network, "")
	if err != nil {
		// Try unprivileged ICMP (Linux 3.0+)
		if d.isIPv6 {
			network = "udp6"
		} else {
			network = "udp4"
		}
		conn, err = icmp.ListenPacket(network, "")
		if err != nil {
			if d.verbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to create ICMP socket: %v\n", err)
			}
			return false
		}
		if d.verbose {
			fmt.Fprintf(os.Stderr, "  Using unprivileged ICMP (%s)\n", network)
		}
	} else if d.verbose {
		fmt.Fprintf(os.Stderr, "  Using raw ICMP socket (%s)\n", network)
	}
	defer conn.Close()

	// Set Don't Fragment bit for IPv4 using the ipv4 package
	if !d.isIPv6 {
		p := conn.IPv4PacketConn()
		if err := setDontFragmentIPv4(p); err != nil && d.verbose {
			fmt.Fprintf(os.Stderr, "Warning: failed to set DF bit: %v\n", err)
		}
	}

	// Build ICMP message
	msg := icmp.Message{
		Type: msgType,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: make([]byte, payloadSize),
		},
	}

	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		if d.verbose {
			fmt.Fprintf(os.Stderr, "  Error marshaling ICMP message: %v\n", err)
		}
		return false
	}

	// Set deadline
	if err := conn.SetDeadline(time.Now().Add(pingTimeout)); err != nil {
		if d.verbose {
			fmt.Fprintf(os.Stderr, "  Error setting deadline: %v\n", err)
		}
		return false
	}

	// Determine destination address
	var dst net.Addr
	if d.isIPv6 {
		dst = &net.UDPAddr{IP: d.target}
	} else {
		dst = &net.UDPAddr{IP: d.target}
	}

	// For raw sockets, use IPAddr
	if network == "ip4:icmp" || network == "ip6:ipv6-icmp" {
		dst = &net.IPAddr{IP: d.target}
	}

	// Send packet
	if _, err := conn.WriteTo(msgBytes, dst); err != nil {
		if d.verbose {
			fmt.Fprintf(os.Stderr, "  Error sending packet: %v\n", err)
		}
		return false
	}

	// Wait for reply
	reply := make([]byte, 1500)
	n, _, err := conn.ReadFrom(reply)
	if err != nil {
		if d.verbose {
			fmt.Fprintf(os.Stderr, "  Error reading reply: %v\n", err)
		}
		return false
	}

	// Parse reply
	rm, err := icmp.ParseMessage(proto, reply[:n])
	if err != nil {
		if d.verbose {
			fmt.Fprintf(os.Stderr, "  Error parsing reply: %v\n", err)
		}
		return false
	}

	// Check if it's an echo reply
	if d.isIPv6 {
		success := rm.Type == ipv6.ICMPTypeEchoReply
		if d.verbose && !success {
			fmt.Fprintf(os.Stderr, "  Received unexpected ICMP type: %v\n", rm.Type)
		}
		return success
	}
	success := rm.Type == ipv4.ICMPTypeEchoReply
	if d.verbose && !success {
		fmt.Fprintf(os.Stderr, "  Received unexpected ICMP type: %v\n", rm.Type)
	}
	return success
}
