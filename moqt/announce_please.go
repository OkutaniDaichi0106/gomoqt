package moqt

type AnnounceConfig struct {
	TrackPrefix []string
	Parameters  Parameters
}

func (ac AnnounceConfig) String() string {
	return "TrackPrefix: " + TrackPartsString(ac.TrackPrefix) + ", Parameters: " + ac.Parameters.String()
}
