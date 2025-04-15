package moqt

// type announcementsChan struct {
// 	announcements chan []*Announcement
// }

// var _ AnnouncementWriter = (*announcementsChan)(nil)

// func newAnnouncementsChan() *announcementsChan {
// 	return &announcementsChan{
// 		announcements: make(chan []*Announcement),
// 	}
// }

// func (ac *announcementsChan) SendAnnouncements(announcements []*Announcement) error {
// 	ac.announcements <- announcements
// 	return nil
// }

// func (ac *announcementsChan) ReceiveAnnouncements(ctx context.Context) ([]*Announcement, error) {
// 	select {
// 	case <-ctx.Done():
// 		return nil, ctx.Err()
// 	case anns := <-ac.announcements:
// 		return anns, nil
// 	}
// }

// func (ac *announcementsChan) Close() error {
// 	close(ac.announcements)
// 	return nil
// }

// func (ac *announcementsChan) CloseWithError(err error) error {
// 	close(ac.announcements)
// 	return nil
// }
