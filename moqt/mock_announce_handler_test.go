package moqt

type MockAnnouncementHandler struct {
	ServeAnnouncementsFunc func(w AnnouncementWriter, config *AnnounceConfig)
}

func (m *MockAnnouncementHandler) ServeAnnouncements(w AnnouncementWriter, config *AnnounceConfig) {
	if m.ServeAnnouncementsFunc != nil {
		m.ServeAnnouncementsFunc(w, config)
	}
}
