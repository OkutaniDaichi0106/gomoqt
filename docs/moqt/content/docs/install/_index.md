---
title: Install
description: gomoqt の導入方法（アプリへの組み込み / ソースからの実行）
weight: 1
---


`gomoqt` is a project for implementing Media over QUIC in Go.
It provides several Go packages for building MOQ applications, which can be integrated into your Go projects.

## Prerequisites

- Go 1.22 or later (recommended)
- Git (required for building from source)
- (Optional, recommended) Local certificate tool mkcert (for TLS usage)

## Getting Started

### 1. Initialize Go module
If you haven't already, initialize a Go module in your project directory.

```bash
go mod init [module_name]
```

### 2. Get gomoqt

```bash
go get github.com/OkutaniDaichi0106/gomoqt
```

### 3. Importing packages

gomoqt provides several packages that can be imported into your Go application. The main package is `moqt`, which contains the core logic for Media over QUIC. In addition to `moqt`, the following packages are provided:

| Package Name   | Description                                                                 |
|:-------------- |:---------------------------------------------------------------------------|
| `moqt`         | Main package implementing the core logic for Media over QUIC.               |
| `quic`         | Abstraction and interface definitions for QUIC used by moqt.<br>Includes a wrapper for `quic-go/quic-go` which is used in `moqt` by default. |
| `webtransport` | Abstraction and interface definitions for WebTransport used by moqt.<br>Includes a wrapper for `quic-go/webtransport-go` which is used in `moqt` by default. |
| `lomc`         | Package for Low Overhead Media Container.                                  |
| `hang`         | Implementation of the Hang protocol for building conference apps over MOQ.  |
| `catalog`      | Implementation of Common Catalog Format for MOQ.                           |

**Example of importing the `moqt` package:**

```go
import (
	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)
```

## Access from Web