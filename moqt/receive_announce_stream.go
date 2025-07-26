package moqt

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newReceiveAnnounceStream(stream quic.Stream, prefix string, init map[string]*Announcement) *AnnouncementReader {
	annstr := &AnnouncementReader{
		streamCtx:   stream.Context(),
		stream:      stream,
		prefix:      prefix,
		active:      init,
		pendings:    make([]*Announcement, 0),
		announcedCh: make(chan struct{}, 1),
	}

	for _, ann := range init {
		annstr.pendings = append(annstr.pendings, ann)
	}

	slog.Info("announcement reader initialized", "prefix", prefix, "active", len(init))

	// Receive announcements in a separate goroutine
	go annstr.listenAnnouncements()

	return annstr
}

type AnnouncementReader struct {
	stream quic.Stream
	prefix string

	streamCtx context.Context

	// Track Suffix -> Announcement
	active map[string]*Announcement

	pendings    []*Announcement
	announcedCh chan struct{} // notify when new announcement is available
	pendingMu   sync.Mutex

	listenOnce sync.Once
}

func (ras *AnnouncementReader) ReceiveAnnouncement(ctx context.Context) (*Announcement, error) {
	for {
		ras.pendingMu.Lock()

		slog.Info("waiting for announcement", "prefix", ras.prefix)

		streamCtx := ras.streamCtx

		if streamCtx.Err() != nil {
			ras.pendingMu.Unlock()

			reason := context.Cause(streamCtx)
			var strErr *quic.StreamError
			if errors.As(reason, &strErr) {
				return nil, &AnnounceError{
					StreamError: strErr,
				}
			}

			return nil, reason
		}

		slog.Info("pending announcements available", "count", len(ras.pendings))

		if len(ras.pendings) > 0 {
			slog.Info("pending announcements available", "count", len(ras.pendings))
			next := ras.pendings[0]
			ras.pendings = ras.pendings[1:]

			ras.pendingMu.Unlock()

			return next, nil
		}

		announceCh := ras.announcedCh
		ras.pendingMu.Unlock()

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-streamCtx.Done():
			reason := context.Cause(streamCtx)
			return nil, reason
		case <-announceCh:
			// New announcement available, loop to check pendings
			continue
		}
	}
}

func (ras *AnnouncementReader) Close() error {
	ras.pendingMu.Lock()

	if ras.streamCtx.Err() != nil {
		ras.pendingMu.Unlock()
		return nil
	}

	close(ras.announcedCh)
	ras.announcedCh = nil

	ras.pendingMu.Unlock()

	return ras.stream.Close()
}

func (ras *AnnouncementReader) CloseWithError(code AnnounceErrorCode) error {
	ras.pendingMu.Lock()

	if ras.streamCtx.Err() != nil {
		ras.pendingMu.Unlock()
		return nil
	}

	close(ras.announcedCh)
	ras.announcedCh = nil

	ras.pendingMu.Unlock()

	strErrCode := quic.StreamErrorCode(code)
	ras.stream.CancelRead(strErrCode)
	ras.stream.CancelWrite(strErrCode)

	return nil
}

func (ras *AnnouncementReader) listenAnnouncements() {
	ras.listenOnce.Do(func() {
		var am message.AnnounceMessage
		var err error
		ctx := ras.streamCtx
		for {
			// Check if announcement is already closed before decoding
			if ctx.Err() != nil {
				return
			}

			err = am.Decode(ras.stream)
			if err != nil {
				return
			}

			slog.Debug("received announce message", "message", am)

			// Check if announcement is already closed during decoding
			if ctx.Err() != nil {
				return
			}

			ras.pendingMu.Lock()
			old, ok := ras.active[am.TrackSuffix]

			switch am.AnnounceStatus {
			case message.ACTIVE:
				if !ok || (ok && !old.IsActive()) {
					// Create a new announcement
					ann := NewAnnouncement(ras.streamCtx, BroadcastPath(ras.prefix+am.TrackSuffix))
					ras.active[am.TrackSuffix] = ann
					ras.pendings = append(ras.pendings, ann)

					// Notify that new announcement is available
					select {
					case ras.announcedCh <- struct{}{}:
					default:
					}

					ras.pendingMu.Unlock()

					continue
				} else {
					// Release lock before calling CloseWithError to avoid deadlock
					ras.pendingMu.Unlock()
					// Close the stream with an error
					ras.CloseWithError(DuplicatedAnnounceErrorCode)
					return
				}
			case message.ENDED:
				if ok && old.IsActive() {
					// End the existing announcement
					old.End()

					// Remove the announcement from the map
					delete(ras.active, am.TrackSuffix)
					ras.pendingMu.Unlock()
					continue
				} else {
					// Release lock before calling CloseWithError to avoid deadlock
					ras.pendingMu.Unlock()
					ras.CloseWithError(DuplicatedAnnounceErrorCode)
					return
				}
			default:
				ras.pendingMu.Unlock()
				// Unsupported status, close with error
				ras.CloseWithError(InvalidAnnounceStatusErrorCode)
				return
			}
		}
	})
}
