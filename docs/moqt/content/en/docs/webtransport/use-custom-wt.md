---
title: Use Custom WebTransport
weight: 2
---

## Injection

To use a custom WebTransport implementation, you need to provide your own implementation of the `gomoqt/webtransport` interfaces. This can be done by creating a new package that implements the required interfaces and then configuring the `gomoqt/moqt` library to use your implementation.

### Inject to Server

{{%steps%}}

### Implement `webtransport.Server` and constructor

```go {filename="gomoqt/webtransport/server.go",base_url="https://github.com/OkutaniDaichi0106/gomoqt/tree/main/webtransport/server.go"}
type Server interface {
	Upgrade(w http.ResponseWriter, r *http.Request) (quic.Connection, error)
	ServeQUICConn(conn quic.Connection) error
	Close() error
	Shutdown(context.Context) error
}
```

```go
func NewWebTransportServer(checkOrigin func(*http.Request) bool) webtransport.Server {
	return // Return your implementation
}
```

### Set the `NewWebTransportServerFunc` to `moqt.Server`

Set the `NewWebTransportServerFunc` field to your custom WebTransport server constructor.
It will be called once and the initialized WebTransport server will be stored in the `wtServer` field.

```go {filename="gomoqt/moqt/server.go",base_url="https://github.com/OkutaniDaichi0106/gomoqt/tree/main/moqt/server.go"}
type Server struct {
    // ...

    NewWebTransportServerFunc func(checkOrigin func(*http.Request) bool) webtransport.Server
	wtServer                  webtransport.Server

    // ...
}
```

### Run `(*moqt.Server) ListenAndServe`

When a QUIC Connection with "h3" ALPN is established, the stored WebTransport server is used to handle the connection.

```go
server := &moqt.Server{
    Addr:           /* Address to listen on */,
    TLSConfig:      /* Your custom *tls.Config */,
    QUICConfig:     /* Your custom *quic.Config */,
    ListenFunc:     /* Your custom ListenAddrFunc */,
}

// Handle incoming set-up requests

server.ListenAndServe() // or you may use server.ListenAndServeTLS()
```

{{%/steps%}}

### Inject to Client

{{%steps%}}

### Implement `webtransport.DialAddrFunc`

```go {filename="gomoqt/webtransport/dialer.go",base_url="https://github.com/OkutaniDaichi0106/gomoqt/tree/main/webtransport/dialer.go"}
type DialAddrFunc func(ctx context.Context, addr string, tlsConfig *tls.Config,
                quicConfig *Config) (Connection, error)
```

### Set the `DialWebTransportFunc` to `moqt.Client`

Set the `DialWebTransportFunc` field to your custom WebTransport dial function.

```go
client := &moqt.Client{
    // ...

	DialWebTransportFunc webtransport.DialAddrFunc

    // ...
}
```

### Run `(*moqt.Client) Dial` with "https" scheme

When using the "https" scheme, the `DialWebTransportFunc` will be called to establish a WebTransport connection.

```go
var client *moqt.Client

sess, err := client.Dial(ctx, "https://[addr]/[path]", mux)
if err != nil {
    // Handle error
}
```
{{%/steps%}}

By following these steps, you can easily inject your own QUIC implementation into the `gomoqt/moqt` library and optimize its behavior to suit your needs.