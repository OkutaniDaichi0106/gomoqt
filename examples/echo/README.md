# Echo Sample

## Run the example
- **Make cert files**

command

```
mkcert localhost
```

- **Run a server**

command

```
go run ./server/main.go
```

log

```
time=2025-01-24T01:46:26.404+09:00 level=INFO msg="Server runs on path: \"/path\""
time=2025-01-24T01:46:26.423+09:00 level=INFO msg="Starting the server"
time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Waiting for an Announce Stream"

time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Running a subscriber"

time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Opening an Announce Stream"

time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Opened an Announce Stream"

time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Receiving announcements"

time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Received an announce request"
    config=TrackPrefix: [japan, kyoto], Parameters: Parameters: { }
time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Accepted an Announce Stream"

time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Announcing"

time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Successfully Announced"

time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Waiting for a subscribe stream"

time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Received announcements"
    announcements=[Announcement: { AnnounceStatus: ACTIVE, TrackPath: [japan, kyoto, kiu, text], AnnounceParameters: Parameters: { } }]
time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Subscribing"

time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Received a subscribe request"
    config=SubscribeConfig: { TrackPath: [japan, kyoto, kiu, text], TrackPriority: 0, GroupOrder: 0, MinGroupSequence: 0, MaxGroupSequence: 0, SubscribeParameters: Parameters: { } }
time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Accepted a subscribe stream"

time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Subscribed"
    info=Info: { TrackPriority: 0, LatestGroupSequence: 0, GroupOrder: 0 }
time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Receiving data"

time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Received a frame"
    frame=
        ☆         o
         ☆    ☆
          ☆       ☆
        ☆       o   ☆
         ☆       ☆
      o         ☆

time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Received a frame"
    frame=
        ☆            ☆
         ☆   o
       o     ☆
     ☆       ☆   ☆
       ☆   o        ☆
           ☆     ☆

time=2025-01-24T01:46:31+09:00 level=INFO
    msg="Received a frame"
    frame=
      o          ☆
       ☆       ☆
         ☆     ☆
    ☆         ☆      o
       ☆  o         ☆
         ☆
```

- **Run a client**

command

```
go run ./client/main.go
```

log

```
time=2025-01-24T01:46:30.586+09:00 level=INFO msg="Dial to the server"
time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Running a subscriber"

time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Opening an Announce Stream"

time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Running a publisher"

time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Waiting an Announce Stream"

time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Opened an Announce Stream"

time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Receiving announcements"

time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Received an announce request"
    config=TrackPrefix: [japan, kyoto], Parameters: Parameters: { }
time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Accepted an Announce Stream"

time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Send Announcements"

time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Announced"

time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Waiting a subscribe stream"

time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Received announcements"
    announcements=[Announcement: { AnnounceStatus: ACTIVE, TrackPath: [japan, kyoto, kiu, text], AnnounceParameters: Parameters: { } }]
time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Subscribing"

time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Received a subscribe request"
    config=SubscribeConfig: { TrackPath: [japan, kyoto, kiu, text], TrackPriority: 0, GroupOrder: 0, MinGroupSequence: 0, MaxGroupSequence: 0, SubscribeParameters: Parameters: { } }
time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Subscribed"
    info=Info: { TrackPriority: 0, LatestGroupSequence: 0, GroupOrder: 0 }
time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Received data"
    data=
        ☆         o
         ☆    ☆
          ☆       ☆
        ☆       o   ☆
         ☆       ☆
      o         ☆

time=2025-01-24T01:46:30+09:00 level=INFO
    msg="Received data"
    data=
        ☆            ☆
         ☆   o
       o     ☆
     ☆       ☆   ☆
       ☆   o        ☆
           ☆     ☆

time=2025-01-24T01:46:31+09:00 level=INFO
    msg="Received data"
    data=
      o          ☆
       ☆       ☆
         ☆     ☆
    ☆         ☆      o
       ☆  o         ☆
         ☆
```
