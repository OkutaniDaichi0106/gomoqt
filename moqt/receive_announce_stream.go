package moqt

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

type AnnouncementReader interface {
	ReceiveAnnouncements(context.Context) ([]*Announcement, error)
	AnnounceConfig() AnnounceConfig
	Close() error
	CloseWithError(error) error
}

func newReceiveAnnounceStream(stream *internal.ReceiveAnnounceStream) *receiveAnnounceStream {
	annstr := &receiveAnnounceStream{
		internalStream: stream,
		announcements:  make(map[string]*Announcement),
		next:           make([]*Announcement, 0),
		liveCh:         make(chan struct{}),
	}

	go annstr.listenAnnouncements()

	return annstr
}

var _ AnnouncementReader = (*receiveAnnounceStream)(nil)

type receiveAnnounceStream struct {
	internalStream *internal.ReceiveAnnounceStream

	// Track Suffix -> Announcement
	announcements map[string]*Announcement

	next []*Announcement
	mu   sync.RWMutex

	liveCh chan struct{}

	closed   bool
	closeErr error
}

func (ras *receiveAnnounceStream) ReceiveAnnouncements(ctx context.Context) ([]*Announcement, error) {
	ras.mu.RLock()
	defer ras.mu.RUnlock()

	for {
		if len(ras.next) > 0 {
			next := ras.next

			// Clear the next list
			ras.next = ras.next[:0]

			return next, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ras.liveCh:
			continue
		}
	}
}

func (ras *receiveAnnounceStream) AnnounceConfig() AnnounceConfig {
	return AnnounceConfig{
		TrackPattern: ras.internalStream.AnnouncePleaseMessage.TrackPattern,
	}
}

func (ras *receiveAnnounceStream) Close() error {
	ras.mu.Lock()
	defer ras.mu.Unlock()

	if ras.closed {
		if ras.closeErr == nil {
			return fmt.Errorf("stream has already closed due to: %v", ras.closeErr)
		}

		return errors.New("stream has already closed")
	}

	ras.closed = true

	return ras.internalStream.Close()
}

func (ras *receiveAnnounceStream) CloseWithError(err error) error {
	ras.mu.Lock()
	defer ras.mu.Unlock()

	if ras.closed {
		if ras.closeErr == nil {
			return fmt.Errorf("stream has already closed due to: %v", ras.closeErr)
		}

		return errors.New("stream has already closed")
	}

	if err == nil {
		err = ErrInternalError
	}

	ras.closed = true
	ras.closeErr = err

	return ras.internalStream.CloseWithError(err)
}

func (ras *receiveAnnounceStream) listenAnnouncements() {
	prefix := ras.internalStream.AnnouncePleaseMessage.TrackPattern

	var am message.AnnounceMessage
	for {
		err := ras.internalStream.ReceiveAnnounceMessage(&am)
		if err != nil {
			ras.CloseWithError(err) // TODO: is this correct?
			return
		}

		announcement, ok := ras.announcements[am.TrackSuffix]

		switch am.AnnounceStatus {
		case message.ACTIVE:
			if !ok || (ok && !announcement.IsActive()) {
				slog.Debug("active")
				// Create a new announcement
				ann := NewAnnouncement(TrackPath(prefix + am.TrackSuffix))

				ras.addAnnouncement(am.TrackSuffix, ann)
			} else {
				err := errors.New("announcement is already active")
				slog.Error(err.Error(), "track_path", announcement.TrackPath)

				// Close the stream with an error
				ras.CloseWithError(err)
				return
			}
		case message.ENDED:
			if ok && announcement.IsActive() {
				// End the existing announcement
				err := announcement.End()
				if err != nil {
					slog.Error("failed to end a track", "error", err, "track_path", announcement.TrackPath)
				}

				ras.removeAnnouncement(am.TrackSuffix)
			} else {
				err := errors.New("announcement is already ended")
				slog.Error(err.Error(), "track_path", TrackPath(prefix+am.TrackSuffix))

				ras.CloseWithError(err)
				return
			}
		case message.LIVE:
			slog.Debug("all track are announced")
			select {
			case ras.liveCh <- struct{}{}:
			default:
			}
		}
	}
}

func (ras *receiveAnnounceStream) addAnnouncement(suffix string, ann *Announcement) {
	ras.mu.Lock()
	defer ras.mu.Unlock()

	ras.announcements[suffix] = ann
	ras.next = append(ras.next, ann)
}

func (ras *receiveAnnounceStream) removeAnnouncement(suffix string) {
	ras.mu.Lock()
	defer ras.mu.Unlock()

	delete(ras.announcements, suffix)
}
