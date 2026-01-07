# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Wire-seek is a WireGuard MTU optimization tool that discovers the optimal MTU for WireGuard tunnels using ICMP Path MTU Discovery with the Don't Fragment bit.

## Build and Development Commands

```bash
# Build the project
go build ./...

# Run tests
go test ./...

# Run a single test
go test -run TestName ./path/to/package

# Run tests with coverage
go test -cover ./...

# Format code
go fmt ./...

# Vet code for issues
go vet ./...
```

## Usage

```bash
# Basic usage
wire-seek <target-host>

# With options
wire-seek -v -target example.com   # Verbose mode
wire-seek -6 example.com           # Force IPv6
```

## Architecture

- `main.go` - CLI entry point, argument parsing, and result display
- `mtu/` - MTU discovery package
  - `discover.go` - Binary search PMTUD algorithm using ICMP echo with DF bit
  - `wireguard.go` - WireGuard overhead calculations (60 bytes IPv4, 80 bytes IPv6)
  - `df_linux.go`, `df_darwin.go`, `df_other.go` - Platform-specific Don't Fragment bit handling
