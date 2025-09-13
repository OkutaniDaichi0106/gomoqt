---
title: Install
description: gomoqt の導入方法（アプリへの組み込み / ソースからの実行）
weight: 1
---


`gomoqt` provides several Go packages for building MOQ applications, which can be integrated into your Go projects.

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
go get github.com/OkutaniDaichi0106/gomoqt
```

### Importing packages

gomoqt provides several packages that can be imported into your Go application. The main package is `moqt`, which contains the core logic for Media over QUIC. In addition to `moqt`, the following packages are provided:

| Package Name   | Description                                                                 |
|:-------------- |:---------------------------------------------------------------------------|
| `moqt`         | Main package implementing the core logic for Media over QUIC.               |
| `quic`         | Abstraction and interface definitions for QUIC used by moqt.<br>Includes a wrapper for `quic-go/quic-go` which is used in `moqt` by default. |
| `webtransport` | Abstraction and interface definitions for WebTransport used by moqt.<br>Includes a wrapper for `quic-go/webtransport-go` which is used in `moqt` by default. |
| `lomc`         | Package for Low Overhead Media Container.                                  |
| `hang`         | Implementation of the Hang protocol for building conference apps over MOQ.  |
| `catalog`      | Implementation of Common Catalog Format for MOQ.                           |

**Example of importing the `moqt` package**:

```go
import (
	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)
```
{{% /steps %}}

## Access from Web

A significant feature of MoQ is that it is available on web browsers using WebTransport. This allows for real-time media streaming directly in the browser without the need for additional plugins or software.
We provide a JavaScript client library to facilitate this integration.

### Prerequisites

- Node.js (version 14 or later)
- npm (Node Package Manager)

{{% steps %}}

### Initialize npm module

If you haven't already, initialize an npm module in your project directory.

```bash
npm init -y
```

### Install module

```bash
npm install @okutanidaichi/moqt
```

{{% /steps %}}

> [!NOTE] Note: Browser compatibility
> If your browser does not support WebTransport, `moqt` does not work.
> Check the [Can I Use](https://caniuse.com/webtransport) for the latest compatibility information.