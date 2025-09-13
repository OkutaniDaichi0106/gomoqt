---
title: Use Custom QUIC
weight: 3
---

## Injection

To use a custom QUIC implementation, you need to provide your own implementation of the `gomoqt/quic` interfaces. This can be done by creating a new package that implements the required interfaces and then configuring the `gomoqt/moqt` library to use your implementation.

### Injecting to Server

{{%steps%}}

### Implement `quic.Listener` and `quic.ListenAddrFunc`

```go {filename="gomoqt/quic/listener.go",base_url="https://github.com/OkutaniDaichi0106/gomoqt/tree/main/quic/listener.go"}
type ListenAddrFunc func(addr string, tlsConfig *tls.Config,
                quicConfig *Config) (Listener, error)

type Listener interface {
	Accept(ctx context.Context) (Connection, error)
	Addr() net.Addr
	Close() error
}
```

### Set Your Own `quic.ListenAddrFunc` to `moqt.Server`
```go {filename="gomoqt/moqt/server.go",base_url="https://github.com/OkutaniDaichi0106/gomoqt/tree/main/moqt/server.go"}
type Server struct {
    // ...

    /*
	 * Listen QUIC function
	 */
	ListenFunc quic.ListenAddrFunc

    // ...
}
```

### Run `(*moqt.Server) ListenAndServe`
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

### Injecting to Client

{{%steps%}}

### Implement `quic.DialAddrFunc`

```go {filename="gomoqt/quic/dialer.go",base_url="https://github.com/OkutaniDaichi0106/gomoqt/tree/main/quic/dialer.go"}
type DialAddrFunc func(ctx context.Context, addr string, tlsConfig *tls.Config,
                quicConfig *Config) (Connection, error)
```

### Set Your Own `quic.DialAddrFunc` to `moqt.Client`
```go {filename="gomoqt/moqt/client.go",base_url="https://github.com/OkutaniDaichi0106/gomoqt/tree/main/moqt/client.go"}
type Client struct {
    // ...

    /*
	 * Dial QUIC function
	 */
	DialQUICFunc quic.DialAddrFunc

    // ...
}
```

### Run `(*moqt.Client) Dial`
```go
client := &moqt.Client{
    TLSConfig:      /* Your custom *tls.Config */,
    QUICConfig:     /* Your custom *quic.Config */,
    DialQUICFunc:   /* Your custom DialAddrFunc */,
}

sess, err := client.Dial(ctx, [url], mux)
if err != nil {
    // Handle error
}
```
{{%/steps%}}

By following these steps, you can easily inject your own QUIC implementation into the `gomoqt` library and customize its behavior to suit your needs.