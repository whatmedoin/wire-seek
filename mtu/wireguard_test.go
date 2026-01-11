package mtu

import "testing"

func TestCalculateWireGuardMTU(t *testing.T) {
	tests := []struct {
		name     string
		pathMTU  int
		isIPv6   bool
		expected int
	}{
		{
			name:     "standard ethernet IPv4",
			pathMTU:  1500,
			isIPv6:   false,
			expected: 1440, // 1500 - 60
		},
		{
			name:     "standard ethernet IPv6",
			pathMTU:  1500,
			isIPv6:   true,
			expected: 1420, // 1500 - 80
		},
		{
			name:     "minimum MTU IPv4",
			pathMTU:  1280,
			isIPv6:   false,
			expected: 1220, // 1280 - 60
		},
		{
			name:     "minimum MTU IPv6",
			pathMTU:  1280,
			isIPv6:   true,
			expected: 1200, // 1280 - 80
		},
		{
			name:     "jumbo frame IPv4",
			pathMTU:  9000,
			isIPv6:   false,
			expected: 8940, // 9000 - 60
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateWireGuardMTU(tt.pathMTU, tt.isIPv6)
			if result != tt.expected {
				t.Errorf("CalculateWireGuardMTU(%d, %v) = %d, want %d",
					tt.pathMTU, tt.isIPv6, result, tt.expected)
			}
		})
	}
}

func TestWireGuardOverheadConstants(t *testing.T) {
	// Verify overhead constants match the specification:
	// IPv4: 20 (IP) + 8 (UDP) + 32 (WG) = 60
	// IPv6: 40 (IP) + 8 (UDP) + 32 (WG) = 80
	if WireGuardOverheadIPv4 != 60 {
		t.Errorf("WireGuardOverheadIPv4 = %d, want 60", WireGuardOverheadIPv4)
	}
	if WireGuardOverheadIPv6 != 80 {
		t.Errorf("WireGuardOverheadIPv6 = %d, want 80", WireGuardOverheadIPv6)
	}
}

func TestCalculateMTU(t *testing.T) {
	tests := []struct {
		name       string
		pathMTU    int
		isIPv6     bool
		tunnelMode bool
		expected   int
	}{
		{
			name:       "tunnel mode IPv4 - no overhead subtraction",
			pathMTU:    1420,
			isIPv6:     false,
			tunnelMode: true,
			expected:   1420,
		},
		{
			name:       "tunnel mode IPv6 - no overhead subtraction",
			pathMTU:    1400,
			isIPv6:     true,
			tunnelMode: true,
			expected:   1400,
		},
		{
			name:       "endpoint mode IPv4 - subtracts 60 bytes",
			pathMTU:    1500,
			isIPv6:     false,
			tunnelMode: false,
			expected:   1440,
		},
		{
			name:       "endpoint mode IPv6 - subtracts 80 bytes",
			pathMTU:    1500,
			isIPv6:     true,
			tunnelMode: false,
			expected:   1420,
		},
		{
			name:       "tunnel mode preserves exact path MTU",
			pathMTU:    1372,
			isIPv6:     false,
			tunnelMode: true,
			expected:   1372,
		},
		{
			name:       "endpoint mode with reduced path MTU",
			pathMTU:    1400,
			isIPv6:     false,
			tunnelMode: false,
			expected:   1340,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateMTU(tt.pathMTU, tt.isIPv6, tt.tunnelMode)
			if result != tt.expected {
				t.Errorf("CalculateMTU(%d, %v, %v) = %d, want %d",
					tt.pathMTU, tt.isIPv6, tt.tunnelMode, result, tt.expected)
			}
		})
	}
}
