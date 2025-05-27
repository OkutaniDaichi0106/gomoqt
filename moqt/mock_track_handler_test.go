package moqt

var _ TrackHandler = (*MockTrackHandler)(nil)

// MockTrackHandler is a mock implementation of the TrackHandler interface
// It allows customizing behavior for testing different scenarios
type MockTrackHandler struct {
	// Function fields to customize behavior
	ServeTrackFunc func(pub *Publisher)
}

func (m *MockTrackHandler) ServeTrack(pub *Publisher) {
	if m.ServeTrackFunc != nil {
		m.ServeTrackFunc(pub)
	}
}
