package moqt

import (
	"context"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newAnnouncementReader(stream quic.Stream, prefix prefix, initSuffixes []suffix) *AnnouncementReader {
	ar := &AnnouncementReader{
		ctx:         context.WithValue(stream.Context(), &biStreamTypeCtxKey, message.StreamTypeAnnounce),
		stream:      stream,
		prefix:      prefix,
		active:      make(map[suffix]*Announcement),
		pendings:    make([]*Announcement, 0),
		announcedCh: make(chan struct{}, 1),
	}

	for _, suffix := range initSuffixes {
		ann := NewAnnouncement(stream.Context(), BroadcastPath(prefix+suffix))
		ar.active[suffix] = ann
		ar.pendings = append(ar.pendings, ann)
	}

	// Receive announcements in a separate goroutine
	go func() {
		var am message.AnnounceMessage
		var err error

		for {
			err = am.Decode(ar.stream)
			if err != nil {
				return
			}

			slog.Debug("received announce message", "message", am)

			// Check if announcement is already closed during decoding
			if ar.ctx.Err() != nil {
				return
			}

			ar.announcementsMu.Lock()

			old, ok := ar.active[am.TrackSuffix]

			switch am.AnnounceStatus {
			case message.ACTIVE:
				if !ok || (ok && !old.IsActive()) {
					// Create a new announcement
					ann := NewAnnouncement(ar.ctx, BroadcastPath(ar.prefix+am.TrackSuffix))
					ar.active[am.TrackSuffix] = ann
					ar.pendings = append(ar.pendings, ann)

					// Notify that new announcement is available
					select {
					case ar.announcedCh <- struct{}{}:
					default:
					}

					ar.announcementsMu.Unlock()

					continue
				} else {
					// Release lock before calling CloseWithError to avoid deadlock
					ar.announcementsMu.Unlock()

					// Close the stream with an error
					ar.CloseWithError(DuplicatedAnnounceErrorCode)

					return
				}
			case message.ENDED:
				if ok && old.IsActive() {
					// End the existing announcement
					old.End()

					// Remove the announcement from the map
					delete(ar.active, am.TrackSuffix)

					ar.announcementsMu.Unlock()
					continue
				} else {
					// Release lock before calling CloseWithError to avoid deadlock
					ar.announcementsMu.Unlock()
					ar.CloseWithError(DuplicatedAnnounceErrorCode)
					return
				}
			default:
				ar.announcementsMu.Unlock()

				// Unsupported status, close with error
				ar.CloseWithError(InvalidAnnounceStatusErrorCode)
				return
			}
		}
	}()

	return ar
}

type AnnouncementReader struct {
	stream quic.Stream
	prefix prefix

	ctx context.Context

	// Track Suffix -> Announcement
	announcementsMu sync.Mutex

	active map[suffix]*Announcement

	pendings    []*Announcement
	announcedCh chan struct{} // notify when new announcement is available
}

func (ras *AnnouncementReader) ReceiveAnnouncement(ctx context.Context) (*Announcement, error) {
	for {
		ras.announcementsMu.Lock()

		slog.Info("waiting for announcement", "prefix", ras.prefix)

		if ras.ctx.Err() != nil {
			ras.announcementsMu.Unlock()
			return nil, Cause(ras.ctx)
		}

		slog.Info("pending announcements available", "count", len(ras.pendings))

		if len(ras.pendings) > 0 {
			slog.Info("pending announcements available", "count", len(ras.pendings))
			next := ras.pendings[0]
			ras.pendings = ras.pendings[1:]

			ras.announcementsMu.Unlock()

			return next, nil
		}

		announceCh := ras.announcedCh

		ras.announcementsMu.Unlock()

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ras.ctx.Done():
			return nil, Cause(ras.ctx)
		case <-announceCh:
			// New announcement available, loop to check pendings
			continue
		}
	}
}

func (ras *AnnouncementReader) Close() error {
	ras.announcementsMu.Lock()
	defer ras.announcementsMu.Unlock()

	if ras.ctx.Err() != nil {
		return nil
	}

	if ras.announcedCh != nil {
		close(ras.announcedCh)
		ras.announcedCh = nil
	}

	return ras.stream.Close()
}

func (ras *AnnouncementReader) CloseWithError(code AnnounceErrorCode) error {
	ras.announcementsMu.Lock()
	defer ras.announcementsMu.Unlock()

	if ras.ctx.Err() != nil {
		return nil
	}

	close(ras.announcedCh)
	ras.announcedCh = nil

	strErrCode := quic.StreamErrorCode(code)
	ras.stream.CancelRead(strErrCode)
	ras.stream.CancelWrite(strErrCode)

	return nil
}

type suffix = string
type prefix = string
