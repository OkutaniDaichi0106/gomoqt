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

type AnnouncementReader interface {
	ReceiveAnnouncement(context.Context) (*Announcement, error)
	Close() error
	CloseWithError(AnnounceErrorCode) error
}

func newReceiveAnnounceStream(sessCtx context.Context, stream quic.Stream, prefix string) *receiveAnnounceStream {
	ctx, cancel := context.WithCancelCause(sessCtx)
	annstr := &receiveAnnounceStream{
		ctx:      ctx,
		cancel:   cancel,
		stream:   stream,
		prefix:   prefix,
		active:   make(map[string]*Announcement),
		pendings: make([]*Announcement, 0),
		notifyCh: make(chan struct{}, 1),
	}

	// Receive announcements in a separate goroutine
	go func() {
		var am message.AnnounceMessage
		var suffix string
		var err error

		for {
			if annstr.closed {
				return
			}

			_, err = am.Decode(stream)
			if err != nil {
				annstr.mu.Lock()
				defer annstr.mu.Unlock()

				if annstr.closed {
					return
				}

				annstr.closed = true

				if err == io.EOF {
					annstr.closeErr = nil
					annstr.cancel(nil)
					return
				}

				var strErr *quic.StreamError
				if errors.As(err, &strErr) {
					annErr := &AnnounceError{
						StreamError: strErr,
					}
					annstr.closeErr = annErr
					annstr.cancel(annErr)
					return
				}

				annstr.closeErr = err
				annstr.cancel(err)
				return
			}

			suffix = am.TrackSuffix

			old, ok := annstr.active[suffix]
			switch am.AnnounceStatus {
			case message.ACTIVE:
				if !ok || (ok && !old.IsActive()) {
					// Create a new announcement
					ann := NewAnnouncement(annstr.ctx, BroadcastPath(prefix+suffix))
					annstr.mu.Lock()
					annstr.active[suffix] = ann
					annstr.pendings = append(annstr.pendings, ann)
					// Notify that new announcement is available
					if !annstr.closed {
						select {
						case annstr.notifyCh <- struct{}{}:
						default:
						}
					}
					annstr.mu.Unlock()
				} else {
					// Close the stream with an error
					annstr.CloseWithError(DuplicatedAnnounceErrorCode)
					return
				}
			case message.ENDED:
				if ok && old.IsActive() {
					// End the existing announcement
					old.End()

					// Remove the announcement from the map
					delete(annstr.active, suffix)
				} else {
					annstr.CloseWithError(DuplicatedAnnounceErrorCode)
					return
				}
			}
		}
	}()

	return annstr
}

var _ AnnouncementReader = (*receiveAnnounceStream)(nil)

type receiveAnnounceStream struct {
	stream quic.Stream
	prefix string

	ctx    context.Context
	cancel context.CancelCauseFunc

	closeErr error
	closed   bool

	// Track Suffix -> Announcement
	active map[string]*Announcement

	pendings []*Announcement
	notifyCh chan struct{} // notify when new announcement is available
	mu       sync.Mutex
}

func (ras *receiveAnnounceStream) ReceiveAnnouncement(ctx context.Context) (*Announcement, error) {
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
			slog.Error("receive announce stream context done",
				"reason", reason,
			)
			return nil, reason
		case <-ras.notifyCh:
			// New announcement available, loop to check pendings
			continue
		}
	}
}

func (ras *receiveAnnounceStream) Close() error {
	ras.mu.Lock()
	defer ras.mu.Unlock()

	if ras.closed {
		return ras.closeErr
	}

	ras.closed = true
	ras.cancel(nil)

	// Close the notification channel safely
	func() {
		defer func() {
			recover() // Ignore panic if channel is already closed
		}()
		close(ras.notifyCh)
	}()

	err := ras.stream.Close()
	if err != nil {
		return err
	}

	return nil
}

func (ras *receiveAnnounceStream) CloseWithError(code AnnounceErrorCode) error {
	ras.mu.Lock()
	defer ras.mu.Unlock()

	if ras.closed {
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
	ras.cancel(err)

	// Close the notification channel safely
	func() {
		defer func() {
			recover() // Ignore panic if channel is already closed
		}()
		close(ras.notifyCh)
	}()

	strErrCode := quic.StreamErrorCode(code)

	ras.stream.CancelRead(strErrCode)
	ras.stream.CancelWrite(strErrCode)

	return nil
}
