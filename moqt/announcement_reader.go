package moqt

import (
	"context"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/quic"
)

func newAnnouncementReader(stream quic.Stream, prefix prefix, initSuffixes []suffix) *AnnouncementReader {
	if !isValidPrefix(prefix) {
		panic("invalid prefix for AnnouncementReader")
	}

	ar := &AnnouncementReader{
		ctx:         context.WithValue(stream.Context(), &biStreamTypeCtxKey, message.StreamTypeAnnounce),
		stream:      stream,
		prefix:      prefix,
		actives:     make(map[suffix]*Announcement),
		pendings:    make([]*Announcement, 0),
		announcedCh: make(chan struct{}, 1),
	}

	for _, suffix := range initSuffixes {
		ann, _ := NewAnnouncement(stream.Context(), BroadcastPath(prefix+suffix))
		ar.actives[suffix] = ann
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

			old, ok := ar.actives[am.TrackSuffix]

			switch am.AnnounceStatus {
			case message.ACTIVE:
				if !ok || (ok && !old.IsActive()) {
					// Create a new announcement
					ann, _ := NewAnnouncement(ar.ctx, BroadcastPath(ar.prefix+am.TrackSuffix))
					ar.actives[am.TrackSuffix] = ann
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
					old.end()

					// Remove the announcement from the map
					delete(ar.actives, am.TrackSuffix)

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

// AnnouncementReader receives and manages broadcast announcements from a remote peer.
// It maintains a list of active announcements and notifies when new announcements
// are received or existing ones are canceled.
type AnnouncementReader struct {
	stream quic.Stream
	prefix prefix

	ctx context.Context

	// Track Suffix -> Announcement
	announcementsMu sync.Mutex

	actives map[suffix]*Announcement

	pendings    []*Announcement
	announcedCh chan struct{} // notify when new announcement is available
}

func (ras *AnnouncementReader) ReceiveAnnouncement(ctx context.Context) (*Announcement, error) {
	for {
		ras.announcementsMu.Lock()

		if len(ras.pendings) > 0 {
			next := ras.pendings[0]
			ras.pendings = ras.pendings[1:]

			ras.announcementsMu.Unlock()

			return next, nil
		}

		if ras.ctx.Err() != nil {
			ras.announcementsMu.Unlock()
			return nil, Cause(ras.ctx)
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

	if ras.announcedCh != nil {
		close(ras.announcedCh)
		ras.announcedCh = nil
	}

	strErrCode := quic.StreamErrorCode(code)
	ras.stream.CancelRead(strErrCode)
	ras.stream.CancelWrite(strErrCode)

	return nil
}

func (ras *AnnouncementReader) Context() context.Context {
	return ras.ctx
}

type suffix = string
type prefix = string
