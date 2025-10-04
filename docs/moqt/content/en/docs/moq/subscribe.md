---
title: Subscribe
weight: 9
---

Subscribing is the mechanism for receiving media tracks from a broadcast, either as a client or server.

## Subscribe to a Track

To subscribe to a track, you need to call the `(moqt.Session).Subscribe` method.


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

- **Broadcast Path**: The path to the broadcast you want to subscribe to. This is an unique string that identifies the broadcast.
- **Track Name**:
  The track name is an identifier for the track within the broadcast.
- **Track Config**:
  The track config is used to specify additional options for the track subscription, such as the priority or range.


By specifying options in the `moqt.TrackConfig` when calling `(moqt.Session).Subscribe`, you can configure the initial subscription parameters.

### Control Subscription

You can adjust the subscription parameters at any time by calling the `(moqt.TrackReader).Update` method. This allows you to change options such as the priority.

```go
    var tr *moqt.TrackReader

    config := &moqt.TrackConfig{
        // Specify updated track configuration options here
    }
    tr.Update(config)

    // Handle the TrackReader
```

## Announced Broadcasts

Before subscribing to a track, you may want to discover available broadcasts.
To do this, peers can specify the prefix for the broadcast path they are interested in and listen for announcements.

{{<cards>}}
    {{< card link="../announce/#discover-broadcasts" title="Discover Broadcasts" icon="external-link">}}
{{</cards>}}

After receiving an `moqt.Announcement`, broadcast path can be obtained using the `(moqt.Announcement).BroadcastPath` method.

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

### Discover Available Tracks

The Announce mechanism clarifies which broadcasts are available, but does not specify which tracks belong to each broadcast.
It is up to the application logic to associate tracks with their respective broadcasts.
When you access the broadcast's track, you have to know its name in advance or out-of-band.

[Hang protocol](../hang/#protocol) can be helpful as a way to negotiate track names.

## Handle Track

The `Subscribe` method returns a `TrackReader`. The `TrackReader` represents a subscription to a track.

### Managing the Track

`(moqt.TrackReader).Update` can be used to change the subscription parameters (e.g., priority) during the session.
When no longer needed, make sure to call `(moqt.TrackReader).Close` to unsubscribe from the track.
If you specify some reason for closing, `(moqt.TrackReader).CloseWithError` which notifies the reason to the sender can be used.

### Consuming a Track

{{< cards >}}
	{{< card link="../track_group_frame/#consume-a-track" title="Consume a Track" icon="download" subtitle="How to consume a track." >}}
{{< /cards >}}
