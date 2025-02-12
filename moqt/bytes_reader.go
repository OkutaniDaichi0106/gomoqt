package moqt

type directBytesReader interface {
	newBytesReader() reader
}

type reader interface {
	Read(p *[]byte) (int, error)
}
