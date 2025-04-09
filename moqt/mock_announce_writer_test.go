package moqt_test

import "github.com/OkutaniDaichi0106/gomoqt/moqt"

var _ moqt.AnnouncementWriter = (*MockAnnouncementWriter)(nil)

// MockAnnouncementWriter is a mock implementation of the AnnouncementWriter interface
// It tracks announcements for verification in tests
type MockAnnouncementWriter struct {
	ConfigValue     moqt.AnnounceConfig
	AnnouncedTracks []moqt.TrackPath
	Notifications   int
}

func (m *MockAnnouncementWriter) SendAnnouncements(announcements []*moqt.Announcement) error {
	for _, ann := range announcements {
		m.AnnouncedTracks = append(m.AnnouncedTracks, ann.TrackPath())
		m.Notifications++
	}
	return nil
}

func (m *MockAnnouncementWriter) Close() error {
	return nil
}

func (m *MockAnnouncementWriter) CloseWithError(err error) error {
	return err
}
