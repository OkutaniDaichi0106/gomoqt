---
title: Client
weight: 1
---

Client manages client-side operations for the MoQ protocol. It establishes and maintains sessions, handling the lifecycle and communication between the client and server.


{{% details title="Overview" closed="true" %}}

```go
func main() {
    // Create a new client
	client := moqt.Client{
		// Set client options here
	}

	// Dial and establish a session with the server
	sess, err := client.Dial(context.Background(), "https://[addr]:[port]/[path]", nil)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}
	defer sess.Terminate(moq.SessionErrorCodeNoError, "Client shutting down")

	// Use the session (e.g., subscribe to tracks, receive announcements)

	// Gracefully shut down the client when done
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Shutdown(ctx); err != nil {
		log.Printf("Client shutdown error: %v", err)
	}
}
```
{{% /details %}}

## Initialize a Client

There is no dedicated function (such as a constructor) for initializing a client.
Users define the struct directly and assign values to its fields as needed.

```go
    client := moqt.Client{
		// Set client options here
	}
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

{{< tabs items="Default QUIC, Custom QUIC" >}}
{{< tab >}}
**Using Default QUIC**

`quic-go/quic-go` is used internally as the default QUIC implementation when relevant fields which is set for customization are not set or `nil`.

{{<github-readme-stats user="quic-go" repo="quic-go" >}}

{{< /tab >}}
{{< tab >}}
**Using Custom QUIC**

To use a custom QUIC implementation, you need to provide your own implementation of the `quic.DialAddrFunc`. `(moqt.Client).DialQUICFunc` field is set, it is used to dial QUIC connections instead of the default implementation.

```go {filename="gomoqt/moqt/client.go",base_url="https://github.com/OkutaniDaichi0106/gomoqt/tree/main/moqt/client.go"}
type Client struct {
    // ...
	DialQUICFunc quic.DialAddrFunc
    // ...
}
```
{{< /tab >}}

{{< /tabs >}}

{{< tabs items="Default WebTransport, Custom WebTransport" >}}
{{< tab >}}

**Using Default WebTransport**

`quic-go/webtransport-go` is used internally as the default WebTransport implementation when relevant fields which is set for customization are not set or `nil`.

{{<github-readme-stats user="quic-go" repo="webtransport-go" >}}

{{< /tab >}}
{{< tab >}}
**Using Custom WebTransport**

To use a custom WebTransport implementation, you need to provide your own implementation of the `webtransport.DialAddrFunc`. `(moqt.Client).DialWebTransportFunc` field is set, it is used to dial WebTransport connections instead of the default implementation.

```go {filename="gomoqt/moqt/client.go",base_url="https://github.com/OkutaniDaichi0106/gomoqt/tree/main/moqt/client.go"}
type Client struct {
    // ...
	DialWebTransportFunc webtransport.DialAddrFunc
    // ...
}
```
{{< /tab >}}

{{< /tabs >}}

## Dial and Establish Session

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