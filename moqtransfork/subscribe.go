package moqtransfork

type Subscription struct {
	Announcement
	TrackName string
	//Parames           map[uint64]any
}

type SubscribeResponceWriter interface {
	Accept()
	Reject()
}

type SubscribeHandler interface {
	HandleSubscribe(Subscription, SubscribeResponceWriter)
}

var _ SubscribeHandler = (*SubscribeHandlerFunc)(nil)

type SubscribeHandlerFunc func(Subscription, SubscribeResponceWriter)

func (op SubscribeHandlerFunc) HandleSubscribe(s Subscription, srw SubscribeResponceWriter) {
	op(s, srw)
}
