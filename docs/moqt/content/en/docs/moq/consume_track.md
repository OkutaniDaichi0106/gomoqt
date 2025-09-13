---
title: Consume a Track
weight: 6
---

Consuming a track involves reading media data from a `moqt.TrackReader`, which provides access to groups and frames as they are received. This is typically done by the subscriber or receiver.
`moqt.TrackReader` is created when calling `(moqt.Session).Subscribe` method.

**Overview**
```go
    // Create a new session
    // This differs from if it is client-side or server-side
    var sess *moqt.Session
    var config *moqt.TrackConfig
    tr, err := sess.Subscribe("/broadcast_path", "track_name", config)
    if err != nil {
        // Handle error
        return
    }
    defer tr.Close()

    for {
        gr, err := tr.AcceptGroup()
        if err != nil {
            // Handle error
            return
        }

        go func(gr *moqt.GroupReader) { // Read frames in parallel
            defer gr.Close()

            for {
                frame, err := gr.ReadFrame()
                if err != nil {
                    // End of group or error
                    break
                }

                // Process the frame
                // The raw payload data can be accessed with frame.Bytes()
            }
        }(gr)
    }
```

## Subscribe to a Track

By subscribing to a track, a moqt.TrackReader is created, allowing the receiver to read media data from the track.

{{< cards >}}
    {{< card link="../subscribe/#subscribe-to-a-track" title="Subscribe to a Track" icon="external-link">}}
{{</cards>}}

## Accept a Group
To receive the next available group from a track, use `(moqt.TrackReader).AcceptGroup` method. This returns a `moqt.GroupReader` for reading frames from the group.

```go
    var tr *moqt.TrackReader
    group, err := tr.AcceptGroup(ctx)
    if err != nil {
        // End of track or error
        return
    }
    defer group.Close() // Always close when done
```

## Read Frames

To read frames from a group, use `(moqt.GroupReader).ReadFrame` method. Each call decodes the next frame in the group into the internal `moqt.Frame` object. The frame is reused for each call. To cache the frame data, you have to clone it via `(moqt.Frame).Clone` before reading the next frame.

```go
    var group *moqt.GroupReader

    for {
        frame, err := group.ReadFrame()
        if err != nil {
            // End of group or error
            break
        }

        // Process the frame
    }
```

### Clone Frame

```go
    var frame *moqt.Frame
    clone = frame.Clone()
```

## Cancel Group Reading

To cancel a group and stop receiving frames, call `(moqt.GroupReader).CancelRead` method with an error code.

```go
    var group *moqt.GroupReader
    var code moqt.GroupErrorCode
    group.CancelRead(code)
```

- **Group Error Code**:
  The group error code is used to indicate the reason for canceling the group reading. It helps the sender understand why the group was canceled.

## Unsubscribe from a Track

To unsubscribe from a track and stop receiving any further groups or frames with no errors, call `(moqt.TrackReader).Close` method.

```go
    var tr *moqt.TrackReader
    tr.Close()
```

To stop receiving any further groups or frames due to an error, call `(moqt.TrackReader).CloseWithError` method with an error code.

```go
    var tr *moqt.TrackReader
    var code moqt.SubscribeErrorCode
    tr.CloseWithError(code)
```

- **Subscribe Error Code**:
  The subscribe error code is used to indicate the reason for closing the track subscription. It helps the subscriber understand why the subscription was terminated.

If any groups belong to the track, they will be closed automatically with an error code `moqt.SubscribeCanceledErrorCode` when the track is closed. This ensures that all resources are released properly and no unexpected behavior occurs.