---
title: Session
weight: 3
---

MOQ Session is established when a client connects to a QUIC server and offers to set up a new session and the server accepts the request.

## Implementation

### `moqt.Session`

```go
type Session struct {
    // Embedded Fields
    Version          Version          // through internal stream
    Path             string           // through internal stream
    Versions         []Version        // through internal stream
    ClientExtensions *Parameters      // through internal stream
    SetupRequest     *SetupRequest    // through internal stream
    ServerExtensions *Parameters      // through internal stream
    // contains filtered or unexported fields
}

func (s *Session) AcceptAnnounce(prefix string) (*AnnouncementReader, error)
func (s *Session) CloseWithError(code SessionErrorCode, msg string) error
func (s *Session) Context() context.Context  // through internal stream
func (s *Session) Subscribe(path BroadcastPath, name TrackName, config *TrackConfig) (*TrackReader, error)
func (s *Session) Updated() <-chan struct{}  // through internal stream
```

Outgoing requests such as subscribing to tracks or discovering available tracks are handled by the session.

## Subscribe to a Track

{{<cards>}}
    {{< card link="../subscribe/#subscribe-to-a-track" title="Subscribe to a Track" icon="external-link">}}
{{</cards>}}

## Discover Available Broadcasts

{{<cards>}}
    {{< card link="../announce_discover/#discover-broadcasts" title="Discover Broadcasts" icon="external-link">}}
{{</cards>}}

## Session State Updates

Peers can send session state updates (such as bitrate changes) via the SESSION_UPDATE message. The implementation monitors connection bitrate and sends updates when significant shifts are detected.

Session state updates are sourced from QUIC connection statistics and processed internally by the session's bitrate monitor.

### Detect Session Updates

When the remote peer sends a session state update, you can be notified via the `Updated()` channel:

```go
var sess *moqt.Session

select {
case <-sess.Updated():
    // Session state has been updated by the peer
    // React to bitrate changes or other session parameters
case <-sess.Context().Done():
    // Session closed
}
```

The `Updated()` channel signals when the peer has sent a SESSION_UPDATE message, typically indicating a significant change in network conditions or bitrate.

## Terminate Session

{{<cards>}}
    {{< card link="../terminate/#terminate-a-session" title="Terminate a Session" icon="external-link">}}
{{</cards>}}

## Incoming Requests

Incoming requests, such as track subscriptions and discovery broadcasts, are handled internally by the session's `moqt.TrackMux`, not directly by the `moqt.Session` struct. Therefore, there are no dedicated methods for these requests on `moqt.Session`.

### Handle Track Subscriptions

{{<cards>}}
    {{< card link="../publish/#handle-track-subscriptions" title="Handle Track Subscriptions" icon="external-link">}}
{{</cards>}}

### Announce Broadcasts

{{<cards>}}
    {{< card link="../announce/#announce-broadcasts" title="Announce Broadcasts" icon="external-link">}}
{{</cards>}}

## Terminating a Session

To explicitly close a session due to protocol violations, errors, or other reasons, use the `(moqt.Session).CloseWithError` method. This closes all associated streams.

```go
func (s *Session) CloseWithError(code SessionErrorCode, msg string) error
```

- `code`: Error code (e.g., from built-in codes)
- `msg`: Descriptive message

Prefer reserved error codes for standard reasons. See [Built-in Error Codes](http://localhost:1313/gomoqt/docs/moq/errors/#built-in-error-codes) for details.

## 📝 Future Work

- Bitrate Notification: (#XXX)
