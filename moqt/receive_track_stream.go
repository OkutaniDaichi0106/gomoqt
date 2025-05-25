package moqt

type ReceiveTrackStream struct {
	BroadcastPath   BroadcastPath
	TrackName       string
	TrackReader     TrackReader
	SubscribeStream *SendSubscribeStream
}
