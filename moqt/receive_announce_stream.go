package moqt

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newReceiveAnnounceStream(stream quic.Stream, prefix string) *AnnouncementReader {
	ctx, cancel := context.WithCancelCause(context.Background())

	// Propagate the cancellation with AnnounceError
	go func() {
		streamCtx := stream.Context()
		<-streamCtx.Done()
		reason := context.Cause(streamCtx)
		var (
			strErr *quic.StreamError
			appErr *quic.ApplicationError
		)
		if errors.As(reason, &strErr) {
			reason = &SubscribeError{
				StreamError: strErr,
			}
		} else if errors.As(reason, &appErr) {
			reason = &SessionError{
				ApplicationError: appErr,
			}
		}
		cancel(reason)
	}()

	annstr := &AnnouncementReader{
		ctx:         ctx,
		cancel:      cancel,
		stream:      stream,
		prefix:      prefix,
		active:      make(map[string]*Announcement),
		pendings:    make([]*Announcement, 0),
		announcedCh: make(chan struct{}, 1),
	}

	// Receive announcements in a separate goroutine
	go annstr.listenAnnouncements()

	return annstr
}

type AnnouncementReader struct {
	stream quic.Stream
	prefix string

	ctx    context.Context
	cancel context.CancelCauseFunc

	closed bool

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

		annStrCtx := ras.ctx

		if ras.closed {
			ras.pendingMu.Unlock()
			return nil, context.Cause(annStrCtx)
		}

		if len(ras.pendings) > 0 {
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
		case <-annStrCtx.Done():
			reason := context.Cause(annStrCtx)
			return nil, reason
		case <-announceCh:
			// New announcement available, loop to check pendings
			continue
		}
	}
}

func (ras *AnnouncementReader) Close() error {
	ras.pendingMu.Lock()

	if ras.closed {
		ras.pendingMu.Unlock()
		return nil
	}
	ras.closed = true

	close(ras.announcedCh)
	ras.announcedCh = nil

	ras.pendingMu.Unlock()

	ras.cancel(nil)

	return ras.stream.Close()
}

func (ras *AnnouncementReader) CloseWithError(code AnnounceErrorCode) error {
	ras.pendingMu.Lock()

	if ras.closed {
		ras.pendingMu.Unlock()
		return nil
	}
	ras.closed = true

	close(ras.announcedCh)
	ras.announcedCh = nil

	ras.pendingMu.Unlock()

	strErrCode := quic.StreamErrorCode(code)
	ras.stream.CancelRead(strErrCode)
	ras.stream.CancelWrite(strErrCode)

	err := &AnnounceError{
		StreamError: &quic.StreamError{
			StreamID:  ras.stream.StreamID(),
			ErrorCode: strErrCode,
		},
	}

	ras.cancel(err)

	return nil
}

func (ras *AnnouncementReader) listenAnnouncements() {
	ras.listenOnce.Do(func() {
		var am message.AnnounceMessage
		// var suffix string
		var err error
		for {
			// Check if announcement is already closed before decoding
			ras.pendingMu.Lock()
			if ras.closed {
				ras.pendingMu.Unlock()
				return
			}
			ras.pendingMu.Unlock()

			err = am.Decode(ras.stream)
			if err != nil {
				var strErr *quic.StreamError
				if errors.As(err, &strErr) {
					annErr := &AnnounceError{
						StreamError: strErr,
					}
					ras.cancel(annErr)
				} else {
					ras.cancel(err)
				}
				return
			}

			slog.Debug("received announce message", "message", am)

			// Check if announcement is already closed during decoding
			ras.pendingMu.Lock()
			if ras.closed {
				ras.pendingMu.Unlock()
				return
			}

			old, ok := ras.active[am.TrackSuffix]

			switch am.AnnounceStatus {
			case message.ACTIVE:
				if !ok || (ok && !old.IsActive()) {
					// Create a new announcement
					ann := NewAnnouncement(ras.ctx, BroadcastPath(ras.prefix+am.TrackSuffix))
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
