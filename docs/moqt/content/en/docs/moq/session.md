---
title: Session
weight: 3
---

MOQ Session is established when a client connects to a QUIC server and offers to set up a new session and the server accepts the request.

## Implementation

### `moq.Session`

```go
type Session struct {
    // contains filtered or unexported fields
}

func (s *Session) AcceptAnnounce(prefix string) (*AnnouncementReader, error)
func (s *Session) Context() context.Context
func (s *Session) SessionUpdated() <-chan struct{}
func (s *Session) Subscribe(path BroadcastPath, name TrackName, config *TrackConfig) (*TrackReader, error)
func (s *Session) Terminate(code SessionErrorCode, msg string) error
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

## Update Session State 🚧

Peers send their local session state to the other peer such as after a significant change in the session bitrate. This provides them with the necessary information to control the session.
The session state SHOULD be sourced directly from the QUIC congestion controller.

> [!WARNING]
> This feature is not fully implemented and does not work as intended.

## Detect Session Updates

When the session state is updated by the peer, signals can be caught using the `(moqt.Session).SessionUpdated` channel.

```go
    var sess *moqt.Session
    <-sess.SessionUpdated()
```

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
