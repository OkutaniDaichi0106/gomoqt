package moqt

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var _ AnnouncementWriter = (*announcementsBuffer)(nil)
var _ AnnouncementReader = (*announcementsBuffer)(nil)

func newAnnouncementBuffer(config AnnounceConfig) *announcementsBuffer {
	return &announcementsBuffer{
		config:        config,
		announcements: make([]*Announcement, 0),
		notifyCh:      make(chan struct{}, 1),
	}
}

type announcementsBuffer struct {
	config        AnnounceConfig
	announcements []*Announcement
	mu            sync.Mutex
	notifyCh      chan struct{}

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

	ab.announcements = append(ab.announcements, announcements...)

	select {
	case ab.notifyCh <- struct{}{}:
	default:
		// Skip if the channel is already full
	}

	return nil
}

func (ab *announcementsBuffer) ReceiveAnnouncements(ctx context.Context) ([]*Announcement, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	for {
		ab.mu.Lock()

		// Check if the buffer is closed
		if ab.closed {
			err := ab.closedErr
			if ab.closedErr == nil {
				err = errors.New("announcement buffer is already closed")
			}
			ab.mu.Unlock()
			return nil, err
		}

		// Check if there are any announcements
		if len(ab.announcements) > 0 {
			defer ab.mu.Unlock()
			announcements := ab.announcements
			ab.announcements = ab.announcements[:0]
			return announcements, nil
		}

		ab.mu.Unlock()

		// Wait for the next announcement
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ab.notifyCh:
			continue
		}
	}
}

func (ab *announcementsBuffer) AnnounceConfig() AnnounceConfig {
	return ab.config
}

func (ab *announcementsBuffer) Close() error {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	if ab.closed {
		return errors.New("announcement buffer is already closed")
	}

	ab.closed = true
	close(ab.notifyCh)
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
	close(ab.notifyCh)

	return nil
}
