package moqt

var _ TrackHandler = (*MockTrackHandler)(nil)

// MockTrackHandler is a mock implementation of the TrackHandler interface
// It allows customizing behavior for testing different scenarios
type MockTrackHandler struct {
	// Function fields to customize behavior
	ServeTrackFunc func(w TrackWriter, config *SubscribeConfig)
	GetInfoFunc    func(path TrackPath) (Info, error)
}

func (m *MockTrackHandler) ServeTrack(w TrackWriter, config *SubscribeConfig) {
	if m.ServeTrackFunc != nil {
		m.ServeTrackFunc(w, config)
	}
}

func (m *MockTrackHandler) GetInfo(path TrackPath) (Info, error) {
	if m.GetInfoFunc != nil {
		return m.GetInfoFunc(path)
	}
	return Info{
		TrackPriority:       1,
		LatestGroupSequence: 0,
		GroupOrder:          0,
	}, nil
}
