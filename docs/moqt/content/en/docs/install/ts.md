---
title: TypeScript / JavaScript
weight: 1
---


## Building Web Applications

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
