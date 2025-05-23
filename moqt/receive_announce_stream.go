package moqt

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

type AnnouncementReader interface {
	ReceiveAnnouncements(context.Context) ([]*Announcement, error)
	Close() error
	CloseWithError(error) error
}

func newReceiveAnnounceStream(stream quic.Stream, config *AnnounceConfig) *receiveAnnounceStream {
	annstr := &receiveAnnounceStream{
		stream:        stream,
		config:        config,
		announcements: make(map[TrackPath]*Announcement),
		next:          make([]*Announcement, 0),
		liveCh:        make(chan struct{}, 1),
	}

	go annstr.listenAnnouncements()

	return annstr
}

var _ AnnouncementReader = (*receiveAnnounceStream)(nil)

type receiveAnnounceStream struct {
	stream quic.Stream
	config *AnnounceConfig

	closed   bool
	closeErr error

	// Track Suffix -> Announcement
	announcements map[TrackPath]*Announcement

	next []*Announcement
	mu   sync.Mutex

	liveCh chan struct{}
}

func (ras *receiveAnnounceStream) ReceiveAnnouncements(ctx context.Context) ([]*Announcement, error) {
	for {
		ras.mu.Lock()
		if len(ras.next) > 0 {
			next := ras.next

			// Clear the next list
			ras.next = ras.next[:0]

			ras.mu.Unlock()
			return next, nil
		}
		if ras.closed {
			ras.mu.Unlock()
			return nil, ras.closeErr
		}
		ras.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ras.liveCh:

		}

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

	err := ras.stream.Close()
	if err != nil {
		return err
	}

	slog.Debug("closed a receive announce stream",
		slog.Any("stream_id", ras.stream.StreamID()),
	)

	return nil
}

func (ras *receiveAnnounceStream) CloseWithError(err error) error {
	ras.mu.Lock()
	defer ras.mu.Unlock()

	if ras.closed {
		if ras.closeErr != nil {
			return ras.closeErr
		}
		return nil
	}

	if err == nil {
		err = ErrInternalError
	}

	ras.closed = true
	ras.closeErr = err

	var annerr AnnounceError
	if !errors.As(err, &annerr) {
		annerr = ErrInternalError
	}

	code := quic.StreamErrorCode(annerr.AnnounceErrorCode())

	ras.stream.CancelRead(code)
	ras.stream.CancelWrite(code)

	slog.Debug("closed a receive announce stream with an error",
		slog.Any("stream_id", ras.stream.StreamID()),
		slog.String("reason", err.Error()),
	)

	return nil
}

func (ras *receiveAnnounceStream) listenAnnouncements() {
	pattern := ras.config.TrackPattern

	var am message.AnnounceMessage
	for {
		_, err := am.Decode(ras.stream)
		if err != nil {
			slog.Error("failed to decode an ANNOUNCE message", "error", err)
			ras.CloseWithError(err) // TODO: is this correct?
			return
		}

		slog.Debug("received an ANNOUNCE message",
			"announce_message", am,
		)

		path := NewTrackPath(pattern, am.WildcardParameters...)

		old, ok := ras.announcements[path]

		switch am.AnnounceStatus {
		case message.ACTIVE:
			if !ok || (ok && !old.IsActive()) {
				slog.Debug("active")
				// Create a new announcement

				ras.addAnnouncement(NewAnnouncement(context.Background(), path))
			} else {
				slog.Error("announcement is already active", "track_path", old.TrackPath)

				// Close the stream with an error
				ras.CloseWithError(ErrProtocolViolation)
				return
			}
		case message.ENDED:
			if ok && old.IsActive() {
				// End the existing announcement
				old.End()

				// Remove the announcement from the map
				ras.removeAnnouncement(old.TrackPath())
			} else {
				slog.Error("announcement is already ended",
					"track_path", old.TrackPath(),
				)

				ras.CloseWithError(ErrProtocolViolation)
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

func (ras *receiveAnnounceStream) addAnnouncement(ann *Announcement) {
	ras.mu.Lock()
	defer ras.mu.Unlock()

	ras.announcements[ann.TrackPath()] = ann
	ras.next = append(ras.next, ann)
}

func (ras *receiveAnnounceStream) removeAnnouncement(path TrackPath) {
	ras.mu.Lock()
	defer ras.mu.Unlock()

	delete(ras.announcements, path)
}
