# webtransport

Package webtransport provides a WebTransport abstraction layer for the Media over QUIC (MoQ) protocol.

## Overview

This package defines transport-agnostic interfaces that abstract WebTransport sessions, streams, and servers. It does not implement the WebTransport protocol itself, but provides a wrapper around the `quic-go/webtransport-go` library in the `webtransportgo` subpackage.

## Purpose

- **Abstraction Layer**: Provides common interfaces used by the `moqt` package
- **Pluggable Implementations**: Allows custom WebTransport implementations by satisfying the defined interfaces
- **Internal Use**: Not intended for direct external use; consumed by `moqt` package internally

## Custom Implementations

To use a custom WebTransport implementation with gomoqt:

1. Implement the interfaces defined in this package:
   - `Session` - WebTransport session abstraction
   - `Stream` - WebTransport stream abstraction
   - `Server` - WebTransport server abstraction
   - `Dialer` - WebTransport dialer abstraction

2. Pass your implementation to the `moqt` package through its configuration

See `webtransportgo/` subpackage for the reference implementation using `quic-go/webtransport-go`.

## Installation

```go
import "github.com/okdaichi/gomoqt/webtransport"
```
