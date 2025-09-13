---
linkTitle: Produce a Track
title: Produce a Track
weight: 5
---

Producing a track involves writing media data to a `moqt.TrackWriter`, which manages the creation of groups and frames for transmission.
`moqt.TrackWriter` is responsible for handling the track's lifecycle, including opening and closing groups.
`moqt.TrackWriter` is created when being subscribed to a track from the peer on the session. Created `moqt.TrackWriter` is then routed via the session's `moqt.TrackMux` and can be used to publish media data.

**Overview**

```go
    mux := moqt.NewTrackMux()

    // Set the mux to a session
    // This differs from if it is client-side or server-side

    mux.PublishFunc(func(ctx context.Context, tw *moqt.TrackWriter) {
        defer tw.Close() // Always close when done

        var seq GroupSequence = 1
        for {
            gw, err := tw.OpenGroup(seq) // Create a new group
            if err != nil {
                // Handle error
                return
            }

            frame := moqt.NewFrame(int(1024))

            err = gw.WriteFrame(frame)
            if err != nil {
                // Handle error
                return
            }

            seq++
        }
    })
```

## Handle Track Subscriptions

When the other side (client or server) subscribes to a broadcast path, the registered `moqt.TrackHandler` is invoked to serve the subscription. The handler receives a `moqt.TrackWriter` and is responsible for sending the appropriate track data for the requested track name.

{{< cards >}}
    {{< card link="../publish/#handle-track-subscriptions" title="Handle Track Subscriptions" icon="external-link">}}
{{</cards>}}

```go
    var mux *moqt.TrackMux

    // Register the track handler with PublishFunc
    mux.PublishFunc(ctx, "/broadcast_path", func(ctx context.Context, tw *moqt.TrackWriter) {
        defer tw.Close() // Always close when done

        // Handle track subscription
    })
```

```go
    // trackHandler implements moqt.TrackHandler
    var _ moqt.TrackHandler = (*trackHandler)(nil)
    type trackHandler struct{}
    func (h *trackHandler) ServeTrack(ctx context.Context, tw *moqt.TrackWriter) {
        defer tw.Close() // Always close when done

        // Handle track subscription
    }

    var mux *moqt.TrackMux

    // Register the track handler
    mux.Publish(ctx, "/broadcast_path", trackHandler{})

    // Register the track handler using an Announcement
    var ann *moqt.Announcement
    mux.Announce(ann, trackHandler{})
```

## Create a Group

To start a new group within a track, use `(moqt.TrackWriter).OpenGroup` method. This creates a new group with the specified sequence number and returns a `moqt.GroupWriter` for writing frames to that group.


```go
    var tw *moqt.TrackWriter
    gw, err := tw.OpenGroup(1) // Start group with sequence number 1
    if err != nil {
        // Handle error
        return
    }
    defer gw.Close() // Always close when done
    // Use gw to write frames
```

> [!NOTE] Note: Group Ordering
> Groups are supposed to be produced in order, with each group having a unique increasing sequence number and then consumed in order by the subscriber.
> However, groups not in order can be created, but this may lead to complications in playback and synchronization.


> [!NOTE] Note: Sequence Number 0
> Sequence number 0 has a special meaning: it is reserved for special identifiers like "Latest Available Group" or "Final Group", and does not represent a real group.

## Write Frames

To add media data to a group, use `(moqt.GroupWriter).WriteFrame` method. Each frame represents a chunk of media data (audio, video, etc.) to be sent as part of the group.

```go
    var gw *moqt.GroupWriter
    var frame *moqt.Frame
    err := gw.WriteFrame(frame)
    if err != nil {
        // Handle error
    }
```

### Frame Creation

To create a new frame, use `moqt.FrameBuilder`. It provides a convenient way to build frames, allowing for efficient creation and reuse of frame data buffers.

```go
    builder := moqt.NewFrameBuilder(1024)

    builder.Append([]byte("Good Morning"))
    frame := builder.Frame()
    fmt.Println(string(frame.Bytes())) // "Good Morning"

    builder.Reset() // Reset the builder for the next frame
    builder.Append([]byte("Good Evening"))
    fmt.Println(string(frame.Bytes())) // "Good Evening"
```

Frame is designed to be immutable once after creation. This is intentional because MOQ does not allow modifications to the frame data.

> Relays MUST NOT combine, split, or otherwise modify object payloads.<br>
> — <cite>MOQT WG[^1]</cite>
[^1]: [IETF Draft - The Media Over QUIC Transport (moqtransport)](https://www.ietf.org/archive/id/draft-ietf-moq-transport-13.html)

## Finalize a Group

To finalize a group and indicate that no more frames will be sent, call `(moqt.GroupWriter).Close` method.

```go
    var gw *moqt.GroupWriter
    gw.Close()
```

## Cancel Group Writing

To cancel a group and stop receiving frames, call `(moqt.GroupReader).CancelWrite` method with an error code.

```go
    var group *moqt.GroupReader
    var code moqt.GroupErrorCode
    group.CancelWrite(code)
```

## Stop a Track

To stop a track and indicate that no more groups will exist, call `(moqt.TrackWriter).Close` method.

```go
    var tw *moqt.TrackWriter
    tw.Close()
```

To abort a track and indicate that no more groups will exist due to an error, call `(moqt.TrackWriter).CloseWithError` method with an error code.

```go
    var tw *moqt.TrackWriter
    var code moqt.SubscribeErrorCode
    tw.CloseWithError(code)
```

- **Subscribe Error Code**:
  The subscribe error code is used to indicate the reason for closing the track subscription. It helps the publisher understand why the subscription finished.

If any groups belong to the track, they will be closed automatically with an error code `moqt.PublishAbortedErrorCode` when the track is closed. This ensures that all resources are released properly and no data is lost.

