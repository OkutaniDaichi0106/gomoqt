package moqt

type InfoHandler interface {
	HandleInfo(Info, InfoResponcdWriter)
}

type InfoRequestHandler interface {
	HandleInfoRequest(InfoRequest, InfoWriter)
}

type InfoRequest struct{}

type Info struct{}

type InfoWriter interface {
	Answer(Info)
}
