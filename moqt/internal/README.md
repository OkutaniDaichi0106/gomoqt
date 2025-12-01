# internal Package

## Overview

The `internal` package provides the core implementation details for the MOQT
(Media Over QUIC Transfork) protocol. It contains the low-level components that
power the public API exposed through the parent `moqt` package. Following Go's
convention, this package is only accessible to its parent package and sibling
packages, ensuring implementation details remain hidden from external consumers.

## Responsibilities

- Implementing the low-level protocol interactions defined in the MOQT
  specification
- Managing session establishment and maintenance
- Handling message encoding and decoding for all MOQT message types
- Providing transport abstractions for QUIC connections
- Implementing error handling and propagation
- Supporting version negotiation and protocol compatibility

## Key Interfaces and Components

### Session

The `Session` struct represents an established MOQT session and manages the
underlying connection, streams, and state. It handles the sending and receiving
of various MOQT messages and provides methods for track subscription,
announcements, and data transmission.

### message package

The `message` subpackage implements all the message types defined in the MOQT
specification, including:

- `SessionClientMessage`/`SessionServerMessage` - For session establishment
- `AnnounceMessage`/`AnnouncePleaseMessage` - For track announcements
- `FrameMessage` - For media data transmission
- `GroupMessage` - For grouping frames
- `InfoMessage`/`InfoRequestMessage` - For track information

### protocol package

The `protocol` package defines protocol constants, including version identifiers
like `Draft01`, `Draft02`, `Draft03`, and `Develop`.

### transport package

The `transport` package provides abstractions for the underlying QUIC transport,
including connection management and stream handling.

## Interaction with Other Packages

### Dependencies (packages this package depends on)

- `github.com/quic-go/quic-go`: Used for the underlying QUIC implementation
- `github.com/quic-go/quic-go/quicvarint`: Used for variable-length integer
  encoding/decoding
- `golang.org/x/exp/slog`: Used for structured logging
- Standard Go packages: `context`, `sync`, `errors`, `bytes`, etc.

### Dependents (packages that depend on this package)

- `moqt` (parent package): Uses the internal implementation to provide the
  public API
- Sibling packages: May access internal functionality as needed

## Implementation Notes

- The package follows a layered architecture:
  - Transport layer (connection, streams)
  - Protocol layer (message encoding/decoding)
  - Session layer (state management, track handling)
- Error handling follows Go conventions with custom error types for specific
  protocol errors
- Message encoding/decoding is done carefully with attention to binary format
  specifications
- Concurrent operations are protected by appropriate synchronization primitives
  (mutexes, etc.)

## Testing Strategy

Testing the internal package involves:

1. Unit tests for individual components, particularly message encoding/decoding
2. Mock-based tests using the transport mocks (see `mock_connection.go`,
   `mock_stream.go`)
3. Integration tests that verify the correct protocol flows
4. Ensuring thread-safety through concurrent testing scenarios

## Future Extensibility

The package is designed to be extensible in several ways:

- Support for new protocol versions via the version negotiation mechanism
- Addition of new message types as the protocol evolves
- Enhancement of error handling and recovery strategies
- Performance optimizations (e.g., using `sync.Pool` for message buffers)

## References

- [MOQ Transfork Specification](https://kixelated.github.io/moq-drafts/draft-lcurley-moq-transfork.html)
- [quic-go library](https://github.com/quic-go/quic-go)
- [Go Style Guide for gomoqt](../../.github/prompts/style.prompt.md)
