package moqt

import "github.com/stretchr/testify/mock"

var _ AnnouncementWriter = (*MockAnnouncementWriter)(nil)

// MockAnnouncementWriter is a mock implementation of the AnnouncementWriter interface
// It tracks announcements for verification in tests
type MockAnnouncementWriter struct {
	mock.Mock
}

func (m *MockAnnouncementWriter) SendAnnouncement(announcement *Announcement) error {
	args := m.Called(announcement)
	return args.Error(0)
}

func (m *MockAnnouncementWriter) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAnnouncementWriter) CloseWithError(code AnnounceErrorCode) error {
	args := m.Called(code)
	return args.Error(0)
}
