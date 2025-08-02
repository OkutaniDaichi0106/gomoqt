package moqt

import (
	"context"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newSessionStream(stream quic.Stream, version protocol.Version, path string, clientParams, serverParams *Parameters) *sessionStream {
	sessStr := &sessionStream{
		ctx:              context.WithValue(stream.Context(), &biStreamTypeCtxKey, message.StreamTypeSession),
		updatedCh:        make(chan struct{}, 1),
		stream:           stream,
		version:          version,
		path:             path,
		clientParameters: clientParams,
		serverParameters: serverParams,
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

			// Notify that the session has been updated
			select {
			case sessStr.updatedCh <- struct{}{}:
			default:
			}

			sessStr.mu.Unlock()

		}

		sessStr.mu.Lock()

		if sessStr.updatedCh != nil {
			close(sessStr.updatedCh)
			sessStr.updatedCh = nil
		}

		sessStr.mu.Unlock()
	}()

	return sessStr
}

type sessionStream struct {
	ctx       context.Context
	updatedCh chan struct{}

	localBitrate  uint64 // The bitrate set by the local
	remoteBitrate uint64 // The bitrate set by the remote

	stream quic.Stream

	mu sync.Mutex

	path string

	// Version of the protocol used in this session
	version protocol.Version

	// Parameters specified by the client and server
	clientParameters *Parameters

	// Parameters specified by the server
	serverParameters *Parameters
}

func (ss *sessionStream) updateSession(bitrate uint64) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	err := message.SessionUpdateMessage{
		Bitrate: bitrate,
	}.Encode(ss.stream)
	if err != nil {
		return Cause(ss.ctx)
	}

	ss.localBitrate = bitrate

	return nil
}

func (ss *sessionStream) SessionUpdated() <-chan struct{} {
	return ss.updatedCh
}

func (ss *sessionStream) Context() context.Context {
	return ss.ctx
}
