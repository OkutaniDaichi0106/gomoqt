package moqt

var _ TrackHandler = (*MockTrackHandler)(nil)

// MockTrackHandler is a mock implementation of the TrackHandler interface
// It allows customizing behavior for testing different scenarios
type MockTrackHandler struct {
	// Function fields to customize behavior
	ServeTrackFunc func(w TrackWriter, sub ReceivedSubscription)
	GetInfoFunc    func() (Info, bool)
	SendGapFunc    func(gap Gap) error
}

func (m *MockTrackHandler) HandleTrack(w TrackWriter, sub ReceivedSubscription) {
	if m.ServeTrackFunc != nil {
		m.ServeTrackFunc(w, sub)
	}
}

func (m *MockTrackHandler) Info() (Info, bool) {
	if m.GetInfoFunc != nil {
		return m.GetInfoFunc()
	}
	return Info{}, true
}
