# gomoqt

A Go implementation of Media over QUIC Transport (MOQT), specifically implementing the MOQ Lite specification for efficient media streaming over QUIC.

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

The core package implementing the MOQ Lite protocol interactions, including:
- Session establishment and management
- Track publishing and subscription
- Announcement handling
- Stream multiplexing and routing

This implementation is specifically designed for the MOQ Lite specification, focusing on simplicity and performance for real-time media streaming applications.

### lomc (Coming Soon)

Implementation of the Low Overhead Media Container.
**Note:** This package is currently under development.

### catalog (Coming Soon)

Implementation of the MOQ Catalog for content detection and management.
**Note:** This package is currently under development.

## Development

### Prerequisites

- Go 1.22 or later
- [just](https://github.com/casey/just) command runner

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

3. Set up the development environment:
   ```bash
   just dev-setup
   ```

This command will perform the following:
- Install the required certificate tools (mkcert).
- Install development tools (goimports, golangci-lint).
- Download project dependencies.
- Generate development certificates.

### Development Commands

#### Running Examples
```bash
# Start the echo server
just run-echo-server

# In another terminal, run the echo client
just run-echo-client
```

#### Code Quality
```bash
# Format code
just fmt

# Run linter
just lint

# Run tests
just test

# Perform overall quality checks (formatting and linting)
just check
```

#### Build & Clean
```bash
# Build the code
just build

# Clean up generated files
just clean
```

### Examples

The [examples](examples) directory includes sample applications demonstrating how to use gomoqt:

- **Echo Server and Client** (`echo/`): A simple echo server and client implementation.
- More samples coming soon…

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
   just fmt
   just lint
   just test
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










