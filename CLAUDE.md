# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

### Build Commands
```bash
# Build for current platform
go build

# Build for Raspberry Pi (ARM)
env GOOS=linux GOARCH=arm GOARM=6 go build

# Build specific exporters
go build ./cmd/elexporter
go build ./cmd/smartmeter-exporter
```

### Testing
```bash
# Run all tests
go test ./...

# Run medium tests (requires BP35C2 emulator)
go test ./... -tags medium

# Run BP35C2 emulator for medium tests
cd tools/bp35c2-emulator
go run main.go
```

### Setup Dependencies
```bash
# Install required dependencies
go get github.com/golang/mock/gomock
go install github.com/golang/mock/mockgen
go get github.com/SonyCSL/ECHONETLite-ObjectDatabase
```

## Architecture Overview

This is a Go-based EchonetLite controller with Prometheus exporter functionality for smart home device monitoring, particularly air conditioners and smart meters.

### Core Packages

- **echonetlite**: Main EchonetLite protocol implementation
  - `controller_node.go`: Core controller node for multicast/unicast communication
  - `electricity_controller_node.go`: Specialized controller for smart meter communication
  - `frame.go`: EchonetLite frame parsing and generation
  - `class_dictionary.go`: Device class definitions and property mappings
  - `object.go`, `property.go`: Device object and property management

- **wisun**: Wi-SUN communication layer for smart meters
  - `bp35c2_client.go`: ROHM BP35C2 module client for Wi-SUN B-route
  - `bp35c2_emulator.go`: Emulator for testing without hardware

- **transport**: Communication abstraction layer
  - `serial.go`: Serial communication implementation
  - `transport.go`: Generic transport interface
  - Platform-specific implementations for Windows/Unix

### Main Applications

- **cmd/elexporter**: Prometheus exporter for EchonetLite devices (air conditioners)
  - Runs on port 8083 by default
  - Polls air conditioner state every 30 seconds
  - Exposes temperature metrics via `/metrics`

- **cmd/smartmeter-exporter**: Smart meter data exporter via Wi-SUN
  - Connects to smart meter using BP35C2 module
  - Requires B-route ID and password
  - Default serial port: `/dev/ttyUSB0`
  - Default exporter port: 8080

### Testing Strategy

The codebase uses standard Go testing with mock generation via gomock. Medium tests require the BP35C2 emulator to simulate hardware communication. All core packages have corresponding test files.

### Deployment

Deployment scripts are provided in `deployments/` for Raspberry Pi installation, including systemd service files for both exporters.