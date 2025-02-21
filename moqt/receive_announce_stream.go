package moqt

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

type AnnouncementReader interface {
	NextAnnouncements(context.Context) ([]*Announcement, error)
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

	next   []*Announcement
	liveCh chan struct{}

	closed   bool
	closeErr error
	mu       sync.RWMutex
}

func (ras *receiveAnnounceStream) NextAnnouncements(ctx context.Context) ([]*Announcement, error) {
	ras.mu.RLock()
	defer ras.mu.RUnlock()

	for {
		if len(ras.next) > 0 {
			next := ras.next
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
		TrackPrefix: ras.internalStream.AnnouncePleaseMessage.TrackPrefix,
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
		return ras.Close()
	}

	ras.closed = true
	ras.closeErr = err

	return ras.internalStream.CloseWithError(err)
}

func (ras *receiveAnnounceStream) listenAnnouncements() {
	prefix := ras.internalStream.AnnouncePleaseMessage.TrackPrefix

	var am message.AnnounceMessage
	for {
		err := ras.internalStream.ReadAnnounceMessage(&am)
		if err != nil {
			ras.CloseWithError(err) // TODO: is this correct?
			return
		}
		switch am.AnnounceStatus {
		case message.ACTIVE:
			_, ok := ras.announcements[am.TrackSuffix]
			if !ok {
				ann := NewAnnouncement(TrackPath(prefix + am.TrackSuffix))
				ras.announcements[am.TrackSuffix] = ann
				ras.next = append(ras.next, ann)
			} else {
				// Activate the existing announcement
				ras.announcements[am.TrackSuffix].activate()
			}
		case message.ENDED:
			_, ok := ras.announcements[am.TrackSuffix]
			if ok {
				// End the existing announcement
				ras.announcements[am.TrackSuffix].end()
			} else {
				continue
				// TODO: handle this case as an error
			}
		case message.LIVE:
			select {
			case ras.liveCh <- struct{}{}:
			default:
			}
		}
	}
}
