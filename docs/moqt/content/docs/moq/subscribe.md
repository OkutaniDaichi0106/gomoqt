---
title: Subscribe
weight: 6
---

Subscribing is the mechanism for receiving media tracks from a broadcast, either as a client or server. To subscribe, use the `Subscribe` method on a `Session`, specifying the broadcast path and track name.

## Subscribe to a Track

To subscribe to a track, you need to call the `Subscribe` method on a `Session` struct. This method opens a Subscribe Stream which is to control subscription and manages incoming Group Streams.

```go
    var session *moqt.Session
    config := &moqt.TrackConfig{
        // Specify track configuration options here
    }
    tr, err := session.Subscribe("/broadcast_path", "track_name", config)
    if err != nil {
        // Handle error
        return
    }
    defer tr.Close() // Make sure to close the TrackReader when done

    // Handle the TrackReader
```

## Discover Available Broadcasts

To discover available broadcasts, peers can specify the prefix for the broadcast path they are interested in and listen for announcements.

{{<cards>}}
    {{< card link="../announce/#discover-broadcasts" title="Discover Broadcasts" icon="external-link">}}
{{</cards>}}

The Announce mechanism clarifies which broadcasts are available, but does not specify which tracks belong to each broadcast.
It is up to the application logic to associate tracks with their respective broadcasts.
When you access the broadcast's track, you have to know its name in advance.

```go
    var ann *moqt.Announcement

    path := ann.BroadcastPath()

    tr, err := sess.Subscribe(path, /* specific track name */, nil)
    if err != nil {
        // Handle error
        return
    }

    // Handle the TrackReader
```

## Handle Track

The `Subscribe` method returns a `TrackReader`. The `TrackReader` represents a subscription to a track.

### Managing the Track

`(*moqt.TrackReader) Update` can be used to change the subscription parameters (e.g., priority) during the session.
When no longer needed, make sure to call `(*moqt.TrackReader) Close` to unsubscribe from the track.
If you specify some reason for closing, `(*moqt.TrackReader) CloseWithError` which notifies the reason to the sender can be used.

### Consuming a Track

{{< cards >}}
	{{< card link="../track_group_frame/#consume-a-track" title="Consume a Track" icon="download" subtitle="How to consume a track." >}}
{{< /cards >}}
