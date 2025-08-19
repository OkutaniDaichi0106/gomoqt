---
title: Session
weight: 3
---

MOQ Session is established when a client connects to a QUIC server and offers to set up a new session on a specific stream (Session Stream).

## Methods

```go
func (s *Session) AcceptAnnounce(prefix string) (*moqt.AnnouncementReader, error)
func (s *Session) Context() context.Context
func (s *Session) SessionUpdated() <-chan struct{}
func (s *Session) Subscribe(path moqt.BroadcastPath, name moqt.TrackName,
     config *moqt.TrackConfig) (*moqt.TrackReader, error)
func (s *Session) Terminate(code moqt.SessionErrorCode, msg string) error
```

## Outgoing Requests

Outgoing requests such as subscribing to tracks or discovering available tracks are handled by the session.

### Subscribe to a Track

{{<cards>}}
    {{< card link="../subscribe/#subscribe-to-a-track" title="Subscribe to a Track" icon="external-link">}}
{{</cards>}}

### Discover Available Tracks

{{<cards>}}
    {{< card link="../announce/#receive-announcements" title="Receive Announcements" icon="external-link">}}
{{</cards>}}


## Incoming Requests

Incoming requests such as track subscriptions and discovery broadcasts are handled by the internal `moqt.TrackMux` of the session. So, there is no method for handling these requests directly on the `Session` struct.