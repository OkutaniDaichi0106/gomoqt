---
title: Use Default WebTransport
weight: 1
---

`quic-go/webtransport-go` is used internally as the default WebTransport implementation when relevant fields which is set for customization are not set or `nil`.

{{<github-readme-stats user="quic-go" repo="webtransport-go" >}}

> [!Note]
> This implementation is based on the draft-02 of the WebTransport specification which is supported by the latest versions of major browsers

> [!Warning]
> The draft-02 of the WebTransport specification is supported by the latest versions of major browsers but is not guaranteed to be stable.
> It may change in the future, so please be aware of this when using it.