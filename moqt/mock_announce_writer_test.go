package moqt

var _ AnnouncementWriter = (*MockAnnouncementWriter)(nil)

// MockAnnouncementWriter is a mock implementation of the AnnouncementWriter interface
// It tracks announcements for verification in tests
type MockAnnouncementWriter struct {
	SendAnnouncementsFunc func(announcements []*Announcement) error
	CloseFunc             func() error
	CloseWithErrorFunc    func(err error) error
}

func (m *MockAnnouncementWriter) SendAnnouncements(announcements []*Announcement) error {
	if m.SendAnnouncementsFunc != nil {
		return m.SendAnnouncementsFunc(announcements)
	}
	return nil
}

func (m *MockAnnouncementWriter) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockAnnouncementWriter) CloseWithError(err error) error {
	if m.CloseWithErrorFunc != nil {
		return m.CloseWithErrorFunc(err)
	}
	return nil
}
