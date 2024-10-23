package moqtransfork

type AnnounceStream Stream

type Announcement struct {
	TrackNamespace    []string
	AuthorizationInfo string
	//Parames           map[uint64]any
}

type AnnounceResponceWriter interface {
	Accept()
	Reject(AnnounceError)
}

type AnnounceHandler interface {
	HandleAnnounce(Announcement, AnnounceResponceWriter)
}

type AnnounceHandlerFunc func(Announcement, AnnounceResponceWriter)

func (op AnnounceHandlerFunc) HandleAnnounce(a Announcement, arw AnnounceResponceWriter) {
	op(a, arw)

}
