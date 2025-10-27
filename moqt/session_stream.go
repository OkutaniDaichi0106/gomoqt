package moqt

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/message"
	"github.com/OkutaniDaichi0106/gomoqt/quic"
)

func newSessionStream(stream quic.Stream, req *SetupRequest) *sessionStream {
	ss := &sessionStream{
		ctx:          context.WithValue(stream.Context(), &biStreamTypeCtxKey, message.StreamTypeSession),
		stream:       stream,
		Version:      DefaultServerVersion, // Default version before setup
		SetupRequest: req,
		updatedCh:    make(chan struct{}, 1),
	}
	return ss
}

type sessionStream struct {
	ctx       context.Context
	updatedCh chan struct{}

	localBitrate  uint64 // The bitrate set by the local
	remoteBitrate uint64 // The bitrate set by the remote

	stream quic.Stream

	mu sync.Mutex

	// Version of the protocol used in this session
	Version Version

	// Setup request from the client
	*SetupRequest

	// Parameters specified by the server
	ServerExtensions *Parameters

	listenOnce sync.Once
}

func newResponse(sessStr *sessionStream) *response {
	return &response{
		sessionStream: sessStr,
	}
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
		r.ServerExtensions = &Parameters{sum.Parameters}

		r.handleUpdates()
	})

	return err
}

var _ SetupResponseWriter = (*responseWriter)(nil)

func newResponseWriter(conn quic.Connection, sessStr *sessionStream, connLogger *slog.Logger, server *Server) *responseWriter {
	return &responseWriter{
		sessionStream: sessStr,
		conn:          conn,
		connLogger:    connLogger,
		server:        server,
	}
}

type responseWriter struct {
	*sessionStream
	conn       quic.Connection
	connLogger *slog.Logger
	server     *Server
	onceSetup  sync.Once
}

func (w *responseWriter) SelectVersion(v Version) error {
	if !slices.Contains(w.Versions, v) {
		return fmt.Errorf("version %d not supported by client", v)
	}
	w.Version = v
	return nil
}

func (w *responseWriter) SetExtensions(extensions *Parameters) {
	w.ServerExtensions = extensions
}

func (w *responseWriter) accept(mux *TrackMux) (*Session, error) {
	var err error
	w.onceSetup.Do(func() {
		// TODO: Implement setup logic if needed
		var paramMsg message.Parameters
		if w.ServerExtensions != nil {
			paramMsg = w.ServerExtensions.paramMap
		}
		err = message.SessionServerMessage{
			SelectedVersion: w.Version,
			Parameters:      paramMsg,
		}.Encode(w.stream)
		if err != nil {
			return
		}

		// Start listening for updates
		w.handleUpdates()
	})

	if err != nil {
		return nil, err
	}

	var sess *Session
	sess = newSession(w.conn, w.sessionStream, mux, w.connLogger, func() { w.server.removeSession(sess) })
	w.server.addSession(sess)

	return sess, nil
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

// handleUpdates triggers the goroutine to start listening for session updates
func (ss *sessionStream) handleUpdates() {
	// Safe to call multiple times
	ss.listenOnce.Do(func() {
		go func() {
			var sum message.SessionUpdateMessage
			var err error

			for {
				err = sum.Decode(ss.stream)
				if err != nil {
					break
				}

				ss.mu.Lock()
				ss.remoteBitrate = sum.Bitrate
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
	})
}

func (ss *sessionStream) SessionUpdated() <-chan struct{} {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	return ss.updatedCh
}

func (ss *sessionStream) Context() context.Context {
	return ss.ctx
}

func Accept(w SetupResponseWriter, r *SetupRequest, mux *TrackMux) (*Session, error) {
	if rsp, ok := w.(*responseWriter); ok {
		return rsp.accept(mux)
	} else {
		return nil, fmt.Errorf("moq: invalid response writer type %T", w)
	}
}
