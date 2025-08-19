---
linkTitle: Track/Group/Frame
title: Track, Group, and Frame
weight: 7
---

## Data Model

Media streams are organized into a hierarchy of tracks, groups, and frames.

### Track

A track is a continuous media stream, such as video or audio, within a broadcast path. Each track has a name that distinguishes it within the broadcast path, often corresponding to a specific media layer (e.g., SVC).

| Implementations       | Description                                 |
|-----------------------|---------------------------------------------|
| `moqt.TrackWriter`    | The TrackWriter handles all aspects of writing data to a track, including managing the lifecycle of the track, opening and closing groups, transmitting media data, and dealing with errors. |
| `moqt.TrackReader`    | The TrackReader is responsible for reading data from a track, receiving groups and frames in order, and managing any errors that occur during the process. |

### Group

A group is a self-contained segment of a track, often corresponding to a time-based unit like a video GOP or an audio packet group. Groups are processed and transmitted independently, and may contain frames that are either standalone or interdependent (for example, I/P/B frames in video).

| Implementations       | Description                                  |
|-----------------------|----------------------------------------------|
| `moqt.GroupWriter`    | The GroupWriter is used to write frames to a group, taking care of starting and ending groups, adding frames, and handling any errors that arise. |
| `moqt.GroupReader`    | The GroupReader allows for reading frames from a group in sequence, finalizing or canceling groups, and managing errors throughout the process. |

### Frame

A frame is the smallest unit of media data, such as a single video image or audio sample. Frames can be independent (like keyframes) or rely on other frames (like delta frames in video), and together they form the building blocks of a track.

| Implementations  | Description                                      |
|------------------|--------------------------------------------------|
| `moqt.Frame`     | The Frame struct represents the smallest unit of media data. It ensures that data is stored immutably, provides safe access to the payload, and supports efficient memory reuse. |

## Produce a Track

Producing a track involves writing media data to a `TrackWriter`, which manages the creation of groups and frames for transmission.
`moqt.TrackWriter` is responsible for handling the track's lifecycle, including opening and closing groups.
`moqt.TrackWriter` is created when being subscribed to a track from the peer on the session. Created `moqt.TrackWriter` is then routed via the session's `TrackMux` and can be used to publish media data.

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

            frame := moqt.NewFrame([]byte{/* raw frame data */})

            err = gw.WriteFrame(frame)
            if err != nil {
                // Handle error
                return
            }

            seq++
        }
    })
```

### Create a Group

To start a new group within a track, use `(*moqt.TrackWriter).OpenGroup` method. This creates a new group with the specified sequence number and returns a `GroupWriter` for writing frames to that group.


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

### Write Frames

To add media data to a group, use `(*moqt.GroupWriter).WriteFrame` method. Each frame represents a chunk of media data (audio, video, etc.) to be sent as part of the group.

```go
    var gw *moqt.GroupWriter
    frame := moqt.NewFrame([]byte{/* raw frame data */})
    err := gw.WriteFrame(frame)
    if err != nil {
        // Handle error
    }
```


#### Frame Creation
`moqt.NewFrame` function creates a new frame from the given byte slice.

```go
    frame := moqt.NewFrame([]byte{0x01, 0x02, 0x03})
```

Frame is designed to be immutable after creation. This is intentional.
Even when accessing the payload with `(*moqt.Frame).Bytes`, a copy of the byte slice is returned, ensuring the original data remains unchanged.

> MoQ relays do not modify media data in transit.

> [!WARNING] Warning: Payload Slice
> When you call `moqt.NewFrame` method, the frame internally holds a reference to the byte slice you provide. Do not modify the original byte slice after creating the frame, as changes will affect the frame's contents.

> [!TIP] Tip: Frame Reuse
> Creating a new Frame each time results in frequent memory allocations, which is inefficient; consider reusing Frame instances. For details, see [Reusing Frames↓](#reusing-frames) section.


### Finalize a Group

To finalize a group and indicate that no more frames will be sent, call `(*moqt.GroupWriter).Close` method.

```go
    var gw *moqt.GroupWriter
    gw.Close()
```

### Cancel Group Writing

To cancel a group and stop receiving frames, call `(*moqt.GroupReader).CancelWrite` method with an error code.

```go
    var group *moqt.GroupReader
    var code moqt.GroupErrorCode
    group.CancelWrite(code)
```

## Consume a Track

Consuming a track involves reading media data from a `TrackReader`, which provides access to groups and frames as they are received. This is typically done by the subscriber or receiver.
`moqt.TrackReader` is created when calling `(*moqt.Session).Subscribe` method.

**Overview**
```go
    // Create a new session
    // This differs from if it is client-side or server-side
    var sess *moqt.Session

    tr, err := sess.Subscribe()
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

            buf := make([]byte, 0, 1024) // Preallocate 1KB buffer for frame data
            frame := moqt.NewFrame(buf)

            for {
                err := gr.ReadFrame(frame)
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

### Accept a Group
To receive the next available group from a track, use `(*moqt.TrackReader).AcceptGroup`. This returns a `moqt.GroupReader` for reading frames from the group.

```go
    var tr *moqt.TrackReader
    group, err := tr.AcceptGroup(ctx)
    if err != nil {
        // End of track or error
        return
    }
    defer group.Close() // Always close when done
```

### Read Frames
To read frames from a group, use `(*moqt.GroupReader).ReadFrame` method. Each call decodes the next frame in the group into the provided `moqt.Frame` object.

```go
    var group *moqt.GroupReader

    frame := moqt.NewFrame(make([]byte, 0, 1024)) // Preallocate 1KB buffer for frame data

    for {
        err := group.ReadFrame(frame)
        if err != nil {
            // End of group or error
            break
        }

        // Process the frame
    }
```

> [!NOTE] Note: BYOB Reading
> When you pass a `moqt.Frame` to `(*moqt.GroupReader).ReadFrame`, the function fills it with the decoded data and takes ownership of the object.
> For best performance, reuse the same `moqt.Frame` instance when reading multiple frames in a group — this helps avoid unnecessary memory allocations.

### Cancel Group Reading

To cancel a group and stop receiving frames, call `(*moqt.GroupReader).CancelRead` method with an error code.

```go
    var group *moqt.GroupReader
    var code moqt.GroupErrorCode
    group.CancelRead(code)
```

## Reusing Frames

To improve performance and reduce memory allocations, you can reuse `moqt.Frame` instances. This helps minimize garbage collection overhead and improve cache locality.

`gomoqt` does not currently provide a built-in mechanism for frame pooling, so you may need to implement your own pooling strategy if desired.

### Reuse `moqt.Frame`

When frames are just used for reading or relaying, it is beneficial to reuse `moqt.Frame` instances. This helps minimize garbage collection overhead and improve cache locality.

> [!NOTE] Note: Frame Referencing
> `(moqt.GroupReader).ReadFrame` overwrites the contents of the provided `moqt.Frame` instance with the decoded frame data.
> This means you should not reuse a `moqt.Frame` instance across multiple calls to `(moqt.GroupReader).ReadFrame` unless you are certain the previous contents are no longer needed.

**Example Pooling Frame**

```go
var framePool = sync.Pool{
    New: func() interface{} {
        return moqt.NewFrame(make([]byte, 0, 1024))
    },
}

func putFrame(frame *moqt.Frame) {
    framePool.Put(frame)
}

func getFrame() *moqt.Frame {
    return framePool.Get().(*moqt.Frame)
}
```

### Reuse `[]byte`

When you need to create a new frame to generate a contents in the Go application, it is recommended to pool byte slices for better performance.

**Example Bytes Pooling**

```go
var bytesPool = sync.Pool{
    New: func() interface{} {
        b := make([]byte, 0, 1024)
        return &b
    },
}

func putBytes(b []byte) {
    bytesPool.Put(&b)
}

func getBytes() []byte {
    return *bytesPool.Get().(*[]byte)
}
```
**Usage Example**
```go
    b := getBytes() // Get a byte slice from the pool
    // Write frame data to the byte slice

    frame := moqt.NewFrame(b)

    defer putBytes(b) // Return the byte slice to the pool
```
