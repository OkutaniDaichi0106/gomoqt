---
title: Go
weight: 1
---

## Building Go Applications

### Prerequisites

- Go 1.22 or later (recommended)
- Git (required for building from source)
- (Optional, recommended) Local certificate tool mkcert (for TLS usage)

{{% steps %}}

### Initialize Go module

If you haven't already, initialize a Go module in your project directory.

```bash
go mod init [module_name]
```

### Get gomoqt

Download the package in your Go environment.

```bash
go get github.com/okdaichi/gomoqt
```

### Importing packages

gomoqt provides several packages that can be imported into your Go application. The main package is `moqt`, which contains the core logic for Media over QUIC. In addition to `moqt`, the following packages are provided:

| Package Name   | Description                                                                 |
|:-------------- |:---------------------------------------------------------------------------|
| `moqt`         | Main package implementing the core logic for Media over QUIC.               |
| `quic`         | Abstraction and interface definitions for QUIC used by moqt.<br>Includes a wrapper for `quic-go/quic-go` which is used in `moqt` by default. |
| `webtransport` | Abstraction and interface definitions for WebTransport used by moqt.<br>Includes a wrapper for `quic-go/webtransport-go` which is used in `moqt` by default. |

**Example of importing the `moqt` package**:

```go
import (
	"github.com/okdaichi/gomoqt/moqt"
)
```
{{% /steps %}}
