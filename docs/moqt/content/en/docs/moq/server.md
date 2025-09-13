---
title: Server
weight: 2
---

Server manages server-side operations for the MoQ protocol. It listens for incoming connections, establishes sessions, relays data, and manages their lifecycle.

**Overview:**

Here's a complete example of a MoQ relay server using WebTransport, based on the example in `examples/relay/server/main.go`:

```go
func main() {
    server := moqt.Server{
        Addr: "moqt.example.com:9000",
        TLSConfig: &tls.Config{
            NextProtos:         []string{"h3", "moq-00"},
            Certificates:       []tls.Certificate{loadCert()},
            InsecureSkipVerify: false,
        },
        QUICConfig: &quic.Config{
            Allow0RTT:       true,
            EnableDatagrams: true,
        },
        Config: &moqt.Config{
            CheckHTTPOrigin: func(r *http.Request) bool {
                return r.Header.Get("Origin") == "https://trusted.example.com"
            },
        },
        Logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelInfo,
        })),
    }

    path := "/relay"

    // Handle WebTransport connections
    http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
        err := server.ServeWebTransport(w, r)
        if err != nil {
            slog.Error("Failed to serve MoQ over WebTransport", "error", err)
        }
    })

    // Set up MoQ handler
    moqt.HandleFunc(path, func(w moqt.SetupResponseWriter, r *moqt.SetupRequest) {
        sess, err := server.Accept(w, r, nil)
        if err != nil {
            slog.Error("Failed to accept session", "error", err)
            return
        }

        slog.Info("New session established")

        // Handle announcements and tracks...
    })

    err := server.ListenAndServe()
    if err != nil {
        slog.Error("Failed to start server", "error", err)
    }
}
```

## Initialize a Server

There is no dedicated function (such as a constructor) for initializing a server.
Users define the struct directly and assign values to its fields as needed.


**Detailed Configuration Example:**

For a production-ready server with TLS, QUIC settings, and logging:

```go
server := moqt.Server{
    Addr: "moqt.example.com:9000",
    TLSConfig: &tls.Config{
        NextProtos:         []string{"h3", "moq-00"},
        Certificates:       []tls.Certificate{loadCert()},
        InsecureSkipVerify: false, // Recommended for production
    },
    QUICConfig: &quic.Config{
        Allow0RTT:       true,
        EnableDatagrams: true,
    },
    Config: &moqt.Config{
        CheckHTTPOrigin: func(r *http.Request) bool {
            // Implement proper origin check for security
            return r.Header.Get("Origin") == "https://trusted.example.com"
        },
    },
    Logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    })),
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

## Listen and Accept Sessions

### Route Set-Up Requests

Before establishing sessions, servers have to handle incoming set-up requests for a specific path and route them to appropriate handlers.
`(Server).SetupHandler` field is used for this purpose.

```go  {filename="gomoqt/moqt/server.go",base_url="https://github.com/OkutaniDaichi0106/gomoqt/tree/main/moqt/server.go"}
type Server struct {
    // ...

    SetupHandler SetupHandler

    // ...
}
```

```go {filename="gomoqt/moqt/router.go",base_url="https://github.com/OkutaniDaichi0106/gomoqt/tree/main/moqt/router.go"}
type SetupHandler interface {
    ServeMOQ(SetupResponseWriter, *SetupRequest)
}

type SetupHandlerFunc func(SetupResponseWriter, *SetupRequest)
```

- **Use Default Router**

When `(Server).SetupHandler` is not set, `moqt.DefaultRouter` is the default router used by the server.
To use the default router, you can register your handlers with it directly:

```go
    moqt.DefaultRouter.HandleFunc("/path", handlerFunc)
    moqt.DefaultRouter.Handle("/path", handler)
```
We also provide global function to register handlers with the default router:

```go
    moqt.HandleFunc("/path", handlerFunc)
    moqt.Handle("/path", handler)
```

- **Use Custom Router**

If you need more control over routing, you can create a custom router and set it as the server's handler:

```go
    router := moqt.NewRouter()
    router.HandleFunc("/path", handlerFunc)
    router.Handle("/path", handler)
    server.Handler = router
```

After a set-up request is routed to a specific handler and is accepted, a session is established.

```go
    var server *moqt.Server
    var mux *moqt.TrackMux

    moqt.HandleFunc("/path", func(w moqt.SetupResponseWriter, r *moqt.SetupRequest) {
        sess, err := server.Accept(w, r, mux)
        if err != nil {
            // Handle error
            return
        }

        // Handle the established session
    })
```

The `Accept` method establishes a new MoQ session by accepting the setup request. It takes:
- `w SetupResponseWriter`: Writer to send the server response
- `r *SetupRequest`: The client's setup request
- `mux *TrackMux`: Multiplexer for track management (can be nil for default handling)

On success, it returns a `*Session` for managing the established connection.

## Handle WebTransport Connections

For WebTransport-based MoQ sessions, integrate the server with an HTTP server using `ServeWebTransport`.

**Using with net/http:**

```go
http.HandleFunc("/moq", func(w http.ResponseWriter, r *http.Request) {
    err := server.ServeWebTransport(w, r)
    if err != nil {
        http.Error(w, "Failed to serve MoQ over WebTransport", http.StatusInternalServerError)
        slog.Error("Failed to serve MoQ over WebTransport", "error", err)
    }

    // Fallback to another protocol if not WebTransport
})

// Start HTTP/3 server (requires HTTP/3 support)
http.ListenAndServeTLS(":443", "cert.pem", "key.pem", nil)
```

The `ServeWebTransport` method upgrades the HTTP/3 connection to WebTransport, accepts the session stream, and routes the setup request to the configured `SetupHandler`.

## Run the Server

`(moqt.Server).ListenAndServe()` starts the server listening for incoming connections and setting up sessions.

```go
    server.ListenAndServe()
```

For more advanced use cases:
- `ListenAndServeTLS(certFile, keyFile string)`: Starts the server with TLS certificates loaded from files.
- `ServeQUICListener(ln quic.Listener)`: Serves on an existing QUIC listener.
- `ServeQUICConn(conn quic.Connection)`: Handles a single QUIC connection directly.


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