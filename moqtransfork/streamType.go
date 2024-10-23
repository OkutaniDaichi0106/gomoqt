package moqtransfork

type StreamType byte

const (
	SESSION   StreamType = 0x0
	ANNOUNCE  StreamType = 0x1
	SUBSCRIBE StreamType = 0x2
	FETCH     StreamType = 0x3
	INFO      StreamType = 0x4
)
