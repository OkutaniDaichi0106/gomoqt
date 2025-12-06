package quic_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"

	"github.com/okdaichi/gomoqt/quic"
	"github.com/okdaichi/gomoqt/quic/quicgo"
)

// Example demonstrates basic QUIC connection setup.
func Example() {
	// Create a TLS configuration
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
	}

	// Create a QUIC configuration
	quicConfig := &quic.Config{}

	// Listen for incoming QUIC connections
	listener, err := quicgo.ListenAddrEarly("localhost:4433", tlsConfig, quicConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	// Accept a connection
	conn, err := listener.Accept(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Accepted connection from: %s\n", conn.RemoteAddr())
}

// ExampleConnection_OpenStreamSync demonstrates opening a bidirectional stream.
func ExampleConnection_OpenStreamSync() {
	// Assume we have a connection (conn) from a dial or accept operation
	var conn quic.Connection

	// Open a new bidirectional stream
	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	// Write data to the stream
	_, err = stream.Write([]byte("Hello, QUIC!"))
	if err != nil {
		log.Fatal(err)
	}

	// Read response
	buf := make([]byte, 1024)
	n, err := stream.Read(buf)
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}

	fmt.Printf("Received: %s\n", buf[:n])
}

// ExampleConnection_OpenUniStreamSync demonstrates opening a unidirectional stream.
func ExampleConnection_OpenUniStreamSync() {
	// Assume we have a connection (conn) from a dial or accept operation
	var conn quic.Connection

	// Open a unidirectional stream for sending
	sendStream, err := conn.OpenUniStreamSync(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer sendStream.Close()

	// Write data
	_, err = sendStream.Write([]byte("One-way message"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Sent one-way message")
}

// ExampleListener demonstrates accepting multiple connections.
func ExampleListener() {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
	}

	listener, err := quicgo.ListenAddrEarly("localhost:4433", tlsConfig, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Printf("Listening on: %s\n", listener.Addr())

	// Accept connections in a loop
	for range 3 {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}

		// Handle connection in a goroutine
		go func(c quic.Connection) {
			defer func() {
				_ = c.CloseWithError(0, "done")
			}()
			fmt.Printf("Handling connection from: %s\n", c.RemoteAddr())
		}(conn)
	}
}
