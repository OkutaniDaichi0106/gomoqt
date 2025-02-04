package moqt

type AnnounceConfig struct {
	TrackPrefix []string
	Parameters  Parameters
}

func (ac AnnounceConfig) String() string {
	return "TrackPrefix: " + TrackPath(ac.TrackPrefix).String() + ", Parameters: " + ac.Parameters.String()
}
