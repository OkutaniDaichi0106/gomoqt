---
title: Agnostic Design
weight: 1
---

To abstract the underlying WebTransport implementation details, interfaces and wrappers are implemented in `gomoqt/webtransport` package.

## Interfaces

### `Server`

```go {filename="gomoqt/webtransport/server.go",base_url="https://github.com/OkutaniDaichi0106/gomoqt/tree/main/webtransport/server.go"}
type Server interface {
	Upgrade(w http.ResponseWriter, r *http.Request) (quic.Connection, error)
	ServeQUICConn(conn quic.Connection) error
	Close() error
	Shutdown(context.Context) error
}
```

`Server` handles incoming QUIC connections with "h3" ALPN as HTTP/3 session on `ServeQUICConn` method and serves WebTransport. The connection is upgraded from HTTP/3 to WebTransport session and exposed through the `Upgrade` method.