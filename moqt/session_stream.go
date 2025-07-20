package moqt

import (
	"errors"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newSessionStream(stream quic.Stream) *sessionStream {
	sessStr := &sessionStream{
		updatedCh: make(chan struct{}, 1),
		stream:    stream,
	}

	go func() {
		var sum message.SessionUpdateMessage
		var err error

		for {
			err = sum.Decode(stream)
			if err != nil {
				break
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

		close(sessStr.updatedCh)
	}()

	return sessStr
}

type sessionStream struct {
	updatedCh chan struct{}

	localBitrate  uint64 // The bitrate set by the local
	remoteBitrate uint64 // The bitrate set by the remote

	stream quic.Stream

	mu sync.Mutex
}

func (ss *sessionStream) updateSession(bitrate uint64) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	err := message.SessionUpdateMessage{
		Bitrate: bitrate,
	}.Encode(ss.stream)
	if err != nil {
		var appErr *quic.ApplicationError
		if errors.As(err, &appErr) {
			return &SessionError{
				ApplicationError: appErr,
			}
		}

		return err
	}

	ss.localBitrate = bitrate

	return nil
}

func (ss *sessionStream) SessionUpdated() <-chan struct{} {
	return ss.updatedCh
}
