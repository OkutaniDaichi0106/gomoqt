package moqt

type AnnounceConfig struct {
	TrackPrefix string
}

func (ac AnnounceConfig) String() string {
	return "TrackPrefix: " + TrackPath(ac.TrackPrefix).String()
}
