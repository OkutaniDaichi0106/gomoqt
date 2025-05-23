package moqt

const (
	anyTrackPattern = "/**"
)

var (
	anyAnnounceConfig = &AnnounceConfig{
		TrackPattern: anyTrackPattern,
	}
)

type AnnounceConfig struct {
	TrackPattern string
}

func (ac AnnounceConfig) String() string {
	return "TrackPattern: " + ac.TrackPattern
}
