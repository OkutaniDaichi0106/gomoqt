package moqt

type CacheManager interface {
	GetGroupData(string, string, GroupSequence) []byte
}

//TODO:
