package moqt

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newSession(conn quic.Connection, version protocol.Version, path string, clientParams, serverParams *Parameters,
	stream quic.Stream, mux *TrackMux, logger *slog.Logger) *Session {
	if mux == nil {
		mux = DefaultMux
	}

	sessStream := newSessionStream(conn.Context(), stream)

	sess := &Session{
		sessionStream:    sessStream,
		path:             path,
		version:          version,
		clientParameters: clientParams,
		serverParameters: serverParams,
		logger:           logger,
		conn:             conn,
		mux:              mux,
		trackReceivers:   make(map[SubscribeID]*trackReceiver),
		trackSenders:     make(map[SubscribeID]*trackSender),
	}

	sess.wg.Add(3)
	// Listen for session stream closure
	go func() {
		defer sess.wg.Done()
		<-sess.Context().Done()
		if !sess.terminating() {
			sess.Terminate(ProtocolViolationErrorCode, "session stream closed unexpectedly")
		}
	}()

	// Listen bidirectional streams
	go func() {
		defer sess.wg.Done()
		sess.handleBiStreams()
	}()

	// Listen unidirectional streams
	go func() {
		defer sess.wg.Done()
		sess.handleUniStreams()
	}()

	return sess
}

type Session struct {
	*sessionStream

	wg sync.WaitGroup // WaitGroup for session cleanup

	path string

	// Version of the protocol used in this session
	version protocol.Version

	// Parameters specified by the client and server
	clientParameters *Parameters

	// Parameters specified by the server
	serverParameters *Parameters

	// bitrate atomic.Uint64 // Bitrate in bits per second

	logger *slog.Logger

	conn   quic.Connection
	connMu sync.Mutex

	mux *TrackMux // TODO

	subscribeIDCounter atomic.Uint64

	trackReceivers        map[SubscribeID]*trackReceiver
	receiveGroupMapLocker sync.RWMutex

	trackSenders       map[SubscribeID]*trackSender
	sendGroupMapLocker sync.RWMutex

	isTerminating atomic.Bool
	termErr       error
}

func (s *Session) terminating() bool {
	return s.isTerminating.Load()
}

func (s *Session) Context() context.Context {
	return s.ctx
}

func (s *Session) Terminate(code SessionErrorCode, msg string) error {
	if s.terminating() {
		return s.termErr
	}

	s.isTerminating.Store(true)

	if s.logger != nil {
		s.logger.Debug("terminating a session",
			"code", code,
			"message", msg,
		)
	}

	s.connMu.Lock()
	defer s.connMu.Unlock()
	err := s.conn.CloseWithError(quic.ConnectionErrorCode(code), msg)
	if err != nil {
		var appErr *quic.ApplicationError
		if errors.As(err, &appErr) {
			reason := &SessionError{
				ApplicationError: appErr,
			}
			s.termErr = reason
			return reason
		}
		s.termErr = err
		return err
	}

	s.termErr = &SessionError{
		ApplicationError: &quic.ApplicationError{
			ErrorCode:    quic.ApplicationErrorCode(code),
			ErrorMessage: msg,
		},
	}

	s.cancel(s.termErr)

	// Wait for finishing handling streams
	s.wg.Wait()

	if s.logger != nil {
		s.logger.Debug("terminated a session")
	}

	return nil
}

func (s *Session) OpenTrackStream(path BroadcastPath, name TrackName, config *SubscribeConfig) (*Subscriber, error) {
	if config == nil {
		config = &SubscribeConfig{}
	}
	id := s.nextSubscribeID()

	if s.logger != nil {
		s.logger.Debug("opening track stream",
			"subscribe_config", config.String(),
			"broadcast_path", path,
			"track_name", name,
			"subscribe_id", id)
	}

	s.connMu.Lock()
	stream, err := s.conn.OpenStream()
	if err != nil {
		var appErr *quic.ApplicationError
		if errors.As(err, &appErr) {
			s.connMu.Unlock()
			return nil, &SessionError{
				ApplicationError: appErr,
			}
		}
		s.connMu.Unlock()
		return nil, err
	}
	s.connMu.Unlock()

	// Send a SUBSCRIBE message
	sm := message.SubscribeMessage{
		SubscribeID:      message.SubscribeID(id),
		BroadcastPath:    string(path),
		TrackName:        string(name),
		TrackPriority:    message.TrackPriority(config.TrackPriority),
		MinGroupSequence: message.GroupSequence(config.MinGroupSequence),
		MaxGroupSequence: message.GroupSequence(config.MaxGroupSequence),
	}
	_, err = sm.Encode(stream)
	if err != nil {
		var strErr *quic.StreamError
		if errors.As(err, &strErr) && strErr.Remote {
			//
			stream.CancelRead(strErr.ErrorCode)

			return nil, &SubscribeError{
				StreamError: strErr,
			}
		}

		strErrCode := quic.StreamErrorCode(InternalSubscribeErrorCode)
		stream.CancelWrite(strErrCode)
		stream.CancelRead(strErrCode)

		return nil, err
	}

	var subok message.SubscribeOkMessage
	_, err = subok.Decode(stream)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, err
		}

		var strErr *quic.StreamError
		if errors.As(err, &strErr) {
			strErrCode := quic.StreamErrorCode(strErr.ErrorCode)
			stream.CancelWrite(strErrCode)

			return nil, &SubscribeError{
				StreamError: strErr,
			}
		}

		strErrCode := quic.StreamErrorCode(InternalSubscribeErrorCode)
		stream.CancelWrite(strErrCode)
		stream.CancelRead(strErrCode)

		return nil, err
	}

	substr := newSendSubscribeStream(s.ctx, id, stream, config)

	// Create a receive group stream queue
	trackReceiver := newTrackReceiver(substr)
	//
	s.receiveGroupMapLocker.Lock()
	s.trackReceivers[id] = trackReceiver
	s.receiveGroupMapLocker.Unlock()

	return &Subscriber{
		BroadcastPath:   path,
		TrackName:       name,
		SubscribeStream: substr,
		TrackReader:     trackReceiver,
	}, nil
}

func (s *Session) nextSubscribeID() SubscribeID {
	// Increment and return the previous value atomically
	return SubscribeID(s.subscribeIDCounter.Add(1))
}

// // TODO: Implement this method and use it
// func (sess *Session) updateSession(bitrate uint64) error {
// 	if sess.logger != nil {
// 		sess.logger.Debug("updating a session",
// 			"bitrate", bitrate,
// 		)
// 	}

// 	// Send a SESSION_UPDATE message
// 	err := sess.sessionStream.updateSession(bitrate)
// 	if err != nil {
// 		sess.logger.Error("failed to update a session",
// 			"error", err,
// 		)
// 		return err
// 	}

// 	// Update the bitrate
// 	// sess.ctx.bitrate.Store(bitrate)
// 	if sess.logger != nil {
// 		sess.logger.Debug("session updated successfully",
// 			"bitrate", bitrate,
// 		)
// 	}

// 	return nil
// }

func (sess *Session) OpenAnnounceStream(prefix string) (AnnouncementReader, error) {
	sess.connMu.Lock()
	stream, err := sess.conn.OpenStream()
	if err != nil {
		var appErr *quic.ApplicationError
		if errors.As(err, &appErr) {
			sess.connMu.Unlock()
			return nil, &SessionError{
				ApplicationError: appErr,
			}
		}
		sess.connMu.Unlock()
		return nil, err
	}
	sess.connMu.Unlock()

	st := message.StreamTypeMessage{
		StreamType: stream_type_announce,
	}
	_, err = st.Encode(stream)
	if err != nil {
		var strErr *quic.StreamError
		if errors.As(err, &strErr) {
			strErrCode := quic.StreamErrorCode(InternalAnnounceErrorCode)
			stream.CancelRead(strErrCode)

			return nil, &AnnounceError{
				StreamError: strErr,
			}
		}

		return nil, err
	}

	apm := message.AnnouncePleaseMessage{
		TrackPrefix: prefix,
	}
	_, err = apm.Encode(stream)
	if err != nil {
		var strErr *quic.StreamError
		if errors.As(err, &strErr) {
			strErrCode := quic.StreamErrorCode(InternalAnnounceErrorCode)
			stream.CancelRead(strErrCode)

			return nil, &AnnounceError{
				StreamError: strErr,
			}
		}

		strErrCode := quic.StreamErrorCode(InternalAnnounceErrorCode)
		stream.CancelWrite(strErrCode)
		stream.CancelRead(strErrCode)

		return nil, err
	}

	return newReceiveAnnounceStream(sess.ctx, stream, prefix), nil
}

func (sess *Session) goAway(uri string) {
	// TODO
}

// listenBiStreams accepts bidirectional streams and handles them based on their type.
// It listens for incoming streams and processes them in separate goroutines.
// The function handles session streams, announce streams, subscribe streams, and info streams.
// It also handles errors and terminates the session if an unknown stream type is encountered.
func (sess *Session) handleBiStreams() {
	for { // Accept a bidirectional stream
		stream, err := sess.conn.AcceptStream(sess.ctx)
		if err != nil {
			return
		}

		// Handle the stream
		go sess.processBiStream(stream)
	}
}

func (sess *Session) processBiStream(stream quic.Stream) {
	var stm message.StreamTypeMessage
	n, err := stm.Decode(stream)
	if err != nil {
		if errors.Is(err, io.EOF) || n == 0 {

			sess.Terminate(ProtocolViolationErrorCode, err.Error())
			return
		}

		var strErr *quic.StreamError
		if errors.As(err, &strErr) && strErr.Remote {
		}

		sess.Terminate(ProtocolViolationErrorCode, err.Error())
		return
	}

	switch stm.StreamType {
	case stream_type_announce:
		var apm message.AnnouncePleaseMessage
		_, err := apm.Decode(stream)
		if err != nil {
			if errors.Is(err, io.EOF) {
				sess.Terminate(ProtocolViolationErrorCode, err.Error())
				return
			}

			var strErr *quic.StreamError
			if errors.As(err, &strErr) && strErr.Remote {
			}

			sess.Terminate(ProtocolViolationErrorCode, err.Error())
			return
		}

		prefix := apm.TrackPrefix

		annstr := newSendAnnounceStream(stream, prefix)

		sess.mux.ServeAnnouncements(annstr, prefix)
	case stream_type_subscribe:
		var sm message.SubscribeMessage
		_, err := sm.Decode(stream)
		if err != nil {
			if errors.Is(err, io.EOF) {
				sess.Terminate(ProtocolViolationErrorCode, err.Error())
				return
			}

			var strErr *quic.StreamError
			if errors.As(err, &strErr) && strErr.Remote {
			}

			sess.Terminate(ProtocolViolationErrorCode, err.Error())

			return
		}

		// Create a receiveSubscribeStream
		id := SubscribeID(sm.SubscribeID)
		path := BroadcastPath(sm.BroadcastPath)
		name := TrackName(sm.TrackName)
		config := &SubscribeConfig{
			TrackPriority:    TrackPriority(sm.TrackPriority),
			MinGroupSequence: GroupSequence(sm.MinGroupSequence),
			MaxGroupSequence: GroupSequence(sm.MaxGroupSequence),
		}

		node := sess.mux.findRoutingNode(path)

		if node == nil || node.handler == nil || node.announcement == nil || !node.announcement.IsActive() || node.info == nil {
			strErrCode := quic.StreamErrorCode(TrackNotFoundErrorCode)

			stream.CancelWrite(strErrCode)

			stream.CancelRead(strErrCode)

			return
		}

		som := message.SubscribeOkMessage{
			GroupOrder: message.GroupOrder(node.info.GroupOrder),
		}
		_, err = som.Encode(stream)
		if err != nil {
			var strErr *quic.StreamError
			if errors.As(err, &strErr) && strErr.Remote {
				stream.CancelRead(strErr.ErrorCode)

				return
			}

			code := quic.StreamErrorCode(InternalSubscribeErrorCode)
			stream.CancelWrite(code)
			stream.CancelRead(code)

			return
		}

		openStreamFunc := func(seq GroupSequence) (*sendGroupStream, error) {
			stream, err := sess.conn.OpenStream()
			if err != nil {
				var appErr *quic.ApplicationError
				if errors.As(err, &appErr) {
					return nil, &SessionError{
						ApplicationError: appErr,
					}
				}

				return nil, err
			}

			stm := message.StreamTypeMessage{
				StreamType: stream_type_group,
			}
			_, err = stm.Encode(stream)
			if err != nil {
				var strErr *quic.StreamError
				if errors.As(err, &strErr) {
					return nil, &GroupError{StreamError: strErr}
				}

				strErrCode := quic.StreamErrorCode(InternalGroupErrorCode)
				stream.CancelWrite(strErrCode)

				return nil, GroupError{
					StreamError: &quic.StreamError{
						StreamID:  stream.StreamID(),
						ErrorCode: strErrCode,
					},
				}
			}

			gm := message.GroupMessage{
				SubscribeID:   sm.SubscribeID,
				GroupSequence: message.GroupSequence(seq),
			}
			_, err = gm.Encode(stream)
			if err != nil {
				var strErr *quic.StreamError
				if errors.As(err, &strErr) {
					return nil, &GroupError{StreamError: strErr}
				}

				strErr = &quic.StreamError{
					StreamID:  stream.StreamID(),
					ErrorCode: quic.StreamErrorCode(InternalGroupErrorCode),
				}

				return nil, GroupError{StreamError: strErr}
			}

			return newSendGroupStream(stream, seq), nil
		}

		substr := newReceiveSubscribeStream(id, stream, config)

		trackSender := newTrackSender(substr, openStreamFunc)
		sess.sendGroupMapLocker.Lock()
		sess.trackSenders[id] = trackSender
		sess.sendGroupMapLocker.Unlock()

		node.handler.ServeTrack(&Publisher{
			BroadcastPath:   path,
			TrackName:       name,
			SubscribeStream: substr,
			TrackWriter:     trackSender,
		})
	default:
		sess.Terminate(ProtocolViolationErrorCode, fmt.Sprintf("unknown bidirectional stream type: %v", stm.StreamType))

		return
	}
}

func (sess *Session) handleUniStreams() {
	for {
		/*
		 * Accept a unidirectional stream
		 */
		stream, err := sess.conn.AcceptUniStream(sess.ctx)
		if err != nil {
			return
		}

		// Handle the stream
		go sess.processUniStream(stream)
	}
}

func (sess *Session) processUniStream(stream quic.ReceiveStream) {
	logger := sess.logger.With(
		"stream_id", stream.StreamID(),
	)
	/*
	 * Get a Stream Type ID
	 */
	var stm message.StreamTypeMessage
	_, err := stm.Decode(stream)
	if err != nil {
		if logger != nil {
			logger.Error("failed to get a Stream Type ID", "error", err)
		}
		return
	}

	// Handle the stream by the Stream Type ID
	switch stm.StreamType {
	case stream_type_group:
		if logger != nil {
			logger.Debug("group stream was opened")
		}

		var gm message.GroupMessage
		_, err := gm.Decode(stream)
		if err != nil {
			if logger != nil {
				logger.Error("failed to get a group", "error", err)
			}
			return
		}

		id := SubscribeID(gm.SubscribeID)

		receiver, ok := sess.trackReceivers[id]
		if !ok {
			stream.CancelRead(quic.StreamErrorCode(InvalidSubscribeIDErrorCode))
			return
		}

		group := newReceiveGroupStream(GroupSequence(gm.GroupSequence), stream)
		// Enqueue the receiver
		receiver.enqueueGroup(group)
	default:
		if logger != nil {
			logger.Debug("An unknown type of unidirectional stream was opened")
		}

		// Terminate the session
		sess.Terminate(ProtocolViolationErrorCode, fmt.Sprintf("unknown unidirectional stream type: %v", stm.StreamType))

		return
	}
}
