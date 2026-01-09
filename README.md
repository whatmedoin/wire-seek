# Wire-Seek

[![CI](https://github.com/yeya/wire-seek/actions/workflows/ci.yml/badge.svg)](https://github.com/yeya/wire-seek/actions/workflows/ci.yml)
[![Release](https://github.com/yeya/wire-seek/actions/workflows/release.yml/badge.svg)](https://github.com/yeya/wire-seek/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/yeya/wire-seek)](https://goreportcard.com/report/github.com/yeya/wire-seek)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/yeya/wire-seek)](https://go.dev/)

A WireGuard MTU optimization tool that discovers the optimal MTU for WireGuard tunnels using ICMP Path MTU Discovery.

## Why?

If your WireGuard MTU is too high, packets get fragmented (slow). If it's too low, you're wasting bandwidth on overhead. Wire-Seek finds the sweet spot automatically.

## How It Works

1. **Path MTU Discovery** - Sends ICMP Echo packets with the Don't Fragment (DF) bit set
2. **Binary Search** - Efficiently finds the optimal MTU in ~8 probes instead of 200+
3. **WireGuard Calculation** - Subtracts the correct overhead for your tunnel:
   - IPv4: 60 bytes (20 IP + 8 UDP + 32 WireGuard)
   - IPv6: 80 bytes (40 IP + 8 UDP + 32 WireGuard)

## Installation

### From Releases

Download the latest binary from [Releases](https://github.com/yeya/wire-seek/releases).

```bash
# Linux (amd64)
curl -LO https://github.com/yeya/wire-seek/releases/latest/download/wire-seek-linux-amd64
chmod +x wire-seek-linux-amd64
sudo mv wire-seek-linux-amd64 /usr/local/bin/wire-seek
```

### From Source

```bash
go install github.com/yeya/wire-seek@latest
```

Or clone and build:

```bash
git clone https://github.com/yeya/wire-seek.git
cd wire-seek
go build -o wire-seek .
```

## Usage

```bash
# Basic usage (requires root for raw ICMP sockets)
sudo wire-seek <target-host>

# Examples
sudo wire-seek 10.0.0.1              # WireGuard peer IP
sudo wire-seek vpn.example.com       # Hostname
sudo wire-seek -v 8.8.8.8            # Verbose mode
sudo wire-seek -6 2001:db8::1        # Force IPv6
```

### Options

| Flag | Description |
|------|-------------|
| `-target` | Target host or IP address |
| `-6` | Use IPv6 instead of IPv4 |
| `-v` | Verbose output (shows binary search progress) |

### Example Output

```
Wire-Seek: WireGuard MTU Optimizer
Target: 10.0.0.1 (10.0.0.1)
Protocol: IPv4

Discovering path MTU (range: 1280-1500)...

Results:
  Path MTU:      1500 bytes
  WireGuard MTU: 1440 bytes

Add to your WireGuard config:
  MTU = 1440
```

## Applying the MTU

Add the discovered MTU to your WireGuard configuration:

```ini
[Interface]
PrivateKey = ...
Address = 10.0.0.2/24
MTU = 1440  # <-- Add this line
```

Then restart the interface:

```bash
sudo wg-quick down wg0
sudo wg-quick up wg0
```

## Supported Platforms

| Platform | Architecture | Status |
|----------|--------------|--------|
| Linux | amd64, arm64 | ✅ Full support |
| macOS | amd64, arm64 | ✅ Full support |
| FreeBSD | amd64 | ✅ Full support |
| Windows | amd64 | ✅ Full support |

## License

MIT License - see [LICENSE](LICENSE) for details.
