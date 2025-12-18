# quic

Package quic provides a QUIC abstraction layer for the Media over QUIC (MoQ) protocol.

## Overview

This package defines transport-agnostic interfaces that abstract QUIC connections, streams, and listeners. It does not implement the QUIC protocol itself, but provides a wrapper around the `quic-go/quic-go` library in the `quicgo` subpackage.

## Purpose

- **Abstraction Layer**: Provides common interfaces used by the `moqt` package
- **Pluggable Implementations**: Allows custom QUIC implementations by satisfying the defined interfaces
- **Internal Use**: Not intended for direct external use; consumed by `moqt` package internally

## Custom Implementations

To use a custom QUIC implementation with gomoqt:

1. Implement the interfaces defined in this package:
   - `Connection` - QUIC connection abstraction
   - `Stream` - QUIC stream abstraction  
   - `Listener` - QUIC listener abstraction
   - `Dialer` - QUIC dialer abstraction

2. Pass your implementation to the `moqt` package through its configuration

See `quicgo/` subpackage for the reference implementation using `quic-go/quic-go`.

## Installation

```go
import "github.com/okdaichi/gomoqt/quic"
```
