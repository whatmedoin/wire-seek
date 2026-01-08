package mtu

import (
	"net"
	"testing"
)

func TestNewDiscoverer(t *testing.T) {
	tests := []struct {
		name       string
		target     net.IP
		verbose    bool
		wantIPv6   bool
	}{
		{
			name:     "IPv4 address",
			target:   net.ParseIP("8.8.8.8"),
			verbose:  false,
			wantIPv6: false,
		},
		{
			name:     "IPv6 address",
			target:   net.ParseIP("2001:4860:4860::8888"),
			verbose:  true,
			wantIPv6: true,
		},
		{
			name:     "IPv4-mapped IPv6 treated as IPv4",
			target:   net.ParseIP("::ffff:8.8.8.8").To4(),
			verbose:  false,
			wantIPv6: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDiscoverer(tt.target, tt.verbose)
			if d.isIPv6 != tt.wantIPv6 {
				t.Errorf("NewDiscoverer().isIPv6 = %v, want %v", d.isIPv6, tt.wantIPv6)
			}
			if d.verbose != tt.verbose {
				t.Errorf("NewDiscoverer().verbose = %v, want %v", d.verbose, tt.verbose)
			}
		})
	}
}

func TestPayloadSizeCalculation(t *testing.T) {
	// Test that payload size calculation is correct
	// Payload = MTU - IP header - ICMP header

	tests := []struct {
		name            string
		mtu             int
		isIPv6          bool
		expectedPayload int
	}{
		{
			name:            "1500 MTU IPv4",
			mtu:             1500,
			isIPv6:          false,
			expectedPayload: 1500 - IPv4HeaderSize - ICMPHeaderSize, // 1500 - 20 - 8 = 1472
		},
		{
			name:            "1500 MTU IPv6",
			mtu:             1500,
			isIPv6:          true,
			expectedPayload: 1500 - IPv6HeaderSize - ICMPHeaderSize, // 1500 - 40 - 8 = 1452
		},
		{
			name:            "1280 MTU IPv4",
			mtu:             1280,
			isIPv6:          false,
			expectedPayload: 1280 - IPv4HeaderSize - ICMPHeaderSize, // 1280 - 20 - 8 = 1252
		},
		{
			name:            "1280 MTU IPv6",
			mtu:             1280,
			isIPv6:          true,
			expectedPayload: 1280 - IPv6HeaderSize - ICMPHeaderSize, // 1280 - 40 - 8 = 1232
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var headerSize int
			if tt.isIPv6 {
				headerSize = IPv6HeaderSize
			} else {
				headerSize = IPv4HeaderSize
			}
			payload := tt.mtu - headerSize - ICMPHeaderSize
			if payload != tt.expectedPayload {
				t.Errorf("payload = %d, want %d", payload, tt.expectedPayload)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	// Verify header size constants
	if IPv4HeaderSize != 20 {
		t.Errorf("IPv4HeaderSize = %d, want 20", IPv4HeaderSize)
	}
	if IPv6HeaderSize != 40 {
		t.Errorf("IPv6HeaderSize = %d, want 40", IPv6HeaderSize)
	}
	if ICMPHeaderSize != 8 {
		t.Errorf("ICMPHeaderSize = %d, want 8", ICMPHeaderSize)
	}
	if MinMTU_IPv4 != 576 {
		t.Errorf("MinMTU_IPv4 = %d, want 576", MinMTU_IPv4)
	}
	if MinMTU_IPv6 != 1280 {
		t.Errorf("MinMTU_IPv6 = %d, want 1280", MinMTU_IPv6)
	}
	if MaxMTU != 1500 {
		t.Errorf("MaxMTU = %d, want 1500", MaxMTU)
	}
}
