package moqt

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newReceiveAnnounceStream(stream quic.Stream, prefix string, suffixes []string) *AnnouncementReader {
	ar := &AnnouncementReader{
		ctx:         context.WithValue(stream.Context(), &biStreamTypeCtxKey, message.StreamTypeAnnounce),
		stream:      stream,
		prefix:      prefix,
		active:      make(map[string]*Announcement),
		pendings:    make([]*Announcement, 0),
		announcedCh: make(chan struct{}, 1),
	}

	for _, suffix := range suffixes {
		ann := NewAnnouncement(stream.Context(), BroadcastPath(prefix+suffix))
		ar.active[suffix] = ann
		ar.pendings = append(ar.pendings, ann)
	}

	// Receive announcements in a separate goroutine
	go func() {
		var am message.AnnounceMessage
		var err error
		ctx := ar.ctx
		for {
			// Check if announcement is already closed before decoding
			if ctx.Err() != nil {
				return
			}

			err = am.Decode(ar.stream)
			if err != nil {
				return
			}

			slog.Debug("received announce message", "message", am)

			// Check if announcement is already closed during decoding
			if ctx.Err() != nil {
				return
			}

			ar.pendingMu.Lock()
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

					ar.pendingMu.Unlock()

					continue
				} else {
					// Release lock before calling CloseWithError to avoid deadlock
					ar.pendingMu.Unlock()
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
					ar.pendingMu.Unlock()
					continue
				} else {
					// Release lock before calling CloseWithError to avoid deadlock
					ar.pendingMu.Unlock()
					ar.CloseWithError(DuplicatedAnnounceErrorCode)
					return
				}
			default:
				ar.pendingMu.Unlock()
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
	prefix string

	ctx context.Context

	// Track Suffix -> Announcement
	active map[string]*Announcement

	pendings    []*Announcement
	announcedCh chan struct{} // notify when new announcement is available
	pendingMu   sync.Mutex
}

func (ras *AnnouncementReader) ReceiveAnnouncement(ctx context.Context) (*Announcement, error) {
	for {
		ras.pendingMu.Lock()

		slog.Info("waiting for announcement", "prefix", ras.prefix)

		streamCtx := ras.ctx

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

	if ras.ctx.Err() != nil {
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

	if ras.ctx.Err() != nil {
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
