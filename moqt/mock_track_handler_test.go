package moqt_test

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

var _ moqt.TrackHandler = (*MockTrackHandler)(nil)

// MockTrackHandler is a mock implementation of the TrackHandler interface
// It allows customizing behavior for testing different scenarios
type MockTrackHandler struct {
	// Function fields to customize behavior
	ServeTrackFunc         func(w moqt.TrackWriter, config *moqt.SubscribeConfig)
	GetInfoFunc            func(path moqt.TrackPath) (moqt.Info, error)
	ServeAnnouncementsFunc func(w moqt.AnnouncementWriter, config *moqt.AnnounceConfig)
}

func (m *MockTrackHandler) ServeTrack(w moqt.TrackWriter, config *moqt.SubscribeConfig) {
	if m.ServeTrackFunc != nil {
		m.ServeTrackFunc(w, config)
	}
}

func (m *MockTrackHandler) GetInfo(path moqt.TrackPath) (moqt.Info, error) {
	if m.GetInfoFunc != nil {
		return m.GetInfoFunc(path)
	}
	return moqt.Info{
		TrackPriority:       1,
		LatestGroupSequence: 0,
		GroupOrder:          0,
	}, nil
}

func (m *MockTrackHandler) ServeAnnouncements(w moqt.AnnouncementWriter, config *moqt.AnnounceConfig) {
	if m.ServeAnnouncementsFunc != nil {
		m.ServeAnnouncementsFunc(w, config)
	}
}
