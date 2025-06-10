package moqt

import (
	"context"
	"errors"
	"fmt"
	"io"
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
		ctx:           ctx,
		cancel:        cancel,
		stream:        stream,
		prefix:        prefix,
		announcements: make(map[string]*Announcement),
		pendings:      make([]*Announcement, 0),
	}

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
				if err == io.EOF {
					annstr.cancel(nil)
					return
				}

				var strErr *quic.StreamError
				if errors.As(err, &strErr) {
					err = &AnnounceError{
						StreamError: strErr,
					}
					annstr.cancel(err)
					return
				}

				annstr.cancel(err)
				return
			}

			suffix = am.TrackSuffix

			old, ok := annstr.announcements[suffix]
			switch am.AnnounceStatus {
			case message.ACTIVE:
				if !ok || (ok && !old.IsActive()) {
					// Create a new announcement
					ann := NewAnnouncement(annstr.ctx, BroadcastPath(prefix+suffix))
					annstr.mu.Lock()
					annstr.announcements[suffix] = ann
					annstr.pendings = append(annstr.pendings, ann)
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
					delete(annstr.announcements, suffix)
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
	announcements map[string]*Announcement

	pendings []*Announcement
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
			return nil, fmt.Errorf("receive announce stream is closed: %v", ras.closeErr)
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
			return nil, ras.ctx.Err()
		}
	}
}

func (ras *receiveAnnounceStream) Close() error {
	ras.mu.Lock()
	defer ras.mu.Unlock()

	if ras.closed {
		return ras.closeErr
	}

	ras.cancel(nil)

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

	err := &AnnounceError{
		StreamError: &quic.StreamError{
			StreamID:  ras.stream.StreamID(),
			ErrorCode: quic.StreamErrorCode(code),
		},
	}

	ras.cancel(err)

	strErrCode := quic.StreamErrorCode(code)

	ras.stream.CancelRead(strErrCode)
	ras.stream.CancelWrite(strErrCode)

	return nil
}
