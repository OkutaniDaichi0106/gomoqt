# gomoqt

A Go implementation of Media over QUIC Transfork (MOQT), designed for efficient media streaming over QUIC.

## Overview

This implementation follows the [MOQ Transfork specification](https://kixelated.github.io/moq-drafts/draft-lcurley-moq-transfork.html), providing a foundation for building media streaming applications using QUIC transport.

## Features

- **MOQT Protocol**: Core implementation of the MOQTransfork protocol.
- **WebTransport Support**: Supports both WebTransport and raw QUIC connections.
- **Sample Implementations**: Provides several code examples for common use cases.

## Components

### moqt

Implements the main interactions defined by MOQ Transfork.

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
- [Specification](https://kixelated.github.io/moq-drafts/draft-lcurley-moq-transfork.html)

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
- [moq-drafts](https://github.com/kixelated/moq-drafts) - MOQ Transfork specification











