---
title: Agnostic Design
weight: 1
---

To abstract the underlying QUIC implementation details, interfaces and wrappers are implemented in `gomoqt/quic` package.

## Interfaces

### `Listener`

```go {filename="gomoqt/quic/listener.go",base_url="https://github.com/OkutaniDaichi0106/gomoqt/tree/main/quic/listener.go"}
type Listener interface {
	Accept(ctx context.Context) (Connection, error)
	Addr() net.Addr
	Close() error
}
```

`Listener` listens for incoming QUIC connections and exposes them through the `Accept` method.

### `Connection`

```go {filename="gomoqt/quic/connection.go",base_url="https://github.com/OkutaniDaichi0106/gomoqt/tree/main/quic/connection.go"}
type Connection interface {
	AcceptStream(ctx context.Context) (Stream, error)
	AcceptUniStream(ctx context.Context) (ReceiveStream, error)
	CloseWithError(code ApplicationErrorCode, msg string) error
	ConnectionState() ConnectionState
	Context() context.Context
	LocalAddr() net.Addr
	OpenStream() (Stream, error)
	OpenStreamSync(ctx context.Context) (Stream, error)
	OpenUniStream() (SendStream, error)
	OpenUniStreamSync(ctx context.Context) (str SendStream, err error)
	RemoteAddr() net.Addr
}
```

`Connection` is the QUIC connection and manages streams belonging to it. Outgoing streams are created with the `OpenStream` and `OpenUniStream` methods, while incoming streams are accepted with the `AcceptStream` and `AcceptUniStream` methods.

> [!Note]
> The `OpenStreamSync` and `OpenUniStreamSync` methods which is block until the stream is established and useful for synchronous operations are implemented in the `Connection` interface. However, these methods are not used in the current MOQ implementation. So if you try to use your own QUIC implementation, you do not need to implement these methods.
> These may be needed in the future or may be deprecated.

### `Stream`, `SendStream` and `ReceiveStream`

```go {filename="gomoqt/quic/stream.go",base_url="https://github.com/OkutaniDaichi0106/gomoqt/tree/main/quic/stream.go"}
type Stream interface {
	SendStream
	ReceiveStream
	SetDeadline(time.Time) error
}

type SendStream interface {
	io.Writer
	io.Closer
	StreamID() StreamID
	CancelWrite(StreamErrorCode)
	SetWriteDeadline(time.Time) error
	Context() context.Context
}

type ReceiveStream interface {
	io.Reader
	StreamID() StreamID
	CancelRead(StreamErrorCode)
	SetReadDeadline(time.Time) error
}
```

`Stream` is the QUIC bidirectional stream, `SendStream` is the outgoing QUIC unidirectional stream, and `ReceiveStream` is the incoming QUIC unidirectional stream.
`Close` method which is for graceful closure sends `FIN_STREAM` frame.
`CancelWrite` method which is for aborting writes sends `RESET_STREAM` frame.
`CancelRead` method which is for aborting reads sends `STOP_SENDING` frame.