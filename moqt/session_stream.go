package moqt

import (
	"context"
	"errors"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newSessionStream(connCtx context.Context, stream quic.Stream) *sessionStream {
	ctx, cancel := context.WithCancelCause(connCtx)

	sessStr := &sessionStream{
		ctx:       ctx,
		cancel:    cancel,
		updatedCh: make(chan struct{}, 1),
		stream:    stream,
	}

	go func() {
		var sum message.SessionUpdateMessage
		var err error

		for {
			err = sum.Decode(sessStr.stream)
			if err != nil {
				return
			}

			// Update the session bitrate
			sessStr.mu.Lock()
			sessStr.remoteBitrate = sum.Bitrate
			sessStr.mu.Unlock()

			// Notify that the session has been updated
			select {
			case sessStr.updatedCh <- struct{}{}:
			default:
			}
		}

		sessStr.close()
	}()

	return sessStr
}

type sessionStream struct {
	ctx    context.Context
	cancel context.CancelCauseFunc

	updatedCh chan struct{}
	closed    bool // Track if the channel is closed

	localBitrate  uint64 // The bitrate set by the local
	remoteBitrate uint64 // The bitrate set by the remote

	stream quic.Stream
	mu     sync.Mutex
}

func (ss *sessionStream) updateSession(bitrate uint64) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	sum := message.SessionUpdateMessage{
		Bitrate: bitrate,
	}
	err := sum.Encode(ss.stream)
	if err != nil {
		var strErr *quic.StreamError
		if errors.As(err, &strErr) && strErr.Remote {
		}

		return err
	}

	ss.localBitrate = bitrate

	return nil
}

func (ss *sessionStream) SessionUpdated() <-chan struct{} {
	return ss.updatedCh
}

func (ss *sessionStream) close() error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	// If already closed, return early without error
	if ss.closed {
		return nil
	}
	ss.closed = true

	err := ss.stream.Close()

	ss.cancel(nil)

	close(ss.updatedCh)

	return err
}
