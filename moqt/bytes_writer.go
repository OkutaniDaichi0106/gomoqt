package moqt

type directBytesWriter interface {
	newBytesWriter() writer
}

type writer interface {
	Write(p *[]byte) (int, error)
}
