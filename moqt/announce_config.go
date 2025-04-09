package moqt

type AnnounceConfig struct {
	TrackPattern string
}

func (ac AnnounceConfig) String() string {
	return "TrackPattern: " + ac.TrackPattern
}
