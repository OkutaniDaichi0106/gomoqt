package moqt

var _ Handler = (*MockTrackHandler)(nil)

// MockTrackHandler is a mock implementation of the TrackHandler interface
// It allows customizing behavior for testing different scenarios
type MockTrackHandler struct {
	// Function fields to customize behavior
	ServeTrackFunc         func(w TrackWriter, config *SubscribeConfig)
	GetInfoFunc            func(path TrackPath) (Info, error)
	ServeAnnouncementsFunc func(w AnnouncementWriter, config *AnnounceConfig)
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

func (m *MockTrackHandler) ServeAnnouncements(w AnnouncementWriter, config *AnnounceConfig) {
	if m.ServeAnnouncementsFunc != nil {
		m.ServeAnnouncementsFunc(w, config)
	}
}
