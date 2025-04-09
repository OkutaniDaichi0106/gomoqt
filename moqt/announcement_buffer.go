package moqt

import (
	"errors"
	"fmt"
	"sync"
)

var _ AnnouncementWriter = (*announcementsBuffer)(nil)

func newAnnouncementBuffer(config *AnnounceConfig) *announcementsBuffer {
	return &announcementsBuffer{
		config:        config,
		announcements: make([]*Announcement, 0),
		cond:          sync.NewCond(&sync.Mutex{}),
	}
}

type announcementsBuffer struct {
	config        *AnnounceConfig
	announcements []*Announcement
	mu            sync.RWMutex
	cond          *sync.Cond

	closed    bool
	closedErr error
}

func (ab *announcementsBuffer) SendAnnouncements(announcements []*Announcement) error {
	if len(announcements) == 0 {
		return nil
	}

	ab.mu.Lock()
	defer ab.mu.Unlock()

	if ab.closed {
		if ab.closedErr == nil {
			return errors.New("announcement buffer is already closed")
		}
		return ab.closedErr
	}

	// Append the new active announcements
	for _, ann := range announcements {
		if ann != nil && ann.IsActive() {
			ab.announcements = append(ab.announcements, ann)
		}
	}

	ab.cond.Broadcast()

	return nil
}

func (ab *announcementsBuffer) ServeAnnouncements(w AnnouncementWriter, config *AnnounceConfig) {
	pos := 0
	for {
		for pos > len(ab.announcements) {
			ab.cond.Wait()
		}

		// Check if the buffer is closed
		if ab.closed {
			err := ab.closedErr
			if ab.closedErr == nil {
				w.Close()
				return
			}

			w.CloseWithError(err)
			return
		}

		// Send the announcements from the current position to the end of the buffer
		anns := ab.announcements[pos:]
		err := w.SendAnnouncements(anns)
		if err != nil {
			w.CloseWithError(err)
			return
		}

		pos = len(ab.announcements)
	}
}

func (ab *announcementsBuffer) Close() error {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	if ab.closed {
		return errors.New("announcement buffer is already closed")
	}

	ab.closed = true

	ab.cond.Broadcast()

	return nil
}

func (ab *announcementsBuffer) CloseWithError(err error) error {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	if ab.closed {
		if ab.closedErr != nil {
			return fmt.Errorf("announcement buffer is already closed with error: %w", ab.closedErr)
		}
		return errors.New("announcement buffer is already closed")
	}

	// If no error is provided, use a default error
	if err == nil {
		err = ErrInternalError
	}

	ab.closedErr = err
	ab.closed = true

	ab.cond.Broadcast()

	return nil
}
