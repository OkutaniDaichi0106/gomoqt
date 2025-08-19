---
title: Client
weight: 1
---

In `gomoqt`, `*moqt.Client` is implemented to manage client-side operations for the MoQ Lite protocol.
`*moqt.Client` establishes sessions using QUIC or WebTransport and manages their lifecycle.

## Initialize a Client

There is no dedicated function (such as a constructor) for initializing a client.
Users define the struct directly and assign values to its fields as needed.

```go
    client := moqt.Client{}
```

### Configuration

Clients can be configured by setting fields on the `Client` struct.

```go
	client := moqt.Client{
		// Set client options here
	}
```

The following table describes the public fields of the `Client` struct:

| Field                  | Type                        | Description                                 |
|------------------------|-----------------------------|---------------------------------------------|
| `TLSConfig`            | [`*tls.Config`](https://pkg.go.dev/crypto/tls#Config) | TLS configuration for secure connections    |
| `QUICConfig`           | [`*quic.Config`](https://pkg.go.dev/github.com/OkutaniDaichi0106/gomoqt/quic#Config)              | QUIC protocol configuration                 |
| `Config`               | [`*moqt.Config`](https://pkg.go.dev/github.com/OkutaniDaichi0106/gomoqt/moqt#Config)                   | MOQ protocol configuration                  |
| `DialQUICFunc`         | [`quic.DialAddrFunc`](https://pkg.go.dev/github.com/OkutaniDaichi0106/gomoqt/quic#DialAddrFunc)         | Function to dial QUIC connection            |
| `DialWebTransportFunc` | [`webtransport.DialAddrFunc`](https://pkg.go.dev/github.com/OkutaniDaichi0106/gomoqt/webtransport#DialAddrFunc) | Function to dial WebTransport connection    |
| `Logger`               | [`*slog.Logger`](https://pkg.go.dev/log/slog#Logger)              | Logger for client events and errors         |

## Dial

Clients can initiate a connection and establish a session with a server using the `(*moqt.Client).Dial` method.

```go
	var mux *TrackMux
	sess, err := client.Dial(ctx, "https://[addr]:[port]/[path]", mux)
	if err != nil {
		// Handle error
		return
	}

	// Handle session
```

> [!NOTE] Note: Nil TrackMux
> When set nil for `mux`, the `DefaultTrackMux` will be used by default.
> Ensure that the `mux` is properly configured for your use case to avoid unexpected behavior.

## Terminate and Shut Down Client

Clients can terminate all active sessions and shut down using two main methods, each suited for different operational needs:

### Immediate Shutdown

Calling `Client.Close()` will immediately terminate all active sessions and release resources. This is a forceful shutdown: all sessions are closed using `Session.Terminate` with a no-error code, and any in-flight operations are interrupted. If shutdown is already in progress, further calls are ignored. After shutdown, all sessions and streams are closed.

```go
	client.Close() // Immediately closes all sessions.
	// Use with care, as abrupt disconnection may occur.
```

### Graceful Shutdown

client also provides a `Shutdown` method for graceful termination.
This method takes a context and when it is canceled or times out, it will forcefully close all sessions.

```go
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := client.Shutdown(ctx) // Waits for sessions to close gracefully,
	// or forces termination on timeout/cancel.

	if err != nil {
		// Forced termination occurred, or shutdown timed out.
	}
```

> [!NOTE] Note: No GOAWAY message
> The current implementation does not send a `GOAWAY` message. `GOAWAY` notifies peers of upcoming shutdown, so they can stop opening new streams and finish processing existing ones before disconnecting. Immediate session closure occurs when the context is canceled. This will be updated once the `GOAWAY` message specification is finalized.

In both cases, after shutdown, all sessions and streams are closed. For most use cases, prefer graceful shutdown to ensure a smooth experience for connected clients.