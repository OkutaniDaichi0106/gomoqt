// Package quic provides a QUIC transport abstraction layer for the gomoqt library.
//
// This package defines interfaces and types that abstract QUIC connections,
// streams, and listeners, allowing the moqt package to work with different
// QUIC implementations without direct dependencies.
//
// # Interfaces
//
// The package provides the following core interfaces:
//
//   - Connection: Represents a QUIC connection with stream management
//   - Stream: Bidirectional QUIC stream for reading and writing
//   - SendStream: Unidirectional QUIC stream for sending data
//   - ReceiveStream: Unidirectional QUIC stream for receiving data
//   - Listener: Accepts incoming QUIC connections
//
// # Implementations
//
// The package includes a concrete implementation using quic-go:
//   - quicgo subpackage: Wraps github.com/quic-go/quic-go types
//
// # Basic Usage
//
// To create a QUIC listener:
//
//	listener, err := quicgo.Listen("udp", "localhost:4433", tlsConfig, quicConfig)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer listener.Close()
//
//	for {
//	    conn, err := listener.Accept(ctx)
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    go handleConnection(conn)
//	}
//
// To dial a QUIC connection:
//
//	conn, err := quicgo.DialAddr(ctx, "localhost:4433", tlsConfig, quicConfig)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer conn.CloseWithError(0, "done")
//
// # Stream Management
//
// Connections support both bidirectional and unidirectional streams:
//
//	// Open a bidirectional stream
//	stream, err := conn.OpenStreamSync(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Open a unidirectional stream
//	sendStream, err := conn.OpenUniStreamSync(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Configuration
//
// The Config type provides QUIC-specific configuration options:
//   - MaxIdleTimeout: Connection idle timeout
//   - MaxIncomingStreams: Maximum concurrent incoming streams
//   - MaxStreamReceiveWindow: Stream-level flow control window
//   - MaxConnectionReceiveWindow: Connection-level flow control window
//
// For more information about QUIC, see RFC 9000:
// https://datatracker.ietf.org/doc/html/rfc9000
package quic
