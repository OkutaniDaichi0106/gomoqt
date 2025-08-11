package moqt

import (
	"context"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/quic"
)

func newSessionStream(stream quic.Stream, req *Request) *sessionStream {
	ss := &sessionStream{
		ctx:       context.WithValue(stream.Context(), &biStreamTypeCtxKey, message.StreamTypeSession),
		stream:    stream,
		Request:   req,
		setupDone: make(chan struct{}),
		updatedCh: make(chan struct{}, 1),
	}

	go func() {
		<-ss.setupDone

		var sum message.SessionUpdateMessage
		var err error

		for {
			err = sum.Decode(ss.stream)
			if err != nil {
				break
			}

			// Update the session bitrate
			ss.mu.Lock()
			ss.remoteBitrate = sum.Bitrate

			// Notify that the session has been updated
			select {
			case ss.updatedCh <- struct{}{}:
			default:
			}

			ss.mu.Unlock()
		}

		ss.mu.Lock()

		if ss.updatedCh != nil {
			close(ss.updatedCh)
		}

		ss.mu.Unlock()
	}()

	return ss
}

var _ ResponseWriter = (*responseWriter)(nil)

type sessionStream struct {
	ctx       context.Context
	updatedCh chan struct{}

	localBitrate  uint64 // The bitrate set by the local
	remoteBitrate uint64 // The bitrate set by the remote

	stream quic.Stream

	mu sync.Mutex

	// Version of the protocol used in this session
	Version protocol.Version

	// Parameters specified by the client and server

	*Request

	// Parameters specified by the server
	serverParameters *Parameters

	setupDone chan struct{}
}

type response struct {
	*sessionStream
	onceSetup sync.Once
}

func (r *response) AwaitAccepted() error {
	var err error
	r.onceSetup.Do(func() {
		var sum message.SessionServerMessage
		err = sum.Decode(r.stream)
		if err != nil {
			return
		}
		r.Version = sum.SelectedVersion
		r.serverParameters = &Parameters{sum.Parameters}

		r.listenUpdates()
	})

	return err
}

type responseWriter struct {
	*sessionStream
	conn      quic.Connection
	onceSetup sync.Once
}

func (w *responseWriter) Accept(v Version, extensions *Parameters) error {
	var err error
	w.onceSetup.Do(func() {
		// TODO: Implement setup logic if needed
		var paramMsg message.Parameters
		if extensions != nil {
			paramMsg = extensions.paramMap
		}
		err = message.SessionServerMessage{
			SelectedVersion: v,
			Parameters:      paramMsg,
		}.Encode(w.stream)
		if err != nil {
			return
		}

		w.Version = v
		w.serverParameters = extensions

		// Start listening for updates
		w.listenUpdates()
	})
	return err
}

func (w *responseWriter) Reject(code SessionErrorCode) error {
	w.stream.CancelWrite(quic.StreamErrorCode(code))
	w.stream.CancelRead(quic.StreamErrorCode(code))
	return nil
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

// listenUpdates triggers the goroutine to start listening for session updates
func (ss *sessionStream) listenUpdates() {
	// Safe to call multiple times
	select {
	case ss.setupDone <- struct{}{}:
		// Successfully triggered
	default:
		// Already triggered or closed
	}
}

func (ss *sessionStream) SessionUpdated() <-chan struct{} {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	return ss.updatedCh
}

func (ss *sessionStream) Context() context.Context {
	return ss.ctx
}
