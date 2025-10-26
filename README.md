# gomoqt

A Go implementation of Media over QUIC Transport (MOQT), specifically implementing the MOQ Lite specification for efficient media streaming over QUIC.

[![codecov](https://codecov.io/gh/OkutaniDaichi0106/gomoqt/branch/main/graph/badge.svg?token=4LZCD3FEU3)](https://codecov.io/gh/OkutaniDaichi0106/gomoqt)

## Overview

This implementation follows the [MOQ Lite specification](https://kixelated.github.io/moq-drafts/draft-lcurley-moq-transfork.html), providing a foundation for building real-time media streaming applications using QUIC transport. MOQ Lite is a simplified version of the Media over QUIC Transport protocol, designed for lower latency and reduced complexity while maintaining the core benefits of QUIC-based media delivery.

## Features

- **MOQ Lite Protocol**: Core implementation of the MOQ Lite specification
- **WebTransport Support**: Full support for WebTransport connections in browsers
- **Raw QUIC Support**: Direct QUIC connections for native applications
- **Track Management**: Publisher/Subscriber pattern for media track handling
- **Multiplexed Streaming**: Efficient multiplexing of multiple media tracks
- **Sample Applications**: Complete examples demonstrating various use cases

## Components

### moqt

The core Go package implementing the MOQ Lite protocol interactions, including:
- Session establishment and management
- Track publishing and subscription
- Announcement handling
- Stream multiplexing and routing

This implementation is specifically designed for the MOQ Lite specification, focusing on simplicity and performance for real-time media streaming applications.

### moq-web

TypeScript/JavaScript implementation providing WebTransport support for browsers, including:
- WebTransport session management
- Stream ID calculation and tracking
- Reader/Writer interfaces for QUIC streams
- Integration with browser WebTransport API

### interop

Interoperability testing tools and examples for validating MOQ implementations across different platforms.

## Development

### Prerequisites

- Go 1.25.0 or later
- [Mage](https://magefile.org/) build tool (install with `go install github.com/magefile/mage@latest`)

### Getting Started

1. Clone the repository:
   ```bash
   git clone https://github.com/OkutaniDaichi0106/gomoqt.git
   cd gomoqt
   ```

2. Install the package:
   ```bash
   go get github.com/OkutaniDaichi0106/gomoqt
   ```

3. Install Mage build tool:
   ```bash
   go install github.com/magefile/mage@latest
   ```

Note: Development setup commands (dev-setup, certificate generation, etc.) are still available via the Justfile. The core build commands (test, lint, fmt, build, clean) have been migrated to Mage.

### Development Commands

#### Running Examples
```bash
# Start the interop server
just interop-server

# In another terminal, run the interop client
just interop-client
```

#### Code Quality
```bash
# Format code
mage fmt

# Run linter (requires golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
mage lint

# Run tests
mage test
```

#### Build & Clean
```bash
# Build the code
mage build

# Clean up generated files
mage clean

# Show available commands
mage help
```

### Examples

The [examples](examples) directory includes sample applications demonstrating how to use gomoqt:

- **Interop Server and Client** (`interop/`): Interoperability testing between different MOQ implementations
- **Broadcast Example** (`examples/broadcast/`): Broadcasting functionality demonstration
- **Echo Example** (`examples/echo/`): Simple echo server and client implementation
- **Native QUIC** (`examples/native_quic/`): Direct QUIC connection examples
- **Relay** (`examples/relay/`): Relay functionality for media streaming

### Documentation

- [GoDoc](https://pkg.go.dev/github.com/OkutaniDaichi0106/gomoqt)
- [MOQ Lite Specification](https://kixelated.github.io/moq-drafts/draft-lcurley-moq-transfork.html)
- [Implementation Status](moqt/README.md) - Detailed implementation progress

## Specification Compliance

This implementation targets the MOQ Lite specification, which provides a simplified approach to Media over QUIC Transport. The current implementation status can be found in the [moqt package README](moqt/README.md), which includes detailed tracking of implemented features according to the specification sections.

## Contributing

We welcome contributions! Here's how you can help:

1. Fork the repository.
2. Create a feature branch (`git checkout -b feature/amazing-feature`).
3. Make your changes.
4. Verify code quality:
   ```bash
   mage fmt
   mage lint
   mage test
   ```
5. Commit your changes (`git commit -m 'Add amazing feature'`).
6. Push your branch (`git push origin feature/amazing-feature`).
7. Open a Pull Request.

## License

This project is licensed under the MIT License; see [LICENSE](LICENSE) for details.

## Acknowledgments

- [quic-go](https://github.com/quic-go/quic-go) - QUIC implementation in Go
- [webtransport-go](https://github.com/quic-go/webtransport-go) - WebTransport implementation in Go
- [MOQ Lite Specification](https://kixelated.github.io/moq-drafts/draft-lcurley-moq-transfork.html) - The specification this implementation follows










