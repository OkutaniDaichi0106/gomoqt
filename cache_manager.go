package moqt

type CacheManager interface {
	GetFrameData(string, string, GroupSequence, uint64) []byte
}

//TODO:
