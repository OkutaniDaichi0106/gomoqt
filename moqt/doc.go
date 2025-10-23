// Package moqt implements the MOQ Lite specification for Media over QUIC Transport.
//
// MOQ Lite is a simplified version of the Media over QUIC Transport protocol,
// designed for lower latency and reduced complexity while maintaining the core
// benefits of QUIC-based media delivery. This implementation follows the
// MOQ Lite specification (draft-lcurley-moq-transfork).
//
// # Key Features
//
//   - Session establishment and management for both WebTransport and raw QUIC
//   - Track publishing and subscription with the Publisher/Subscriber pattern
//   - Announcement handling for track discovery
//   - Stream multiplexing and routing for efficient media delivery
//   - Group and frame-based media data transmission
//
// # Basic Usage
//
// To create a MOQ server:
//
//	server := &moqt.Server{
//	    Addr:       ":4433",
//	    TLSConfig:  tlsConfig,
//	    SetupHandler: setupHandler,
//	}
//	if err := server.ListenAndServe(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
// To create a MOQ client:
//
//	client := &moqt.Client{
//	    TLSConfig: tlsConfig,
//	}
//	session, err := client.DialWebTransport(ctx, "https://example.com:4433")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Core Components
//
//   - Server: MOQ Lite server with WebTransport and QUIC support
//   - Client: MOQ Lite client for establishing sessions
//   - Session: Manages bidirectional communication between client and server
//   - TrackWriter: Publishes media data to a track
//   - TrackReader: Subscribes to and consumes media data from a track
//   - Mux: Routes announcements and subscriptions to appropriate handlers
//
// # Specification Compliance
//
// This package implements the MOQ Lite specification, which includes:
//   - Session establishment via WebTransport or QUIC (Section 3)
//   - Version and extension negotiation (Section 3.2)
//   - Track announcement and subscription (Sections 5.1, 5.2)
//   - Group and frame-based data transmission (Sections 8, 9)
//   - Control messages for session and subscription management (Section 7)
//
// For detailed specification status and implementation progress, see the
// package README.md file.
//
// # Performance Considerations
//
// The implementation is optimized for real-time media streaming with:
//   - Minimal latency for group and frame delivery
//   - Efficient stream multiplexing for concurrent tracks
//   - Resource pooling for reduced allocations
//   - Support for track priority control (in development)
//
// # Examples
//
// See the examples/ directory for complete working examples including:
//   - Echo server/client for basic request-response patterns
//   - Broadcast server/client for one-to-many streaming
//   - Relay server for multi-hop media delivery
//
// For more information, visit: https://github.com/OkutaniDaichi0106/gomoqt
package moqt
