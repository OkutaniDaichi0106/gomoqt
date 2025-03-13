package moqt

type AnnounceConfig struct {
	TrackPattern string
}

func (ac AnnounceConfig) String() string {
	return "TrackPrefix: " + TrackPath(ac.TrackPattern).String()
}
