package moqtransfork

type Handler interface {
	SetupHandler
	AnnounceHandler
	SubscribeHandler
	//TrackStatusHandler
}
