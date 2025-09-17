---
title: Announce & Discover
weight: 8
---

Announcements notify which broadcasts exist and whether they are active. They are essential for track discovery in a session. You can send announcements from a `*moqt.Session` by using the `Publish`, `PublishFunc`, or `Announce` methods on the associated `*moqt.TrackMux`.

When a track is published on a `moqt.TrackMux`, it distributes `Announcement` to all listeners.


## Announce Broadcasts

Broadcasts are announced when they are registered with the `moqt.TrackMux`. This can be done using the `(*moqt.TrackMux).Announce` method, or by using `(*moqt.TrackMux).Publish` or `(*moqt.TrackMux).PublishFunc`, which internally create the announcement.

```go
    var mux *moqt.TrackMux
    var sess *moqt.Session // Holds the mux

    // Register and announce a broadcast
    ann, end := NewAnnouncement(ctx, "/broadcast_path")
    mux.Announce(ann, trackHandler)
    defer end() // Cleanup when done

    // Or use Publish and PublishFunc (serves Announcement internally)
    mux.Publish(ctx, "/broadcast_path", trackHandler)
    mux.PublishFunc(ctx, "/broadcast_path", trackHandleFunc)
```

## Discover Broadcasts

Peers can discover available broadcasts by specifying the prefix for the broadcast path they are interested in and listening for announcements.

To be able to listen for announcements with a specific prefix, use the `(*moqt.Session).AcceptAnnounce` method. This returns a `*moqt.AnnouncementReader`, which allows you to read incoming announcements.

```go
var sess *moqt.Session

ar, err := sess.AcceptAnnounce("/prefix/")
if err != nil {
    // Handle error
}

for {
    ann, err := ar.ReceiveAnnouncement()
    if err != nil {
        // Handle error
        break
    }
    // Handle announcement
}
```
