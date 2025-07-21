package moqt

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

// type AnnouncementReader interface {
// 	ReceiveAnnouncement(context.Context) (*Announcement, error)
// 	Close() error
// 	CloseWithError(AnnounceErrorCode) error
// }

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

	closeErr error
	closed   bool

	// Track Suffix -> Announcement
	active map[string]*Announcement

	pendings    []*Announcement
	announcedCh chan struct{} // notify when new announcement is available
	mu          sync.Mutex

	listenOnce sync.Once
}

func (ras *AnnouncementReader) ReceiveAnnouncement(ctx context.Context) (*Announcement, error) {
	for {
		ras.mu.Lock()

		if ras.closed {
			ras.mu.Unlock()
			if ras.closeErr != nil {
				return nil, ras.closeErr
			}
			return nil, fmt.Errorf("receive announce stream is closed")
		}

		if len(ras.pendings) > 0 {
			next := ras.pendings[0]
			ras.pendings = ras.pendings[1:]
			ras.mu.Unlock()
			return next, nil
		}

		ras.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ras.ctx.Done():
			reason := context.Cause(ras.ctx)
			return nil, reason
		case <-ras.announcedCh:
			// New announcement available, loop to check pendings
			continue
		}
	}
}

func (ras *AnnouncementReader) Close() error {
	ras.mu.Lock()

	if ras.closed {
		ras.mu.Unlock()
		return ras.closeErr
	}

	ras.closed = true

	// Close the notification channel safely
	select {
	case <-ras.announcedCh:
	default:
		close(ras.announcedCh)
	}

	ras.mu.Unlock()

	ras.cancel(nil)

	return ras.stream.Close()
}

func (ras *AnnouncementReader) CloseWithError(code AnnounceErrorCode) error {
	ras.mu.Lock()

	if ras.closed {
		ras.mu.Unlock()
		return ras.closeErr
	}

	ras.closed = true

	err := &AnnounceError{
		StreamError: &quic.StreamError{
			StreamID:  ras.stream.StreamID(),
			ErrorCode: quic.StreamErrorCode(code),
		},
	}
	ras.closeErr = err

	select {
	case <-ras.announcedCh:
	default:
		close(ras.announcedCh)
	}

	strErrCode := quic.StreamErrorCode(code)
	ras.stream.CancelRead(strErrCode)
	ras.stream.CancelWrite(strErrCode)

	ras.mu.Unlock()
	ras.cancel(err)

	return nil
}

func (ras *AnnouncementReader) listenAnnouncements() {
	ras.listenOnce.Do(func() {
		var am message.AnnounceMessage
		// var suffix string
		var err error
		for {
			// Check if closed under lock
			ras.mu.Lock()
			if ras.closed {
				ras.mu.Unlock()
				return
			}
			ras.mu.Unlock()

			err = am.Decode(ras.stream)
			if err != nil {
				ras.mu.Lock()
				if ras.closed {
					ras.mu.Unlock()
					return
				}

				ras.closed = true

				if err == io.EOF {
					ras.closeErr = io.EOF
				} else {
					var strErr *quic.StreamError
					if errors.As(err, &strErr) {
						ras.closeErr = &AnnounceError{
							StreamError: strErr,
						}
					} else {
						ras.closeErr = err
					}
				}

				closeErr := ras.closeErr

				// Close the notification channel safely
				select {
				case <-ras.announcedCh:
				default:
					close(ras.announcedCh)
				}

				ras.mu.Unlock()

				ras.cancel(closeErr)
				return
			}

			slog.Debug("received announce message", "message", am)

			// suffix = am.TrackSuffix

			ras.mu.Lock()
			old, ok := ras.active[am.TrackSuffix]

			switch am.AnnounceStatus {
			case message.ACTIVE:
				if !ok || (ok && !old.IsActive()) {
					// Create a new announcement
					ann := NewAnnouncement(ras.ctx, BroadcastPath(ras.prefix+am.TrackSuffix))
					ras.active[am.TrackSuffix] = ann
					ras.pendings = append(ras.pendings, ann)
					// Notify that new announcement is available
					if !ras.closed {
						select {
						case ras.announcedCh <- struct{}{}:
						default:
						}
					}
					ras.mu.Unlock()
				} else {
					// Release lock before calling CloseWithError to avoid deadlock
					ras.mu.Unlock()
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
					ras.mu.Unlock()
				} else {
					// Release lock before calling CloseWithError to avoid deadlock
					ras.mu.Unlock()
					ras.CloseWithError(DuplicatedAnnounceErrorCode)
					return
				}
			default:
				ras.mu.Unlock()
			}
		}
	})
}
