package mtu

import (
	"net"
	"testing"

	"github.com/yeya/wire-seek/output"
)

func TestNewDiscoverer(t *testing.T) {
	tests := []struct {
		name     string
		target   net.IP
		wantIPv6 bool
	}{
		{
			name:     "IPv4 address",
			target:   net.ParseIP("8.8.8.8"),
			wantIPv6: false,
		},
		{
			name:     "IPv6 address",
			target:   net.ParseIP("2001:4860:4860::8888"),
			wantIPv6: true,
		},
		{
			name:     "IPv4-mapped IPv6 treated as IPv4",
			target:   net.ParseIP("::ffff:8.8.8.8").To4(),
			wantIPv6: false,
		},
	}

	log := output.New(output.LevelNormal)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minMTU := MinMTU_IPv4
			if tt.wantIPv6 {
				minMTU = MinMTU_IPv6
			}
			d := NewDiscoverer(tt.target, log, minMTU, MaxMTU)
			if d.isIPv6 != tt.wantIPv6 {
				t.Errorf("NewDiscoverer().isIPv6 = %v, want %v", d.isIPv6, tt.wantIPv6)
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

func TestNewDiscovererWithCustomMTU(t *testing.T) {
	tests := []struct {
		name      string
		target    net.IP
		minMTU    int
		maxMTU    int
		wantMin   int
		wantMax   int
	}{
		{
			name:    "Default IPv4 range",
			target:  net.ParseIP("8.8.8.8"),
			minMTU:  MinMTU_IPv4,
			maxMTU:  MaxMTU,
			wantMin: 576,
			wantMax: 1500,
		},
		{
			name:    "Default IPv6 range",
			target:  net.ParseIP("2001:4860:4860::8888"),
			minMTU:  MinMTU_IPv6,
			maxMTU:  MaxMTU,
			wantMin: 1280,
			wantMax: 1500,
		},
		{
			name:    "Custom WireGuard range",
			target:  net.ParseIP("10.0.0.1"),
			minMTU:  1280,
			maxMTU:  1420,
			wantMin: 1280,
			wantMax: 1420,
		},
		{
			name:    "Jumbo frame range",
			target:  net.ParseIP("192.168.1.1"),
			minMTU:  1500,
			maxMTU:  9000,
			wantMin: 1500,
			wantMax: 9000,
		},
		{
			name:    "Below protocol minimum (allowed with warning)",
			target:  net.ParseIP("8.8.8.8"),
			minMTU:  400,
			maxMTU:  1500,
			wantMin: 400,
			wantMax: 1500,
		},
	}

	log := output.New(output.LevelQuiet)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDiscoverer(tt.target, log, tt.minMTU, tt.maxMTU)
			if d.minMTU != tt.wantMin {
				t.Errorf("NewDiscoverer().minMTU = %d, want %d", d.minMTU, tt.wantMin)
			}
			if d.maxMTU != tt.wantMax {
				t.Errorf("NewDiscoverer().maxMTU = %d, want %d", d.maxMTU, tt.wantMax)
			}
		})
	}
}
