---
title: Server
weight: 2
---
## Initialize a Server
In `gomoqt`, `*moqt.Server` is implemented to manage server-side operations for the MoQ Lite protocol.
`*moqt.Server` listens for incoming connections, establishes sessions using QUIC or WebTransport, and manages their lifecycle.

```go
    server := moqt.Server{
        Addr: "[addr]:[port]",
    }
```

### Configuration
The following table describes the public fields of the `Server` struct:

| Field                  | Type                        | Description                                 |
|------------------------|-----------------------------|---------------------------------------------|
| `Addr`                 | `string`                    | Server address and port                     |
| `TLSConfig`            | [`*tls.Config`](https://pkg.go.dev/crypto/tls#Config) | TLS configuration for secure connections    |
| `QUICConfig`           | [`*quic.Config`](https://pkg.go.dev/github.com/OkutaniDaichi0106/gomoqt/quic#Config)              | QUIC protocol configuration                 |
| `Config`               | [`*moqt.Config`](https://pkg.go.dev/github.com/OkutaniDaichi0106/gomoqt/moqt#Config)                   | MOQ protocol configuration                  |
| `Handler`              | [`moqt.Handler`](https://pkg.go.dev/github.com/OkutaniDaichi0106/gomoqt/moqt#Handler)                 | Set-up Request handler for routing                 |
| `ListenFunc`           | [`quic.ListenAddrFunc`](https://pkg.go.dev/github.com/OkutaniDaichi0106/gomoqt/quic#ListenAddrFunc)   | Function to listen for QUIC connections     |
| `Logger`               | [`*slog.Logger`](https://pkg.go.dev/log/slog#Logger)              | Logger for server events and errors         |

## Listen and Serve

```go
    server.ListenAndServe()
```
## Set Up Handling and Routing

### Use DefaultRouter

```go
    moqt.HandleFunc("/path", handler)
```

### Use Custom Router

```go
    router := moqt.NewRouter()
    router.HandleFunc("/path", handler)
    server.Handler = router
```

## Terminate and Shut Down Server

Servers can terminate all active sessions and shut down using two main methods, each suited for different operational needs:

### Immediate Shutdown

Calling `Server.Close()` will immediately terminate all active sessions and close all listeners. This is a forceful shutdown: all sessions are closed using `Session.Terminate` with a no-error code, and any in-flight operations are interrupted. If shutdown is already in progress, further calls are ignored. After shutdown, all sessions, streams, and listeners are closed.

```go
server.Close() // Immediately closes all sessions and listeners.
// Use with care, as clients may be disconnected abruptly.
```

### Graceful Shutdown

server also provides a `Shutdown` method for graceful termination.
This method takes a context and when it is canceled or times out, it will forcefully close all sessions.

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
err := server.Shutdown(ctx) // Waits for sessions to close gracefully,
// or forces termination on timeout/cancel.
if err != nil {
    // Forced termination occurred, or shutdown timed out.
}
```

> [!NOTE] Note: GOAWAY message
> The current implementation does not send a GOAWAY message during shutdown. Immediate session closure occurs when the context is canceled. This will be updated once the GOAWAY message specification is finalized.

In both cases, after shutdown, all sessions, streams, and listeners are closed. For most use cases, prefer graceful shutdown to ensure a smooth experience for connected clients.