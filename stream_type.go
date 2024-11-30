package moqt

type StreamType byte

const (
	stream_type_session   StreamType = 0x0
	stream_type_announce  StreamType = 0x1
	stream_type_subscribe StreamType = 0x2
	stream_type_fetch     StreamType = 0x3
	stream_type_info      StreamType = 0x4
)
