package moqt

const (
	AnyTrackPattern = "/**"
)

type AnnounceConfig struct {
	TrackPattern string
}

func (ac AnnounceConfig) String() string {
	return "TrackPattern: " + ac.TrackPattern
}
