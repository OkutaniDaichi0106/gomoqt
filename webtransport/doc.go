// Package webtransport provides a WebTransport abstraction layer for the gomoqt library.
//
// WebTransport is a protocol framework that enables clients constrained by the
// Web security model to communicate with a remote server using a secure multiplexed
// transport. This package abstracts WebTransport connections to allow the moqt
// package to support both browser-based and native QUIC connections.
//
// # Interfaces
//
// The package defines interfaces for WebTransport functionality:
//   - Server: WebTransport server for accepting browser connections
//   - DialAddrFunc: Function type for establishing WebTransport connections
//
// # Implementations
//
// The package includes a concrete implementation:
//   - webtransportgo subpackage: Wraps WebTransport functionality
//
// # Basic Usage
//
// To create a WebTransport server:
//
//	server := &webtransportgo.Server{
//	    Addr:      ":4433",
//	    TLSConfig: tlsConfig,
//	    Handler:   handler,
//	}
//	if err := server.ListenAndServe(); err != nil {
//	    log.Fatal(err)
//	}
//
// To dial a WebTransport connection from a client:
//
//	conn, err := webtransportgo.DialAddr(ctx, "https://example.com:4433", tlsConfig)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer conn.Close()
//
// # Browser Compatibility
//
// WebTransport is designed to work with modern web browsers that support
// the WebTransport API. This enables Go servers to communicate directly
// with JavaScript clients running in browsers without requiring WebSocket
// or HTTP/2 workarounds.
//
// # Relationship to QUIC
//
// WebTransport runs over HTTP/3, which itself runs over QUIC. The package
// provides a bridge between the quic package abstractions and WebTransport
// semantics, enabling transparent protocol negotiation.
//
// For more information about WebTransport, see:
// https://www.w3.org/TR/webtransport/
package webtransport
